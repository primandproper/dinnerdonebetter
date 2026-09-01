package mealplanning

import (
	"context"

	"github.com/primandproper/platform-go/v13/uploads/registry"
)

type (
	// UploadedMediaFetcher fetches uploaded media by IDs (used for enriching preparations/ingredients with media).
	UploadedMediaFetcher interface {
		GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*registry.Object, error)
	}

	ValidEnumerationDataManager interface {
		ValidIngredientGroupDataManager
		ValidIngredientMeasurementUnitDataManager
		ValidIngredientPreparationDataManager
		ValidPrepTaskConfigDataManager
		ValidIngredientDataManager
		ValidIngredientStateIngredientDataManager
		ValidIngredientStateDataManager
		ValidMeasurementUnitDataManager
		ValidInstrumentDataManager
		ValidMeasurementUnitConversionDataManager
		ValidPreparationInstrumentDataManager
		ValidPreparationDataManager
		ValidPreparationVesselDataManager
		ValidVesselDataManager
		PreparationMediaDataManager
		IngredientMediaDataManager
		UploadedMediaFetcher
	}
)
