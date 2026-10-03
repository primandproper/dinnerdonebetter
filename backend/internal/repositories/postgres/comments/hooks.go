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

Both are written from platform's comments.Hooks, which the store calls on the
caller's transaction once each write has landed. A hook's error fails the
write, so the comment, the entry and the event commit together or not at all.
*/
package comments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	platformcomments "github.com/primandproper/platform-go/v14/comments"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// resourceTypeComments is what an audit entry about a comment names.
const resourceTypeComments = "comments"

// hooks records every comment write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformcomments.Hooks = (*hooks)(nil)

// AfterCreateComment records a comment being written.
func (h *hooks) AfterCreateComment(ctx context.Context, tx database.Tx, _ tenancy.Scope, comment *platformcomments.Comment) error {
	return h.record(ctx, tx, comment, platformaudit.EventCreated, ddbcomments.CommentCreatedServiceEventType, nil)
}

// AfterUpdateComment records a comment being revised, and which of its fields
// moved.
func (h *hooks) AfterUpdateComment(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformcomments.Comment) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated comment")
	}

	return h.record(ctx, tx, after, platformaudit.EventUpdated, ddbcomments.CommentUpdatedServiceEventType, changes)
}

// AfterArchiveComment records a comment being removed from the discussion. The
// entry names whoever wrote it, which the row platform hands the hook still
// says, rather than whoever removed it.
func (h *hooks) AfterArchiveComment(ctx context.Context, tx database.Tx, _ tenancy.Scope, comment *platformcomments.Comment) error {
	return h.record(ctx, tx, comment, platformaudit.EventArchived, ddbcomments.CommentArchivedServiceEventType, nil)
}

// AfterDeleteCommentsForTarget records nothing, deliberately: a sweep was never
// recorded here, and nothing in this application sweeps a target's comments yet.
// The write that removes the target is the one that records the removal.
func (*hooks) AfterDeleteCommentsForTarget(context.Context, database.Tx, tenancy.Scope, platformcomments.Target, int64) error {
	return nil
}

// AfterDeleteCommentsByAuthor records nothing, deliberately: an erasure was never
// recorded here, and an entry naming the author would put back into the audit
// log the reference the erasure exists to remove.
func (*hooks) AfterDeleteCommentsByAuthor(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// record writes the audit entry and enqueues the data change event on the
// transaction the write ran in. changes is the field-level diff an update
// carries, and nil for every other write.
//
// The entry belongs to the comment's author rather than to whoever made the
// request, so that what happened to somebody's words is answerable from the
// audit log under their name.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	comment *platformcomments.Comment,
	auditEventType platformaudit.EventType,
	changeEventType string,
	changes map[string]platformaudit.Change,
) error {
	logger := h.logger.WithValue(commentskeys.CommentIDKey, comment.ID)

	entry := audit.NewEntry(comment.Author, "", resourceTypeComments, comment.ID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, "", map[string]any{
		commentskeys.CommentIDKey: comment.ID,
	})
}
