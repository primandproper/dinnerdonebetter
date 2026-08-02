package grpcapi

import (
	"context"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	paymentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/payments"

	cachecfg "github.com/primandproper/platform-go/v9/cache/config"
	distributedlockcfg "github.com/primandproper/platform-go/v9/distributedlock/config"
	idempotencycfg "github.com/primandproper/platform-go/v9/idempotency/config"
	idempotencygrpc "github.com/primandproper/platform-go/v9/idempotency/grpc"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// enabledIdempotencyConfig is a memory-backed manager. That is wrong in production for the
// reason IdempotencyConfig.Enabled documents, and exactly right here: one process, one test.
func enabledIdempotencyConfig() *config.APIServiceConfig {
	return &config.APIServiceConfig{
		Idempotency: config.IdempotencyConfig{
			Enabled: true,
			Manager: idempotencycfg.Config{
				KeyPrefix:   "test.",
				Cache:       cachecfg.Config{Provider: cachecfg.ProviderMemory, Expiry: time.Hour},
				Lock:        distributedlockcfg.Config{Provider: distributedlockcfg.MemoryProvider},
				TTL:         time.Hour,
				InFlightTTL: time.Minute,
			},
		},
	}
}

func withPrincipal(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, sessions.SessionContextDataKey, &sessions.ContextData{
		Requester: sessions.RequesterInfo{UserID: userID},
	})
}

func withIdempotencyKey(ctx context.Context, key string) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs(idempotencygrpc.MetadataKey, key))
}

func buildTestIdempotencyInterceptor(t *testing.T) grpc.UnaryServerInterceptor {
	t.Helper()

	interceptor, err := ProvideIdempotencyInterceptor(
		t.Context(),
		enabledIdempotencyConfig(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		nil,
	)
	require.NoError(t, err)

	return interceptor
}

func TestProvideIdempotencyInterceptor(T *testing.T) {
	T.Parallel()

	T.Run("replays a repeated key instead of re-running the handler", func(t *testing.T) {
		t.Parallel()

		interceptor := buildTestIdempotencyInterceptor(t)

		calls := 0
		handler := func(context.Context, any) (any, error) {
			calls++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}
		ctx := withIdempotencyKey(withPrincipal(t.Context(), "user_1"), "key_1")

		first, err := interceptor(ctx, &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)
		require.NotNil(t, first)

		second, err := interceptor(ctx, &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)
		require.NotNil(t, second)

		assert.Equal(t, 1, calls, "the handler ran twice for one idempotency key")
	})

	T.Run("with a different key the handler runs again", func(t *testing.T) {
		t.Parallel()

		interceptor := buildTestIdempotencyInterceptor(t)

		calls := 0
		handler := func(context.Context, any) (any, error) {
			calls++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}

		_, err := interceptor(withIdempotencyKey(withPrincipal(t.Context(), "user_1"), "key_1"), &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)

		_, err = interceptor(withIdempotencyKey(withPrincipal(t.Context(), "user_1"), "key_2"), &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)

		assert.Equal(t, 2, calls)
	})

	T.Run("with the same key from a different principal the request is refused", func(t *testing.T) {
		t.Parallel()

		interceptor := buildTestIdempotencyInterceptor(t)

		calls := 0
		handler := func(context.Context, any) (any, error) {
			calls++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}

		_, err := interceptor(withIdempotencyKey(withPrincipal(t.Context(), "user_1"), "key_1"), &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)

		// The principal is part of the fingerprint, so a second user reusing the same key
		// reads as a key reused for a different request. That is refused, which is the
		// property that matters: the only unacceptable outcome here would be handing
		// user_2 the response recorded for user_1.
		reply, err := interceptor(withIdempotencyKey(withPrincipal(t.Context(), "user_2"), "key_1"), &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, reply, "another principal's recorded response was returned")

		assert.Equal(t, 1, calls)
	})

	T.Run("with a method outside the filter the handler always runs", func(t *testing.T) {
		t.Parallel()

		interceptor := buildTestIdempotencyInterceptor(t)

		calls := 0
		handler := func(context.Context, any) (any, error) {
			calls++

			return &paymentssvc.GetProductsResponse{}, nil
		}

		// Reads are not idempotency-eligible: replaying them buys nothing and would serve
		// stale data for the whole TTL.
		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_GetProducts_FullMethodName}
		ctx := withIdempotencyKey(withPrincipal(t.Context(), "user_1"), "key_1")

		for range 2 {
			_, err := interceptor(ctx, &paymentssvc.GetProductsRequest{}, info, handler)
			require.NoError(t, err)
		}

		assert.Equal(t, 2, calls)
	})

	T.Run("with no key the handler always runs", func(t *testing.T) {
		t.Parallel()

		interceptor := buildTestIdempotencyInterceptor(t)

		calls := 0
		handler := func(context.Context, any) (any, error) {
			calls++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}
		ctx := withPrincipal(t.Context(), "user_1")

		for range 2 {
			_, err := interceptor(ctx, &paymentssvc.CreateSubscriptionRequest{}, info, handler)
			require.NoError(t, err)
		}

		assert.Equal(t, 2, calls, "a call with no key was treated as idempotent")
	})

	T.Run("when disabled the handler always runs", func(t *testing.T) {
		t.Parallel()

		interceptor, err := ProvideIdempotencyInterceptor(
			t.Context(),
			&config.APIServiceConfig{},
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
			nil,
		)
		require.NoError(t, err)

		calls := 0
		handler := func(context.Context, any) (any, error) {
			calls++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}
		ctx := withIdempotencyKey(withPrincipal(t.Context(), "user_1"), "key_1")

		for range 2 {
			_, err = interceptor(ctx, &paymentssvc.CreateSubscriptionRequest{}, info, handler)
			require.NoError(t, err)
		}

		assert.Equal(t, 2, calls)
	})
}
