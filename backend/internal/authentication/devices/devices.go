/*
Package devices reads where a sign-in came from, for "where you're signed in".

Everything else about the device list is platform's authentication/signin/devices: the table, the
hook that records a row on every mint and deletes it when the login ends, the annotator that
answers it back on the listing RPCs, the privacy adapter and the sweep. What platform leaves to
the consumer is the one decision that depends on the deployment — which parts of a request to
trust — and Extract is this application's answer to it.

# What is read, and how far to trust it

The address, the user agent, and a device name. Each is read from the request that minted the
token, in this order:

  - What the client says it is forwarding for somebody else — ClientAddressMetadataKey,
    ClientUserAgentMetadataKey — which is how the web apps, which call this server from their own
    servers, pass on the browser behind a sign-in. Without it every web sign-in would be recorded
    as the web app's.
  - What the edge stamped: the last X-Forwarded-For entry, which Caddy writes and a client cannot.
  - The connection itself, and the client's own user-agent.

A device name is only ever what a client says it is: DeviceNameMetadataKey, which the iOS app
sends.

What a client says about itself is display, not evidence. Somebody signing in to their own
account can make their own screen say what they like, which tells nobody anything; this is never
read for a decision. The rate limiter in front of the sign-in doors reads only what the edge
stamped, for exactly that reason.

The values are bounded — trimmed, made valid UTF-8, cut to a length — by platform's store when
they are written, not here.
*/
package devices

import (
	"context"
	"net"
	"slices"
	"strings"

	platformdevices "github.com/primandproper/platform-go/v15/authentication/signin/devices"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	// ClientAddressMetadataKey is the address of the person a client is calling on behalf of,
	// which a web app's server forwards for the browser it is serving.
	ClientAddressMetadataKey = "x-client-address"
	// ClientUserAgentMetadataKey is the user agent of the person a client is calling on behalf
	// of, which a web app's server forwards for the browser it is serving.
	ClientUserAgentMetadataKey = "x-client-user-agent"
	// DeviceNameMetadataKey is what a client calls the device it is running on.
	DeviceNameMetadataKey = "x-device-name"

	forwardedForMetadataKey = "x-forwarded-for"
	userAgentMetadataKey    = "user-agent"
)

var _ platformdevices.Extractor = Extract

// Extract reads the origin of a gRPC request: the platformdevices.Extractor this application's
// sign-in hooks are built with. Every field it cannot read is empty.
func Extract(ctx context.Context) platformdevices.Origin {
	md, _ := metadata.FromIncomingContext(ctx)

	origin := platformdevices.Origin{
		IPAddress:  firstOf(md, ClientAddressMetadataKey),
		UserAgent:  firstOf(md, ClientUserAgentMetadataKey),
		DeviceName: firstOf(md, DeviceNameMetadataKey),
	}

	if origin.IPAddress == "" {
		origin.IPAddress = lastForwardedFor(md.Get(forwardedForMetadataKey))
	}

	if origin.IPAddress == "" {
		origin.IPAddress = peerAddress(ctx)
	}

	if origin.UserAgent == "" {
		origin.UserAgent = firstOf(md, userAgentMetadataKey)
	}

	return origin
}

func firstOf(md metadata.MD, key string) string {
	for _, value := range md.Get(key) {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

// lastForwardedFor is the rightmost address in a set of X-Forwarded-For values: the one the edge
// wrote. Everything to its left is whatever the client sent.
func lastForwardedFor(values []string) string {
	for _, value := range slices.Backward(values) {
		parts := strings.Split(value, ",")
		if last := strings.TrimSpace(parts[len(parts)-1]); last != "" {
			return last
		}
	}

	return ""
}

func peerAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return ""
	}

	address := p.Addr.String()
	if host, _, err := net.SplitHostPort(address); err == nil {
		address = host
	}

	return address
}
