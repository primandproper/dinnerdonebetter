package manager

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/keys"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	o11yName = "comments_data_manager"
)

var _ CommentsDataManager = (*commentsManager)(nil)

type commentsManager struct {
	tracer               tracing.Tracer
	logger               logging.Logger
	repo                 comments.Repository
	dataChangesPublisher messagequeue.Publisher
}

// NewCommentsDataManager returns a new CommentsDataManager.
func NewCommentsDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo comments.Repository,
	cfg *msgconfig.QueuesConfig,
	publisherProvider messagequeue.PublisherProvider,
) (CommentsDataManager, error) {
	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide publisher for data changes topic: %w", err)
	}

	return &commentsManager{
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:               logging.NewNamedLogger(logger, o11yName),
		repo:                 repo,
		dataChangesPublisher: dataChangesPublisher,
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
	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, comments.CommentCreatedServiceEventType, map[string]any{
		commentskeys.CommentIDKey: created.ID,
	}))

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

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, comments.CommentUpdatedServiceEventType, map[string]any{
		commentskeys.CommentIDKey: id,
	}))

	return nil
}

func (m *commentsManager) ArchiveComment(ctx context.Context, id string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(commentskeys.CommentIDKey, id)
	tracing.AttachToSpan(span, commentskeys.CommentIDKey, id)

	if err := m.repo.ArchiveComment(ctx, id); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, comments.CommentArchivedServiceEventType, map[string]any{
		commentskeys.CommentIDKey: id,
	}))

	return nil
}

func (m *commentsManager) ArchiveCommentsForReference(ctx context.Context, targetType, referencedID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.ArchiveCommentsForReference(ctx, targetType, referencedID)
}
