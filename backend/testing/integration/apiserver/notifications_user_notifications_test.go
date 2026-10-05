package integration

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	notificationspb "github.com/primandproper/platform-go/v15/notifications/notificationspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The inbox's behavior is asserted by platform's notifications conformance suite, run against
// this deployment in conformance_test.go. What remains here is this application's own: the
// audit entries its store records, and the count MarkAllNotificationsRead reports, which the suite leaves to each deployment because it
// is exact only in an inbox nothing else writes to.

// createUserNotificationForTest files a notification the way this application's own
// announcements are filed: through the decorated inbox, under the one topic it uses.
func createUserNotificationForTest(t *testing.T, forUser string) *platformnotifications.Notification {
	t.Helper()

	created, err := createUserNotification(t.Context(), forUser)
	require.NoError(t, err)
	assert.NotNil(t, created)

	return created
}

func createUserNotification(ctx context.Context, forUser string) (*platformnotifications.Notification, error) {
	var created *platformnotifications.Notification

	err := databaseClient.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		created, writeErr = notifsInbox.CreateNotification(ctx, tx, tenancy.Global(), &platformnotifications.Notification{
			ID:        identifiers.New(),
			Principal: forUser,
			Topic:     notifications.DefaultTopic,
			Title:     fake.BuildFakeID(),
		})

		return writeErr
	})

	return created, err
}

func TestUserNotifications_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		retrieved, err := testClient.GetNotification(ctx, &notificationspb.GetNotificationRequest{NotificationId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved.GetResult())

		assert.Equal(t, created.Title, retrieved.GetResult().GetTitle())

		// The write that put it there is in the log. Marking it read would not be, on
		// purpose: somebody opening their own inbox is not a change anybody investigates
		// later, and an entry per read would bury the ones that matter under the traffic
		// of an inbox being opened. read_at is the record, and it is on the row. See
		// notificationsstore/writes.go, which names this and the three other writes it
		// leaves unrecorded.
		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "user_notifications", RelevantID: created.ID},
		})
	})
}

func TestUserNotifications_MarkingRead(T *testing.T) {
	T.Parallel()

	T.Run("marking everything read reports how many it moved", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		for range exampleQuantity {
			createUserNotificationForTest(t, user.ID)
		}

		// At least as many as were written here rather than exactly, so that anything
		// else this deployment files into a new user's inbox does not break it.
		response, err := testClient.MarkAllNotificationsRead(ctx, &notificationspb.MarkAllNotificationsReadRequest{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response.GetMarked(), int64(exampleQuantity))
	})
}

func TestUserNotifications_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		_, err := testClient.ArchiveNotification(ctx, &notificationspb.ArchiveNotificationRequest{
			NotificationId: created.ID,
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "user_notifications", RelevantID: created.ID},
		})
	})
}
