package mealplanning

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
)

// IngredientMediaRow represents a row from ingredient_media table.
type IngredientMediaRow struct {
	CreatedAt         time.Time
	ArchivedAt        *time.Time
	ID                string
	ValidIngredientID string
	UploadedMediaID   string
	Index             int32
}

// IngredientMediaDataManager describes a structure capable of storing ingredient media associations.
type IngredientMediaDataManager interface {
	AddIngredientMedia(ctx context.Context, validIngredientID, uploadedMediaID string, index int32) error
	GetIngredientMediaByIngredient(ctx context.Context, validIngredientID string) ([]*IngredientMediaRow, error)
}

// IngredientMediaWithUploadedMedia pairs an ingredient media row with its uploaded media.
type IngredientMediaWithUploadedMedia struct {
	Row           *IngredientMediaRow
	UploadedMedia *uploadedmedia.UploadedMedia
}
