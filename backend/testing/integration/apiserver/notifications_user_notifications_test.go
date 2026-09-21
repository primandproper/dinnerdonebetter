package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"

	notificationspb "github.com/primandproper/platform-go/v14/notifications/notificationspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The inbox is platform's now, and one RPC has no successor: UpdateUserNotification.
//
// It took a status and wrote whatever it was given, which made "read" a string a client
// could set to anything. Platform models the two things that actually happen to a
// notification as their own calls — MarkNotificationRead, which stamps when it was read,
// and ArchiveNotification, which takes it off the list — so the tests below exercise those
// rather than an update that can no longer be expressed.

func createUserNotificationForTest(t *testing.T, forUser string) *notifications.UserNotification {
	t.Helper()

	ctx := t.Context()

	creationRequestInput := fakes.BuildFakeUserNotification()
	input := converters.ConvertUserNotificationToUserNotificationDatabaseCreationInput(creationRequestInput)
	input.BelongsToUser = forUser

	created, err := notifsRepo.CreateUserNotification(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, created)

	return created
}

func TestUserNotifications_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		retrieved, err := testClient.GetNotification(ctx, &notificationspb.GetNotificationRequest{NotificationId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved.GetResult())

		assert.Equal(t, created.ID, retrieved.GetResult().GetId())
		// The content is the headline. platform splits a notification in two and this
		// application only ever writes the first half — see notificationsstore.Adapter.
		assert.Equal(t, created.Content, retrieved.GetResult().GetTitle())
		// Unread until somebody says otherwise, which is the state the mark-read case below
		// moves it out of.
		assert.Nil(t, retrieved.GetResult().GetReadAt())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetNotification(ctx, &notificationspb.GetNotificationRequest{NotificationId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetNotification(ctx, &notificationspb.GetNotificationRequest{NotificationId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestUserNotifications_MarkingRead(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		_, err := testClient.MarkNotificationRead(ctx, &notificationspb.MarkNotificationReadRequest{
			NotificationId: created.ID,
		})
		require.NoError(t, err)

		// The stamp is the whole state change: a read notification is one with a time on
		// it, rather than one whose status string happens to say "read".
		retrieved, err := testClient.GetNotification(ctx, &notificationspb.GetNotificationRequest{NotificationId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved.GetResult().GetReadAt())

		// And it is off the unread list, which is the read a badge count comes from.
		unread, err := testClient.ListUnreadNotifications(ctx, &notificationspb.ListUnreadNotificationsRequest{})
		require.NoError(t, err)
		for _, notification := range unread.GetResults() {
			assert.NotEqual(t, created.ID, notification.GetId(), "a notification marked read is still unread")
		}

		// The write that put it there is in the log. The read that marked it is not, on
		// purpose: somebody opening their own inbox is not a change anybody investigates
		// later, and an entry per read would bury the ones that matter under the traffic
		// of an inbox being opened. read_at above is the record, and it is on the row.
		// See notificationsstore/writes.go, which names this and the three other writes
		// it leaves unrecorded.
		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "user_notifications", RelevantID: created.ID},
		})
	})

	T.Run("marking everything read reports how many it moved", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		for range exampleQuantity {
			createUserNotificationForTest(t, user.ID)
		}

		response, err := testClient.MarkAllNotificationsRead(ctx, &notificationspb.MarkAllNotificationsReadRequest{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response.GetMarked(), int64(exampleQuantity))

		unread, err := testClient.ListUnreadNotifications(ctx, &notificationspb.ListUnreadNotificationsRequest{})
		require.NoError(t, err)
		assert.Empty(t, unread.GetResults())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.MarkNotificationRead(ctx, &notificationspb.MarkNotificationReadRequest{
			NotificationId: created.ID,
		})
		assert.Error(t, err)
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

		_, err = testClient.GetNotification(ctx, &notificationspb.GetNotificationRequest{NotificationId: created.ID})
		require.Error(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "user_notifications", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveNotification(ctx, &notificationspb.ArchiveNotificationRequest{
			NotificationId: created.ID,
		})
		assert.Error(t, err)
	})
}

func TestUserNotifications_Listing(T *testing.T) {
	T.Parallel()

	u, testClient := createUserAndClientForTest(T)
	createdUserNotifications := []*notifications.UserNotification{}
	for range exampleQuantity {
		created := createUserNotificationForTest(T, u.ID)
		createdUserNotifications = append(createdUserNotifications, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.ListNotifications(ctx, &notificationspb.ListNotificationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.GetResults()), len(createdUserNotifications))

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, u.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "user_notifications"},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ListNotifications(ctx, &notificationspb.ListNotificationsRequest{})
		assert.Error(t, err)
	})
}
