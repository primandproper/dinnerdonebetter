/*
Package comments records what a comment write means to the rest of this
application. The comments themselves are platform-go's: the schema, the paging,
the thread depth, the tenancy column and the erasure all live there, and this
package neither reimplements nor wraps them.

What it adds is the half platform cannot know about — an audit log entry naming
who did what, and a data change event on the outbox that the webhook dispatcher
fans out. comment_created, comment_updated and comment_archived are all in the
webhook event catalog, so a subscriber can already ask for them; a write that
skipped the pair would be a row with no provenance and a subscriber that never
heard.

# The transaction the events are not in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one cannot: platform's
CreateComment, UpdateComment and ArchiveComment own their transactions and take
no executor, so the audit entry and the event are a second transaction after the
first has committed.

The gap that opens is the ordinary one — the comment lands, the process dies, and
nothing is recorded about it. It is narrow and it is one-directional: a comment
can exist with no event, but no event can name a comment that was not written.
Closing it needs platform's write methods to accept a database.Tx the way
DeleteCommentsForTarget and DeleteCommentsByAuthor already do. That is filed
upstream as platform-go #457 rather than worked around here — a gap papered over
locally stops being a gap anyone remembers.
*/
package comments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
)

// resourceTypeComments is what an audit entry about a comment names.
const resourceTypeComments = "comments"

var _ platformcomments.Store = (*repository)(nil)

// CreateComment writes the comment, then records it.
func (q *repository) CreateComment(ctx context.Context, comment *platformcomments.Comment) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if err := q.Store.CreateComment(ctx, comment); err != nil {
		return err
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, comment.ID)

	return q.record(ctx, comment.ID, comment.Author, audit.AuditLogEventTypeCreated, ddbcomments.CommentCreatedServiceEventType)
}

// UpdateComment revises the body, then records it.
func (q *repository) UpdateComment(ctx context.Context, comment *platformcomments.Comment) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if err := q.Store.UpdateComment(ctx, comment); err != nil {
		return err
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, comment.ID)

	return q.record(ctx, comment.ID, comment.Author, audit.AuditLogEventTypeUpdated, ddbcomments.CommentUpdatedServiceEventType)
}

// ArchiveComment removes the comment from the discussion, then records it.
//
// The author is read before the archive rather than after, because an audit entry
// names who the row belonged to and the archived row is the one this method is
// about. A read that fails is the archive's failure too: platform answers an
// absent, archived, or other-scope comment as ErrCommentNotFound either way, so
// returning it from here is the same answer one call earlier.
func (q *repository) ArchiveComment(ctx context.Context, scope tenancy.Scope, commentID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, commentID)

	comment, err := q.GetComment(ctx, scope, commentID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching comment for archive")
	}

	if err = q.Store.ArchiveComment(ctx, scope, commentID); err != nil {
		return err
	}

	return q.record(ctx, commentID, comment.Author, audit.AuditLogEventTypeArchived, ddbcomments.CommentArchivedServiceEventType)
}

// record writes the audit entry and enqueues the data change event, in one
// transaction of their own.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (q *repository) record(ctx context.Context, commentID, author, auditEventType, changeEventType string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithSpan(span).WithValue(commentskeys.CommentIDKey, commentID)

	return q.client.WithTransaction(ctx, func(tx database.Tx) error {
		if err := q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceTypeComments,
			RelevantID:    commentID,
			EventType:     auditEventType,
			BelongsToUser: author,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
		}

		if err := q.events.Emit(ctx, tx, logger, changeEventType, "", map[string]any{
			commentskeys.CommentIDKey: commentID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "enqueuing data change event")
		}

		return nil
	})
}
