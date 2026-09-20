package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// The writes that owe a record, and the three that do not.
//
// MarkNotificationRead and MarkAllNotificationsRead are deliberately unrecorded.
// Somebody reading their own notification is not a change anybody investigates
// later, and an entry per read would bury the ones that matter under the
// traffic of an inbox being opened. The row's read_at is the record, and it is
// on the row rather than in the log.
//
// DeleteNotificationsForPrincipal and DeleteDevicesForPrincipal are unrecorded
// for the opposite reason: they are the erasure, and the erasure writes its own
// entry through dataprivacy. A second one naming the rows it destroyed would be
// personal data surviving the erasure that removed it.

// CreateNotification writes the notification, then records it.
func (i *inbox) CreateNotification(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	notification *platformnotifications.Notification,
) (*platformnotifications.Notification, error) {
	ctx, span := i.tracer.StartSpan(ctx)
	defer span.End()

	created, err := i.Inbox.CreateNotification(ctx, tx, scope, notification)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, notificationskeys.UserNotificationIDKey, created.ID)

	if err = i.record(ctx, tx, created.ID, created.Principal,
		resourceTypeUserNotifications,
		audit.AuditLogEventTypeCreated,
		ddbnotifications.UserNotificationCreatedServiceEventType,
		notificationskeys.UserNotificationIDKey); err != nil {
		return nil, err
	}

	return created, nil
}

// ArchiveNotification hides the notification, then records it.
func (i *inbox) ArchiveNotification(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	principal, notificationID string,
) (*platformnotifications.Notification, error) {
	ctx, span := i.tracer.StartSpan(ctx)
	defer span.End()

	archived, err := i.Inbox.ArchiveNotification(ctx, tx, scope, principal, notificationID)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, notificationskeys.UserNotificationIDKey, notificationID)

	if err = i.record(ctx, tx, notificationID, principal,
		resourceTypeUserNotifications,
		audit.AuditLogEventTypeArchived,
		ddbnotifications.UserNotificationUpdatedServiceEventType,
		notificationskeys.UserNotificationIDKey); err != nil {
		return nil, err
	}

	return archived, nil
}

// RegisterDevice records the handset, then records that.
func (r *registry) RegisterDevice(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	device *platformnotifications.Device,
) (*platformnotifications.Device, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	registered, err := r.Registry.RegisterDevice(ctx, tx, scope, device)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, notificationskeys.UserDeviceTokenIDKey, registered.ID)

	if err = r.record(ctx, tx, registered.ID, registered.Principal,
		resourceTypeUserDeviceTokens,
		audit.AuditLogEventTypeCreated,
		ddbnotifications.UserDeviceTokenCreatedServiceEventType,
		notificationskeys.UserDeviceTokenIDKey); err != nil {
		return nil, err
	}

	return registered, nil
}

// RevokeDevice removes the handset, then records it.
//
// The device is returned by the revoke, which is what lets the entry name whose
// handset it was: platform revokes by removal rather than by a flag, so a read
// afterwards would find nothing.
func (r *registry) RevokeDevice(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	principal, deviceID string,
) (*platformnotifications.Device, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	revoked, err := r.Registry.RevokeDevice(ctx, tx, scope, principal, deviceID)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, notificationskeys.UserDeviceTokenIDKey, deviceID)

	if err = r.record(ctx, tx, deviceID, principal,
		resourceTypeUserDeviceTokens,
		audit.AuditLogEventTypeArchived,
		ddbnotifications.UserDeviceTokenArchivedServiceEventType,
		notificationskeys.UserDeviceTokenIDKey); err != nil {
		return nil, err
	}

	return revoked, nil
}

// record writes the audit entry and enqueues the data change event, inside the
// caller's transaction, so they commit with the row they describe or not at all.
func (i *inbox) record(
	ctx context.Context,
	tx database.Tx,
	relevantID, principal, resourceType, auditEventType, changeEventType, logKey string,
) error {
	return recordOne(ctx, tx, i.tracer, i.logger, i.recorder,
		relevantID, principal, resourceType, auditEventType, changeEventType, logKey)
}

func (r *registry) record(
	ctx context.Context,
	tx database.Tx,
	relevantID, principal, resourceType, auditEventType, changeEventType, logKey string,
) error {
	return recordOne(ctx, tx, r.tracer, r.logger, r.recorder,
		relevantID, principal, resourceType, auditEventType, changeEventType, logKey)
}

// recordOne is the body both halves share.
//
// The entry belongs to the principal rather than to whoever made the request.
// Most notifications are written by a background worker with no session at all,
// and an entry filed under nobody is an entry no investigation finds; filing it
// under the person it is about is what keeps "what was this user told, and when"
// answerable.
func recordOne(
	ctx context.Context,
	tx database.Tx,
	tracer tracing.Tracer,
	logger logging.Logger,
	recorder *recording.Recorder,
	relevantID, principal, resourceType, auditEventType, changeEventType, logKey string,
) error {
	ctx, span := tracer.StartSpan(ctx)
	defer span.End()

	return recorder.RecordAndEmit(ctx, tx, logger.WithSpan(span).WithValue(logKey, relevantID), &audit.AuditLogEntry{
		ID:            identifiers.New(),
		ResourceType:  resourceType,
		RelevantID:    relevantID,
		EventType:     auditEventType,
		BelongsToUser: principal,
	}, changeEventType, "", map[string]any{logKey: relevantID})
}
