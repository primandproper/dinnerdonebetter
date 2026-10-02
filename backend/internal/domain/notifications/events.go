package notifications

// The data change events a notification write emits. The store emits them —
// see internal/repositories/postgres/notificationsstore — and the webhook event
// catalog collects them from here.
const (
	// UserNotificationCreatedServiceEventType indicates a user notification was created.
	UserNotificationCreatedServiceEventType = "user_notification_created"
	// UserNotificationUpdatedServiceEventType indicates a user notification was updated.
	UserNotificationUpdatedServiceEventType = "user_notification_updated"

	// UserDeviceTokenCreatedServiceEventType indicates a user device token was created.
	UserDeviceTokenCreatedServiceEventType = "user_device_token_created"
	// UserDeviceTokenArchivedServiceEventType indicates a user device token was archived.
	UserDeviceTokenArchivedServiceEventType = "user_device_token_archived"
)

// DefaultTopic is the category every notification this application writes is filed under.
//
// platform requires one — a client groups, mutes and routes by it, and an inbox where every
// row is uncategorized gives a client nothing to offer somebody who wants fewer of one kind.
// This application has never had the concept: a notification here is a line of text, written
// by one path, and inventing several categories to look richer would be describing a
// distinction the writers do not make.
//
// So there is one, named rather than left empty, and it is the seam to widen the day a
// notification is genuinely a different kind of thing.
const DefaultTopic = "general"
