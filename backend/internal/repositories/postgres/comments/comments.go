package comments

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeComments = "comments"
)

var (
	_ types.CommentDataManager = (*repository)(nil)
)

func targetTypeToGenerated(s string) generated.CommentTargetType {
	return generated.CommentTargetType(s)
}

func convertCommentFromGenerated(c *generated.Comments) *types.Comment {
	var parentID *string
	if c.ParentCommentID.Valid {
		parentID = &c.ParentCommentID.String
	}
	return &types.Comment{
		ID:              c.ID,
		Content:         c.Content,
		TargetType:      string(c.TargetType),
		ReferencedID:    c.ReferencedID,
		ParentCommentID: parentID,
		BelongsToUser:   c.BelongsToUser,
		CreatedAt:       c.CreatedAt,
		LastUpdatedAt:   database.TimePointerFromNullTime(c.LastUpdatedAt),
		ArchivedAt:      database.TimePointerFromNullTime(c.ArchivedAt),
	}
}

func convertRowToComment(r *generated.GetCommentsForReferenceRow) *types.Comment {
	var parentID *string
	if r.ParentCommentID.Valid {
		parentID = &r.ParentCommentID.String
	}
	return &types.Comment{
		ID:              r.ID,
		Content:         r.Content,
		TargetType:      string(r.TargetType),
		ReferencedID:    r.ReferencedID,
		ParentCommentID: parentID,
		BelongsToUser:   r.BelongsToUser,
		CreatedAt:       r.CreatedAt,
		LastUpdatedAt:   database.TimePointerFromNullTime(r.LastUpdatedAt),
		ArchivedAt:      database.TimePointerFromNullTime(r.ArchivedAt),
	}
}

// CreateComment creates a comment in the database.
func (q *repository) CreateComment(ctx context.Context, input *types.CommentDatabaseCreationInput) (*types.Comment, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, "comment_id", input.ID)
	logger := q.logger.WithValue("comment_id", input.ID)

	var parentCommentID sql.NullString
	if input.ParentCommentID != nil && *input.ParentCommentID != "" {
		parentCommentID = sql.NullString{String: *input.ParentCommentID, Valid: true}
	}

	x := &types.Comment{
		ID:              input.ID,
		Content:         input.Content,
		TargetType:      input.TargetType,
		ReferencedID:    input.ReferencedID,
		ParentCommentID: input.ParentCommentID,
		BelongsToUser:   input.BelongsToUser,
		CreatedAt:       q.CurrentTime(),
	}

	if err := q.WithTransaction(ctx, func(tx database.Tx) error {
		if err := q.generatedQuerier.CreateComment(ctx, tx, &generated.CreateCommentParams{
			ID:              input.ID,
			Content:         input.Content,
			TargetType:      targetTypeToGenerated(input.TargetType),
			ReferencedID:    input.ReferencedID,
			ParentCommentID: parentCommentID,
			BelongsToUser:   input.BelongsToUser,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing comment creation query")
		}

		if err := q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceTypeComments,
			RelevantID:    x.ID,
			EventType:     audit.AuditLogEventTypeCreated,
			BelongsToUser: x.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.CommentCreatedServiceEventType, "", map[string]any{
			commentskeys.CommentIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, "comment_id", x.ID)
	logger.Info("comment created")

	return x, nil
}

// GetComment fetches a comment from the database.
func (q *repository) GetComment(ctx context.Context, id string) (*types.Comment, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if id == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(commentskeys.CommentIDKey, id)
	tracing.AttachToSpan(span, commentskeys.CommentIDKey, id)

	result, err := q.generatedQuerier.GetComment(ctx, q.readDB, id)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching comment")
	}

	return convertCommentFromGenerated(result), nil
}

// GetCommentsForReference fetches comments for a reference (including replies).
func (q *repository) GetCommentsForReference(ctx context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Comment], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if targetType == "" || referencedID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("target_type", targetType).WithValue("referenced_id", referencedID)
	tracing.AttachToSpan(span, "target_type", targetType)
	tracing.AttachToSpan(span, "referenced_id", referencedID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetCommentsForReference(ctx, q.readDB, &generated.GetCommentsForReferenceParams{
		CreatedAfter:    filterArgs.CreatedAfter,
		CreatedBefore:   filterArgs.CreatedBefore,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		IncludeArchived: filterArgs.IncludeArchived,
		TargetType:      targetTypeToGenerated(targetType),
		ReferencedID:    referencedID,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing comments list retrieval query")
	}

	var (
		data                      = []*types.Comment{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, convertRowToComment(result))
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Comment) string { return t.ID },
		filter,
	), nil
}

// GetCommentsForUser fetches every comment authored by a user.
func (q *repository) GetCommentsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Comment], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("user_id", userID)
	tracing.AttachToSpan(span, "user_id", userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetCommentsForUser(ctx, q.readDB, &generated.GetCommentsForUserParams{
		CreatedAfter:    filterArgs.CreatedAfter,
		CreatedBefore:   filterArgs.CreatedBefore,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		IncludeArchived: filterArgs.IncludeArchived,
		BelongsToUser:   userID,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing user comments list retrieval query")
	}

	var (
		data                      = []*types.Comment{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, convertRowToComment(&generated.GetCommentsForReferenceRow{
			ID:              result.ID,
			Content:         result.Content,
			TargetType:      result.TargetType,
			ReferencedID:    result.ReferencedID,
			ParentCommentID: result.ParentCommentID,
			BelongsToUser:   result.BelongsToUser,
			CreatedAt:       result.CreatedAt,
			LastUpdatedAt:   result.LastUpdatedAt,
			ArchivedAt:      result.ArchivedAt,
		}))
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Comment) string { return t.ID },
		filter,
	), nil
}

// UpdateComment updates a comment in the database.
func (q *repository) UpdateComment(ctx context.Context, id, belongsToUser, content string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" || belongsToUser == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(commentskeys.CommentIDKey, id)
	tracing.AttachToSpan(span, commentskeys.CommentIDKey, id)

	return q.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, err := q.generatedQuerier.UpdateComment(ctx, tx, &generated.UpdateCommentParams{
			Content:       content,
			ID:            id,
			BelongsToUser: belongsToUser,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating comment")
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceTypeComments,
			RelevantID:    id,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: belongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.CommentUpdatedServiceEventType, "", map[string]any{
			commentskeys.CommentIDKey: id,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	})
}

// ArchiveComment archives a comment.
func (q *repository) ArchiveComment(ctx context.Context, id string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(commentskeys.CommentIDKey, id)
	tracing.AttachToSpan(span, commentskeys.CommentIDKey, id)

	comment, getErr := q.GetComment(ctx, id)
	if getErr != nil {
		return observability.PrepareAndLogError(getErr, logger, span, "fetching comment for archive")
	}

	return q.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, err := q.generatedQuerier.ArchiveComment(ctx, tx, id)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving comment")
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceTypeComments,
			RelevantID:    id,
			EventType:     audit.AuditLogEventTypeArchived,
			BelongsToUser: comment.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.CommentArchivedServiceEventType, "", map[string]any{
			commentskeys.CommentIDKey: id,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	})
}
