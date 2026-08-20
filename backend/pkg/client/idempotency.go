package client

import (
	"context"

	"github.com/primandproper/platform-go/v12/idempotency"
	idempotencygrpc "github.com/primandproper/platform-go/v12/idempotency/grpc"

	"google.golang.org/grpc"
)

// WithIdempotency returns a DialOption that stamps the idempotency key from the context onto
// every outgoing call that has one.
//
// The interceptor never invents a key. It cannot: from inside a call, a retry and a deliberate
// second purchase are byte-identical, so a key minted per invocation would look like protection
// and provide none. Only the caller knows where a logical operation begins, which is what
// NewIdempotencyContext expresses.
//
//	conn, _ := client.BuildTLSGRPCClient(addr, client.WithIdempotency())
//
//	ctx, _ = client.NewIdempotencyContext(ctx)   // once, OUTSIDE the retry loop
//	for attempt := range maxAttempts {           // every attempt carries the same key
//		_, err := c.CreateSubscription(ctx, req)
//		...
//	}
//
// A key minted inside the retry loop is a new key per attempt. Nothing on the server can detect
// that mistake, so the ordering above is the whole contract.
func WithIdempotency(opts ...idempotencygrpc.ClientOption) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(idempotencygrpc.NewUnaryClientInterceptor(opts...))
}

// NewIdempotencyContext returns a context carrying a freshly minted idempotency key, and the key
// itself for logging or for reuse across a later process.
//
// Call it once per logical operation, outside whatever loop retries it. Because a request that
// timed out never returned anything, there is deliberately no round trip to acquire a key — the
// client already has everything it needs to mint one.
func NewIdempotencyContext(ctx context.Context) (keyed context.Context, key string) {
	keyed, minted := idempotency.WithNewKey(ctx)

	return keyed, string(minted)
}

// WithIdempotencyKey returns a context carrying a caller-supplied idempotency key.
//
// Use this when the key outlives the process — read back from a job record or a client-side
// queue — so a retry after a restart still resolves to the same operation.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return idempotency.WithKey(ctx, idempotency.Key(key))
}
