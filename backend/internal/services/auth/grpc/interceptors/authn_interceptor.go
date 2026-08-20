package interceptors

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"
	"github.com/primandproper/platform-go/v12/authentication/tokens"
	errorsgrpc "github.com/primandproper/platform-go/v12/errors/grpc"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	o11yName = "auth_interceptor"

	authHeaderName = "Authorization"
	tokenPrefix    = "Bearer "

	// TODO: organize this so that the API client gets the same source.
	zuckModeUserHeader    = "X-Zuck-Mode-User"
	zuckModeAccountHeader = "X-Zuck-Mode-Account"

	// runModeEnvVarKey is the environment variable describing the current run mode (development,
	// testing, or production). Kept as a literal here to avoid importing internal/config.
	runModeEnvVarKey = "DINNER_DONE_BETTER_META_RUN_MODE"
)

type AuthInterceptor struct {
	tracer                      tracing.Tracer
	logger                      logging.Logger
	identityDataManager         identitymanager.IdentityDataManager
	sessionDataManager          auth.UserSessionDataManager
	methodPermissions           map[string][]authorization.Permission
	oauth2Server                *oauth2server.Server
	tokenIssuer                 tokens.Issuer
	oauth2Resource              string
	unauthenticatedRoutes       []string
	passwordChangeAllowedRoutes []string
	methodScopesHat             sync.Mutex
}

// MethodPermissionsMap is a map of gRPC method full names to the permissions required to call them.
// This type is used for dependency injection of aggregated service permissions.
type MethodPermissionsMap map[string][]authorization.Permission

func ProvideAuthInterceptor(
	tracerProvider tracing.Provider,
	logger logging.Logger,
	identityDataManager identitymanager.IdentityDataManager,
	sessionDataManager auth.UserSessionDataManager,
	oauth2Server *oauth2server.Server,
	oauth2Resource string,
	tokenIssuer tokens.Issuer,
	aggregatedPermissions MethodPermissionsMap,
) *AuthInterceptor {
	// TODO: configure this elsewhere
	unauthenticatedRoutes := []string{
		"/auth.AuthService/AdminLoginForToken",
		"/auth.AuthService/BeginPasskeyAuthentication",
		"/auth.AuthService/FinishPasskeyAuthentication",
		"/identity.IdentityService/CreateUser",
		"/auth.AuthService/VerifyTOTPSecret",
		"/auth.AuthService/LoginForToken",
		"/auth.AuthService/RequestPasswordResetToken",
		"/auth.AuthService/RedeemPasswordResetToken",
		// A user who forgot their username can't authenticate first.
		"/auth.AuthService/RequestUsernameReminder",
		"/auth.AuthService/VerifyEmailAddress",
		// Analytics proxy: anonymous events (no auth)
		"/analytics.AnalyticsService/TrackAnonymousEvent",
	}

	// gRPC reflection exposes the full service catalog to unauthenticated callers. It's handy for
	// local tooling (grpcurl, k6) but should not be reachable in production, so only allow-list it
	// outside of production run mode.
	if grpcReflectionEnabled() {
		unauthenticatedRoutes = append(unauthenticatedRoutes,
			"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		)
	}

	return &AuthInterceptor{
		tracer:              tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:              logging.NewNamedLogger(logger, o11yName),
		identityDataManager: identityDataManager,
		sessionDataManager:  sessionDataManager,
		oauth2Server:        oauth2Server,
		oauth2Resource:      oauth2Resource,
		tokenIssuer:         tokenIssuer,
		methodPermissions:   aggregatedPermissions,
		// Routes allowed when requires_password_change is true.
		passwordChangeAllowedRoutes: []string{
			"/auth.AuthService/UpdatePassword",
			"/auth.AuthService/RevokeCurrentSession",
			"/auth.AuthService/GetAuthStatus",
			"/auth.AuthService/GetActiveAccount",
		},
		unauthenticatedRoutes: unauthenticatedRoutes,
	}
}

// grpcReflectionEnabled reports whether gRPC reflection should be reachable without authentication.
// It is enabled only outside of production run mode (production is the default when the env var is unset).
func grpcReflectionEnabled() bool {
	switch strings.TrimSpace(strings.ToLower(os.Getenv(runModeEnvVarKey))) {
	case "development", "testing":
		return true
	default:
		return false
	}
}

var (
	// ErrUserNotAuthorizedToImpersonateOthers is returned when a user is not authorized to impersonate others.
	ErrUserNotAuthorizedToImpersonateOthers = errors.New("user not authorized to impersonate others")
)

func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}

func (s *AuthInterceptor) determineZuckMode(ctx context.Context, metaData metadata.MD, sessionContextData *sessions.ContextData) (userID, accountID string, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if zuckUserHeaders := metaData.Get(zuckModeUserHeader); len(zuckUserHeaders) > 0 {
		var (
			zuckUserID    = zuckUserHeaders[0]
			zuckAccountID string
		)

		if !sessionContextData.ServiceRolePermissionChecker().CanImpersonateUsers() {
			return "", "", ErrUserNotAuthorizedToImpersonateOthers
		}

		if _, err = s.identityDataManager.GetUser(ctx, zuckUserID); err != nil {
			return "", "", observability.PrepareError(err, span, "fetching user info")
		}

		if zuckAccountIDs := metaData.Get(zuckModeAccountHeader); len(zuckAccountIDs) > 0 {
			zuckAccountID = zuckAccountIDs[0]

			// Honor the specifically requested impersonation account instead of always falling back
			// to the user's default. BuildSessionContextDataForUser validates that the impersonated
			// user is actually a member of the requested account and errors out otherwise.
			if _, err = s.identityDataManager.BuildSessionContextDataForUser(ctx, zuckUserID, zuckAccountID); err != nil {
				return "", "", observability.PrepareError(err, span, "validating impersonated account membership")
			}
		}

		return zuckUserID, zuckAccountID, nil
	}

	return "", "", nil
}

func (s *AuthInterceptor) extractSessionContextData(ctx context.Context, metaData metadata.MD) (*sessions.ContextData, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	authHeader := metaData.Get("authorization")
	if len(authHeader) == 0 {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(status.Error(codes.Unauthenticated, "missing authorization header"), logger, span, codes.Unauthenticated, "missing authorization header")
	}

	accessToken := strings.TrimPrefix(authHeader[0], tokenPrefix)

	// Try OAuth2 token first. The token is opaque, so this is a store lookup rather than a
	// signature check — which is what makes a revoked token stop working on the next request
	// rather than at the end of its lifetime.
	if token, err := s.oauth2Server.Authenticate(ctx, accessToken); err == nil {
		if audErr := s.checkAudience(token); audErr != nil {
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(audErr, logger, span, codes.Unauthenticated, "token audience does not name this resource server")
		}

		if userID := token.Subject.ID; userID != "" {
			// The user's current default account, not the one named in the token's claims.
			//
			// The authorization server does record which account the authorization was granted
			// against — see the account_id claim it mints — and pinning to it would be the more
			// literal reading of a scoped token. It is not what this server can do: an access
			// token is opaque and long-lived relative to a session, SetDefaultAccount and
			// ChangeActiveAccount are how a user moves between accounts, and there is no way to
			// re-mint an OAuth2 access token when they do. Pinning would mean an account switch
			// silently not applying until the next full authorization.
			//
			// So the claim is recorded and not spent. Honoring it needs a way for a client to
			// ask for a token on a named account and a way to notice when that account is no
			// longer the one in use — neither of which exists yet.
			sessionCtxData, sessionErr := s.identityDataManager.BuildSessionContextDataForUser(ctx, userID, "")
			if sessionErr != nil {
				return nil, observability.PrepareAndLogError(sessionErr, logger, span, "fetching user info for oauth2 token")
			}
			return s.applyZuckMode(ctx, metaData, sessionCtxData)
		}
	}

	// Fallback: treat Bearer token as JWT (e.g. from LoginForToken with DesiredAccountID).
	claims, parseErr := s.tokenIssuer.ParseToken(ctx, accessToken)
	if parseErr == nil {
		userID := claims.Subject()
		if userID != "" {
			accountID, _ := claims.GetString("account_id")
			// Validate session if token has a session ID.
			sessionID, _ := claims.GetString("sid")
			if sessionID != "" {
				jti := claims.JTI()
				if jti != "" {
					if _, sessErr := s.sessionDataManager.GetUserSessionBySessionTokenID(ctx, jti); sessErr != nil {
						return nil, Unauthenticated("session has been revoked or expired")
					}
					// Touch last active asynchronously so it doesn't block the request. Use a local error
					// variable so the goroutine doesn't race on the outer err captured by the closure.
					touchJTI := jti
					touchCtx := context.WithoutCancel(ctx)
					go func() {
						if touchErr := s.sessionDataManager.TouchSessionLastActive(touchCtx, touchJTI); touchErr != nil {
							logger.Error("touch session last active failed", touchErr)
						}
					}()
				}
			}

			sessionCtxData, sessionErr := s.identityDataManager.BuildSessionContextDataForUser(ctx, userID, accountID)
			if sessionErr != nil {
				return nil, observability.PrepareAndLogError(sessionErr, logger, span, "fetching user info from token")
			}
			sessionCtxData.SessionID = sessionID
			return s.applyZuckMode(ctx, metaData, sessionCtxData)
		}
	}

	return nil, Unauthenticated("invalid or expired token")
}

func (s *AuthInterceptor) applyZuckMode(ctx context.Context, metaData metadata.MD, sessionCtxData *sessions.ContextData) (*sessions.ContextData, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	zuckUserID, zuckAccountID, zuckErr := s.determineZuckMode(ctx, metaData, sessionCtxData)
	if zuckErr != nil {
		return nil, observability.PrepareAndLogError(zuckErr, logger, span, "fetching user info for zuck mode")
	}

	if zuckUserID != "" {
		sessionCtxData.Requester.UserID = zuckUserID
	}

	if zuckAccountID != "" {
		sessionCtxData.ActiveAccountID = zuckAccountID
		sessionCtxData.AccountPermissions[zuckAccountID] = authorization.NewAccountRolePermissionChecker(nil)
	}

	return sessionCtxData, nil
}

func (s *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		logger := s.logger.WithValue("grpc.method", info.FullMethod)

		if slices.Contains(s.unauthenticatedRoutes, info.FullMethod) {
			logger.Info("skipping authentication for method")
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, Unauthenticated("missing metadata")
		}

		authHeader := md.Get(authHeaderName)
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		sessionContextData, err := s.extractSessionContextData(ctx, md)
		if err != nil {
			// Propagate the original status (e.g. a genuine Unauthenticated) rather than masking every
			// failure as Internal, which breaks client token-refresh retry logic.
			if _, isStatusErr := status.FromError(err); isStatusErr {
				return nil, err
			}
			return nil, status.Error(codes.Internal, "building session context data for user")
		}

		proceed := true
		permissionEvaluation := map[string]bool{}

		s.methodScopesHat.Lock()
		if requiredPermissions, methodHasDefinedScopes := s.methodPermissions[info.FullMethod]; methodHasDefinedScopes {
			for _, scope := range requiredPermissions {
				hasPerm := sessionContextData.ServiceRolePermissionChecker().HasPermission(scope) || sessionContextData.AccountRolePermissionsChecker().HasPermission(scope)
				permissionEvaluation[string(scope)] = hasPerm

				if !hasPerm {
					proceed = false
				}
			}
		} else {
			logger.Info(fmt.Sprintf("missing required permissions for method %q", info.FullMethod))
			proceed = false
		}
		s.methodScopesHat.Unlock()

		if !proceed {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		requiresChange, pcErr := s.identityDataManager.UserRequiresPasswordChange(ctx, sessionContextData.GetUserID())
		if pcErr != nil {
			return nil, status.Error(codes.Internal, "checking password change requirement")
		}
		if requiresChange && !slices.Contains(s.passwordChangeAllowedRoutes, info.FullMethod) {
			return nil, status.Error(codes.FailedPrecondition, "password change required")
		}

		ctx = sessions.AttachToContext(ctx, sessionContextData)

		return handler(ctx, req)
	}
}

// serverStreamWithContext wraps grpc.ServerStream to inject a modified context.
type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStreamWithContext) Context() context.Context {
	return s.ctx
}

// StreamServerInterceptor returns an interceptor that authenticates and authorizes streaming RPCs.
// Without this, streaming RPCs (e.g. UploadedMediaService.Upload) bypass auth and session context is never set.
func (s *AuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger := s.logger.WithValue("grpc.method", info.FullMethod)

		if slices.Contains(s.unauthenticatedRoutes, info.FullMethod) {
			logger.Info("skipping authentication for streaming method")
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return Unauthenticated("missing metadata")
		}

		authHeader := md.Get(authHeaderName)
		if len(authHeader) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		sessionContextData, err := s.extractSessionContextData(ss.Context(), md)
		if err != nil {
			// Propagate the original status (e.g. a genuine Unauthenticated) rather than masking every
			// failure as Internal, which breaks client token-refresh retry logic.
			if _, isStatusErr := status.FromError(err); isStatusErr {
				return err
			}
			return status.Error(codes.Internal, "building session context data for user")
		}

		proceed := true
		s.methodScopesHat.Lock()
		if requiredPermissions, methodHasDefinedScopes := s.methodPermissions[info.FullMethod]; methodHasDefinedScopes {
			for _, scope := range requiredPermissions {
				hasPerm := sessionContextData.ServiceRolePermissionChecker().HasPermission(scope) ||
					sessionContextData.AccountRolePermissionsChecker().HasPermission(scope)
				if !hasPerm {
					proceed = false
					break
				}
			}
		} else {
			logger.Info(fmt.Sprintf("missing required permissions for streaming method %q", info.FullMethod))
			proceed = false
		}
		s.methodScopesHat.Unlock()

		if !proceed {
			return status.Error(codes.PermissionDenied, "permission denied")
		}

		requiresChange, pcErr := s.identityDataManager.UserRequiresPasswordChange(ss.Context(), sessionContextData.GetUserID())
		if pcErr != nil {
			return status.Error(codes.Internal, "checking password change requirement")
		}
		if requiresChange && !slices.Contains(s.passwordChangeAllowedRoutes, info.FullMethod) {
			return status.Error(codes.FailedPrecondition, "password change required")
		}

		newCtx := sessions.AttachToContext(ss.Context(), sessionContextData)
		wrappedStream := &serverStreamWithContext{ServerStream: ss, ctx: newCtx}

		return handler(srv, wrappedStream)
	}
}

// UnauthenticatedRoutes returns the methods this interceptor lets through without a session.
//
// It exists so the platform's authorization enforcer can be built from the same list rather than
// a second copy of it. Two allow-lists that are supposed to agree and are maintained separately
// is how a method ends up public in one and not the other.
func (s *AuthInterceptor) UnauthenticatedRoutes() []string {
	return slices.Clone(s.unauthenticatedRoutes)
}

// errWrongAudience is the refusal the store cannot make.
//
// Expiry and revocation are already Authenticate's answer, so this is the only condition left
// for a resource server to check for itself: a token minted for a different resource — the MCP
// server, say, which shares this database and therefore this store — must not be spendable
// here. RFC 8707 exists to make that detectable and explicitly leaves the check to whoever is
// being handed the token.
var errWrongAudience = errors.New("token audience does not name this resource server")

// checkAudience refuses a token whose audience names somewhere that is not this server.
//
// An empty audience is accepted: a client that sends no resource parameter gets a token with
// none, and refusing those would make every such client unable to call anything. What must not
// be accepted is an audience that names a different resource.
//
// An unset oauth2Resource disables the check rather than failing every request, because a
// deployment that has not declared its own identifier cannot say whether a token names it.
func (s *AuthInterceptor) checkAudience(token *oauth2server.AccessToken) error {
	if s.oauth2Resource == "" || len(token.Audience) == 0 {
		return nil
	}

	if !slices.Contains(token.Audience, s.oauth2Resource) {
		return errWrongAudience
	}

	return nil
}
