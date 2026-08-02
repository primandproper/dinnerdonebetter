package recipevalidator

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
)

// RecipeValidator validates and populates recipe creation inputs using bridge table data.
type RecipeValidator struct {
	validIngredientPreparations     map[string]*mealplanning.ValidIngredientPreparation
	validIngredientMeasurementUnits map[string]*mealplanning.ValidIngredientMeasurementUnit
	validPreparationInstruments     map[string]*mealplanning.ValidPreparationInstrument
	validPreparationVessels         map[string]*mealplanning.ValidPreparationVessel
}

// NewRecipeValidator creates a new RecipeValidator with the provided bridge table maps.
func NewRecipeValidator(
	validIngredientPreparations map[string]*mealplanning.ValidIngredientPreparation,
	validIngredientMeasurementUnits map[string]*mealplanning.ValidIngredientMeasurementUnit,
	validPreparationInstruments map[string]*mealplanning.ValidPreparationInstrument,
	validPreparationVessels map[string]*mealplanning.ValidPreparationVessel,
) *RecipeValidator {
	return &RecipeValidator{
		validIngredientPreparations:     validIngredientPreparations,
		validIngredientMeasurementUnits: validIngredientMeasurementUnits,
		validPreparationInstruments:     validPreparationInstruments,
		validPreparationVessels:         validPreparationVessels,
	}
}

// ValidateAndPopulate validates a recipe creation input and populates derived fields.
// It returns an error if any bridge table ID is invalid or doesn't match the step's preparation.
func (v *RecipeValidator) ValidateAndPopulate(input *mealplanning.RecipeDatabaseCreationInput) error {
	if input == nil {
		return fmt.Errorf("input is nil")
	}

	for stepIdx, step := range input.Steps {
		if err := v.validateStep(stepIdx, step); err != nil {
			return err
		}
	}

	return nil
}

// validateStep validates all ingredients, instruments, and vessels in a step.
func (v *RecipeValidator) validateStep(stepIdx int, step *mealplanning.RecipeStepDatabaseCreationInput) error {
	if step == nil {
		return nil
	}

	// Validate and populate ingredients
	for ingredientIdx, ingredient := range step.Ingredients {
		if err := v.validateAndPopulateIngredient(stepIdx, ingredientIdx, step.PreparationID, ingredient); err != nil {
			return err
		}
	}

	// Validate and populate instruments
	for instrumentIdx, instrument := range step.Instruments {
		if err := v.validateAndPopulateInstrument(stepIdx, instrumentIdx, step.PreparationID, instrument); err != nil {
			return err
		}
	}

	// Validate and populate vessels
	for vesselIdx, vessel := range step.Vessels {
		if err := v.validateAndPopulateVessel(stepIdx, vesselIdx, step.PreparationID, vessel); err != nil {
			return err
		}
	}

	// Validate option grouping for ingredients, instruments, and vessels
	if err := v.validateOptionGrouping(stepIdx, step); err != nil {
		return err
	}

	return nil
}

// isRecipeStepProduct returns true if the ingredient is a recipe step product (output from a previous step).
func isRecipeStepProductIngredient(ingredient *mealplanning.RecipeStepIngredientDatabaseCreationInput) bool {
	if ingredient == nil {
		return false
	}
	// A recipe step product is indicated by having a RecipeStepProductID, ProductOfRecipeStepIndex, or RecipeStepProductRecipeID set
	return (ingredient.RecipeStepProductID != nil && *ingredient.RecipeStepProductID != "") ||
		ingredient.ProductOfRecipeStepIndex != nil ||
		(ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != "")
}

// isRecipeStepProductInstrument returns true if the instrument is a recipe step product (output from a previous step).
func isRecipeStepProductInstrument(instrument *mealplanning.RecipeStepInstrumentDatabaseCreationInput) bool {
	if instrument == nil {
		return false
	}
	return (instrument.RecipeStepProductID != nil && *instrument.RecipeStepProductID != "") ||
		instrument.ProductOfRecipeStepIndex != nil
}

// isRecipeStepProductVessel returns true if the vessel is a recipe step product (output from a previous step).
func isRecipeStepProductVessel(vessel *mealplanning.RecipeStepVesselDatabaseCreationInput) bool {
	if vessel == nil {
		return false
	}
	return (vessel.RecipeStepProductID != nil && *vessel.RecipeStepProductID != "") ||
		vessel.ProductOfRecipeStepIndex != nil
}

// validateAndPopulateIngredient validates the bridge table IDs for an ingredient and populates derived fields.
func (v *RecipeValidator) validateAndPopulateIngredient(
	stepIdx,
	ingredientIdx int,
	preparationID string,
	ingredient *mealplanning.RecipeStepIngredientDatabaseCreationInput,
) error {
	if ingredient == nil {
		return nil
	}

	// Check if this is a recipe step product (from same recipe or another recipe)
	isRecipeStepProduct := isRecipeStepProductIngredient(ingredient)
	isFromAnotherRecipe := ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != ""

	// Validate ValidIngredientPreparationID if provided (do this before early return
	// so IngredientID is set even for recipe step products, which is needed for
	// VesselType products that can't set RecipeStepProductID)
	if ingredient.ValidIngredientPreparationID != nil && *ingredient.ValidIngredientPreparationID != "" {
		vipID := *ingredient.ValidIngredientPreparationID
		vip, ok := v.validIngredientPreparations[vipID]
		if !ok {
			return fmt.Errorf("step %d ingredient %d: ValidIngredientPreparation %q not found", stepIdx, ingredientIdx, vipID)
		}

		// Validate that the VIP's preparation matches the step's preparation
		if vip.Preparation.ID != preparationID {
			return fmt.Errorf("step %d ingredient %d: ValidIngredientPreparation %q is for preparation %q, but step uses preparation %q",
				stepIdx, ingredientIdx, vipID, vip.Preparation.ID, preparationID)
		}
		// Ensure Ingredient.ID is not empty
		if vip.Ingredient.ID == "" {
			return fmt.Errorf("step %d ingredient %d: ValidIngredientPreparation %q has empty Ingredient.ID", stepIdx, ingredientIdx, vipID)
		}
		ingredient.IngredientID = &vip.Ingredient.ID
	}

	// Validate ValidIngredientMeasurementUnitID if provided (do this before early return
	// so IngredientID and MeasurementUnitID are set even for recipe step products)
	if ingredient.ValidIngredientMeasurementUnitID != nil && *ingredient.ValidIngredientMeasurementUnitID != "" {
		vimuID := *ingredient.ValidIngredientMeasurementUnitID
		vimu, ok := v.validIngredientMeasurementUnits[vimuID]
		if !ok {
			return fmt.Errorf("step %d ingredient %d: ValidIngredientMeasurementUnit %q not found", stepIdx, ingredientIdx, vimuID)
		}

		// If we have a VIP, validate that the VIMU's ingredient matches
		if ingredient.IngredientID != nil && *ingredient.IngredientID != "" {
			if vimu.Ingredient.ID != *ingredient.IngredientID {
				return fmt.Errorf("step %d ingredient %d: ValidIngredientMeasurementUnit %q is for ingredient %q, but ingredient %q was specified",
					stepIdx, ingredientIdx, vimuID, vimu.Ingredient.ID, *ingredient.IngredientID)
			}
		} else {
			// If no VIP was provided but VIMU was, populate IngredientID from VIMU
			ingredient.IngredientID = &vimu.Ingredient.ID
		}

		ingredient.MeasurementUnitID = vimu.MeasurementUnit.ID
	}

	// Final check: if we have bridge table IDs but IngredientID is still not set, this is an error
	// This ensures the database constraint (ingredient_id OR recipe_step_product_id must be set) is satisfied
	// This check must happen before the early return for recipe step products
	if (ingredient.ValidIngredientPreparationID != nil && *ingredient.ValidIngredientPreparationID != "") ||
		(ingredient.ValidIngredientMeasurementUnitID != nil && *ingredient.ValidIngredientMeasurementUnitID != "") {
		if (ingredient.IngredientID == nil || *ingredient.IngredientID == "") &&
			(ingredient.RecipeStepProductID == nil || *ingredient.RecipeStepProductID == "") {
			return fmt.Errorf("step %d ingredient %d: bridge table IDs provided but IngredientID was not populated and RecipeStepProductID is not set", stepIdx, ingredientIdx)
		}
	}

	// If it's a recipe step product from the same recipe, skip remaining validation
	if isRecipeStepProduct && !isFromAnotherRecipe {
		return nil
	}

	// For ingredients from another recipe, skip all bridge table validation
	// as the product was already validated in its own recipe
	if isFromAnotherRecipe {
		return nil
	}

	return nil
}

// validateAndPopulateInstrument validates the bridge table ID for an instrument and populates derived fields.
func (v *RecipeValidator) validateAndPopulateInstrument(
	stepIdx, instrumentIdx int,
	preparationID string,
	instrument *mealplanning.RecipeStepInstrumentDatabaseCreationInput,
) error {
	if instrument == nil || isRecipeStepProductInstrument(instrument) {
		return nil
	}

	// Validate ValidPreparationInstrumentID if provided
	if instrument.ValidPreparationInstrumentID != nil && *instrument.ValidPreparationInstrumentID != "" {
		vpiID := *instrument.ValidPreparationInstrumentID
		vpi, ok := v.validPreparationInstruments[vpiID]
		if !ok {
			return fmt.Errorf("step %d instrument %d: ValidPreparationInstrument %q not found", stepIdx, instrumentIdx, vpiID)
		}

		// Validate that the VPI's preparation matches the step's preparation
		if vpi.Preparation.ID != preparationID {
			return fmt.Errorf("step %d instrument %d: ValidPreparationInstrument %q is for preparation %q, but step uses preparation %q",
				stepIdx, instrumentIdx, vpiID, vpi.Preparation.ID, preparationID)
		}

		// Populate InstrumentID from the VPI
		instrument.InstrumentID = &vpi.Instrument.ID
	}

	return nil
}

// validateAndPopulateVessel validates the bridge table ID for a vessel and populates derived fields.
func (v *RecipeValidator) validateAndPopulateVessel(
	stepIdx, vesselIdx int,
	preparationID string,
	vessel *mealplanning.RecipeStepVesselDatabaseCreationInput,
) error {
	if vessel == nil || isRecipeStepProductVessel(vessel) {
		return nil
	}

	// Validate ValidPreparationVesselID if provided
	if vessel.ValidPreparationVesselID != nil && *vessel.ValidPreparationVesselID != "" {
		vpvID := *vessel.ValidPreparationVesselID
		vpv, ok := v.validPreparationVessels[vpvID]
		if !ok {
			return fmt.Errorf("step %d vessel %d: ValidPreparationVessel %q not found", stepIdx, vesselIdx, vpvID)
		}

		// Validate that the VPV's preparation matches the step's preparation
		if vpv.Preparation.ID != preparationID {
			return fmt.Errorf("step %d vessel %d: ValidPreparationVessel %q is for preparation %q, but step uses preparation %q",
				stepIdx, vesselIdx, vpvID, vpv.Preparation.ID, preparationID)
		}

		// Populate VesselID from the VPV
		vessel.VesselID = &vpv.Vessel.ID
	}

	return nil
}

// validateOptionGrouping validates that options within the same index are properly grouped.
// It ensures that OptionIndex values are sequential within each Index group (0, 1, 2... not 0, 5, 10).
func (v *RecipeValidator) validateOptionGrouping(stepIdx int, step *mealplanning.RecipeStepDatabaseCreationInput) error {
	// Validate ingredient option grouping
	if err := v.validateIngredientOptionGrouping(stepIdx, step.Ingredients); err != nil {
		return err
	}

	// Validate instrument option grouping
	if err := v.validateInstrumentOptionGrouping(stepIdx, step.Instruments); err != nil {
		return err
	}

	// Validate vessel option grouping
	if err := v.validateVesselOptionGrouping(stepIdx, step.Vessels); err != nil {
		return err
	}

	return nil
}

// validateIngredientOptionGrouping validates that ingredient options within the same index are properly grouped.
func (v *RecipeValidator) validateIngredientOptionGrouping(stepIdx int, ingredients []*mealplanning.RecipeStepIngredientDatabaseCreationInput) error {
	// Check for duplicate (index, option_index) combinations
	indexOptionMap := make(map[uint16]map[uint16]bool)
	for _, ingredient := range ingredients {
		if indexOptionMap[ingredient.Index] == nil {
			indexOptionMap[ingredient.Index] = make(map[uint16]bool)
		}
		if indexOptionMap[ingredient.Index][ingredient.OptionIndex] {
			return fmt.Errorf("step %d ingredients: duplicate (index, option_index) combination: index %d, option_index %d", stepIdx, ingredient.Index, ingredient.OptionIndex)
		}
		indexOptionMap[ingredient.Index][ingredient.OptionIndex] = true
	}

	// Group ingredients by index
	indexGroups := make(map[uint16][]*mealplanning.RecipeStepIngredientDatabaseCreationInput)
	for _, ingredient := range ingredients {
		indexGroups[ingredient.Index] = append(indexGroups[ingredient.Index], ingredient)
	}

	// Validate each index group
	for index, group := range indexGroups {
		if len(group) <= 1 {
			continue // No grouping needed for single items
		}

		// Collect option indices for this index group
		optionIndices := make(map[uint16]bool)
		for _, ingredient := range group {
			optionIndices[ingredient.OptionIndex] = true
		}

		// Check that option indices are sequential starting from 0
		expectedCount := len(optionIndices)
		for i := uint16(0); i < uint16(expectedCount); i++ {
			if !optionIndices[i] {
				return fmt.Errorf("step %d ingredients: option indices for index %d must be sequential starting from 0, but found gap at option index %d", stepIdx, index, i)
			}
		}
	}

	return nil
}

// validateInstrumentOptionGrouping validates that instrument options within the same index are properly grouped.
func (v *RecipeValidator) validateInstrumentOptionGrouping(stepIdx int, instruments []*mealplanning.RecipeStepInstrumentDatabaseCreationInput) error {
	// Check for duplicate (index, option_index) combinations
	indexOptionMap := make(map[uint16]map[uint16]bool)
	for _, instrument := range instruments {
		if indexOptionMap[instrument.Index] == nil {
			indexOptionMap[instrument.Index] = make(map[uint16]bool)
		}
		if indexOptionMap[instrument.Index][instrument.OptionIndex] {
			return fmt.Errorf("step %d instruments: duplicate (index, option_index) combination: index %d, option_index %d", stepIdx, instrument.Index, instrument.OptionIndex)
		}
		indexOptionMap[instrument.Index][instrument.OptionIndex] = true
	}

	// Group instruments by index
	indexGroups := make(map[uint16][]*mealplanning.RecipeStepInstrumentDatabaseCreationInput)
	for _, instrument := range instruments {
		indexGroups[instrument.Index] = append(indexGroups[instrument.Index], instrument)
	}

	// Validate each index group
	for index, group := range indexGroups {
		if len(group) <= 1 {
			continue // No grouping needed for single items
		}

		// Collect option indices for this index group
		optionIndices := make(map[uint16]bool)
		for _, instrument := range group {
			optionIndices[instrument.OptionIndex] = true
		}

		// Check that option indices are sequential starting from 0
		expectedCount := len(optionIndices)
		for i := uint16(0); i < uint16(expectedCount); i++ {
			if !optionIndices[i] {
				return fmt.Errorf("step %d instruments: option indices for index %d must be sequential starting from 0, but found gap at option index %d", stepIdx, index, i)
			}
		}
	}

	return nil
}

// validateVesselOptionGrouping validates that vessel options within the same index are properly grouped.
func (v *RecipeValidator) validateVesselOptionGrouping(stepIdx int, vessels []*mealplanning.RecipeStepVesselDatabaseCreationInput) error {
	// Check for duplicate (index, option_index) combinations
	indexOptionMap := make(map[uint16]map[uint16]bool)
	for _, vessel := range vessels {
		if indexOptionMap[vessel.Index] == nil {
			indexOptionMap[vessel.Index] = make(map[uint16]bool)
		}
		if indexOptionMap[vessel.Index][vessel.OptionIndex] {
			return fmt.Errorf("step %d vessels: duplicate (index, option_index) combination: index %d, option_index %d", stepIdx, vessel.Index, vessel.OptionIndex)
		}
		indexOptionMap[vessel.Index][vessel.OptionIndex] = true
	}

	// Group vessels by index
	indexGroups := make(map[uint16][]*mealplanning.RecipeStepVesselDatabaseCreationInput)
	for _, vessel := range vessels {
		indexGroups[vessel.Index] = append(indexGroups[vessel.Index], vessel)
	}

	// Validate each index group
	for index, group := range indexGroups {
		if len(group) <= 1 {
			continue // No grouping needed for single items
		}

		// Collect option indices for this index group
		optionIndices := make(map[uint16]bool)
		for _, vessel := range group {
			optionIndices[vessel.OptionIndex] = true
		}

		// Check that option indices are sequential starting from 0
		expectedCount := len(optionIndices)
		for i := uint16(0); i < uint16(expectedCount); i++ {
			if !optionIndices[i] {
				return fmt.Errorf("step %d vessels: option indices for index %d must be sequential starting from 0, but found gap at option index %d", stepIdx, index, i)
			}
		}
	}

	return nil
}
