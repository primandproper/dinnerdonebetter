package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	// resourceTypeUserNotifications is what an audit entry about an inbox row names.
	resourceTypeUserNotifications = "user_notifications"
	// resourceTypeUserDeviceTokens is what an audit entry about a handset names.
	resourceTypeUserDeviceTokens = "user_device_tokens"
)

// hooks records the notification writes that owe a record.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
//
// Four of the eight record nothing, deliberately, and each says why.
// MarkNotificationRead and MarkAllNotificationsRead are somebody reading their
// own inbox, which is not a change anybody investigates later; the erasures
// write their own entry through dataprivacy.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformnotifications.Hooks = (*hooks)(nil)

// AfterCreateNotification records somebody being told something.
func (h *hooks) AfterCreateNotification(ctx context.Context, tx database.Tx, _ tenancy.Scope, n *platformnotifications.Notification) error {
	return h.recordNotification(ctx, tx, n, platformaudit.EventCreated, ddbnotifications.UserNotificationCreatedServiceEventType)
}

// AfterMarkNotificationRead records nothing. An entry per read would bury the
// entries that matter under the traffic of an inbox being opened; the row's
// read_at is the record, and it is on the row rather than in the log.
func (*hooks) AfterMarkNotificationRead(context.Context, database.Tx, tenancy.Scope, *platformnotifications.Notification, *platformnotifications.Notification) error {
	return nil
}

// AfterMarkAllNotificationsRead records nothing, for AfterMarkNotificationRead's reason.
func (*hooks) AfterMarkAllNotificationsRead(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// AfterArchiveNotification records a notification being dismissed.
func (h *hooks) AfterArchiveNotification(ctx context.Context, tx database.Tx, _ tenancy.Scope, n *platformnotifications.Notification) error {
	return h.recordNotification(ctx, tx, n, platformaudit.EventArchived, ddbnotifications.UserNotificationUpdatedServiceEventType)
}

// AfterDeleteNotificationsForPrincipal records nothing. It is the erasure, and
// the erasure writes its own entry through dataprivacy; a second one naming the
// rows it destroyed would be personal data surviving the erasure that removed it.
func (*hooks) AfterDeleteNotificationsForPrincipal(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// AfterRegisterDevice records a handset being registered. A re-registration
// in this scope carries what it changed — a principal that moved is a handset
// that changed hands — and a first registration carries no diff, as a creation
// does nowhere else in this application either.
func (h *hooks) AfterRegisterDevice(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformnotifications.Device) error {
	var changes map[string]platformaudit.Change

	if before != nil {
		var err error
		if changes, err = platformaudit.Diff(before, after); err != nil {
			return platformerrors.Wrap(err, "diffing the re-registered device")
		}
	}

	return h.record(ctx, tx, after.ID, after.Principal,
		resourceTypeUserDeviceTokens,
		platformaudit.EventCreated,
		ddbnotifications.UserDeviceTokenCreatedServiceEventType,
		notificationskeys.UserDeviceTokenIDKey,
		changes)
}

// AfterRevokeDevice records a handset being signed out. The row is the one from
// before the removal, which is what lets the entry name whose handset it was:
// platform revokes by deleting, so a read afterwards would find nothing.
func (h *hooks) AfterRevokeDevice(ctx context.Context, tx database.Tx, _ tenancy.Scope, d *platformnotifications.Device) error {
	return h.record(ctx, tx, d.ID, d.Principal,
		resourceTypeUserDeviceTokens,
		platformaudit.EventArchived,
		ddbnotifications.UserDeviceTokenArchivedServiceEventType,
		notificationskeys.UserDeviceTokenIDKey,
		nil)
}

// AfterDeleteDevicesForPrincipal records nothing, for
// AfterDeleteNotificationsForPrincipal's reason.
func (*hooks) AfterDeleteDevicesForPrincipal(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// recordNotification is the record for a write to one inbox row.
func (h *hooks) recordNotification(
	ctx context.Context,
	tx database.Tx,
	n *platformnotifications.Notification,
	auditEventType platformaudit.EventType,
	changeEventType string,
) error {
	return h.record(ctx, tx, n.ID, n.Principal,
		resourceTypeUserNotifications,
		auditEventType,
		changeEventType,
		notificationskeys.UserNotificationIDKey,
		nil)
}

// record writes the audit entry and enqueues the data change event on the
// transaction the write ran in, so they commit with the row they describe or
// not at all. changes is the field-level diff a re-registration carries, and nil
// for every other write.
//
// The entry belongs to the principal rather than to whoever made the request.
// Most notifications are written by a background worker with no session at all,
// and an entry filed under nobody is an entry no investigation finds; filing it
// under the person it is about is what keeps "what was this user told, and when"
// answerable.
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	relevantID, principal, resourceType string,
	auditEventType platformaudit.EventType,
	changeEventType, logKey string,
	changes map[string]platformaudit.Change,
) error {
	entry := audit.NewEntry(principal, "", resourceType, relevantID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, h.logger.WithValue(logKey, relevantID), entry, changeEventType, "", map[string]any{logKey: relevantID})
}
