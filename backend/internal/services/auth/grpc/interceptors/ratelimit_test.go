package interceptors

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset/passwordresetpb"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/ratelimiting"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// newOneShotLimiter admits one call per key and then refuses for the rest of the test.
func newOneShotLimiter(t *testing.T) ratelimiting.RateLimiter {
	t.Helper()

	limiter, err := ratelimiting.NewInMemoryRateLimiter(0.0001, 1)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, limiter.Close()) })

	return limiter
}

// fromAddress is a call that arrived through Caddy from address, from inside the cluster.
func fromAddress(ctx context.Context, forwardedFor string) context.Context {
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 443}})
	if forwardedFor == "" {
		return ctx
	}

	return metadata.NewIncomingContext(ctx, metadata.Pairs(forwardedForKey, forwardedFor))
}

func callThrottled(t *testing.T, throttle grpc.UnaryServerInterceptor, ctx context.Context, method string) error {
	t.Helper()

	_, err := throttle(ctx, nil, &grpc.UnaryServerInfo{FullMethod: method}, func(context.Context, any) (any, error) {
		return nil, nil
	})

	return err
}

func TestNewAnonymousDoorThrottle(T *testing.T) {
	T.Parallel()

	T.Run("throttles every anonymous door", func(t *testing.T) {
		t.Parallel()

		throttle, err := NewAnonymousDoorThrottle(newOneShotLimiter(t), noop.NewLogger(), nil, nil)
		require.NoError(t, err)

		ctx := fromAddress(t.Context(), gofakeit.IPv4Address())

		for _, method := range ThrottledMethods() {
			require.NoError(t, callThrottled(t, throttle, ctx, method), method)
			assert.Equal(t, codes.ResourceExhausted, status.Code(callThrottled(t, throttle, ctx, method)), method)
		}
	})

	T.Run("covers the doors the platform says need one", func(t *testing.T) {
		t.Parallel()

		assert.Subset(t, ThrottledMethods(), []string{
			passwordresetpb.PasswordResetService_RequestPasswordReset_FullMethodName,
			signinpb.SignInService_RequestHandleReminder_FullMethodName,
			signinpb.SignInService_RequestVerificationEmailByAddress_FullMethodName,
			signinpb.SignInService_LoginForToken_FullMethodName,
			signinpb.SignInService_AdminLoginForToken_FullMethodName,
			signinpb.SignInService_Register_FullMethodName,
		})
	})

	T.Run("gives each address a budget of its own", func(t *testing.T) {
		t.Parallel()

		throttle, err := NewAnonymousDoorThrottle(newOneShotLimiter(t), noop.NewLogger(), nil, nil)
		require.NoError(t, err)

		method := signinpb.SignInService_LoginForToken_FullMethodName

		require.NoError(t, callThrottled(t, throttle, fromAddress(t.Context(), gofakeit.IPv4Address()), method))
		require.NoError(t, callThrottled(t, throttle, fromAddress(t.Context(), gofakeit.IPv4Address()), method))
	})

	T.Run("reads only the address Caddy stamped", func(t *testing.T) {
		t.Parallel()

		throttle, err := NewAnonymousDoorThrottle(newOneShotLimiter(t), noop.NewLogger(), nil, nil)
		require.NoError(t, err)

		method := signinpb.SignInService_LoginForToken_FullMethodName
		address := gofakeit.IPv4Address()

		require.NoError(t, callThrottled(t, throttle, fromAddress(t.Context(), address), method))

		// A client prepending addresses of its own does not get a fresh bucket.
		spoofed := fromAddress(t.Context(), gofakeit.IPv4Address()+", "+address)
		assert.Equal(t, codes.ResourceExhausted, status.Code(callThrottled(t, throttle, spoofed, method)))
	})

	T.Run("counts a call from inside the cluster against its connection", func(t *testing.T) {
		t.Parallel()

		throttle, err := NewAnonymousDoorThrottle(newOneShotLimiter(t), noop.NewLogger(), nil, nil)
		require.NoError(t, err)

		method := signinpb.SignInService_LoginForToken_FullMethodName

		require.NoError(t, callThrottled(t, throttle, fromAddress(t.Context(), ""), method))
		assert.Equal(t, codes.ResourceExhausted, status.Code(callThrottled(t, throttle, fromAddress(t.Context(), ""), method)))
	})

	T.Run("leaves every other method alone", func(t *testing.T) {
		t.Parallel()

		throttle, err := NewAnonymousDoorThrottle(newOneShotLimiter(t), noop.NewLogger(), nil, nil)
		require.NoError(t, err)

		ctx := fromAddress(t.Context(), gofakeit.IPv4Address())

		for range 3 {
			require.NoError(t, callThrottled(t, throttle, ctx, signinpb.SignInService_ExchangeRefreshToken_FullMethodName))
		}
	})
}

func TestNewAuthorizeFormThrottle(T *testing.T) {
	T.Parallel()

	T.Run("throttles the login form by address", func(t *testing.T) {
		t.Parallel()

		middleware, err := NewAuthorizeFormThrottle(newOneShotLimiter(t), noop.NewLogger(), nil, nil)
		require.NoError(t, err)

		handler := middleware(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
			res.WriteHeader(http.StatusFound)
		}))

		post := func(address string) int {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", http.NoBody)
			req.Header.Set("X-Forwarded-For", address)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)

			return res.Code
		}

		address := gofakeit.IPv4Address()

		assert.Equal(t, http.StatusFound, post(address))
		assert.Equal(t, http.StatusTooManyRequests, post(address))
		assert.Equal(t, http.StatusFound, post(gofakeit.IPv4Address()))
	})
}
