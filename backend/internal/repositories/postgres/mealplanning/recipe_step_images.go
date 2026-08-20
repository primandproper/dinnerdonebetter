package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/identifiers"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

var _ mealplanning.RecipeStepImageDataManager = (*repository)(nil)

// AddRecipeStepImage adds an uploaded media image to a recipe step.
func (q *repository) AddRecipeStepImage(ctx context.Context, recipeStepID, uploadedMediaID, uploadedByUser string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	if uploadedMediaID == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	if uploadedByUser == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if err := q.generatedQuerier.CreateRecipeStepImage(ctx, q.writeDB, &generated.CreateRecipeStepImageParams{
		ID:                  identifiers.New(),
		BelongsToRecipeStep: recipeStepID,
		UploadedMediaID:     uploadedMediaID,
		UploadedByUser:      uploadedByUser,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating recipe step image")
	}

	return nil
}

// GetRecipeStepImagesByStep returns all images for a recipe step.
func (q *repository) GetRecipeStepImagesByStep(ctx context.Context, recipeStepID string) ([]*mealplanning.RecipeStepImageRow, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	rows, err := q.generatedQuerier.GetRecipeStepImagesByStep(ctx, q.readDB, recipeStepID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting recipe step images")
	}

	result := make([]*mealplanning.RecipeStepImageRow, len(rows))
	for i, r := range rows {
		result[i] = &mealplanning.RecipeStepImageRow{
			ID:                  r.ID,
			BelongsToRecipeStep: r.BelongsToRecipeStep,
			UploadedMediaID:     r.UploadedMediaID,
			UploadedByUser:      r.UploadedByUser,
			CreatedAt:           r.CreatedAt,
		}
		if r.ArchivedAt.Valid {
			result[i].ArchivedAt = &r.ArchivedAt.Time
		}
	}
	return result, nil
}
