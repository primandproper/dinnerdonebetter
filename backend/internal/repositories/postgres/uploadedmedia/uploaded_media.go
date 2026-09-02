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

# The transaction the events are not in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one cannot: platform's RecordObject
and ArchiveObject own their transactions and take no executor, so the audit
entry and the event are a second transaction after the first has committed.

The gap that opens is the ordinary one — the row lands, the process dies, and
nothing is recorded about it. It is narrow and one-directional: an object can
exist with no event, but no event can name an object that was not registered. It
is the same gap platform-go #457 describes for the comments store, and it is
left open here for the same reason: papering over it locally is how a gap stops
being one anyone remembers.

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

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

// resourceTypeUploadedMedia is what an audit entry about an uploaded object names.
const resourceTypeUploadedMedia = "uploaded_media"

var _ registry.Store = (*repository)(nil)

// RecordObject registers the object, then records it.
func (r *repository) RecordObject(ctx context.Context, object *registry.Object) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.RecordObject(ctx, object); err != nil {
		return err
	}

	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, object.ID)

	return r.record(ctx, object.ID, object.OwnerID, audit.AuditLogEventTypeCreated, ddbuploadedmedia.UploadedMediaCreatedServiceEventType)
}

// ArchiveObject hides the row, then records it.
//
// The owner is read before the archive rather than after, because an audit entry
// names who the row belonged to and the archived row is the one this method is
// about. A read that fails is the archive's failure too: platform answers an
// absent, archived, or other-scope object as ErrObjectNotFound either way, so
// returning it from here is the same answer one call earlier.
func (r *repository) ArchiveObject(ctx context.Context, scope tenancy.Scope, objectID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, objectID)

	object, err := r.GetObject(ctx, scope, objectID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching uploaded media for archive")
	}

	if err = r.Store.ArchiveObject(ctx, scope, objectID); err != nil {
		return err
	}

	return r.record(ctx, objectID, object.OwnerID, audit.AuditLogEventTypeArchived, ddbuploadedmedia.UploadedMediaArchivedServiceEventType)
}

// record writes the audit entry and enqueues the data change event, in one
// transaction of their own.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (r *repository) record(ctx context.Context, objectID, ownerID, auditEventType, changeEventType string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(uploadedmediakeys.UploadedMediaIDKey, objectID)

	return r.client.WithTransaction(ctx, func(tx database.Tx) error {
		return r.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceTypeUploadedMedia,
			RelevantID:    objectID,
			EventType:     auditEventType,
			BelongsToUser: ownerID,
		}, changeEventType, "", map[string]any{
			uploadedmediakeys.UploadedMediaIDKey: objectID,
		})
	})
}
