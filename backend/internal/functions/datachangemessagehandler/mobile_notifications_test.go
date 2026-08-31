package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	domainnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	notifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	noopnotifications "github.com/primandproper/platform-go/v13/notifications/mobile/noop"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withFanoutOver hands the handler a push fan-out reading from tokens, so a test can assert what
// reached the devices rather than only what the router decided.
func withFanoutOver(t *testing.T, handler *AsyncDataChangeMessageHandler, tokens *notificationsmock.RepositoryMock) {
	t.Helper()

	fanout, err := push.NewFanout(
		loggingnoop.NewLogger(),
		tokens,
		noopnotifications.NewPushNotificationSender(),
		metricsnoop.NewMetricsProvider(),
	)
	require.NoError(t, err)

	handler.pushFanout = fanout
}

func TestMobileNotificationsEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("household invitation accepted pushes to every recipient", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		recipient := fake.BuildFakeID()
		tokens := &notificationsmock.RepositoryMock{}
		withFanoutOver(t, handler, tokens)

		deviceToken := &domainnotifications.UserDeviceToken{
			ID:            fake.BuildFakeID(),
			DeviceToken:   strings.Repeat("a", 64),
			Platform:      domainnotifications.UserDeviceTokenPlatformIOS,
			BelongsToUser: recipient,
		}
		tokens.GetUserDeviceTokensFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[domainnotifications.UserDeviceToken], error) {
			assert.Equal(t, recipient, userID)
			assert.Nil(t, platformFilter)

			return &filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{
				Data: []*domainnotifications.UserDeviceToken{deviceToken},
			}, nil
		}

		req := notifications.MobileNotificationRequest{
			RequestType:      identity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
			RecipientUserIDs: []string{recipient},
			Title:            "Someone joined",
			Body:             "They accepted your invitation",
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		require.NoError(t, handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw))
		assert.Len(t, tokens.GetUserDeviceTokensCalls(), 1)
	})

	// A recipient with no registered device is not a failure. Nothing is owed to somebody who
	// has never opened the app on a phone, and retrying would never produce one.
	t.Run("household invitation accepted succeeds with no registered devices", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		tokens := &notificationsmock.RepositoryMock{}
		withFanoutOver(t, handler, tokens)

		tokens.GetUserDeviceTokensFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter, _ *string) (*filtering.QueryFilteredResult[domainnotifications.UserDeviceToken], error) {
			return &filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{}, nil
		}

		req := notifications.MobileNotificationRequest{
			RequestType:      identity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
			RecipientUserIDs: []string{fake.BuildFakeID()},
			Title:            "Someone joined",
			Body:             "They accepted your invitation",
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		require.NoError(t, handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw))
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), []byte("not json"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding")
	})

	t.Run("missing title", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      identity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
			RecipientUserIDs: []string{fake.BuildFakeID()},
			Title:            "",
			Body:             "body",
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		err = handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("missing body", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      identity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
			RecipientUserIDs: []string{fake.BuildFakeID()},
			Title:            "title",
			Body:             "",
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		err = handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "body")
	})

	t.Run("missing request type", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RecipientUserIDs: []string{fake.BuildFakeID()},
			Title:            "title",
			Body:             "body",
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		err = handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "request type")
	})

	// Meal plan task reminders used to route here. They are claimed from a work queue now, so
	// a message still carrying that type is one nothing should be publishing — and it is
	// rejected rather than quietly delivered by a route left standing for it.
	t.Run("unknown request type", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      "meal_plan_task",
			RecipientUserIDs: []string{fake.BuildFakeID()},
			Title:            "title",
			Body:             "body",
		}
		raw, err := json.Marshal(req)
		require.NoError(t, err)

		err = handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown request type")
	})
}
