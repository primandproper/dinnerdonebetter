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

# The transaction the events are not in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one cannot: platform's writes own
their transactions and take no executor, so the audit entry and the event are a
second transaction after the first has committed.

The gap that opens is the ordinary one — the row lands, the process dies, and
nothing is recorded about it. It is narrow and it is one-directional: a value can
exist with no event, but no event can name a value that was not written. Closing
it needs platform's write methods to accept a database.Tx, which is the same gap
comments has (platform-go #457) and waitlists has (platform-go #458). It is filed
for this package as platform-go #460 rather than worked around here — a gap
papered over locally stops being a gap anyone remembers.

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

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformsettings "github.com/primandproper/platform-go/v13/settings"
	"github.com/primandproper/platform-go/v13/tenancy"
)

const (
	// resourceTypeSettingDefinitions is what an audit entry about the catalog names.
	resourceTypeSettingDefinitions = "setting_definitions"
	// resourceTypeSettingValues is what an audit entry about somebody's answer names.
	resourceTypeSettingValues = "setting_values"
)

var _ platformsettings.Store = (*repository)(nil)

// CreateDefinition adds the setting to the catalog, then records it.
func (r *repository) CreateDefinition(ctx context.Context, scope tenancy.Scope, definition *platformsettings.Definition) (*platformsettings.Definition, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateDefinition(ctx, scope, definition)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, created.ID)
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, created.Name)

	if err = r.recordDefinition(ctx, created, audit.AuditLogEventTypeCreated, ddbsettings.SettingDefinitionCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateDefinition rewrites the setting, then records it.
func (r *repository) UpdateDefinition(ctx context.Context, scope tenancy.Scope, definition *platformsettings.Definition) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.UpdateDefinition(ctx, scope, definition); err != nil {
		return err
	}

	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, definition.ID)
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, definition.Name)

	return r.recordDefinition(ctx, definition, audit.AuditLogEventTypeUpdated, ddbsettings.SettingDefinitionUpdatedServiceEventType)
}

// ArchiveDefinition retires the setting, then records it.
//
// The definition is read before the store runs, because the event names the
// setting rather than only the row: a subscriber that heard "some definition was
// archived" would have to look up a row that the archive has already hidden from
// every read that does not ask for archived ones.
func (r *repository) ArchiveDefinition(ctx context.Context, scope tenancy.Scope, definitionID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, definitionID)

	definition, err := r.GetDefinition(ctx, scope, definitionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching setting definition to record")
	}

	if err = r.Store.ArchiveDefinition(ctx, scope, definitionID); err != nil {
		return err
	}

	return r.recordDefinition(ctx, definition, audit.AuditLogEventTypeArchived, ddbsettings.SettingDefinitionArchivedServiceEventType)
}

// SetValue stores the subject's answer, then records it.
func (r *repository) SetValue(ctx context.Context, scope tenancy.Scope, subject platformsettings.Subject, name, raw string) (*platformsettings.Value, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, settingskeys.SettingNameKey, name)

	value, err := r.Store.SetValue(ctx, scope, subject, name, raw)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingValueIDKey, value.ID)

	if err = r.recordValue(ctx, value, name, audit.AuditLogEventTypeUpdated, ddbsettings.SettingValueSetServiceEventType); err != nil {
		return nil, err
	}

	return value, nil
}

// ClearValue takes the subject's answer back, then records it.
//
// The value is read before the store runs, so the entry and the event can name
// the row that was cleared. Read afterwards it would still be there — clearing
// archives rather than deletes — but the read that found it would have to ask for
// archived rows, which is a different question than "what did this person
// answer".
func (r *repository) ClearValue(ctx context.Context, scope tenancy.Scope, subject platformsettings.Subject, name string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, settingskeys.SettingNameKey, name)

	value, err := r.GetValue(ctx, scope, subject, name)
	if err != nil {
		return observability.PrepareError(err, span, "fetching setting value to record")
	}

	if err = r.Store.ClearValue(ctx, scope, subject, name); err != nil {
		return err
	}

	tracing.AttachToSpan(span, settingskeys.SettingValueIDKey, value.ID)

	return r.recordValue(ctx, value, name, audit.AuditLogEventTypeArchived, ddbsettings.SettingValueClearedServiceEventType)
}

// recordDefinition writes the audit entry and the data change event for a write
// to the catalog.
//
// The entry names no user, because a definition is an administrative row that
// belongs to nobody: who wrote it is the actor on the context, which is what the
// audit recorder resolves. That is the same shape the table this replaced
// recorded under.
func (r *repository) recordDefinition(ctx context.Context, definition *platformsettings.Definition, auditEventType, changeEventType string) error {
	return r.record(ctx, "", resourceTypeSettingDefinitions, definition.ID, auditEventType, changeEventType, map[string]any{
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
func (r *repository) recordValue(ctx context.Context, value *platformsettings.Value, name, auditEventType, changeEventType string) error {
	return r.record(ctx, value.Subject.ID, resourceTypeSettingValues, value.ID, auditEventType, changeEventType, map[string]any{
		settingskeys.SettingValueIDKey:      value.ID,
		settingskeys.SettingDefinitionIDKey: value.DefinitionID,
		settingskeys.SettingNameKey:         name,
	})
}

// record writes the audit entry and enqueues the data change event, in one
// transaction of their own.
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
	userID, resourceType, relevantID, auditEventType, changeEventType string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, metadata[settingskeys.SettingNameKey])

	return r.client.WithTransaction(ctx, func(tx database.Tx) error {
		return r.recordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceType,
			RelevantID:    relevantID,
			EventType:     auditEventType,
			BelongsToUser: userID,
		}, changeEventType, "", metadata)
	})
}
