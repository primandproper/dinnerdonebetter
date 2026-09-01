package mealplanning

import (
	"context"
	"time"

	"github.com/primandproper/platform-go/v13/uploads/registry"
)

// RecipeStepImageRow represents a row from recipe_step_images table.
type RecipeStepImageRow struct {
	CreatedAt           time.Time
	ArchivedAt          *time.Time
	ID                  string
	BelongsToRecipeStep string
	UploadedMediaID     string
	UploadedByUser      string
}

// RecipeStepImageDataManager describes a structure capable of storing recipe step image associations.
type RecipeStepImageDataManager interface {
	AddRecipeStepImage(ctx context.Context, recipeStepID, uploadedMediaID, uploadedByUser string) error
	GetRecipeStepImagesByStep(ctx context.Context, recipeStepID string) ([]*RecipeStepImageRow, error)
}

// RecipeStepImageWithUploadedMedia pairs a recipe step image row with its uploaded media.
type RecipeStepImageWithUploadedMedia struct {
	Row           *RecipeStepImageRow
	UploadedMedia *registry.Object
}
