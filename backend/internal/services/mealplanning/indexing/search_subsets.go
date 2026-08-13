package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	textsearch "github.com/primandproper/platform-go/v10/search/text"
)

// NamedID pairs an identifier with a human-readable name.
type NamedID struct {
	_ struct{} `json:"-"`

	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ValidMeasurementUnitTextSearcher textsearch.Index[ValidMeasurementUnitSearchSubset]

// ValidMeasurementUnitSearchSubset represents the subset of values suitable to index for search.
type ValidMeasurementUnitSearchSubset struct {
	_ struct{} `json:"-"`

	Name        string `json:"name,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	PluralName  string `json:"pluralName,omitempty"`
}

// ConvertValidMeasurementUnitToValidMeasurementUnitSearchSubset converts a ValidMeasurementUnit to a ValidMeasurementUnitSearchSubset.
func ConvertValidMeasurementUnitToValidMeasurementUnitSearchSubset(x *mealplanning.ValidMeasurementUnit) *ValidMeasurementUnitSearchSubset {
	return &ValidMeasurementUnitSearchSubset{
		ID:          x.ID,
		Name:        x.Name,
		PluralName:  x.PluralName,
		Description: x.Description,
	}
}

type MealTextSearcher textsearch.Index[MealSearchSubset]

// MealSearchSubset represents the subset of values suitable to index for search.
type MealSearchSubset struct {
	_ struct{} `json:"-"`

	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Recipes     []NamedID `json:"recipes,omitempty"`
}

func ConvertMealToMealSearchSubset(r *mealplanning.Meal) *MealSearchSubset {
	x := &MealSearchSubset{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}

	for _, component := range r.Components {
		x.Recipes = append(x.Recipes, NamedID{ID: component.Recipe.ID, Name: component.Recipe.Name})
	}

	return x
}

type RecipeTextSearcher textsearch.Index[RecipeSearchSubset]

// RecipeSearchSubset represents the subset of values suitable to index for search.
type RecipeSearchSubset struct {
	_ struct{} `json:"-"`

	ID          string                    `json:"id,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Description string                    `json:"description,omitempty"`
	Steps       []*RecipeStepSearchSubset `json:"steps,omitempty"`
}

// ConvertRecipeToRecipeSearchSubset converts a Recipe to a RecipeSearchSubset.
func ConvertRecipeToRecipeSearchSubset(r *mealplanning.Recipe) *RecipeSearchSubset {
	x := &RecipeSearchSubset{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}

	for _, step := range r.Steps {
		x.Steps = append(x.Steps, ConvertRecipeStepToRecipeStepSearchSubset(step))
	}

	return x
}

type RecipeStepTextSearcher textsearch.Index[RecipeStepSearchSubset]

// RecipeStepSearchSubset represents the subset of values suitable to index for search.
type RecipeStepSearchSubset struct {
	_ struct{} `json:"-"`

	Preparation string    `json:"preparation,omitempty"`
	Ingredients []NamedID `json:"ingredients,omitempty"`
	Instruments []NamedID `json:"instruments,omitempty"`
	Vessels     []NamedID `json:"vessels,omitempty"`
}

func ConvertRecipeStepToRecipeStepSearchSubset(x *mealplanning.RecipeStep) *RecipeStepSearchSubset {
	stepSubset := &RecipeStepSearchSubset{
		Preparation: x.Preparation.Name,
	}

	for _, ingredient := range x.Ingredients {
		stepSubset.Ingredients = append(stepSubset.Ingredients, NamedID{ID: ingredient.ID, Name: ingredient.Name})
	}

	for _, instrument := range x.Instruments {
		stepSubset.Instruments = append(stepSubset.Instruments, NamedID{ID: instrument.ID, Name: instrument.Name})
	}

	for _, vessel := range x.Vessels {
		stepSubset.Vessels = append(stepSubset.Vessels, NamedID{ID: vessel.ID, Name: vessel.Name})
	}

	return stepSubset
}

type ValidIngredientTextSearcher textsearch.Index[ValidIngredientSearchSubset]

// ValidIngredientSearchSubset represents the subset of values suitable to index for search.
type ValidIngredientSearchSubset struct {
	_ struct{} `json:"-"`

	PluralName          string `json:"pluralName,omitempty"`
	Name                string `json:"name,omitempty"`
	ID                  string `json:"id,omitempty"`
	Description         string `json:"description,omitempty"`
	ShoppingSuggestions string `json:"shoppingSuggestions,omitempty"`
}

// ConvertValidIngredientToValidIngredientSearchSubset converts a ValidIngredient to a ValidIngredientSearchSubset.
func ConvertValidIngredientToValidIngredientSearchSubset(x *mealplanning.ValidIngredient) *ValidIngredientSearchSubset {
	return &ValidIngredientSearchSubset{
		ID:                  x.ID,
		Name:                x.Name,
		PluralName:          x.PluralName,
		Description:         x.Description,
		ShoppingSuggestions: x.ShoppingSuggestions,
	}
}

type ValidIngredientStateTextSearcher textsearch.Index[ValidIngredientStateSearchSubset]

// ValidIngredientStateSearchSubset represents the subset of values suitable to index for search.
type ValidIngredientStateSearchSubset struct {
	_ struct{} `json:"-"`

	ID            string `json:"id,omitempty"`
	PastTense     string `json:"pastTense,omitempty"`
	Description   string `json:"description,omitempty"`
	Name          string `json:"name,omitempty"`
	AttributeType string `json:"attributeType,omitempty"`
}

// ConvertValidIngredientStateToValidIngredientStateSearchSubset converts a ValidIngredientState to a ValidIngredientStateSearchSubset.
func ConvertValidIngredientStateToValidIngredientStateSearchSubset(x *mealplanning.ValidIngredientState) *ValidIngredientStateSearchSubset {
	return &ValidIngredientStateSearchSubset{
		ID:            x.ID,
		Name:          x.Name,
		PastTense:     x.PastTense,
		Description:   x.Description,
		AttributeType: x.AttributeType,
	}
}

type ValidInstrumentTextSearcher textsearch.Index[ValidInstrumentSearchSubset]

// ValidInstrumentSearchSubset represents the subset of values suitable to index for search.
type ValidInstrumentSearchSubset struct {
	_ struct{} `json:"-"`

	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	PluralName  string `json:"pluralName,omitempty"`
	Description string `json:"description,omitempty"`
}

// ConvertValidInstrumentToValidInstrumentSearchSubset converts a ValidInstrument to a ValidInstrumentSearchSubset.
func ConvertValidInstrumentToValidInstrumentSearchSubset(x *mealplanning.ValidInstrument) *ValidInstrumentSearchSubset {
	return &ValidInstrumentSearchSubset{
		ID:          x.ID,
		Name:        x.Name,
		PluralName:  x.PluralName,
		Description: x.Description,
	}
}

type ValidPreparationTextSearcher textsearch.Index[ValidPreparationSearchSubset]

// ValidPreparationSearchSubset represents the subset of values suitable to index for search.
type ValidPreparationSearchSubset struct {
	_ struct{} `json:"-"`

	PastTense   string `json:"pastTense,omitempty"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ConvertValidPreparationToValidPreparationSearchSubset converts a ValidPreparation to a ValidPreparationSearchSubset.
func ConvertValidPreparationToValidPreparationSearchSubset(x *mealplanning.ValidPreparation) *ValidPreparationSearchSubset {
	return &ValidPreparationSearchSubset{
		ID:          x.ID,
		Name:        x.Name,
		PastTense:   x.PastTense,
		Description: x.Description,
	}
}

type ValidVesselTextSearcher textsearch.Index[ValidVesselSearchSubset]

// ValidVesselSearchSubset represents the subset of values suitable to index for search.
type ValidVesselSearchSubset struct {
	_ struct{} `json:"-"`

	ID               string  `json:"id,omitempty"`
	Name             string  `json:"name,omitempty"`
	PluralName       string  `json:"pluralName,omitempty"`
	Description      string  `json:"description,omitempty"`
	CapacityUnitName string  `json:"capacityUnitName"`
	Capacity         float32 `json:"capacity,omitempty"`
}

// ConvertValidVesselToValidVesselSearchSubset converts a ValidVessel to a ValidVesselSearchSubset.
func ConvertValidVesselToValidVesselSearchSubset(x *mealplanning.ValidVessel) *ValidVesselSearchSubset {
	return &ValidVesselSearchSubset{
		ID:          x.ID,
		Name:        x.Name,
		PluralName:  x.PluralName,
		Description: x.Description,
	}
}
