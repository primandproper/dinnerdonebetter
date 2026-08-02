package mealplanning

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
)

// PreparationMediaRow represents a row from preparation_media table.
type PreparationMediaRow struct {
	CreatedAt          time.Time
	ForIngredientID    *string
	ArchivedAt         *time.Time
	ID                 string
	ValidPreparationID string
	UploadedMediaID    string
	Index              int32
}

// PreparationMediaDataManager describes a structure capable of storing preparation media associations.
type PreparationMediaDataManager interface {
	AddPreparationMedia(ctx context.Context, validPreparationID string, forIngredientID *string, uploadedMediaID string, index int32) error
	GetPreparationMediaByPreparation(ctx context.Context, validPreparationID string) ([]*PreparationMediaRow, error)
	GetPreparationMediaByPreparationAndIngredient(ctx context.Context, validPreparationID string, forIngredientID *string) ([]*PreparationMediaRow, error)
}

// PreparationMediaWithUploadedMedia pairs a preparation media row with its uploaded media.
type PreparationMediaWithUploadedMedia struct {
	Row           *PreparationMediaRow
	UploadedMedia *uploadedmedia.UploadedMedia
}
