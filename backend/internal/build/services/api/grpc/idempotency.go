package grpcapi

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"

	"github.com/primandproper/platform-go/v12/database"
	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/idempotency"
	idempotencycfg "github.com/primandproper/platform-go/v12/idempotency/config"
	idempotencygrpc "github.com/primandproper/platform-go/v12/idempotency/grpc"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"google.golang.org/grpc"
)

// idempotentMethods are the calls where running the work twice costs real money, so a client
// that retries after a timeout must not be charged again.
//
// Reads are absent because replaying them buys nothing: they have no effect to suppress, and
// recording their responses would spend store capacity on data that is stale by definition.
// Product mutations are absent because they are admin catalog edits, not purchases.
var idempotentMethods = map[string]bool{
	paymentssvc.PaymentsService_CreateSubscription_FullMethodName:  true,
	paymentssvc.PaymentsService_UpdateSubscription_FullMethodName:  true,
	paymentssvc.PaymentsService_ArchiveSubscription_FullMethodName: true,
}

// principalFromContext scopes an idempotency key to the user who sent it.
//
// Without this, two users who happen to mint the same key would collide, and the second would
// be handed the first one's recorded response — a cross-account data leak, not just a wrong
// answer. An unauthenticated call has no principal and is therefore not idempotency-eligible.
func principalFromContext(ctx context.Context) (string, error) {
	sessionCtxData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return "", err
	}

	if sessionCtxData == nil || sessionCtxData.GetUserID() == "" {
		return "", platformerrors.ErrEmptyInputParameter
	}

	return sessionCtxData.GetUserID(), nil
}

// ProvideIdempotencyInterceptor builds the unary interceptor that makes the payment mutations
// at-most-once per client-supplied key.
//
// A call without the idempotency-key metadata passes straight through, so this is opt-in from
// the client's side and changes nothing for callers that do not use it. The key is minted by the
// client once, outside its retry loop — a key minted per attempt looks like protection and
// provides none, and nothing here can detect the mistake.
func ProvideIdempotencyInterceptor(
	ctx context.Context,
	cfg *config.APIServiceConfig,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	dbClient database.Client,
) (grpc.UnaryServerInterceptor, error) {
	// Disabled means no interceptor at all rather than an interceptor over a store that
	// cannot do the job. A pass-through keeps the chain's shape stable.
	if !cfg.Idempotency.Enabled {
		logger.Info("payment idempotency is disabled; retried payment mutations will re-execute")

		return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}, nil
	}

	// The gRPC recordable rule is passed explicitly because the config-level constructor
	// knows nothing about status codes: left to itself it records every outcome, so one
	// Unavailable would replay for the whole TTL and the client could never succeed.
	manager, err := idempotencycfg.NewManager[idempotencygrpc.Response](
		ctx,
		&cfg.Idempotency.Manager,
		dbClient,
		idempotencycfg.WithLogger(logger),
		idempotencycfg.WithTracerProvider(tracerProvider),
		idempotencycfg.WithMetricsProvider(metricsProvider),
		idempotencycfg.WithManagerOptions(idempotency.WithRecordable(idempotencygrpc.Recordable)),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building payments idempotency manager")
	}

	return idempotencygrpc.NewUnaryServerInterceptor(
		manager,
		idempotencygrpc.WithMethodFilter(func(fullMethod string) bool { return idempotentMethods[fullMethod] }),
		idempotencygrpc.WithPrincipalExtractor(principalFromContext),
		idempotencygrpc.WithLogger(logger),
		idempotencygrpc.WithTracerProvider(tracerProvider),
		idempotencygrpc.WithMetricsProvider(metricsProvider),
	)
}
