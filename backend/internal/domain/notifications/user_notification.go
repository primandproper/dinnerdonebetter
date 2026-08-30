package notifications

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v13/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// UserNotificationCreatedServiceEventType indicates a user notification was created.
	UserNotificationCreatedServiceEventType = "user_notification_created"
	// UserNotificationUpdatedServiceEventType indicates a user notification was updated.
	UserNotificationUpdatedServiceEventType = "user_notification_updated"

	// UserNotificationStatusTypeUnread represents the user notification status type for unread.
	UserNotificationStatusTypeUnread = "unread"
	// UserNotificationStatusTypeRead represents the user notification status type for read.
	UserNotificationStatusTypeRead = "read"
	// UserNotificationStatusTypeDismissed represents the user notification status type for dismissed.
	UserNotificationStatusTypeDismissed = "dismissed"
)

func init() {
	gob.Register(new(UserNotification))
	gob.Register(new(UserNotificationCreationRequestInput))
	gob.Register(new(UserNotificationUpdateRequestInput))
}

type (
	// UserNotification represents a user notification.
	//
	// It is not a duplicate of platform-go's notifications/mobile, despite the shared word.
	// That package is push delivery: it hands a title and a body to APNs or FCM and is done,
	// and this application consumes it directly, unwrapped, in
	// internal/functions/datachangemessagehandler. This is an in-app resource — a row that
	// belongs to a user, carries a Status the user moves between unread, read, and dismissed,
	// and is listed and updated through the API. A push has no status to move and no row to
	// list. Nothing to merge; they are two ways of telling someone something, and an event
	// often does both.
	UserNotification struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time  `json:"createdAt"`
		LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
		ID            string     `json:"id"`
		Content       string     `json:"content"`
		Status        string     `json:"status"`
		BelongsToUser string     `json:"belongsToUser"`
	}

	// UserNotificationCreationRequestInput represents what a user could set as input for creating user notifications.
	UserNotificationCreationRequestInput struct {
		_ struct{} `json:"-"`

		Content       string `json:"content"`
		Status        string `json:"status"`
		BelongsToUser string `json:"belongsToUser"`
	}

	// UserNotificationDatabaseCreationInput represents what a user could set as input for creating user notifications.
	UserNotificationDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID            string `json:"-"`
		Content       string `json:"-"`
		BelongsToUser string `json:"-"`
	}

	// UserNotificationUpdateRequestInput represents what a user could set as input for updating user notifications.
	UserNotificationUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Status *string `json:"status,omitempty"`
	}

	// UserNotificationDataManager describes a structure capable of storing user notifications permanently.
	UserNotificationDataManager interface {
		UserNotificationExists(ctx context.Context, userID, userNotificationID string) (bool, error)
		GetUserNotification(ctx context.Context, userID, userNotificationID string) (*UserNotification, error)
		GetUserNotifications(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[UserNotification], error)
		CreateUserNotification(ctx context.Context, input *UserNotificationDatabaseCreationInput) (*UserNotification, error)
		UpdateUserNotification(ctx context.Context, updated *UserNotification) error
	}
)

// Update merges an UserNotificationUpdateRequestInput with a user notification.
func (x *UserNotification) Update(input *UserNotificationUpdateRequestInput) {
	if input.Status != nil && *input.Status != x.Status {
		x.Status = *input.Status
	}
}

var _ validation.ValidatableWithContext = (*UserNotificationCreationRequestInput)(nil)

// ValidateWithContext validates a UserNotificationCreationRequestInput.
func (x *UserNotificationCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Content, validation.Required),
		validation.Field(&x.Status, validation.In(
			UserNotificationStatusTypeUnread,
			UserNotificationStatusTypeRead,
			UserNotificationStatusTypeDismissed,
		)),
	)
}

var _ validation.ValidatableWithContext = (*UserNotificationDatabaseCreationInput)(nil)

// ValidateWithContext validates a UserNotificationDatabaseCreationInput.
func (x *UserNotificationDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Content, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*UserNotificationUpdateRequestInput)(nil)

// ValidateWithContext validates a UserNotificationUpdateRequestInput.
func (x *UserNotificationUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Status, validation.Required, validation.In(
			UserNotificationStatusTypeUnread,
			UserNotificationStatusTypeRead,
			UserNotificationStatusTypeDismissed,
		)),
	)
}
