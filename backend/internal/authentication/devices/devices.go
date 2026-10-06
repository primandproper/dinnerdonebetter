/*
Package devices records where each sign-in came from, and answers it back on "where you're signed
in".

platform's sign-in lists a person's logins and records how each one happened, and stores nothing
about the device behind one: whether a device is recorded at all, under which keys and for how
long is the consumer's. This is this application's answer. A row per login, keyed by its refresh
token family, written by the AfterIssueToken hook on the token's own transaction and renewed on
every refresh; read back by the annotator that fills each listed login's attributes; exported
with the rest of what this application holds about a person; and swept by the db-cleaner job once
the login could no longer be alive.

# What is recorded, and how far to trust it

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
*/
package devices

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/primitives-go/v2/database"

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

	// AttributeIPAddress is the listed login attribute naming the address it was last renewed
	// from.
	AttributeIPAddress = "ip_address"
	// AttributeUserAgent is the listed login attribute naming the user agent it was last
	// renewed by.
	AttributeUserAgent = "user_agent"
	// AttributeDeviceName is the listed login attribute naming the device that holds it, when
	// the client said.
	AttributeDeviceName = "device_name"

	// maxFieldLength bounds every recorded field. They arrive from clients, and a user agent a
	// client chose to make a megabyte long is not one to store a megabyte of.
	maxFieldLength = 512
)

// Device is where one login was last renewed from.
type Device struct {
	_ struct{} `json:"-"`

	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	FamilyID   string    `json:"familyID"`
	UserID     string    `json:"-"`
	IPAddress  string    `json:"ipAddress,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	DeviceName string    `json:"deviceName,omitempty"`
}

// Attributes renders a device as a listed login's attributes, naming only what is known.
func (d *Device) Attributes() map[string]string {
	attributes := map[string]string{}

	for key, value := range map[string]string{
		AttributeIPAddress:  d.IPAddress,
		AttributeUserAgent:  d.UserAgent,
		AttributeDeviceName: d.DeviceName,
	} {
		if value != "" {
			attributes[key] = value
		}
	}

	return attributes
}

// Store keeps devices.
type Store interface {
	// RecordSignInDevice writes where a login was renewed from, on the executor it is handed,
	// replacing what was recorded for it before.
	RecordSignInDevice(ctx context.Context, q database.SQLQueryExecutor, device *Device) error
	// GetSignInDevicesForFamilies reads the devices recorded for a user's logins among
	// familyIDs. A login with nothing recorded is absent from the answer.
	GetSignInDevicesForFamilies(ctx context.Context, userID string, familyIDs []string) ([]*Device, error)
	// GetSignInDevicesForUser reads every device recorded for a user.
	GetSignInDevicesForUser(ctx context.Context, userID string) ([]*Device, error)
}

// FromIncomingContext reads the device behind a gRPC request. Every field it cannot read is empty.
func FromIncomingContext(ctx context.Context) *Device {
	md, _ := metadata.FromIncomingContext(ctx)

	device := &Device{
		IPAddress:  firstOf(md, ClientAddressMetadataKey),
		UserAgent:  firstOf(md, ClientUserAgentMetadataKey),
		DeviceName: firstOf(md, DeviceNameMetadataKey),
	}

	if device.IPAddress == "" {
		device.IPAddress = lastForwardedFor(md.Get(forwardedForMetadataKey))
	}

	if device.IPAddress == "" {
		device.IPAddress = peerAddress(ctx)
	}

	if device.UserAgent == "" {
		device.UserAgent = firstOf(md, userAgentMetadataKey)
	}

	return device
}

// ForSignIn is the device a sign-in was just issued for, read from the request issuing it.
func ForSignIn(ctx context.Context, signIn *signin.SignIn) *Device {
	device := FromIncomingContext(ctx)

	device.FamilyID = signIn.FamilyID
	if signIn.Principal != nil && signIn.Principal.User != nil {
		device.UserID = signIn.Principal.User.ID
	}

	// The row lives as long as the login could: until its refresh token would lapse, or, for a
	// login that has none, until its access token does.
	device.ExpiresAt = signIn.RefreshTokenExpiresAt
	if signIn.ExpiresAt.After(device.ExpiresAt) {
		device.ExpiresAt = signIn.ExpiresAt
	}

	return device
}

func firstOf(md metadata.MD, key string) string {
	for _, value := range md.Get(key) {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return truncate(trimmed)
		}
	}

	return ""
}

// lastForwardedFor is the rightmost address in a set of X-Forwarded-For values: the one the edge
// wrote. Everything to its left is whatever the client sent.
func lastForwardedFor(values []string) string {
	for i := len(values) - 1; i >= 0; i-- {
		parts := strings.Split(values[i], ",")
		if last := strings.TrimSpace(parts[len(parts)-1]); last != "" {
			return truncate(last)
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

	return truncate(address)
}

func truncate(value string) string {
	if len(value) <= maxFieldLength {
		return value
	}

	return value[:maxFieldLength]
}
