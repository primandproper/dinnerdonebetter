package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformnotifs "github.com/primandproper/platform-go/v14/notifications"
	platformnotificationsmock "github.com/primandproper/platform-go/v14/notifications/mock"
	"github.com/primandproper/platform-go/v14/notifications/push"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	notifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
	noopnotifications "github.com/primandproper/primitives-go/v2/notifications/mobile/noop"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withFanoutOver hands the handler a push fan-out reading from tokens, so a test can assert what
// reached the devices rather than only what the router decided.
func withFanoutOver(t *testing.T, handler *AsyncDataChangeMessageHandler, devices *platformnotificationsmock.RegistryMock) {
	t.Helper()

	fanout, err := push.NewFanout(
		devices,
		noopnotifications.NewPushNotificationSender(),
		push.WithLogger(loggingnoop.NewLogger()),
		push.WithMetricsProvider(metricsnoop.NewMetricsProvider()),
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
		devices := &platformnotificationsmock.RegistryMock{}
		withFanoutOver(t, handler, devices)

		device := &platformnotifs.Device{
			ID:        fake.BuildFakeID(),
			Token:     strings.Repeat("a", 64),
			Platform:  platformnotifs.PlatformIOS,
			Principal: recipient,
		}
		// One read for the whole recipient set, which is the read platform's fanout makes.
		devices.ListDevicesByPrincipalsFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, principals []string) ([]*platformnotifs.Device, error) {
			assert.Equal(t, []string{recipient}, principals)

			return []*platformnotifs.Device{device}, nil
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
		assert.Len(t, devices.ListDevicesByPrincipalsCalls(), 1)
	})

	// A recipient with no registered device is not a failure. Nothing is owed to somebody who
	// has never opened the app on a phone, and retrying would never produce one.
	t.Run("household invitation accepted succeeds with no registered devices", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		devices := &platformnotificationsmock.RegistryMock{}
		withFanoutOver(t, handler, devices)

		devices.ListDevicesByPrincipalsFunc = func(context.Context, database.SQLQueryExecutor, tenancy.Scope, []string) ([]*platformnotifs.Device, error) {
			return nil, nil
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
