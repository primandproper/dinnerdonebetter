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
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	passkeysbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passkeys"
	passwordresetbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passwordreset"
	signinbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/signin"
	waitlistsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/identity/identitypb"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/database"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	o11yName = "auth_interceptor"

	authHeaderName = "Authorization"
	tokenPrefix    = "Bearer "

	// runModeEnvVarKey is the environment variable describing the current run mode (development,
	// testing, or production). Kept as a literal here to avoid importing internal/config.
	runModeEnvVarKey = "DINNER_DONE_BETTER_META_RUN_MODE"
)

type AuthInterceptor struct {
	tracer                      tracing.Tracer
	logger                      logging.Logger
	directory                   platformidentity.Store
	db                          database.Client
	sessions                    *identitybuild.SessionBuilder
	methodPermissions           map[string][]authorization.Permission
	oauth2Server                *oauth2server.Server
	tokenIssuer                 tokens.Issuer
	signIns                     SignInChecker
	oauth2Resource              string
	unauthenticatedRoutes       []string
	optionalRoutes              []string
	passwordChangeAllowedRoutes []string
	methodScopesHat             sync.Mutex
}

// SignInChecker reads whether the login a platform sign-in token belongs to is still live —
// signin.Service.CheckSignIn.
type SignInChecker interface {
	CheckSignIn(ctx context.Context, scope tenancy.Scope, familyID, tokenID string) error
}

// MethodPermissionsMap is a map of gRPC method full names to the permissions required to call them.
// This type is used for dependency injection of aggregated service permissions.
type MethodPermissionsMap map[string][]authorization.Permission

func ProvideAuthInterceptor(
	tracerProvider tracing.Provider,
	logger logging.Logger,
	directory platformidentity.Store,
	db database.Client,
	sessionBuilder *identitybuild.SessionBuilder,
	oauth2Server *oauth2server.Server,
	oauth2Resource string,
	tokenIssuer tokens.Issuer,
	signIns SignInChecker,
	aggregatedPermissions MethodPermissionsMap,
) *AuthInterceptor {
	unauthenticatedRoutes := []string{
		// Analytics proxy: anonymous events (no auth)
		"/analytics.AnalyticsService/TrackAnonymousEvent",
	}

	// platform's SignInService: its two sign-in doors, the refresh exchange, and sign-out.
	// See internal/build/signin for which of its RPCs are exposed at all.
	unauthenticatedRoutes = append(unauthenticatedRoutes, signinbuild.AnonymousMethods()...)

	// platform's PasswordResetService, every RPC of which is for somebody who cannot sign in,
	// and PasskeysService's login ceremony. See internal/build/passwordreset and
	// internal/build/passkeys.
	unauthenticatedRoutes = append(unauthenticatedRoutes, passwordresetbuild.AnonymousMethods()...)
	unauthenticatedRoutes = append(unauthenticatedRoutes, passkeysbuild.AnonymousMethods()...)

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
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:            logging.NewNamedLogger(logger, o11yName),
		directory:         directory,
		db:                db,
		sessions:          sessionBuilder,
		oauth2Server:      oauth2Server,
		oauth2Resource:    oauth2Resource,
		tokenIssuer:       tokenIssuer,
		signIns:           signIns,
		methodPermissions: aggregatedPermissions,
		// Routes allowed when requires_password_change is true.
		passwordChangeAllowedRoutes: []string{
			// A client told to change its password has to be able to learn whose password it
			// is changing.
			identitypb.IdentityService_GetPrincipal_FullMethodName,
			// The change itself.
			signinpb.SignInService_UpdatePassword_FullMethodName,
			// And signing out instead, which a person handed somebody else's device needs.
			signinpb.SignInService_SignOutEverywhere_FullMethodName,
		},
		unauthenticatedRoutes: unauthenticatedRoutes,
		// The signup page answers a visitor and reads a signed-in caller's session when one
		// is sent: see internal/build/waitlists.
		optionalRoutes: append(signinbuild.OptionallyAuthenticatedMethods(), waitlistsbuild.PublicMethods()...),
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

// errPasswordChangeRequired refuses a caller who owes a password change, carrying platform's
// sentinel so the error encoder attaches its client-safe reason — PASSWORD_CHANGE_REQUIRED, the
// one a client sends them to the form on — whichever door they signed in through.
func errPasswordChangeRequired() error {
	return observability.GRPCStatusError(signin.ErrPasswordChangeRequired, codes.FailedPrecondition, "password change required")
}

func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
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
			sessionCtxData, sessionErr := s.sessions.BuildSessionContextDataForUser(ctx, userID, "")
			if sessionErr != nil {
				return nil, observability.PrepareAndLogError(sessionErr, logger, span, "fetching user info for oauth2 token")
			}

			return sessionCtxData, nil
		}
	}

	// Otherwise the token has to be one platform's sign-in minted. It says which directory it
	// was issued in, and the claim cannot be added to a token by anybody but the server, which
	// signs it. It is read by its presence: the directory is named by its owner, and the
	// global directory this application's is has none, so the claim is present and empty.
	claims, parseErr := s.tokenIssuer.ParseToken(ctx, accessToken)
	if parseErr == nil {
		if userID := claims.Subject(); userID != "" {
			if issuedIn, ok := claims.GetString(signin.ClaimScope); ok {
				return s.signInSessionContextData(ctx, claims, userID, issuedIn)
			}
		}
	}

	return nil, Unauthenticated("invalid or expired token")
}

// signInSessionContextData turns a token platform's sign-in minted into a caller.
//
// The token names a login — its sid is the refresh token family — and the login is read on
// every request through signin.Service.CheckSignIn, so a sign-out, an ended sign-in or a
// detected refresh-token reuse stops the access token on its next request rather than when it
// expires. The read is on the write pool, for the reason CheckSignIn gives: a refresh's
// successor is presented at once, and a lagging replica would refuse it.
//
// So the checks are the signature, which ParseToken has already made; the directory, which
// has to be this application's; the login, which has to be named, because a token that names
// none is not one platform issued; that the login is still live; and, on an impersonation,
// that the operator still stands.
func (s *AuthInterceptor) signInSessionContextData(
	ctx context.Context,
	claims tokens.Claims,
	userID, issuedIn string,
) (*sessions.ContextData, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if issuedIn != ddbidentity.Scope().Owner() {
		return nil, Unauthenticated("token was issued in another directory")
	}

	familyID, _ := claims.GetString(signin.ClaimFamilyID)
	if familyID == "" {
		return nil, Unauthenticated("token names no login")
	}

	// The login has to still be live, read on every request, so a sign-out, an ended sign-in
	// or a detected reuse stops this access token at once rather than when it expires — the
	// same promise a token this application mints keeps through its session row.
	if s.signIns != nil {
		if err := s.signIns.CheckSignIn(ctx, ddbidentity.Scope(), familyID, claims.JTI()); err != nil {
			logger.WithValue("signin.family_id", familyID).Info("refusing a token whose sign-in has ended")
			return nil, Unauthenticated("sign-in has ended")
		}
	}

	accountID, _ := claims.GetString(signin.ClaimAccountID)

	sessionCtxData, err := s.sessions.BuildSessionContextDataForUser(ctx, userID, accountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user info from sign-in token")
	}

	// Which login is asking, so the doors that act on "this one" or "every other one" can tell.
	sessionCtxData.SignInFamilyID = familyID

	if err = s.applyImpersonator(ctx, claims, sessionCtxData); err != nil {
		return nil, err
	}

	return sessionCtxData, nil
}

// applyImpersonator puts the operator acting through an impersonation token on the session, and
// leaves every other token's session as it is.
//
// The request stays the subject's: their identity, their account, their memberships, their
// rows. What the token adds is who is really at the keyboard — signin.ClaimActor, which signin
// stamps on an impersonation and strips from everything else, so a claims builder can neither
// forge nor drop it — and the session carries it so that the audit log records the operator
// beside the subject.
//
// It also carries the operator's service-level grants in place of the subject's, which is this
// application's answer to the question platform leaves to it. An operator acting in somebody's
// account is still an operator — verifying that account's audit chain, say, is operator work —
// and a customer's service role grants nothing an operator lacks. What the subject's account
// lets them do is unchanged: that is the subject's membership, read from the subject's
// principal.
//
// The operator is read again on every request, in the scope the token says they are in, for
// platform's extractor's reason: suspending an operator mid-impersonation has to stop them on
// their next request, not when the token lapses.
func (s *AuthInterceptor) applyImpersonator(ctx context.Context, claims tokens.Claims, session *sessions.ContextData) error {
	actorID, _ := claims.GetString(signin.ClaimActor)
	if actorID == "" || actorID == session.GetUserID() {
		return nil
	}

	actorScope, present := claims.GetString(signin.ClaimActorScope)
	if !present {
		return Unauthenticated("token names an operator and no directory for them")
	}

	// This application has one directory, operators and customers alike.
	if actorScope != ddbidentity.Scope().Owner() {
		return Unauthenticated("token names an operator in another directory")
	}

	operator, err := s.sessions.BuildSessionContextDataForUser(ctx, actorID, "")
	if err != nil {
		return observability.PrepareError(err, nil, "resolving an impersonation's operator")
	}

	session.ImpersonatorID = actorID
	session.Requester.ServicePermissions = operator.Requester.ServicePermissions

	return nil
}

func (s *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		logger := s.logger.WithValue("grpc.method", info.FullMethod)

		if slices.Contains(s.unauthenticatedRoutes, info.FullMethod) {
			logger.Info("skipping authentication for method")
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)

		if slices.Contains(s.optionalRoutes, info.FullMethod) {
			return s.optionallyAuthenticated(ctx, md, req, handler)
		}

		if !ok {
			return nil, Unauthenticated("missing metadata")
		}

		authHeader := md.Get(authHeaderName)
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		sessionContextData, err := s.extractSessionContextData(ctx, md)
		if err != nil {
			return nil, sessionFailure(err)
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

		requiresChange, pcErr := s.userRequiresPasswordChange(ctx, sessionContextData.GetUserID())
		if pcErr != nil {
			return nil, status.Error(codes.Internal, "checking password change requirement")
		}
		if requiresChange && !slices.Contains(s.passwordChangeAllowedRoutes, info.FullMethod) {
			return nil, errPasswordChangeRequired()
		}

		ctx = sessions.AttachToContext(ctx, sessionContextData)

		return handler(ctx, req)
	}
}

// SessionFromAuthorization builds the caller an Authorization header names, the way the gRPC
// interceptors do for a request's metadata.
//
// It is the HTTP routes' way in. The platform surfaces this server mounts on its router —
// privacy requests, the operations that fulfill them, the object read — resolve their caller
// from the session context the same as every gRPC surface does, and a second token parser for
// HTTP would be a second place for the two protocols to disagree about who somebody is.
func (s *AuthInterceptor) SessionFromAuthorization(ctx context.Context, header string) (*sessions.ContextData, error) {
	return s.extractSessionContextData(ctx, metadata.Pairs(authHeaderName, header))
}

// RequiresPasswordChange reports whether userID has been told to change their password, which
// the gRPC interceptor answers by refusing everything but the change itself. It is exported so
// the HTTP routes can hold a caller to the same rule.
func (s *AuthInterceptor) RequiresPasswordChange(ctx context.Context, userID string) (bool, error) {
	return s.userRequiresPasswordChange(ctx, userID)
}

// sessionFailure is the status a caller is answered with when their session could not be built.
//
// A status the extraction already chose (a genuine Unauthenticated, say) is propagated rather
// than masked, because masking every failure as Internal breaks a client's token-refresh retry.
// A user whose account status does not admit sign-in — banned, terminated — is a caller the
// directory refuses rather than a server fault, and is answered PermissionDenied: the token is
// genuine, so Unauthenticated would send the client to refresh a token that works. Anything
// else is Internal.
func sessionFailure(err error) error {
	if _, isStatusErr := status.FromError(err); isStatusErr {
		return err
	}

	if errors.Is(err, platformidentity.ErrSignInNotAdmitted) {
		return status.Error(codes.PermissionDenied, "this account may not sign in")
	}

	return status.Error(codes.Internal, "building session context data for user")
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
			return sessionFailure(err)
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

		requiresChange, pcErr := s.userRequiresPasswordChange(ss.Context(), sessionContextData.GetUserID())
		if pcErr != nil {
			return status.Error(codes.Internal, "checking password change requirement")
		}
		if requiresChange && !slices.Contains(s.passwordChangeAllowedRoutes, info.FullMethod) {
			return errPasswordChangeRequired()
		}

		newCtx := sessions.AttachToContext(ss.Context(), sessionContextData)
		wrappedStream := &serverStreamWithContext{ServerStream: ss, ctx: newCtx}

		return handler(srv, wrappedStream)
	}
}

// optionallyAuthenticated serves a method that answers anonymous and signed-in callers alike.
//
// A call with no token is anonymous and goes straight through. A call with one is held to it:
// a token that no longer works is refused as Unauthenticated, exactly as on any other method,
// rather than quietly answered as though nobody had asked. The difference matters to
// GetAuthStatus — a client whose access token has expired has to be told so, so that it
// refreshes, and not told that it is signed out — and to the waitlist signup page, where a
// signed-in caller's join is theirs and an expired token must not quietly make it a visitor's.
//
// No permission is asked for. These methods are public by design; what a caller may do on
// them is decided by the handler, which reads the session when there is one.
func (s *AuthInterceptor) optionallyAuthenticated(ctx context.Context, md metadata.MD, req any, handler grpc.UnaryHandler) (any, error) {
	if len(md.Get(authHeaderName)) == 0 {
		return handler(ctx, req)
	}

	sessionContextData, err := s.extractSessionContextData(ctx, md)
	if err != nil {
		return nil, sessionFailure(err)
	}

	return handler(sessions.AttachToContext(ctx, sessionContextData), req)
}

// UnauthenticatedRoutes returns the methods this interceptor lets through without a session:
// those that never read one, and those that read one only when it is sent.
//
// It exists so the platform's authorization enforcer can be built from the same list rather than
// a second copy of it. Two allow-lists that are supposed to agree and are maintained separately
// is how a method ends up public in one and not the other.
func (s *AuthInterceptor) UnauthenticatedRoutes() []string {
	return append(slices.Clone(s.unauthenticatedRoutes), s.optionalRoutes...)
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

// userRequiresPasswordChange reports whether this user must change their password before
// they may do anything else.
//
// It is a field on the row rather than a method of its own now, so this is a read of the
// user. The read this replaced was a dedicated query, which is the shape a repository of
// this application's own could have and a general directory should not: platform answers
// with the user and lets a caller ask whatever it wanted to know.
func (s *AuthInterceptor) userRequiresPasswordChange(ctx context.Context, userID string) (bool, error) {
	user, err := s.directory.GetUser(ctx, s.db.Reader(), tenancy.Global(), userID)
	if err != nil {
		return false, err
	}

	return user.RequiresPasswordChange, nil
}
