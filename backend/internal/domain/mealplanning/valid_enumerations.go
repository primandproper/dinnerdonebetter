package mealplanning

import (
	"context"

	"github.com/primandproper/platform-go/v14/mediaregistry"
)

type (
	// UploadedMediaFetcher fetches uploaded media by IDs (used for enriching preparations/ingredients with media).
	UploadedMediaFetcher interface {
		GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*mediaregistry.Object, error)
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
