package mealplanning

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/identifiers"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

var _ mealplanning.PreparationMediaDataManager = (*repository)(nil)

// AddPreparationMedia adds an uploaded media to a preparation, optionally scoped to an ingredient.
func (q *repository) AddPreparationMedia(ctx context.Context, validPreparationID string, forIngredientID *string, uploadedMediaID string, index int32) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	if uploadedMediaID == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	arg := &generated.CreatePreparationMediaParams{
		ID:                 identifiers.New(),
		ValidPreparationID: validPreparationID,
		UploadedMediaID:    uploadedMediaID,
		Index:              index,
	}
	if forIngredientID != nil && *forIngredientID != "" {
		arg.ForIngredientID = sql.NullString{String: *forIngredientID, Valid: true}
	}

	if err := q.generatedQuerier.CreatePreparationMedia(ctx, q.writeDB, arg); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating preparation media")
	}

	return nil
}

// GetPreparationMediaByPreparation returns all media for a preparation.
func (q *repository) GetPreparationMediaByPreparation(ctx context.Context, validPreparationID string) ([]*mealplanning.PreparationMediaRow, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	rows, err := q.generatedQuerier.GetPreparationMediaByPreparation(ctx, q.readDB, validPreparationID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting preparation media")
	}

	result := make([]*mealplanning.PreparationMediaRow, len(rows))
	for i, r := range rows {
		result[i] = convertPreparationMediaToRow(r)
	}
	return result, nil
}

// GetPreparationMediaByPreparationAndIngredient returns media for a prep+ingredient (ingredient-specific or general).
func (q *repository) GetPreparationMediaByPreparationAndIngredient(ctx context.Context, validPreparationID string, forIngredientID *string) ([]*mealplanning.PreparationMediaRow, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	arg := &generated.GetPreparationMediaByPreparationAndIngredientParams{
		ValidPreparationID: validPreparationID,
	}
	if forIngredientID != nil && *forIngredientID != "" {
		arg.ForIngredientID = sql.NullString{String: *forIngredientID, Valid: true}
	}

	rows, err := q.generatedQuerier.GetPreparationMediaByPreparationAndIngredient(ctx, q.readDB, arg)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting preparation media by ingredient")
	}

	result := make([]*mealplanning.PreparationMediaRow, len(rows))
	for i, r := range rows {
		result[i] = convertPreparationMediaToRow(r)
	}
	return result, nil
}

func convertPreparationMediaToRow(r *generated.PreparationMedia) *mealplanning.PreparationMediaRow {
	row := &mealplanning.PreparationMediaRow{
		ID:                 r.ID,
		ValidPreparationID: r.ValidPreparationID,
		UploadedMediaID:    r.UploadedMediaID,
		Index:              r.Index,
		CreatedAt:          r.CreatedAt,
	}
	if r.ForIngredientID.Valid {
		row.ForIngredientID = new(r.ForIngredientID.String)
	}
	if r.ArchivedAt.Valid {
		row.ArchivedAt = &r.ArchivedAt.Time
	}
	return row
}
