package grpcapi

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	platformgrpc "github.com/primandproper/primitives-go/v2/server/grpc"

	"github.com/samber/do/v2"
	grpc "google.golang.org/grpc"
)

// RegisterExtras registers the server's interceptor chains, its permission table and enforcer,
// and the registrations that mount every surface — see surfaces.go for the list.
func RegisterExtras(i do.Injector) {
	do.Provide(i, func(do.Injector) (interceptors.MethodPermissionsMap, error) {
		return MethodPermissions(), nil
	})

	// One enforcer, shared by both chains, so a stream and a unary call are refused by the same
	// table.
	do.Provide(i, func(i do.Injector) (*authzgrpc.Enforcer, error) {
		return ProvideAuthorizationEnforcer(
			MethodPermissionFragments(),
			MethodPermissionOverrides(),
			do.MustInvoke[*interceptors.AuthInterceptor](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[metrics.Provider](i),
			auditOnlyAuthorization,
		)
	})

	do.Provide(i, func(i do.Injector) ([]grpc.UnaryServerInterceptor, error) {
		logger := do.MustInvoke[logging.Logger](i)
		authInterceptor := do.MustInvoke[*interceptors.AuthInterceptor](i)
		authzEnforcer := do.MustInvoke[*authzgrpc.Enforcer](i)

		idempotencyInterceptor, err := ProvideIdempotencyInterceptor(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*config.APIServiceConfig](i),
			logger,
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
		if err != nil {
			return nil, err
		}

		throttle, err := interceptors.NewAnonymousDoorThrottle(
			do.MustInvoke[ratelimiting.RateLimiter](i),
			logger,
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
		if err != nil {
			return nil, err
		}

		return BuildUnaryServerInterceptors(throttle, authInterceptor, authzEnforcer, idempotencyInterceptor), nil
	})

	do.Provide(i, func(i do.Injector) ([]grpc.StreamServerInterceptor, error) {
		return BuildStreamServerInterceptors(
			do.MustInvoke[*interceptors.AuthInterceptor](i),
			do.MustInvoke[*authzgrpc.Enforcer](i),
		), nil
	})

	do.Provide(i, func(i do.Injector) ([]platformgrpc.RegistrationFunc, error) {
		return BuildRegistrationFuncs(i), nil
	})
}

// The interceptors this server adds, inside the two primitives-go's server installs ahead of every
// list it is given: RecoveryInterceptor, outermost, and the logging interceptor after it. Recovery
// is not repeated here. A second one would sit inside the first and catch nothing it does not.
//
// The first of this server's own strips the encoded error chain. The error encoding interceptors
// put a failure on the wire twice: once as a client-safe status message, and once as the whole
// wrapped chain, encoded into the status details so a trusted peer can reconstruct it with
// errorsgrpc.DecodeErrorFromStatus. That second copy is unredacted — table names, the rule a query
// broke, which of two refusals signin deliberately answers identically — and primitives-go is
// explicit that a server reachable by untrusted clients must strip it at the edge.
//
// This server is that edge. The iOS app and the web frontend dial it directly, and nothing between
// the handler and their transport removes a detail. Nor is there a trusted peer on the far side to
// keep it for: nothing that calls this server decodes the chain. So it is stripped here, for every
// method, rather than per service or behind a flag somebody has to remember to set.
//
// Only the encoded chain goes. The code, the message and the google.rpc.ErrorInfo a client branches
// on are left exactly as the encoder built them — see errorsgrpc.StripEncodedErrorDetail.

// BuildUnaryServerInterceptors is the unary chain this server adds, outermost first.
func BuildUnaryServerInterceptors(
	throttle grpc.UnaryServerInterceptor,
	authInterceptor *interceptors.AuthInterceptor,
	authzEnforcer *authzgrpc.Enforcer,
	idempotencyInterceptor grpc.UnaryServerInterceptor,
) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		// First, so nothing any interceptor below returns reaches a client with the encoded error
		// chain still attached. It must sit outside the error encoder, which is what attaches it.
		errorsgrpc.StripEncodedErrorDetailUnaryServerInterceptor(),
		// Outside authentication, so the refusals the interceptors below make are encoded the
		// way a handler's are — with a client-safe reason where the error names one, which is
		// how a forced password change says PASSWORD_CHANGE_REQUIRED.
		errorsgrpc.UnaryErrorEncodingInterceptor(),
		// Ahead of authentication: the doors it throttles read no credential, and a caller it
		// refuses should cost nothing past the bucket.
		throttle,
		authInterceptor.UnaryServerInterceptor(),
		// Runs after the interceptor above so it sees the session that one established. It is
		// the only permission check: the interceptor above decides who is calling, not what
		// they may do.
		authzEnforcer.UnaryServerInterceptor(),
		// after auth, because the key is scoped to the authenticated principal, and inside the
		// error encoder, because it records the handler's status code rather than a rendered one.
		idempotencyInterceptor,
	}
}

// BuildStreamServerInterceptors is the stream chain this server adds, outermost first.
//
// It enforces the method permission table exactly as the unary chain does. A stream —
// MediaRegistryService.UploadObject is one — is no less a call than a unary one, and a chain
// without the enforcer would admit any signed-in caller to it.
func BuildStreamServerInterceptors(
	authInterceptor *interceptors.AuthInterceptor,
	authzEnforcer *authzgrpc.Enforcer,
) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		// First, for the reason the unary chain's strip is.
		errorsgrpc.StripEncodedErrorDetailStreamServerInterceptor(),
		errorsgrpc.StreamErrorEncodingInterceptor(),
		authInterceptor.StreamServerInterceptor(),
		// After authentication, for the reason the unary chain's enforcer is.
		authzEnforcer.StreamServerInterceptor(),
	}
}
