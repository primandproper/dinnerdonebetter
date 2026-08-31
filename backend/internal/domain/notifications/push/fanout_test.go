package push

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	platformnotifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const requestType = "test_notification"

// stubTokens is the slice of the notifications manager this package reads, recording what it was
// asked to archive.
type stubTokens struct {
	byUser   map[string][]*notifications.UserDeviceToken
	readErr  error
	archived []string
}

func (s *stubTokens) GetUserDeviceTokens(_ context.Context, userID string, _ *filtering.QueryFilter, _ *string) (*filtering.QueryFilteredResult[notifications.UserDeviceToken], error) {
	if s.readErr != nil {
		return nil, s.readErr
	}

	return &filtering.QueryFilteredResult[notifications.UserDeviceToken]{Data: s.byUser[userID]}, nil
}

func (s *stubTokens) ArchiveUserDeviceToken(_ context.Context, _, tokenID string) error {
	s.archived = append(s.archived, tokenID)

	return nil
}

type stubSender struct {
	err func(token string) error
}

func (s *stubSender) SendPush(_ context.Context, _, token string, _ platformnotifications.PushMessage) error {
	if s.err == nil {
		return nil
	}

	return s.err(token)
}

func buildTestFanout(t *testing.T, tokens *stubTokens, sender *stubSender) *Fanout {
	t.Helper()

	fanout, err := NewFanout(loggingnoop.NewLogger(), tokens, sender, metricsnoop.NewMetricsProvider())
	require.NoError(t, err)

	return fanout
}

func tokenFor(userID string) *notifications.UserDeviceToken {
	return &notifications.UserDeviceToken{
		ID:            fake.BuildFakeID(),
		DeviceToken:   fake.BuildFakeID(),
		Platform:      notifications.UserDeviceTokenPlatformIOS,
		BelongsToUser: userID,
	}
}

func TestFanout_Send(t *testing.T) {
	t.Parallel()

	t.Run("pushes to every device of every recipient", func(t *testing.T) {
		t.Parallel()

		userA, userB := fake.BuildFakeID(), fake.BuildFakeID()
		tokens := &stubTokens{byUser: map[string][]*notifications.UserDeviceToken{
			userA: {tokenFor(userA), tokenFor(userA)},
			userB: {tokenFor(userB)},
		}}

		result, err := buildTestFanout(t, tokens, &stubSender{}).
			Send(t.Context(), requestType, []string{userA, userB}, platformnotifications.PushMessage{Title: "hi", Body: "there"})

		require.NoError(t, err)
		assert.Equal(t, Result{Devices: 3, Delivered: 3}, result)
		assert.True(t, result.Reached())
		assert.False(t, result.Unreachable())
	})

	// The distinction the callers act on: nobody to push to at all, rather than pushes that
	// failed. One is permanent and the other is worth retrying.
	t.Run("reports recipients with no registered devices as unreachable", func(t *testing.T) {
		t.Parallel()

		result, err := buildTestFanout(t, &stubTokens{}, &stubSender{}).
			Send(t.Context(), requestType, []string{fake.BuildFakeID()}, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.True(t, result.Unreachable())
		assert.False(t, result.Reached())
	})

	t.Run("reports devices that all refused as reachable but undelivered", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		tokens := &stubTokens{byUser: map[string][]*notifications.UserDeviceToken{userID: {tokenFor(userID)}}}
		sender := &stubSender{err: func(string) error { return errors.New("APNs is having a moment") }}

		result, err := buildTestFanout(t, tokens, sender).
			Send(t.Context(), requestType, []string{userID}, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.Equal(t, Result{Devices: 1, Delivered: 0}, result)
		assert.False(t, result.Unreachable())
		assert.False(t, result.Reached())
	})

	// One dead iPad must not cost the live phone its notification.
	t.Run("keeps going past a device that refused", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		dead, live := tokenFor(userID), tokenFor(userID)
		tokens := &stubTokens{byUser: map[string][]*notifications.UserDeviceToken{userID: {dead, live}}}
		sender := &stubSender{err: func(token string) error {
			if token == dead.DeviceToken {
				return errors.New("nope")
			}

			return nil
		}}

		result, err := buildTestFanout(t, tokens, sender).
			Send(t.Context(), requestType, []string{userID}, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.Equal(t, Result{Devices: 2, Delivered: 1}, result)
	})

	// A token APNs has disowned costs a doomed send on every future notification until it is
	// retired, so it is retired the first time it is refused for that reason.
	t.Run("archives a token APNs calls bad", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		token := tokenFor(userID)
		tokens := &stubTokens{byUser: map[string][]*notifications.UserDeviceToken{userID: {token}}}
		sender := &stubSender{err: func(string) error { return errors.New("410 BadDeviceToken") }}

		_, err := buildTestFanout(t, tokens, sender).
			Send(t.Context(), requestType, []string{userID}, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.Equal(t, []string{token.ID}, tokens.archived)
	})

	// A transient refusal is not a dead token, and archiving one would silently stop every
	// future notification to that device.
	t.Run("leaves a token alone when the refusal is not about the token", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		tokens := &stubTokens{byUser: map[string][]*notifications.UserDeviceToken{userID: {tokenFor(userID)}}}
		sender := &stubSender{err: func(string) error { return errors.New("503 ServiceUnavailable") }}

		_, err := buildTestFanout(t, tokens, sender).
			Send(t.Context(), requestType, []string{userID}, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.Empty(t, tokens.archived)
	})

	t.Run("skips a token with no device token", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		tokens := &stubTokens{byUser: map[string][]*notifications.UserDeviceToken{
			userID: {{ID: fake.BuildFakeID(), BelongsToUser: userID}},
		}}

		result, err := buildTestFanout(t, tokens, &stubSender{}).
			Send(t.Context(), requestType, []string{userID}, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.True(t, result.Unreachable())
	})

	t.Run("surfaces a failure to read device tokens", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("the database is having a moment")

		_, err := buildTestFanout(t, &stubTokens{readErr: expected}, &stubSender{}).
			Send(t.Context(), requestType, []string{fake.BuildFakeID()}, platformnotifications.PushMessage{})

		assert.ErrorIs(t, err, expected)
	})

	t.Run("does nothing with no recipients", func(t *testing.T) {
		t.Parallel()

		tokens := &stubTokens{}

		result, err := buildTestFanout(t, tokens, &stubSender{}).
			Send(t.Context(), requestType, nil, platformnotifications.PushMessage{})

		require.NoError(t, err)
		assert.Equal(t, Result{}, result)
	})
}
