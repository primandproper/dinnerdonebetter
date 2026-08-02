package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

var _ mealplanning.IngredientMediaDataManager = (*repository)(nil)

// AddIngredientMedia adds an uploaded media to an ingredient.
func (q *repository) AddIngredientMedia(ctx context.Context, validIngredientID, uploadedMediaID string, index int32) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validIngredientID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	if uploadedMediaID == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	if err := q.generatedQuerier.CreateIngredientMedia(ctx, q.writeDB, &generated.CreateIngredientMediaParams{
		ID:                identifiers.New(),
		ValidIngredientID: validIngredientID,
		UploadedMediaID:   uploadedMediaID,
		Index:             index,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating ingredient media")
	}

	return nil
}

// GetIngredientMediaByIngredient returns all media for an ingredient.
func (q *repository) GetIngredientMediaByIngredient(ctx context.Context, validIngredientID string) ([]*mealplanning.IngredientMediaRow, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	rows, err := q.generatedQuerier.GetIngredientMediaByIngredient(ctx, q.readDB, validIngredientID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting ingredient media")
	}

	result := make([]*mealplanning.IngredientMediaRow, len(rows))
	for i, r := range rows {
		result[i] = &mealplanning.IngredientMediaRow{
			ID:                r.ID,
			ValidIngredientID: r.ValidIngredientID,
			UploadedMediaID:   r.UploadedMediaID,
			Index:             r.Index,
			CreatedAt:         r.CreatedAt,
		}
		if r.ArchivedAt.Valid {
			result[i].ArchivedAt = &r.ArchivedAt.Time
		}
	}
	return result, nil
}
