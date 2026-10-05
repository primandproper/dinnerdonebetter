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

Both are written from platform's settings.Hooks, which the store calls on the
caller's transaction once each write has landed. A hook's error fails the
write, so the row, the entry and the event commit together or not at all.

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
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformsettings "github.com/primandproper/platform-go/v15/settings"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	// resourceTypeSettingDefinitions is what an audit entry about the catalog names.
	resourceTypeSettingDefinitions = "setting_definitions"
	// resourceTypeSettingValues is what an audit entry about somebody's answer names.
	resourceTypeSettingValues = "setting_values"
)

// hooks records every settings write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformsettings.Hooks = (*hooks)(nil)

// AfterCreateDefinition records a setting being added to the catalog.
func (h *hooks) AfterCreateDefinition(ctx context.Context, tx database.Tx, _ tenancy.Scope, definition *platformsettings.Definition) error {
	return h.recordDefinition(ctx, tx, definition, nil, platformaudit.EventCreated, ddbsettings.SettingDefinitionCreatedServiceEventType)
}

// AfterUpdateDefinition records a setting being rewritten, and which of its
// fields moved.
func (h *hooks) AfterUpdateDefinition(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformsettings.Definition) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated setting definition")
	}

	return h.recordDefinition(ctx, tx, after, changes, platformaudit.EventUpdated, ddbsettings.SettingDefinitionUpdatedServiceEventType)
}

// AfterArchiveDefinition records a setting being retired.
//
// The row is the one from before the archive, which is what lets the event
// name the setting rather than only the row: a subscriber that heard "some
// definition was archived" would have to look up a row the archive has already
// hidden from every read.
func (h *hooks) AfterArchiveDefinition(ctx context.Context, tx database.Tx, _ tenancy.Scope, definition *platformsettings.Definition) error {
	return h.recordDefinition(ctx, tx, definition, nil, platformaudit.EventArchived, ddbsettings.SettingDefinitionArchivedServiceEventType)
}

// AfterSetValue records somebody choosing, and what their answer was before.
// A first answer, or one reviving a cleared answer, diffs against nothing.
func (h *hooks) AfterSetValue(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	definition *platformsettings.Definition,
	before, after *platformsettings.Value,
) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated setting value")
	}

	return h.recordValue(ctx, tx, after, definition.Name, changes, platformaudit.EventUpdated, ddbsettings.SettingValueSetServiceEventType)
}

// AfterClearValue records somebody withdrawing a choice. The value is the one
// the clearing archived, which still carries the answer that was withdrawn.
func (h *hooks) AfterClearValue(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	definition *platformsettings.Definition,
	value *platformsettings.Value,
) error {
	return h.recordValue(ctx, tx, value, definition.Name, nil, platformaudit.EventArchived, ddbsettings.SettingValueClearedServiceEventType)
}

// AfterDeleteValuesForSubject records nothing, deliberately: an erasure was
// never recorded here, and an entry naming the subject would put back into the
// audit log the reference the erasure exists to remove.
func (*hooks) AfterDeleteValuesForSubject(context.Context, database.Tx, tenancy.Scope, platformsettings.Subject, int64) error {
	return nil
}

// recordDefinition writes the audit entry and the data change event for a write
// to the catalog.
//
// The entry names no user, because a definition is an administrative row that
// belongs to nobody: who wrote it is the actor on the context, which is what the
// audit recorder resolves. That is the same shape the table this replaced
// recorded under.
func (h *hooks) recordDefinition(
	ctx context.Context,
	tx database.Tx,
	definition *platformsettings.Definition,
	changes map[string]platformaudit.Change,
	auditEventType platformaudit.EventType,
	changeEventType string,
) error {
	return h.record(ctx, tx, "", resourceTypeSettingDefinitions, definition.ID, auditEventType, changeEventType, changes, map[string]any{
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
func (h *hooks) recordValue(
	ctx context.Context,
	tx database.Tx,
	value *platformsettings.Value,
	name string,
	changes map[string]platformaudit.Change,
	auditEventType platformaudit.EventType,
	changeEventType string,
) error {
	return h.record(ctx, tx, value.Subject.ID, resourceTypeSettingValues, value.ID, auditEventType, changeEventType, changes, map[string]any{
		settingskeys.SettingValueIDKey:      value.ID,
		settingskeys.SettingDefinitionIDKey: value.DefinitionID,
		settingskeys.SettingNameKey:         name,
	})
}

// record writes the audit entry and enqueues the data change event on the
// transaction the write ran in. changes is the field-level diff an update
// carries, and nil for every other write.
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
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	userID, resourceType, relevantID string, auditEventType platformaudit.EventType, changeEventType string,
	changes map[string]platformaudit.Change,
	metadata map[string]any,
) error {
	logger := h.logger.WithValue(settingskeys.SettingNameKey, metadata[settingskeys.SettingNameKey])

	entry := audit.NewEntry(userID, "", resourceType, relevantID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, "", metadata)
}
