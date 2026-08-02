package recipevalidator

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeValidator_ValidateAndPopulate(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input is nil")
	})

	T.Run("with valid recipe and all bridge IDs present and matching", func(t *testing.T) {
		t.Parallel()

		// Create a fake preparation that will be used by the step
		preparation := fakes.BuildFakeValidPreparation()

		// Create bridge table entries that use this preparation
		vip := fakes.BuildFakeValidIngredientPreparation()
		vip.Preparation = *preparation

		vimu := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimu.Ingredient = vip.Ingredient // Same ingredient as VIP

		vpi := fakes.BuildFakeValidPreparationInstrument()
		vpi.Preparation = *preparation

		vpv := fakes.BuildFakeValidPreparationVessel()
		vpv.Preparation = *preparation

		// Build the maps
		vipMap := map[string]*mealplanning.ValidIngredientPreparation{vip.ID: vip}
		vimuMap := map[string]*mealplanning.ValidIngredientMeasurementUnit{vimu.ID: vimu}
		vpiMap := map[string]*mealplanning.ValidPreparationInstrument{vpi.ID: vpi}
		vpvMap := map[string]*mealplanning.ValidPreparationVessel{vpv.ID: vpv}

		// Create the recipe input
		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               fakes.BuildFakeID(),
							ValidIngredientPreparationID:     new(vip.ID),
							ValidIngredientMeasurementUnitID: new(vimu.ID),
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           fakes.BuildFakeID(),
							ValidPreparationInstrumentID: new(vpi.ID),
						},
					},
					Vessels: []*mealplanning.RecipeStepVesselDatabaseCreationInput{
						{
							ID:                       fakes.BuildFakeID(),
							ValidPreparationVesselID: new(vpv.ID),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(vipMap, vimuMap, vpiMap, vpvMap)
		err := validator.ValidateAndPopulate(input)

		assert.NoError(t, err)

		// Verify derived IDs are populated
		require.Len(t, input.Steps, 1)
		require.Len(t, input.Steps[0].Ingredients, 1)
		require.Len(t, input.Steps[0].Instruments, 1)
		require.Len(t, input.Steps[0].Vessels, 1)

		assert.NotNil(t, input.Steps[0].Ingredients[0].IngredientID)
		assert.Equal(t, vip.Ingredient.ID, *input.Steps[0].Ingredients[0].IngredientID)
		assert.Equal(t, vimu.MeasurementUnit.ID, input.Steps[0].Ingredients[0].MeasurementUnitID)

		assert.NotNil(t, input.Steps[0].Instruments[0].InstrumentID)
		assert.Equal(t, vpi.Instrument.ID, *input.Steps[0].Instruments[0].InstrumentID)

		assert.NotNil(t, input.Steps[0].Vessels[0].VesselID)
		assert.Equal(t, vpv.Vessel.ID, *input.Steps[0].Vessels[0].VesselID)
	})

	T.Run("with missing ValidIngredientPreparationID (not found)", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                           fakes.BuildFakeID(),
							ValidIngredientPreparationID: new("nonexistent-vip-id"),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 ingredient 0")
		assert.Contains(t, err.Error(), "ValidIngredientPreparation")
		assert.Contains(t, err.Error(), "not found")
	})

	T.Run("with ValidIngredientPreparation with wrong preparation", func(t *testing.T) {
		t.Parallel()

		// Step preparation
		stepPreparation := fakes.BuildFakeValidPreparation()

		// VIP with a DIFFERENT preparation
		vip := fakes.BuildFakeValidIngredientPreparation()
		// vip.Preparation is different from stepPreparation

		vipMap := map[string]*mealplanning.ValidIngredientPreparation{vip.ID: vip}

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: stepPreparation.ID, // Different from vip.Preparation.ID
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                           fakes.BuildFakeID(),
							ValidIngredientPreparationID: new(vip.ID),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(vipMap, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 ingredient 0")
		assert.Contains(t, err.Error(), "ValidIngredientPreparation")
		assert.Contains(t, err.Error(), "is for preparation")
	})

	T.Run("with ValidIngredientMeasurementUnit for wrong ingredient", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		// VIP with ingredient A
		vip := fakes.BuildFakeValidIngredientPreparation()
		vip.Preparation = *preparation

		// VIMU with ingredient B (different from VIP)
		vimu := fakes.BuildFakeValidIngredientMeasurementUnit()
		// vimu.Ingredient is different from vip.Ingredient

		vipMap := map[string]*mealplanning.ValidIngredientPreparation{vip.ID: vip}
		vimuMap := map[string]*mealplanning.ValidIngredientMeasurementUnit{vimu.ID: vimu}

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               fakes.BuildFakeID(),
							ValidIngredientPreparationID:     new(vip.ID),
							ValidIngredientMeasurementUnitID: new(vimu.ID),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(vipMap, vimuMap, nil, nil)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 ingredient 0")
		assert.Contains(t, err.Error(), "ValidIngredientMeasurementUnit")
		assert.Contains(t, err.Error(), "is for ingredient")
	})

	T.Run("with missing ValidPreparationInstrumentID (not found)", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           fakes.BuildFakeID(),
							ValidPreparationInstrumentID: new("nonexistent-vpi-id"),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 instrument 0")
		assert.Contains(t, err.Error(), "ValidPreparationInstrument")
		assert.Contains(t, err.Error(), "not found")
	})

	T.Run("with ValidPreparationInstrument with wrong preparation", func(t *testing.T) {
		t.Parallel()

		stepPreparation := fakes.BuildFakeValidPreparation()

		// VPI with a DIFFERENT preparation
		vpi := fakes.BuildFakeValidPreparationInstrument()
		// vpi.Preparation is different from stepPreparation

		vpiMap := map[string]*mealplanning.ValidPreparationInstrument{vpi.ID: vpi}

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: stepPreparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           fakes.BuildFakeID(),
							ValidPreparationInstrumentID: new(vpi.ID),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, vpiMap, nil)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 instrument 0")
		assert.Contains(t, err.Error(), "ValidPreparationInstrument")
		assert.Contains(t, err.Error(), "is for preparation")
	})

	T.Run("with missing ValidPreparationVesselID (not found)", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Vessels: []*mealplanning.RecipeStepVesselDatabaseCreationInput{
						{
							ID:                       fakes.BuildFakeID(),
							ValidPreparationVesselID: new("nonexistent-vpv-id"),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 vessel 0")
		assert.Contains(t, err.Error(), "ValidPreparationVessel")
		assert.Contains(t, err.Error(), "not found")
	})

	T.Run("with ValidPreparationVessel with wrong preparation", func(t *testing.T) {
		t.Parallel()

		stepPreparation := fakes.BuildFakeValidPreparation()

		// VPV with a DIFFERENT preparation
		vpv := fakes.BuildFakeValidPreparationVessel()
		// vpv.Preparation is different from stepPreparation

		vpvMap := map[string]*mealplanning.ValidPreparationVessel{vpv.ID: vpv}

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: stepPreparation.ID,
					Vessels: []*mealplanning.RecipeStepVesselDatabaseCreationInput{
						{
							ID:                       fakes.BuildFakeID(),
							ValidPreparationVesselID: new(vpv.ID),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, nil, vpvMap)
		err := validator.ValidateAndPopulate(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 vessel 0")
		assert.Contains(t, err.Error(), "ValidPreparationVessel")
		assert.Contains(t, err.Error(), "is for preparation")
	})

	T.Run("recipe step products are skipped (no validation required)", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					// Ingredient that is a recipe step product (has RecipeStepProductID)
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                  fakes.BuildFakeID(),
							RecipeStepProductID: new("some-product-id"),
							// No ValidIngredientPreparationID - should be skipped, not error
						},
					},
					// Instrument that is a recipe step product (has ProductOfRecipeStepIndex)
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                       fakes.BuildFakeID(),
							ProductOfRecipeStepIndex: new(uint64(0)),
							// No ValidPreparationInstrumentID - should be skipped, not error
						},
					},
					// Vessel that is a recipe step product (has ProductOfRecipeStepIndex)
					Vessels: []*mealplanning.RecipeStepVesselDatabaseCreationInput{
						{
							ID:                       fakes.BuildFakeID(),
							ProductOfRecipeStepIndex: new(uint64(0)),
							// No ValidPreparationVesselID - should be skipped, not error
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		// Should succeed because all items are recipe step products and skipped
		assert.NoError(t, err)
	})

	T.Run("derived IDs are correctly populated after validation", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()
		ingredient := fakes.BuildFakeValidIngredient()
		measurementUnit := fakes.BuildFakeValidMeasurementUnit()
		instrument := fakes.BuildFakeValidInstrument()
		vessel := fakes.BuildFakeValidVessel()

		// Create bridge table entries with specific IDs
		vip := &mealplanning.ValidIngredientPreparation{
			ID:          fakes.BuildFakeID(),
			Preparation: *preparation,
			Ingredient:  *ingredient,
		}

		vimu := &mealplanning.ValidIngredientMeasurementUnit{
			ID:              fakes.BuildFakeID(),
			Ingredient:      *ingredient,
			MeasurementUnit: *measurementUnit,
		}

		vpi := &mealplanning.ValidPreparationInstrument{
			ID:          fakes.BuildFakeID(),
			Preparation: *preparation,
			Instrument:  *instrument,
		}

		vpv := &mealplanning.ValidPreparationVessel{
			ID:          fakes.BuildFakeID(),
			Preparation: *preparation,
			Vessel:      *vessel,
		}

		vipMap := map[string]*mealplanning.ValidIngredientPreparation{vip.ID: vip}
		vimuMap := map[string]*mealplanning.ValidIngredientMeasurementUnit{vimu.ID: vimu}
		vpiMap := map[string]*mealplanning.ValidPreparationInstrument{vpi.ID: vpi}
		vpvMap := map[string]*mealplanning.ValidPreparationVessel{vpv.ID: vpv}

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               fakes.BuildFakeID(),
							ValidIngredientPreparationID:     new(vip.ID),
							ValidIngredientMeasurementUnitID: new(vimu.ID),
							// IngredientID and MeasurementUnitID should be nil before validation
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           fakes.BuildFakeID(),
							ValidPreparationInstrumentID: new(vpi.ID),
							// InstrumentID should be nil before validation
						},
					},
					Vessels: []*mealplanning.RecipeStepVesselDatabaseCreationInput{
						{
							ID:                       fakes.BuildFakeID(),
							ValidPreparationVesselID: new(vpv.ID),
							// VesselID should be nil before validation
						},
					},
				},
			},
		}

		// Verify fields are not set before validation
		assert.Nil(t, input.Steps[0].Ingredients[0].IngredientID)
		assert.Empty(t, input.Steps[0].Ingredients[0].MeasurementUnitID)
		assert.Nil(t, input.Steps[0].Instruments[0].InstrumentID)
		assert.Nil(t, input.Steps[0].Vessels[0].VesselID)

		validator := NewRecipeValidator(vipMap, vimuMap, vpiMap, vpvMap)
		err := validator.ValidateAndPopulate(input)
		require.NoError(t, err)

		// Verify fields are correctly populated after validation
		require.NotNil(t, input.Steps[0].Ingredients[0].IngredientID)
		assert.Equal(t, ingredient.ID, *input.Steps[0].Ingredients[0].IngredientID)
		assert.Equal(t, measurementUnit.ID, input.Steps[0].Ingredients[0].MeasurementUnitID)

		require.NotNil(t, input.Steps[0].Instruments[0].InstrumentID)
		assert.Equal(t, instrument.ID, *input.Steps[0].Instruments[0].InstrumentID)

		require.NotNil(t, input.Steps[0].Vessels[0].VesselID)
		assert.Equal(t, vessel.ID, *input.Steps[0].Vessels[0].VesselID)
	})

	T.Run("with empty recipe (no steps)", func(t *testing.T) {
		t.Parallel()

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:    fakes.BuildFakeID(),
			Name:  "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{},
		}

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		assert.NoError(t, err)
	})

	T.Run("with step with no bridge IDs (all nil)", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID: fakes.BuildFakeID(),
							// No bridge IDs set - should pass validation
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID: fakes.BuildFakeID(),
							// No bridge IDs set - should pass validation
						},
					},
					Vessels: []*mealplanning.RecipeStepVesselDatabaseCreationInput{
						{
							ID: fakes.BuildFakeID(),
							// No bridge IDs set - should pass validation
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, nil, nil, nil)
		err := validator.ValidateAndPopulate(input)

		// Should succeed - no bridge IDs to validate
		assert.NoError(t, err)
	})

	T.Run("with VIMU only (no VIP) populates ingredient from VIMU", func(t *testing.T) {
		t.Parallel()

		preparation := fakes.BuildFakeValidPreparation()

		vimu := fakes.BuildFakeValidIngredientMeasurementUnit()

		vimuMap := map[string]*mealplanning.ValidIngredientMeasurementUnit{vimu.ID: vimu}

		input := &mealplanning.RecipeDatabaseCreationInput{
			ID:   fakes.BuildFakeID(),
			Name: "Test Recipe",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            fakes.BuildFakeID(),
					PreparationID: preparation.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID: fakes.BuildFakeID(),
							// Only VIMU, no VIP
							ValidIngredientMeasurementUnitID: new(vimu.ID),
						},
					},
				},
			},
		}

		validator := NewRecipeValidator(nil, vimuMap, nil, nil)
		err := validator.ValidateAndPopulate(input)

		require.NoError(t, err)

		// Should populate IngredientID from VIMU
		require.NotNil(t, input.Steps[0].Ingredients[0].IngredientID)
		assert.Equal(t, vimu.Ingredient.ID, *input.Steps[0].Ingredients[0].IngredientID)
		assert.Equal(t, vimu.MeasurementUnit.ID, input.Steps[0].Ingredients[0].MeasurementUnitID)
	})
}
