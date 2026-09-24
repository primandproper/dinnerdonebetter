package grpcapi

import (
	"bytes"
	"context"
	"slices"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// chainUnary composes interceptors the way grpc.ChainUnaryInterceptor does: the first is outermost.
func chainUnary(chain []grpc.UnaryServerInterceptor, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) grpc.UnaryHandler {
	for _, c := range slices.Backward(chain) {
		interceptor, next := c, handler
		handler = func(ctx context.Context, req any) (any, error) {
			return interceptor(ctx, req, info, next)
		}
	}

	return handler
}

// chainStream is chainUnary for streams.
func chainStream(chain []grpc.StreamServerInterceptor, info *grpc.StreamServerInfo, handler grpc.StreamHandler) grpc.StreamHandler {
	for _, c := range slices.Backward(chain) {
		interceptor, next := c, handler
		handler = func(srv any, ss grpc.ServerStream) error {
			return interceptor(srv, ss, info, next)
		}
	}

	return handler
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *fakeServerStream) Context() context.Context { return s.ctx }

// internalFailure is a failure shaped the way a handler in this codebase returns one: a sentinel
// deep in a chain whose wrapping names internals, handed to PrepareAndLogGRPCStatus with a
// client-safe description.
type internalFailure struct {
	_ struct{} `json:"-"`

	sentinel    error
	internal    string
	description string
}

func newInternalFailure(t *testing.T) *internalFailure {
	t.Helper()

	return &internalFailure{
		sentinel:    platformerrors.New(fake.Sentence()),
		internal:    fake.UUID(),
		description: fake.Sentence(),
	}
}

func (f *internalFailure) err() error {
	return errorsgrpc.PrepareAndLogGRPCStatus(
		platformerrors.Wrap(f.sentinel, f.internal),
		loggingnoop.NewLogger(), nil, codes.FailedPrecondition, "%s", f.description,
	)
}

// requireNoInternalText asserts that nothing on the status a client receives carries the text the
// message channel was careful to leave out — not the message, and not any detail, however encoded.
func (f *internalFailure) requireNoInternalText(t *testing.T, err error) {
	t.Helper()

	st, ok := status.FromError(err)
	require.True(t, ok)

	assert.NotContains(t, st.Message(), f.internal)
	for _, detail := range st.Proto().GetDetails() {
		assert.False(t, bytes.Contains(detail.GetValue(), []byte(f.internal)), "a %s detail carries internal text", detail.GetTypeUrl())
	}

	decoded := errorsgrpc.DecodeErrorFromStatus(t.Context(), err)
	assert.False(t, platformerrors.Is(decoded, f.sentinel), "the sentinel chain is still reconstructable from the status")
}

func buildTestAuthInterceptor() *interceptors.AuthInterceptor {
	return interceptors.ProvideAuthInterceptor(nil, loggingnoop.NewLogger(), nil, nil, nil, nil, nil, "", nil, realMethodPermissions())
}

func TestErrorEncodingInterceptor_leaksWithoutStripping(T *testing.T) {
	T.Parallel()

	// The reason the strip exists, pinned so it cannot quietly stop being true: on its own, the
	// encoder hands whoever is calling the whole chain, and a stock decoder reads it back.
	T.Run("the encoder alone puts the chain on the wire", func(t *testing.T) {
		t.Parallel()

		f := newInternalFailure(t)

		_, err := errorsgrpc.UnaryErrorEncodingInterceptor()(t.Context(), nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
			return nil, f.err()
		})
		require.Error(t, err)

		decoded := errorsgrpc.DecodeErrorFromStatus(t.Context(), err)
		assert.True(t, platformerrors.Is(decoded, f.sentinel))
		assert.Contains(t, decoded.Error(), f.internal)
	})
}

func TestBuildUnaryServerInterceptors_stripsTheEncodedChain(T *testing.T) {
	T.Parallel()

	authInterceptor := buildTestAuthInterceptor()
	enforcer, enforcerErr := ProvideAuthorizationEnforcer(realMethodPermissions(), authInterceptor, loggingnoop.NewLogger(), metricsnoop.NewMetricsProvider(), false)
	require.NoError(T, enforcerErr)

	passthrough := func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}

	chain := BuildUnaryServerInterceptors(loggingnoop.NewLogger(), authInterceptor, enforcer, passthrough)

	// A method any phone may call without a session, so the failure is the handler's and not a
	// refusal from the interceptors in front of it.
	public := authInterceptor.UnauthenticatedRoutes()
	require.NotEmpty(T, public)
	info := &grpc.UnaryServerInfo{FullMethod: public[0]}

	T.Run("the chain never reaches a client", func(t *testing.T) {
		t.Parallel()

		f := newInternalFailure(t)
		_, err := chainUnary(chain, info, func(context.Context, any) (any, error) {
			return nil, f.err()
		})(t.Context(), nil)
		require.Error(t, err)

		f.requireNoInternalText(t, err)
	})

	T.Run("the code and the client-safe message are untouched", func(t *testing.T) {
		t.Parallel()

		f := newInternalFailure(t)
		_, err := chainUnary(chain, info, func(context.Context, any) (any, error) {
			return nil, f.err()
		})(t.Context(), nil)
		require.Error(t, err)

		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
		assert.Equal(t, f.description, status.Convert(err).Message())
	})

	T.Run("a registered reason still reaches the client", func(t *testing.T) {
		t.Parallel()

		f := newInternalFailure(t)
		reason := errorsgrpc.ClientReason{Err: f.sentinel, Reason: fake.UUID(), Domain: fake.DomainName()}
		errorsgrpc.RegisterClientSafeReasons(reason)

		_, err := chainUnary(chain, info, func(context.Context, any) (any, error) {
			return nil, f.err()
		})(t.Context(), nil)
		require.Error(t, err)

		got, ok := errorsgrpc.ClientReasonFromStatus(err)
		require.True(t, ok)
		assert.Equal(t, reason.Reason, got.GetReason())
		assert.Equal(t, reason.Domain, got.GetDomain())

		f.requireNoInternalText(t, err)
	})

	T.Run("success is untouched", func(t *testing.T) {
		t.Parallel()

		want := fake.UUID()
		resp, err := chainUnary(chain, info, func(context.Context, any) (any, error) {
			return want, nil
		})(t.Context(), nil)
		require.NoError(t, err)
		assert.Equal(t, want, resp)
	})
}

func TestBuildStreamServerInterceptors_stripsTheEncodedChain(T *testing.T) {
	T.Parallel()

	authInterceptor := buildTestAuthInterceptor()
	chain := BuildStreamServerInterceptors(loggingnoop.NewLogger(), authInterceptor)

	public := authInterceptor.UnauthenticatedRoutes()
	require.NotEmpty(T, public)
	info := &grpc.StreamServerInfo{FullMethod: public[0]}

	T.Run("the chain never reaches a client, and the code and message do", func(t *testing.T) {
		t.Parallel()

		f := newInternalFailure(t)
		err := chainStream(chain, info, func(any, grpc.ServerStream) error {
			return f.err()
		})(nil, &fakeServerStream{ctx: t.Context()})
		require.Error(t, err)

		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
		assert.Equal(t, f.description, status.Convert(err).Message())
		f.requireNoInternalText(t, err)
	})
}
