/*
Package settings records what a settings write means to the rest of this
application. The catalog and the answers stored against it are platform-go's:
the schema, the paging, the tenancy column, the enumeration every write is
checked against and the guard that refuses an edit stranding stored values all
live there, and this package neither reimplements nor wraps them.

What it adds is the half platform cannot know about — an audit log entry naming
who did what, and a data change event on the outbox that the webhook dispatcher
fans out. Every event this emits is in the webhook event catalog
(internal/domain/webhooks/catalog), so a subscriber can already ask for them; a
write that skipped the pair would be a row with no provenance and a subscriber
that never heard.

# The transaction the events are in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one now does too. It could not
before: platform's writes owned their transactions and took no executor, so the
audit entry and the event were a second transaction after the first had
committed, and a value could exist that nothing had recorded.

As of platform-go v14 a store write takes the caller's database.Tx, so the
write, the entry and the event are one transaction and share one fate. The gap
that used to be filed for this package as platform-go #460 — and for comments
as #457, waitlists as #458 and payments as #466 — is closed by that convention
rather than by anything here, which is why this package has no workaround to
delete.

# Why a cleared value is recorded apart from a set one

SetValue and ClearValue emit different events because they lead somewhere
different: one is somebody choosing, and the other is somebody withdrawing a
choice and falling back to whatever the catalog says. A subscriber acting on a
preference has to be able to tell "they now want the digest weekly" from "they no
longer have an opinion about the digest", and an event that covered both would
make them read the row to find out.
*/
package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/keys"

	platformsettings "github.com/primandproper/platform-go/v14/settings"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	// resourceTypeSettingDefinitions is what an audit entry about the catalog names.
	resourceTypeSettingDefinitions = "setting_definitions"
	// resourceTypeSettingValues is what an audit entry about somebody's answer names.
	resourceTypeSettingValues = "setting_values"
)

var _ platformsettings.Store = (*repository)(nil)

// CreateDefinition adds the setting to the catalog, then records it.
func (r *repository) CreateDefinition(ctx context.Context, tx database.Tx, scope tenancy.Scope, definition *platformsettings.Definition) (*platformsettings.Definition, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateDefinition(ctx, tx, scope, definition)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, created.ID)
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, created.Name)

	if err = r.recordDefinition(ctx, tx, created, audit.AuditLogEventTypeCreated, ddbsettings.SettingDefinitionCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateDefinition rewrites the setting, then records it.
func (r *repository) UpdateDefinition(ctx context.Context, tx database.Tx, scope tenancy.Scope, definition *platformsettings.Definition) (*platformsettings.Definition, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	updated, err := r.Store.UpdateDefinition(ctx, tx, scope, definition)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, definition.ID)
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, definition.Name)

	if err = r.recordDefinition(ctx, tx, updated, audit.AuditLogEventTypeUpdated, ddbsettings.SettingDefinitionUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return updated, nil
}

// ArchiveDefinition retires the setting, then records it.
//
// The definition is read before the store runs, because the event names the
// setting rather than only the row: a subscriber that heard "some definition was
// archived" would have to look up a row that the archive has already hidden from
// every read that does not ask for archived ones. The read runs on the caller's
// transaction, so it sees that transaction's own earlier writes.
func (r *repository) ArchiveDefinition(ctx context.Context, tx database.Tx, scope tenancy.Scope, definitionID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, definitionID)

	definition, err := r.GetDefinition(ctx, tx, scope, definitionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching setting definition to record")
	}

	if err = r.Store.ArchiveDefinition(ctx, tx, scope, definitionID); err != nil {
		return err
	}

	return r.recordDefinition(ctx, tx, definition, audit.AuditLogEventTypeArchived, ddbsettings.SettingDefinitionArchivedServiceEventType)
}

// SetValue stores the subject's answer, then records it.
func (r *repository) SetValue(ctx context.Context, tx database.Tx, scope tenancy.Scope, subject platformsettings.Subject, name, raw string) (*platformsettings.Value, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, settingskeys.SettingNameKey, name)

	value, err := r.Store.SetValue(ctx, tx, scope, subject, name, raw)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingValueIDKey, value.ID)

	if err = r.recordValue(ctx, tx, value, name, audit.AuditLogEventTypeUpdated, ddbsettings.SettingValueSetServiceEventType); err != nil {
		return nil, err
	}

	return value, nil
}

// ClearValue takes the subject's answer back, then records it.
//
// There is no read before the store runs any more. It used to need one so the
// entry and the event could name the row that was cleared; v14's ClearValue
// returns the row it archived, which is the same value without the round trip
// and without the window between the two statements.
func (r *repository) ClearValue(ctx context.Context, tx database.Tx, scope tenancy.Scope, subject platformsettings.Subject, name string) (*platformsettings.Value, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, settingskeys.SettingNameKey, name)

	cleared, err := r.Store.ClearValue(ctx, tx, scope, subject, name)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingValueIDKey, cleared.ID)

	if err = r.recordValue(ctx, tx, cleared, name, audit.AuditLogEventTypeArchived, ddbsettings.SettingValueClearedServiceEventType); err != nil {
		return nil, err
	}

	return cleared, nil
}

// recordDefinition writes the audit entry and the data change event for a write
// to the catalog.
//
// The entry names no user, because a definition is an administrative row that
// belongs to nobody: who wrote it is the actor on the context, which is what the
// audit recorder resolves. That is the same shape the table this replaced
// recorded under.
func (r *repository) recordDefinition(ctx context.Context, tx database.Tx, definition *platformsettings.Definition, auditEventType, changeEventType string) error {
	return r.record(ctx, tx, "", resourceTypeSettingDefinitions, definition.ID, auditEventType, changeEventType, map[string]any{
		settingskeys.SettingDefinitionIDKey: definition.ID,
		settingskeys.SettingNameKey:         definition.Name,
	})
}

// recordValue writes the audit entry and the data change event for a write to
// one person's answer.
//
// The entry belongs to the value's subject rather than to whoever made the
// request. The two are the same today — nobody may write somebody else's setting
// — and filing it under the subject is what keeps "what has this person chosen,
// and when did they change it" answerable if that ever stops being true.
func (r *repository) recordValue(ctx context.Context, tx database.Tx, value *platformsettings.Value, name, auditEventType, changeEventType string) error {
	return r.record(ctx, tx, value.Subject.ID, resourceTypeSettingValues, value.ID, auditEventType, changeEventType, map[string]any{
		settingskeys.SettingValueIDKey:      value.ID,
		settingskeys.SettingDefinitionIDKey: value.DefinitionID,
		settingskeys.SettingNameKey:         name,
	})
}

// record writes the audit entry and enqueues the data change event, inside the
// caller's transaction.
//
// It used to open one of its own, because the write it describes had already
// committed inside the platform store. As of platform-go v14 that store takes
// the caller's executor, so the write, the entry and the event are one
// transaction: there is no longer a window in which a setting changed and
// nothing recorded that it had.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
//
// The event names no account, which is what makes it a service-wide event: these
// rows live in the global scope, so the account it reaches a webhook subscriber
// under is whichever one the requester had active, resolved from the context by
// the emitter.
func (r *repository) record(
	ctx context.Context,
	tx database.Tx,
	userID, resourceType, relevantID, auditEventType, changeEventType string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, metadata[settingskeys.SettingNameKey])

	return r.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
		ID:            identifiers.New(),
		ResourceType:  resourceType,
		RelevantID:    relevantID,
		EventType:     auditEventType,
		BelongsToUser: userID,
	}, changeEventType, "", metadata)
}
