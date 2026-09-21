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

# The transaction the events are in

Every hand-written repository here emits inside the transaction that wrote
the row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one now does too. It could not
before: platform's writes owned their transactions and took no executor, so
the audit entry and the event were a second transaction after the first had
committed, and a registered object could exist that nothing had recorded.

As of platform-go v14 a store write takes the caller's database.Tx, so the
write, the entry and the event are one transaction and share one fate. The
gap platform-go #457 described for the comments store is closed by that
convention rather than by anything here, which is why this package has no
workaround to delete.

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

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// resourceTypeUploadedMedia is what an audit entry about an uploaded object names.
const resourceTypeUploadedMedia = "uploaded_media"

var _ mediaregistry.Store = (*repository)(nil)

// RecordObject registers the object, then records it.
//
//nolint:gocritic // hugeParam: the value receiver is mediaregistry.Store's signature, not ours
func (r *repository) RecordObject(ctx context.Context, tx database.Tx, scope tenancy.Scope, input mediaregistry.ObjectInput) (*mediaregistry.Object, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	object, err := r.Store.RecordObject(ctx, tx, scope, input)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, object.ID)

	if err = r.record(ctx, tx, object.ID, object.OwnerID, audit.AuditLogEventTypeCreated, ddbuploadedmedia.UploadedMediaCreatedServiceEventType); err != nil {
		return nil, err
	}

	return object, nil
}

// ArchiveObject hides the row, then records it.
//
// The owner no longer needs a read of its own: v14's ArchiveObject returns the
// row it archived, so the entry names who it belonged to without the extra round
// trip, and without the window between the read and the archive.
func (r *repository) ArchiveObject(ctx context.Context, tx database.Tx, scope tenancy.Scope, objectID string) (*mediaregistry.Object, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, objectID)

	archived, err := r.Store.ArchiveObject(ctx, tx, scope, objectID)
	if err != nil {
		return nil, err
	}

	if err = r.record(ctx, tx, objectID, archived.OwnerID, audit.AuditLogEventTypeArchived, ddbuploadedmedia.UploadedMediaArchivedServiceEventType); err != nil {
		return nil, err
	}

	return archived, nil
}

// record writes the audit entry and enqueues the data change event, inside the
// caller's transaction.
//
// It used to open one of its own, because the write it describes had already
// committed inside the platform store. As of platform-go v14 that store takes
// the caller's executor, so the row, the entry and the event commit together.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (r *repository) record(ctx context.Context, tx database.Tx, objectID, ownerID, auditEventType, changeEventType string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(uploadedmediakeys.UploadedMediaIDKey, objectID)

	return r.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
		ID:            identifiers.New(),
		ResourceType:  resourceTypeUploadedMedia,
		RelevantID:    objectID,
		EventType:     auditEventType,
		BelongsToUser: ownerID,
	}, changeEventType, "", map[string]any{
		uploadedmediakeys.UploadedMediaIDKey: objectID,
	})
}
