package manager

import (
	"context"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	o11yName = "uploaded_media_data_manager"
)

var (
	_ uploadedmedia.Repository = (*uploadedMediaManager)(nil)
	_ UploadedMediaManager     = (*uploadedMediaManager)(nil)
)

type uploadedMediaManager struct {
	tracer tracing.Tracer
	logger logging.Logger
	repo   uploadedmedia.Repository
}

// NewUploadedMediaDataManager returns a new UploadedMediaManager that wraps the uploaded media repository.
func NewUploadedMediaDataManager(
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo uploadedmedia.Repository,
) UploadedMediaManager {
	return &uploadedMediaManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
	}
}

func (m *uploadedMediaManager) GetUploadedMedia(ctx context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetUploadedMedia(ctx, uploadedMediaID)
}

func (m *uploadedMediaManager) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*uploadedmedia.UploadedMedia, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetUploadedMediaWithIDs(ctx, ids)
}

func (m *uploadedMediaManager) GetUploadedMediaForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetUploadedMediaForUser(ctx, userID, filter)
}

func (m *uploadedMediaManager) CreateUploadedMedia(ctx context.Context, input *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, input.CreatedByUser)

	created, err := m.repo.CreateUploadedMedia(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating uploaded media")
	}

	return created, nil
}

func (m *uploadedMediaManager) UpdateUploadedMedia(ctx context.Context, uploadedMedia *uploadedmedia.UploadedMedia) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.UpdateUploadedMedia(ctx, uploadedMedia)
}

func (m *uploadedMediaManager) ArchiveUploadedMedia(ctx context.Context, uploadedMediaID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.ArchiveUploadedMedia(ctx, uploadedMediaID)
}
