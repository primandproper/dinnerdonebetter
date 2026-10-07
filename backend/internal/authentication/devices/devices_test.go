package devices

import (
	"context"
	"net"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// requestFrom is an incoming gRPC request from addr carrying md.
func requestFrom(ctx context.Context, addr string, md metadata.MD) context.Context {
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP(addr), Port: 50051}})
	return metadata.NewIncomingContext(ctx, md)
}

func TestExtract(T *testing.T) {
	T.Parallel()

	T.Run("prefers what a client forwards for the person behind it", func(t *testing.T) {
		t.Parallel()

		forwarded, userAgent, deviceName := gofakeit.IPv4Address(), gofakeit.UserAgent(), gofakeit.Word()
		ctx := requestFrom(t.Context(), gofakeit.IPv4Address(), metadata.Pairs(
			ClientAddressMetadataKey, forwarded,
			ClientUserAgentMetadataKey, userAgent,
			DeviceNameMetadataKey, deviceName,
			forwardedForMetadataKey, gofakeit.IPv4Address(),
			userAgentMetadataKey, gofakeit.UserAgent(),
		))

		origin := Extract(ctx)

		assert.Equal(t, forwarded, origin.IPAddress)
		assert.Equal(t, userAgent, origin.UserAgent)
		assert.Equal(t, deviceName, origin.DeviceName)
	})

	T.Run("falls back to the address the edge stamped, and the client's own user agent", func(t *testing.T) {
		t.Parallel()

		stamped, userAgent := gofakeit.IPv4Address(), gofakeit.UserAgent()
		ctx := requestFrom(t.Context(), gofakeit.IPv4Address(), metadata.Pairs(
			forwardedForMetadataKey, gofakeit.IPv4Address()+", "+stamped,
			userAgentMetadataKey, userAgent,
		))

		origin := Extract(ctx)

		assert.Equal(t, stamped, origin.IPAddress)
		assert.Equal(t, userAgent, origin.UserAgent)
		assert.Empty(t, origin.DeviceName)
	})

	T.Run("falls back to the connection", func(t *testing.T) {
		t.Parallel()

		address := gofakeit.IPv4Address()

		assert.Equal(t, address, Extract(requestFrom(t.Context(), address, metadata.MD{})).IPAddress)
	})

	T.Run("with no request at all", func(t *testing.T) {
		t.Parallel()

		origin := Extract(t.Context())

		assert.Empty(t, origin.IPAddress)
		assert.Empty(t, origin.UserAgent)
	})
}
