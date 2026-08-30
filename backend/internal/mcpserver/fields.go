package mcpserver

// The property names the hand-written tool schemas below repeat. They are the wire names a
// model reads and writes, so they are spelled once here rather than once per tool.
const (
	fieldArchivedAt                     = "ArchivedAt"
	fieldBelongsToAccount               = "BelongsToAccount"
	fieldBelongsToRecipe                = "BelongsToRecipe"
	fieldBelongsToRecipeStep            = "BelongsToRecipeStep"
	fieldCreatedAt                      = "CreatedAt"
	fieldCreatedByUser                  = "CreatedByUser"
	fieldDescription                    = "Description"
	fieldFilter                         = "Filter"
	fieldIconPath                       = "IconPath"
	fieldIndex                          = "Index"
	fieldIngredient                     = "Ingredient"
	fieldLastUpdatedAt                  = "LastUpdatedAt"
	fieldMaxQuantity                    = "MaxQuantity"
	fieldMaxStorageTemperatureInCelsius = "MaxStorageTemperatureInCelsius"
	fieldMeasurementUnit                = "MeasurementUnit"
	fieldMinQuantity                    = "MinQuantity"
	fieldMinStorageTemperatureInCelsius = "MinStorageTemperatureInCelsius"
	fieldName                           = "Name"
	fieldNotes                          = "Notes"
	fieldOptional                       = "Optional"
	fieldPluralName                     = "PluralName"
	fieldPreparation                    = "Preparation"
	fieldQuery                          = "Query"
	fieldRecipeID                       = "RecipeID"
	fieldRecipeStepID                   = "RecipeStepID"
	fieldRecipeStepProductID            = "RecipeStepProductID"
	fieldResults                        = "Results"
	fieldSlug                           = "Slug"
	fieldStorageInstructions            = "StorageInstructions"
	fieldValidIngredientID              = "ValidIngredientID"
	fieldValidPreparationID             = "ValidPreparationID"
	fieldWaitlistID                     = "WaitlistID"

	// The JSON Schema keywords the same schemas emit by hand.
	keyDescription = "description"
	keyType        = "type"
)
