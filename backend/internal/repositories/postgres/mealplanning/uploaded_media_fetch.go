package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v12/database"
	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/observability"
)

var _ mealplanning.UploadedMediaFetcher = (*repository)(nil)

// GetUploadedMediaWithIDs fetches uploaded media by IDs.
func (q *repository) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*uploadedmedia.UploadedMedia, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue("id_count", len(ids))

	results, err := q.generatedQuerier.GetUploadedMediaWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media with IDs")
	}

	var uploadedMediaList []*uploadedmedia.UploadedMedia
	for _, result := range results {
		uploadedMediaList = append(uploadedMediaList, &uploadedmedia.UploadedMedia{
			ID:            result.ID,
			StoragePath:   result.StoragePath,
			MimeType:      string(result.MimeType),
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser: result.CreatedByUser,
		})
	}

	return uploadedMediaList, nil
}
