package devices

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// fakeStore keeps devices in memory.
type fakeStore struct {
	err      error
	recorded []*Device
}

func (s *fakeStore) RecordSignInDevice(_ context.Context, _ database.SQLQueryExecutor, device *Device) error {
	if s.err != nil {
		return s.err
	}

	s.recorded = append(s.recorded, device)

	return nil
}

func (s *fakeStore) GetSignInDevicesForFamilies(_ context.Context, userID string, familyIDs []string) ([]*Device, error) {
	if s.err != nil {
		return nil, s.err
	}

	var out []*Device
	for _, device := range s.recorded {
		for _, familyID := range familyIDs {
			if device.UserID == userID && device.FamilyID == familyID {
				out = append(out, device)
			}
		}
	}

	return out, nil
}

func (s *fakeStore) GetSignInDevicesForUser(_ context.Context, userID string) ([]*Device, error) {
	if s.err != nil {
		return nil, s.err
	}

	var out []*Device
	for _, device := range s.recorded {
		if device.UserID == userID {
			out = append(out, device)
		}
	}

	return out, nil
}

// fakeHooks counts the AfterIssueToken calls it is handed, and does nothing else.
type fakeHooks struct {
	signin.NoopHooks

	err    error
	issued int
}

func (h *fakeHooks) AfterIssueToken(context.Context, database.Tx, tenancy.Scope, *signin.SignIn) error {
	h.issued++
	return h.err
}

// requestFrom is an incoming gRPC request from addr carrying md.
func requestFrom(ctx context.Context, addr string, md metadata.MD) context.Context {
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP(addr), Port: 50051}})
	return metadata.NewIncomingContext(ctx, md)
}

func buildSignInForTest() *signin.SignIn {
	return &signin.SignIn{
		Principal:             &platformidentity.Principal{User: identityfakes.BuildFakeUser()},
		FamilyID:              identifiers.New(),
		ExpiresAt:             time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
}

func TestFromIncomingContext(T *testing.T) {
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

		device := FromIncomingContext(ctx)

		assert.Equal(t, forwarded, device.IPAddress)
		assert.Equal(t, userAgent, device.UserAgent)
		assert.Equal(t, deviceName, device.DeviceName)
	})

	T.Run("falls back to the address the edge stamped, and the client's own user agent", func(t *testing.T) {
		t.Parallel()

		stamped, userAgent := gofakeit.IPv4Address(), gofakeit.UserAgent()
		ctx := requestFrom(t.Context(), gofakeit.IPv4Address(), metadata.Pairs(
			forwardedForMetadataKey, gofakeit.IPv4Address()+", "+stamped,
			userAgentMetadataKey, userAgent,
		))

		device := FromIncomingContext(ctx)

		assert.Equal(t, stamped, device.IPAddress)
		assert.Equal(t, userAgent, device.UserAgent)
		assert.Empty(t, device.DeviceName)
	})

	T.Run("falls back to the connection", func(t *testing.T) {
		t.Parallel()

		address := gofakeit.IPv4Address()

		assert.Equal(t, address, FromIncomingContext(requestFrom(t.Context(), address, metadata.MD{})).IPAddress)
	})

	T.Run("bounds what a client sends", func(t *testing.T) {
		t.Parallel()

		ctx := requestFrom(t.Context(), gofakeit.IPv4Address(), metadata.Pairs(userAgentMetadataKey, strings.Repeat("a", 10*maxFieldLength)))

		assert.Len(t, FromIncomingContext(ctx).UserAgent, maxFieldLength)
	})

	T.Run("with no request at all", func(t *testing.T) {
		t.Parallel()

		device := FromIncomingContext(t.Context())

		assert.Empty(t, device.IPAddress)
		assert.Empty(t, device.UserAgent)
	})
}

func TestForSignIn(T *testing.T) {
	T.Parallel()

	T.Run("lives as long as the login's refresh token", func(t *testing.T) {
		t.Parallel()

		signIn := buildSignInForTest()

		device := ForSignIn(t.Context(), signIn)

		assert.Equal(t, signIn.FamilyID, device.FamilyID)
		assert.Equal(t, signIn.Principal.User.ID, device.UserID)
		assert.Equal(t, signIn.RefreshTokenExpiresAt, device.ExpiresAt)
	})

	T.Run("or its access token, for a login with no refresh token", func(t *testing.T) {
		t.Parallel()

		signIn := buildSignInForTest()
		signIn.RefreshTokenExpiresAt = time.Time{}

		assert.Equal(t, signIn.ExpiresAt, ForSignIn(t.Context(), signIn).ExpiresAt)
	})
}

func TestNewHooks(T *testing.T) {
	T.Parallel()

	T.Run("records the device a token was issued to", func(t *testing.T) {
		t.Parallel()

		store, inner := &fakeStore{}, &fakeHooks{}
		signIn := buildSignInForTest()
		address := gofakeit.IPv4Address()

		err := NewHooks(inner, store).AfterIssueToken(requestFrom(t.Context(), address, metadata.MD{}), nil, tenancy.Global(), signIn)
		require.NoError(t, err)

		assert.Equal(t, 1, inner.issued)
		require.Len(t, store.recorded, 1)
		assert.Equal(t, signIn.FamilyID, store.recorded[0].FamilyID)
		assert.Equal(t, address, store.recorded[0].IPAddress)
	})

	T.Run("records nothing for an impersonation", func(t *testing.T) {
		t.Parallel()

		store, inner := &fakeStore{}, &fakeHooks{}
		signIn := buildSignInForTest()
		signIn.ActorID = identifiers.New()

		require.NoError(t, NewHooks(inner, store).AfterIssueToken(t.Context(), nil, tenancy.Global(), signIn))

		assert.Equal(t, 1, inner.issued)
		assert.Empty(t, store.recorded)
	})

	T.Run("records nothing when the hooks it wraps refused", func(t *testing.T) {
		t.Parallel()

		expected := errors.New(gofakeit.Sentence())
		store := &fakeStore{}

		err := NewHooks(&fakeHooks{err: expected}, store).AfterIssueToken(t.Context(), nil, tenancy.Global(), buildSignInForTest())

		require.ErrorIs(t, err, expected)
		assert.Empty(t, store.recorded)
	})

	T.Run("fails the sign-in when the device cannot be recorded", func(t *testing.T) {
		t.Parallel()

		expected := errors.New(gofakeit.Sentence())

		err := NewHooks(&fakeHooks{}, &fakeStore{err: expected}).AfterIssueToken(t.Context(), nil, tenancy.Global(), buildSignInForTest())

		require.ErrorIs(t, err, expected)
	})
}

func TestNewAnnotator(T *testing.T) {
	T.Parallel()

	T.Run("answers each login's device as its attributes", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		known := &Device{FamilyID: identifiers.New(), UserID: userID, IPAddress: gofakeit.IPv4Address(), UserAgent: gofakeit.UserAgent()}
		store := &fakeStore{recorded: []*Device{known}}
		unknown := identifiers.New()

		annotations, err := NewAnnotator(store)(t.Context(), tenancy.Global(), userID, []string{known.FamilyID, unknown})
		require.NoError(t, err)

		assert.Equal(t, map[string]string{
			AttributeIPAddress: known.IPAddress,
			AttributeUserAgent: known.UserAgent,
		}, annotations[known.FamilyID])
		assert.NotContains(t, annotations, unknown)
	})

	T.Run("fails the listing when the devices cannot be read", func(t *testing.T) {
		t.Parallel()

		expected := errors.New(gofakeit.Sentence())

		_, err := NewAnnotator(&fakeStore{err: expected})(t.Context(), tenancy.Global(), identifiers.New(), []string{identifiers.New()})

		assert.ErrorIs(t, err, expected)
	})
}

func TestNewCollector(T *testing.T) {
	T.Parallel()

	T.Run("exports every device recorded for the subject", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		device := &Device{FamilyID: identifiers.New(), UserID: userID, IPAddress: gofakeit.IPv4Address()}
		store := &fakeStore{recorded: []*Device{device, {FamilyID: identifiers.New(), UserID: identifiers.New()}}}

		raw, err := NewCollector(store).Collect(t.Context(), tenancy.Global(), platformdataprivacy.Subject{ID: userID})
		require.NoError(t, err)

		var exported []map[string]any
		require.NoError(t, json.Unmarshal(raw, &exported))
		require.Len(t, exported, 1)
		assert.Equal(t, device.IPAddress, exported[0]["ipAddress"])
	})

	T.Run("holds nothing for somebody with no devices", func(t *testing.T) {
		t.Parallel()

		raw, err := NewCollector(&fakeStore{}).Collect(t.Context(), tenancy.Global(), platformdataprivacy.Subject{ID: identifiers.New()})
		require.NoError(t, err)
		assert.Nil(t, raw)
	})
}
