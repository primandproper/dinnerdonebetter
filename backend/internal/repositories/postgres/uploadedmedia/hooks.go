/*
Package uploadedmedia records what an upload means to the rest of this
application. The registry itself is platform-go's: the schema, the paging, the
tenancy column, the key uniqueness and the ownership the reads answer from all
live there, and this package neither reimplements nor wraps them.

What it adds is the half platform cannot know about — an audit log entry naming
who did what, and a data change event on the outbox that the webhook dispatcher
fans out. uploaded_media_created and uploaded_media_archived are both in the
webhook event catalog, so a subscriber can already ask for them; a write that
skipped the pair would be a row with no provenance and a subscriber that never
heard.

Both are written from platform's mediaregistry.Hooks, which the store calls on
the caller's transaction once each write has landed. A hook's error fails the
write, so the row, the entry and the event commit together or not at all — the
property every hand-written repository here has (see
internal/repositories/postgres/events), reached without a wrapper around the
store.

# There is no update

The registry has no statement that assigns a column after the insert, and the
absence is deliberate rather than missing: every column is a fact about bytes
that are already in a bucket, so an "update" that moved a row's key or content
type would be a row that had stopped describing its object. Changing what an
uploaded object is means storing new bytes and registering them.
*/
package uploadedmedia

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediakeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// resourceTypeUploadedMedia is what an audit entry about an uploaded object names.
const resourceTypeUploadedMedia = "uploaded_media"

// hooks records every registry write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ mediaregistry.Hooks = (*hooks)(nil)

// AfterRecordObject records an object being registered.
func (h *hooks) AfterRecordObject(ctx context.Context, tx database.Tx, _ tenancy.Scope, object *mediaregistry.Object) error {
	return h.record(ctx, tx, object, platformaudit.EventCreated, ddbuploadedmedia.UploadedMediaCreatedServiceEventType)
}

// AfterArchiveObject records an object's row being hidden. The row is the one the
// archive left, so the entry names whose it was without a read of its own.
func (h *hooks) AfterArchiveObject(ctx context.Context, tx database.Tx, _ tenancy.Scope, object *mediaregistry.Object) error {
	return h.record(ctx, tx, object, platformaudit.EventArchived, ddbuploadedmedia.UploadedMediaArchivedServiceEventType)
}

// AfterArchiveObjectsForOwner records nothing, deliberately: an erasure's archive
// was never recorded here, and an entry naming the owner would put back into the
// audit log the reference the erasure exists to remove.
func (*hooks) AfterArchiveObjectsForOwner(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// record writes the audit entry and enqueues the data change event on the
// transaction the write ran in.
//
// The entry belongs to whoever uploaded the object, not to whoever happened to
// be signed in when it was written: "what happened to this person's uploads" is
// the question the audit log answers about them.
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	object *mediaregistry.Object,
	auditEventType platformaudit.EventType,
	changeEventType string,
) error {
	logger := h.logger.WithValue(uploadedmediakeys.UploadedMediaIDKey, object.ID)

	return h.recorder.RecordAndEmit(ctx, tx, logger, audit.NewEntry(object.OwnerID, "", resourceTypeUploadedMedia, object.ID, auditEventType), changeEventType, "", map[string]any{
		uploadedmediakeys.UploadedMediaIDKey: object.ID,
	})
}
