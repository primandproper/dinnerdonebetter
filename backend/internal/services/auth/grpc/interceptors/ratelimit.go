package interceptors

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset/passwordresetpb"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	ratelimitinggrpc "github.com/primandproper/primitives-go/v2/ratelimiting/grpc"
	ratelimitinghttp "github.com/primandproper/primitives-go/v2/ratelimiting/http"
	"github.com/primandproper/primitives-go/v2/routing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// forwardedForKey is the header Caddy stamps the client's address in, as gRPC metadata.
	forwardedForKey = "x-forwarded-for"

	// authorizeFormKeyPrefix keeps the login form's buckets apart from the gRPC doors', which
	// share the limiter.
	authorizeFormKeyPrefix = "oauth2.authorize|"
)

// ThrottledMethods are the doors a caller reaches with no credential and that either test one or
// send somebody mail: the two sign-in doors, sign-up, and the three that mail an address on
// request. Each is a way to guess a password, enumerate an account by timing, or have this
// service mail a stranger, and none of them asks who is calling.
//
// platform's passwordreset/grpc documentation makes a limit in front of its request door
// mandatory, and signin's says the same of the rest: they count attempts and refuse none.
// DefaultRequestFloor, which this deployment's doors also have, only sets a minimum response time
// so the answer does not leak by its timing; it throttles nothing.
func ThrottledMethods() []string {
	return []string{
		signinpb.SignInService_LoginForToken_FullMethodName,
		signinpb.SignInService_AdminLoginForToken_FullMethodName,
		signinpb.SignInService_Register_FullMethodName,
		signinpb.SignInService_RequestHandleReminder_FullMethodName,
		signinpb.SignInService_RequestVerificationEmailByAddress_FullMethodName,
		passwordresetpb.PasswordResetService_RequestPasswordReset_FullMethodName,
	}
}

// NewAnonymousDoorThrottle is the unary interceptor that spends a token from the caller's address's
// budget on every call to a ThrottledMethods door, each door with a budget of its own, and answers
// RESOURCE_EXHAUSTED with the delay to wait when there is none. Every other method is exempt.
//
// It belongs ahead of authentication: none of these doors reads a credential, and a refused
// caller should cost nothing past the bucket.
func NewAnonymousDoorThrottle(
	limiter ratelimiting.RateLimiter,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (grpc.UnaryServerInterceptor, error) {
	throttled := ThrottledMethods()

	return ratelimitinggrpc.NewUnaryServerInterceptor(limiter,
		ratelimitinggrpc.PerMethod(func(ctx context.Context, info *grpc.UnaryServerInfo) (string, error) {
			if !slices.Contains(throttled, info.FullMethod) {
				return "", nil
			}

			return callerAddress(ctx, info)
		}),
		ratelimitinggrpc.WithLogger(logger),
		ratelimitinggrpc.WithTracerProvider(tracerProvider),
		ratelimitinggrpc.WithMetricsProvider(metricsProvider),
	)
}

// callerAddress is the address a gRPC call is counted against.
//
// Every public request reaches this server through Caddy, which replaces any X-Forwarded-For a
// client sent with the address it was connected from — it trusts no proxy in front of it — so the
// header's last entry is the client's and cannot be written by them. A call with no header came
// from inside the cluster, and is counted against the connection's own address.
//
// Nothing a client writes is read: a key a caller could choose is a fresh bucket per request.
// That is why the web apps' sign-ins are counted against the address they reach this server from
// — the cluster's egress, shared by everybody signing in through them — rather than against an
// address they could forward on a browser's behalf. The budget is sized with that in mind.
func callerAddress(ctx context.Context, info *grpc.UnaryServerInfo) (string, error) {
	if forwarded := lastForwardedFor(metadata.ValueFromIncomingContext(ctx, forwardedForKey)); forwarded != "" {
		return "ip:" + forwarded, nil
	}

	return ratelimitinggrpc.KeyByPeer()(ctx, info)
}

// lastForwardedFor is the rightmost address in a set of X-Forwarded-For values, or empty.
func lastForwardedFor(values []string) string {
	for i := len(values) - 1; i >= 0; i-- {
		parts := strings.Split(values[i], ",")
		if last := strings.TrimSpace(parts[len(parts)-1]); last != "" {
			return last
		}
	}

	return ""
}

// NewAuthorizeFormThrottle is the HTTP middleware that throttles the OAuth2 login form's POST, by
// the address Caddy stamps, out of the same limiter the gRPC doors spend from. It answers 429 with
// a Retry-After when the budget is spent.
func NewAuthorizeFormThrottle(
	limiter ratelimiting.RateLimiter,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (routing.Middleware, error) {
	address := ratelimitinghttp.KeyByForwardedFor(1)

	return ratelimitinghttp.NewMiddleware(limiter,
		func(req *http.Request) (string, error) {
			key, err := address(req)
			if err != nil || key == "" {
				return "", err
			}

			return authorizeFormKeyPrefix + key, nil
		},
		ratelimitinghttp.WithLogger(logger),
		ratelimitinghttp.WithTracerProvider(tracerProvider),
		ratelimitinghttp.WithMetricsProvider(metricsProvider),
	)
}
