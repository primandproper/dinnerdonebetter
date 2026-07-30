package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/keys"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	o11yName = "comments_data_manager"
)

var _ CommentsDataManager = (*commentsManager)(nil)

type commentsManager struct {
	tracer tracing.Tracer
	logger logging.Logger
	repo   comments.Repository
}

// NewCommentsDataManager returns a new CommentsDataManager.
func NewCommentsDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo comments.Repository,
) (CommentsDataManager, error) {
	return &commentsManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
	}, nil
}

func (m *commentsManager) CreateComment(ctx context.Context, input *comments.CommentCreationRequestInput) (*comments.Comment, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, input.BelongsToUser)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating comment creation input")
	}

	dbInput := &comments.CommentDatabaseCreationInput{
		ID:              identifiers.New(),
		Content:         input.Content,
		TargetType:      input.TargetType,
		ReferencedID:    input.ReferencedID,
		ParentCommentID: input.ParentCommentID,
		BelongsToUser:   input.BelongsToUser,
	}

	created, err := m.repo.CreateComment(ctx, dbInput)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, created.ID)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return created, nil
}

func (m *commentsManager) GetComment(ctx context.Context, id string) (*comments.Comment, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetComment(ctx, id)
}

func (m *commentsManager) GetCommentsForReference(ctx context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetCommentsForReference(ctx, targetType, referencedID, filter)
}

func (m *commentsManager) UpdateComment(ctx context.Context, id, belongsToUser string, input *comments.CommentUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(commentskeys.CommentIDKey, id).WithValue(identitykeys.UserIDKey, belongsToUser)
	tracing.AttachToSpan(span, commentskeys.CommentIDKey, id)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "validating comment update input")
	}

	if err := m.repo.UpdateComment(ctx, id, belongsToUser, input.Content); err != nil {
		return err
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *commentsManager) ArchiveComment(ctx context.Context, id string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(commentskeys.CommentIDKey, id)
	tracing.AttachToSpan(span, commentskeys.CommentIDKey, id)

	if err := m.repo.ArchiveComment(ctx, id); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archive comment")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *commentsManager) ArchiveCommentsForReference(ctx context.Context, targetType, referencedID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.ArchiveCommentsForReference(ctx, targetType, referencedID)
}
