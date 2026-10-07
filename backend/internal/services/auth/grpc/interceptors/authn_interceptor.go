package interceptors

import (
	"context"
	"errors"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	passkeysbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passkeys"
	passwordresetbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passwordreset"
	signinbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/signin"
	waitlistsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"

	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/callers"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/observability/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	o11yName = "auth_interceptor"

	// runModeEnvVarKey is the environment variable describing the current run mode (development,
	// testing, or production). Kept as a literal here to avoid importing internal/config.
	runModeEnvVarKey = "DINNER_DONE_BETTER_META_RUN_MODE"

	// authorizationHeader is where a bearer token travels, as gRPC metadata.
	authorizationHeader = "authorization"
)

// errNoIdentity is a principal that carries no directory answer, which every principal
// platform's extractor resolves does. One that does not came from somewhere this interceptor
// does not know how to render a session for.
var errNoIdentity = errors.New("the caller carries no identity to build a session from")

// AuthInterceptor resolves who is calling and decides whether they may call the method.
//
// Who is calling is platform's: signingrpc.PrincipalExtractor verifies the bearer, checks its
// login is still live, reads the principal, refuses a user whose status does not admit sign-in,
// grants service roles only to a token minted through the administrative door, stands an
// impersonation's operator, falls back to an OAuth2 access token through oauth2server.Verifier,
// and holds a caller who owes a forced password change at the form. What is left here is what
// was never platform's: rendering that principal as this application's session, and the method
// permission table.
type AuthInterceptor struct {
	logger                logging.Logger
	extractor             *signingrpc.PrincipalExtractor
	requirements          *signingrpc.AuthenticationRequirements
	sessions              *identitybuild.SessionBuilder
	methodPermissions     map[string][]authorization.Permission
	unauthenticatedRoutes []string
	optionalRoutes        []string
}

// MethodPermissionsMap is a map of gRPC method full names to the permissions required to call them.
// This type is used for dependency injection of aggregated service permissions.
type MethodPermissionsMap map[string][]authorization.Permission

func ProvideAuthInterceptor(
	logger logging.Logger,
	extractor *signingrpc.PrincipalExtractor,
	sessionBuilder *identitybuild.SessionBuilder,
	aggregatedPermissions MethodPermissionsMap,
) (*AuthInterceptor, error) {
	// platform's SignInService: its two sign-in doors, the refresh exchange, and sign-out.
	// See internal/build/signin for which of its RPCs are exposed at all.
	unauthenticatedRoutes := slices.Clone(signinbuild.AnonymousMethods())

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

	// The signup page answers a visitor and reads a signed-in caller's session when one is
	// sent: see internal/build/waitlists.
	optionalRoutes := append(signinbuild.OptionallyAuthenticatedMethods(), waitlistsbuild.PublicMethods()...)

	requirements, err := buildRequirements(unauthenticatedRoutes, optionalRoutes, aggregatedPermissions)
	if err != nil {
		return nil, err
	}

	return &AuthInterceptor{
		logger:                logging.NewNamedLogger(logger, o11yName),
		extractor:             extractor,
		requirements:          requirements,
		sessions:              sessionBuilder,
		methodPermissions:     aggregatedPermissions,
		unauthenticatedRoutes: unauthenticatedRoutes,
		optionalRoutes:        optionalRoutes,
	}, nil
}

// buildRequirements declares every method this interceptor knows to the extractor: the public
// ones as anonymous, the ones that read a session when one is sent as optional, and every method
// the permission table names as required.
//
// The table is what decides a method exists here. A method it does not name is undeclared, and
// the extractor refuses an undeclared method as PermissionDenied before it reads a credential —
// which is the answer this interceptor always gave one, for the same reason: a method nobody
// declared permissions for is a method nobody decided was safe to expose.
func buildRequirements(anonymous, optional []string, permissions MethodPermissionsMap) (*signingrpc.AuthenticationRequirements, error) {
	declared := map[string]struct{}{}
	builder := signingrpc.NewAuthenticationRequirements()

	declare := func(requirement signingrpc.Authentication, methods ...string) {
		for _, method := range methods {
			if _, taken := declared[method]; taken {
				continue
			}

			declared[method] = struct{}{}
			builder.Declare(requirement, method)
		}
	}

	declare(signingrpc.AuthenticationAnonymous, anonymous...)
	declare(signingrpc.AuthenticationOptional, optional...)

	for method := range permissions {
		declare(signingrpc.AuthenticationRequired, method)
	}

	return builder.Build()
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

func Unauthenticated(msg string) error {
	return status.Error(codes.Unauthenticated, msg)
}

// UnaryServerInterceptor resolves the caller through the extractor, then renders their session
// and checks the method's permissions.
//
// A public method never reaches the extractor: it reads no credential, so there is nothing to
// resolve, and a server built without one — a test's — can still serve it.
func (s *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	var resolve grpc.UnaryServerInterceptor
	if s.extractor != nil {
		resolve = s.extractor.UnaryServerInterceptor(s.requirements)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if slices.Contains(s.unauthenticatedRoutes, info.FullMethod) {
			return handler(ctx, req)
		}

		if resolve == nil {
			return nil, Unauthenticated("no caller can be resolved")
		}

		return resolve(ctx, req, info, func(ctx context.Context, req any) (any, error) {
			ctx, err := s.authorize(ctx, info.FullMethod)
			if err != nil {
				return nil, err
			}

			return handler(ctx, req)
		})
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

// StreamServerInterceptor is UnaryServerInterceptor for streaming RPCs. Without it, streaming RPCs
// (e.g. MediaRegistryService.UploadObject) bypass auth and session context is never set.
func (s *AuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	var resolve grpc.StreamServerInterceptor
	if s.extractor != nil {
		resolve = s.extractor.StreamServerInterceptor(s.requirements)
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if slices.Contains(s.unauthenticatedRoutes, info.FullMethod) {
			return handler(srv, ss)
		}

		if resolve == nil {
			return Unauthenticated("no caller can be resolved")
		}

		return resolve(srv, ss, info, func(srv any, ss grpc.ServerStream) error {
			ctx, err := s.authorize(ss.Context(), info.FullMethod)
			if err != nil {
				return err
			}

			return handler(srv, &serverStreamWithContext{ServerStream: ss, ctx: ctx})
		})
	}
}

// authorize renders the caller the extractor put on ctx as this application's session, checks
// the method's permissions, and attaches the session.
//
// The extractor has already refused a required method's anonymous caller, so a context with
// nobody on it here is an optional method. A visitor goes on as one. A caller who sent a token the
// extractor could not resolve does not: the extractor reads an optional method's dead token as no
// token, and this interceptor refuses it as Unauthenticated as it would on any other method. That
// difference is the point of GetAuthStatus — a client whose access token has expired has to be
// told so, so that it refreshes, and not told that it is signed out — and of Register, where a
// signed-in operator's registration must not quietly become a stranger's.
//
// No permission is asked on an optional method: they are public by design, and what a caller may
// do on them is the handler's to decide from the session when there is one.
func (s *AuthInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
	principal, ok := signingrpc.PrincipalFromContext(ctx)
	if !ok {
		if slices.Contains(s.optionalRoutes, method) && !carriesCredential(ctx) {
			return ctx, nil
		}

		return ctx, Unauthenticated("authentication required")
	}

	session, err := s.sessionFor(ctx, principal)
	if err != nil {
		s.logger.WithValue("grpc.method", method).Error("building session context data", err)
		return ctx, status.Error(codes.Internal, "building session context data for user")
	}

	if !slices.Contains(s.optionalRoutes, method) && !s.permitted(method, session) {
		return ctx, status.Error(codes.PermissionDenied, "permission denied")
	}

	return sessions.AttachToContext(ctx, session), nil
}

// carriesCredential reports whether a gRPC request sent an Authorization header at all.
func carriesCredential(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	for _, value := range md.Get(authorizationHeader) {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}

	return false
}

// permitted reports whether the session holds every permission the method requires, from the
// service role or the active account's membership. A method the table does not name is refused.
func (s *AuthInterceptor) permitted(method string, session *sessions.ContextData) bool {
	required, declared := s.methodPermissions[method]
	if !declared {
		s.logger.WithValue("grpc.method", method).Info("missing required permissions for method")
		return false
	}

	for _, permission := range required {
		if !session.ServiceRolePermissionChecker().HasPermission(permission) &&
			!session.AccountRolePermissionsChecker().HasPermission(permission) {
			return false
		}
	}

	return true
}

// sessionFor renders a resolved principal as session context.
//
// The service roles are the ones the extractor left on the principal, which for a token minted
// through an ordinary door is none: an operator's grants ride only on a token from the
// administrative door, so a password typed into the consumer app's login form carries none of
// them. An impersonation is the subject's session with the operator named beside it, and carries
// the subject's grants and nothing of the operator's — the operator acts as the subject, and
// operator work is done with the operator's own token.
func (s *AuthInterceptor) sessionFor(ctx context.Context, principal callers.Principal) (*sessions.ContextData, error) {
	carrier, ok := principal.(interface {
		Identity() *platformidentity.Principal
	})
	if !ok {
		return nil, errNoIdentity
	}

	session, err := s.sessions.SessionForPrincipal(ctx, carrier.Identity())
	if err != nil {
		return nil, err
	}

	// Which login is asking, so the doors that act on "this one" or "every other one" can tell.
	// An OAuth2 access token names none.
	if family, isSignIn := principal.(signingrpc.FamilyIdentifier); isSignIn {
		session.SignInFamilyID = family.FamilyID()
	}

	// Who is really at the keyboard, which the audit log records beside the subject.
	if delegated, isDelegated := principal.(callers.Delegated); isDelegated {
		session.ImpersonatorID = delegated.ActorID()
	}

	return session, nil
}

// HTTPMiddleware resolves the caller of an HTTP request the way the gRPC interceptors do, and
// attaches their session.
//
// It is the HTTP routes' way in. The platform surfaces this server mounts on its router —
// privacy requests, the operations that fulfill them, the object read — resolve their caller from
// the session context the same as every gRPC surface does, and a second token parser for HTTP
// would be a second place for the two protocols to disagree about who somebody is.
//
// The extractor's middleware treats every request as optional: a request with no bearer token, or
// one that names nobody, goes on with no session, and the surfaces that need a caller refuse its
// absence themselves. It answers a banned user 403 and a directory it cannot read 503, and holds a
// caller who owes a forced password change with a 403, since no HTTP route is the change.
func (s *AuthInterceptor) HTTPMiddleware(next http.Handler) http.Handler {
	attach := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		principal, ok := signingrpc.PrincipalFromContext(req.Context())
		if !ok {
			next.ServeHTTP(res, req)
			return
		}

		session, err := s.sessionFor(req.Context(), principal)
		if err != nil {
			s.logger.WithValue("http.path", req.URL.Path).Error("building session context data", err)
			http.Error(res, "building session context data for user", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(res, req.WithContext(sessions.AttachToContext(req.Context(), session)))
	})

	if s.extractor == nil {
		return next
	}

	return s.extractor.HTTPMiddleware(attach)
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

// AnonymousRoutes returns the methods that never read a credential: the doors a caller signs in,
// signs up, or recovers an account through, and the few that answer anybody.
func (s *AuthInterceptor) AnonymousRoutes() []string {
	return slices.Clone(s.unauthenticatedRoutes)
}
