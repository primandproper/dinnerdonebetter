package mealplanning

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/primandproper/platform-go/v10/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// RecipeCreatedServiceEventType indicates a recipe was created.
	RecipeCreatedServiceEventType = "recipe_created"
	// RecipeUpdatedServiceEventType indicates a recipe was updated.
	RecipeUpdatedServiceEventType = "recipe_updated"
	// RecipeArchivedServiceEventType indicates a recipe was archived.
	RecipeArchivedServiceEventType = "recipe_archived"
	// RecipeClonedServiceEventType indicates a recipe was cloned.
	RecipeClonedServiceEventType = "recipe_cloned"

	RecipeStatusSubmitted     = "submitted"
	RecipeStatusApproved      = "approved"
	RecipeStatusNeedsRevision = "needs_revision"
)

func init() {
	gob.Register(new(Recipe))
	gob.Register(new(RecipeCreationRequestInput))
	gob.Register(new(RecipeUpdateRequestInput))
}

type (
	// Recipe represents a recipe.
	Recipe struct {
		_                    struct{}          `json:"-"`
		CreatedAt            time.Time         `json:"createdAt"`
		MaxEstimatedPortions *float32          `json:"maxEstimatedPortions,omitempty"`
		InspiredByRecipeID   *string           `json:"inspiredByRecipeID"`
		LastUpdatedAt        *time.Time        `json:"lastUpdatedAt"`
		ArchivedAt           *time.Time        `json:"archivedAt"`
		ID                   string            `json:"id"`
		Slug                 string            `json:"slug"`
		Name                 string            `json:"name"`
		PortionName          string            `json:"portionName"`
		Source               string            `json:"source"`
		SourceISBN           string            `json:"sourceISBN"`
		CreatedByUser        string            `json:"createdByUser"`
		PluralPortionName    string            `json:"pluralPortionName"`
		Description          string            `json:"description"`
		YieldsComponentType  string            `json:"yieldsComponentType"`
		Status               string            `json:"status"`
		Steps                []*RecipeStep     `json:"steps"`
		Media                []*RecipeMedia    `json:"media"`
		PrepTasks            []*RecipePrepTask `json:"prepTasks"`
		AssociatedRecipes    []*Recipe         `json:"associatedRecipes"`
		MinEstimatedPortions float32           `json:"minEstimatedPortions"`
		SealOfApproval       bool              `json:"sealOfApproval"`
		EligibleForMeals     bool              `json:"eligibleForMeals"`
	}

	// RecipeCreationRequestInput represents what a user could set as input for creating recipes.
	RecipeCreationRequestInput struct {
		_ struct{} `json:"-"`

		InspiredByRecipeID   *string                                           `json:"inspiredByRecipeID"`
		MaxEstimatedPortions *float32                                          `json:"maxEstimatedPortions,omitempty"`
		Name                 string                                            `json:"name"`
		Source               string                                            `json:"source"`
		SourceISBN           string                                            `json:"sourceISBN"`
		Description          string                                            `json:"description"`
		PluralPortionName    string                                            `json:"pluralPortionName"`
		PortionName          string                                            `json:"portionName"`
		Slug                 string                                            `json:"slug"`
		YieldsComponentType  string                                            `json:"yieldsComponentType"`
		PrepTasks            []*RecipePrepTaskWithinRecipeCreationRequestInput `json:"prepTasks"`
		Steps                []*RecipeStepCreationRequestInput                 `json:"steps"`
		MinEstimatedPortions float32                                           `json:"minEstimatedPortions"`
		AlsoCreateMeal       bool                                              `json:"alsoCreateMeal"`
		EligibleForMeals     bool                                              `json:"eligibleForMeals"`
	}

	// RecipeDatabaseCreationInput represents what a user could set as input for creating recipes.
	RecipeDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		// ClonedFromRecipeID names the recipe this one was copied from, when it was. It
		// exists so CreateRecipe can emit the cloned event inside the transaction that
		// writes the clone — a clone is one write, and the fact that it is a clone should
		// not be announced separately from it.
		ClonedFromRecipeID   *string                                `json:"-"`
		InspiredByRecipeID   *string                                `json:"-"`
		MaxEstimatedPortions *float32                               `json:"-"`
		CreatedByUser        string                                 `json:"-"`
		ID                   string                                 `json:"-"`
		Name                 string                                 `json:"-"`
		Slug                 string                                 `json:"-"`
		Source               string                                 `json:"-"`
		SourceISBN           string                                 `json:"-"`
		PluralPortionName    string                                 `json:"-"`
		PortionName          string                                 `json:"-"`
		Description          string                                 `json:"-"`
		YieldsComponentType  string                                 `json:"-"`
		PrepTasks            []*RecipePrepTaskDatabaseCreationInput `json:"-"`
		Steps                []*RecipeStepDatabaseCreationInput     `json:"-"`
		MinEstimatedPortions float32                                `json:"-"`
		AlsoCreateMeal       bool                                   `json:"-"`
		EligibleForMeals     bool                                   `json:"-"`
	}

	// RecipeUpdateRequestInput represents what a user could set as input for updating recipes.
	RecipeUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name                 *string  `json:"name,omitempty"`
		Slug                 *string  `json:"slug"`
		Source               *string  `json:"source,omitempty"`
		SourceISBN           *string  `json:"sourceISBN,omitempty"`
		Description          *string  `json:"description,omitempty"`
		InspiredByRecipeID   *string  `json:"inspiredByRecipeID,omitempty"`
		MinEstimatedPortions *float32 `json:"minEstimatedPortions,omitempty"`
		MaxEstimatedPortions *float32 `json:"maxEstimatedPortions,omitempty"`
		PortionName          *string  `json:"portionName"`
		PluralPortionName    *string  `json:"pluralPortionName"`
		EligibleForMeals     *bool    `json:"eligibleForMeals"`
		YieldsComponentType  *string  `json:"yieldsComponentType"`
	}

	// RecipeDataManager describes a structure capable of storing recipes permanently.
	RecipeDataManager interface {
		RecipeExists(ctx context.Context, recipeID string) (bool, error)
		GetRecipe(ctx context.Context, recipeID string) (*Recipe, error)
		GetRecipes(ctx context.Context, status string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Recipe], error)
		GetRecipesCreatedByUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Recipe], error)
		SearchForRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Recipe], error)
		SearchForMealEligibleRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Recipe], error)
		SearchForRecipesWithInstrumentOwnership(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Recipe], error)
		CreateRecipe(ctx context.Context, input *RecipeDatabaseCreationInput) (*Recipe, error)
		UpdateRecipe(ctx context.Context, updated *Recipe) error
		UpdateRecipeStatus(ctx context.Context, recipeID, newStatus string) error
		MarkRecipeAsIndexed(ctx context.Context, recipeID string) error
		ArchiveRecipe(ctx context.Context, recipeID, userID string) error
		ScanRecipeIDsForReindex(ctx context.Context, after string, limit int) ([]string, error)
		GetRecipesWithIDs(ctx context.Context, ids []string) ([]*Recipe, error)
		AddRecipeImage(ctx context.Context, recipeID, uploadedMediaID, uploadedByUser string) error
	}
)

// FindStepForIndex finds a step for a given index.
func (x *Recipe) FindStepForIndex(index uint32) *RecipeStep {
	for _, step := range x.Steps {
		if step.Index == index {
			return step
		}
	}

	// we could return an error here, but that would make my life a little harder
	// so if you fuck up and submit a wrong value, and it's your fault here.
	return nil
}

// FindStepByID finds a step for a given ID.
func (x *Recipe) FindStepByID(id string) *RecipeStep {
	for _, step := range x.Steps {
		if step.ID == id {
			return step
		}
	}

	// we could return an error here, but that would make my life a little harder
	// so if you fuck up and submit a wrong value, and it's your fault here.
	return nil
}

// FindStepForRecipeStepProductID finds a step for a given recipe step product ID.
// It searches the main recipe's steps first, then associated recipes' steps.
func (x *Recipe) FindStepForRecipeStepProductID(recipeStepProductID string) *RecipeStep {
	if step := findStepForRecipeStepProductIDInSteps(x.Steps, recipeStepProductID); step != nil {
		return step
	}
	for _, assoc := range x.AssociatedRecipes {
		if step := findStepForRecipeStepProductIDInSteps(assoc.Steps, recipeStepProductID); step != nil {
			return step
		}
	}
	return nil
}

// findStepForRecipeStepProductIDInSteps searches steps for a product ID.
func findStepForRecipeStepProductIDInSteps(steps []*RecipeStep, recipeStepProductID string) *RecipeStep {
	for _, step := range steps {
		for _, product := range step.Products {
			if product.ID == recipeStepProductID {
				return step
			}
		}
	}
	return nil
}

// FindStepForRecipeStepProductIDWithRecipe finds a step for a given product ID and returns
// the step and the recipe it belongs to (main or associated). If sourceRecipeID is non-nil
// and non-empty, that recipe is searched first.
func (x *Recipe) FindStepForRecipeStepProductIDWithRecipe(recipeStepProductID string, sourceRecipeID *string) (*RecipeStep, *Recipe) {
	if sourceRecipeID != nil && *sourceRecipeID != "" && *sourceRecipeID != x.ID {
		for _, assoc := range x.AssociatedRecipes {
			if assoc.ID == *sourceRecipeID {
				if step := findStepForRecipeStepProductIDInSteps(assoc.Steps, recipeStepProductID); step != nil {
					return step, assoc
				}
				return nil, nil
			}
		}
	}
	if step := findStepForRecipeStepProductIDInSteps(x.Steps, recipeStepProductID); step != nil {
		return step, x
	}
	for _, assoc := range x.AssociatedRecipes {
		if step := findStepForRecipeStepProductIDInSteps(assoc.Steps, recipeStepProductID); step != nil {
			return step, assoc
		}
	}
	return nil, nil
}

// GetRelatedRecipeIDs returns all recipe IDs that this recipe references as components.
// It iterates through all steps and their ingredients to find any that reference other recipes
// via RecipeStepProductRecipeID.
func (x *Recipe) GetRelatedRecipeIDs() []string {
	seen := make(map[string]struct{})
	var relatedIDs []string

	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != "" {
				recipeID := *ingredient.RecipeStepProductRecipeID
				if _, exists := seen[recipeID]; !exists {
					seen[recipeID] = struct{}{}
					relatedIDs = append(relatedIDs, recipeID)
				}
			}
		}
	}

	return relatedIDs
}

// GetUnconsumedProducts returns all recipe step products that are not consumed
// (referenced as ingredients, instruments, or vessels) in any later step.
// It considers both main recipe steps and associated recipe steps.
func (x *Recipe) GetUnconsumedProducts() []*RecipeStepProduct {
	// Build a set of all product IDs that are consumed
	consumedProductIDs := make(map[string]struct{})

	// Main recipe steps: mark consumed when a later main step uses the product
	productToMainStepIndex := make(map[string]int)
	for stepIdx, step := range x.Steps {
		for _, product := range step.Products {
			productToMainStepIndex[product.ID] = stepIdx
		}
	}
	for stepIdx, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.RecipeStepProductID != nil && *ingredient.RecipeStepProductID != "" {
				productStepIdx, exists := productToMainStepIndex[*ingredient.RecipeStepProductID]
				if exists && productStepIdx < stepIdx {
					consumedProductIDs[*ingredient.RecipeStepProductID] = struct{}{}
				}
			}
		}
		for _, instrument := range step.Instruments {
			if instrument.RecipeStepProductID != nil && *instrument.RecipeStepProductID != "" {
				productStepIdx, exists := productToMainStepIndex[*instrument.RecipeStepProductID]
				if exists && productStepIdx < stepIdx {
					consumedProductIDs[*instrument.RecipeStepProductID] = struct{}{}
				}
			}
		}
		for _, vessel := range step.Vessels {
			if vessel.RecipeStepProductID != nil && *vessel.RecipeStepProductID != "" {
				productStepIdx, exists := productToMainStepIndex[*vessel.RecipeStepProductID]
				if exists && productStepIdx < stepIdx {
					consumedProductIDs[*vessel.RecipeStepProductID] = struct{}{}
				}
			}
		}
	}

	// Associated recipe steps: mark consumed when a later step in same recipe uses it, or when main uses it
	for _, assoc := range x.AssociatedRecipes {
		productToAssocStepIndex := make(map[string]int)
		for stepIdx, step := range assoc.Steps {
			for _, product := range step.Products {
				productToAssocStepIndex[product.ID] = stepIdx
			}
		}
		for stepIdx, step := range assoc.Steps {
			for _, ingredient := range step.Ingredients {
				if ingredient.RecipeStepProductID != nil && *ingredient.RecipeStepProductID != "" {
					productStepIdx, exists := productToAssocStepIndex[*ingredient.RecipeStepProductID]
					if exists && productStepIdx < stepIdx {
						consumedProductIDs[*ingredient.RecipeStepProductID] = struct{}{}
					}
				}
			}
			for _, instrument := range step.Instruments {
				if instrument.RecipeStepProductID != nil && *instrument.RecipeStepProductID != "" {
					productStepIdx, exists := productToAssocStepIndex[*instrument.RecipeStepProductID]
					if exists && productStepIdx < stepIdx {
						consumedProductIDs[*instrument.RecipeStepProductID] = struct{}{}
					}
				}
			}
			for _, vessel := range step.Vessels {
				if vessel.RecipeStepProductID != nil && *vessel.RecipeStepProductID != "" {
					productStepIdx, exists := productToAssocStepIndex[*vessel.RecipeStepProductID]
					if exists && productStepIdx < stepIdx {
						consumedProductIDs[*vessel.RecipeStepProductID] = struct{}{}
					}
				}
			}
		}
		// Products from this associated recipe consumed by main recipe
		for _, mainStep := range x.Steps {
			for _, ingredient := range mainStep.Ingredients {
				if ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID == assoc.ID &&
					ingredient.RecipeStepProductID != nil && *ingredient.RecipeStepProductID != "" {
					consumedProductIDs[*ingredient.RecipeStepProductID] = struct{}{}
				}
			}
		}
	}

	// Collect all products that are not consumed
	var unconsumedProducts []*RecipeStepProduct
	for _, step := range x.Steps {
		for _, product := range step.Products {
			if _, isConsumed := consumedProductIDs[product.ID]; !isConsumed {
				unconsumedProducts = append(unconsumedProducts, product)
			}
		}
	}
	for _, assoc := range x.AssociatedRecipes {
		for _, step := range assoc.Steps {
			for _, product := range step.Products {
				if _, isConsumed := consumedProductIDs[product.ID]; !isConsumed {
					unconsumedProducts = append(unconsumedProducts, product)
				}
			}
		}
	}

	return unconsumedProducts
}

// FindStepIndexByID finds a step for a given ID.
func (x *Recipe) FindStepIndexByID(id string) int {
	for i, step := range x.Steps {
		if step.ID == id {
			return i
		}
	}

	return -1
}

// Update merges a RecipeUpdateRequestInput with a recipe.
func (x *Recipe) Update(input *RecipeUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Slug != nil && *input.Slug != x.Slug {
		x.Slug = *input.Slug
	}

	if input.Source != nil && *input.Source != x.Source {
		x.Source = *input.Source
	}

	if input.SourceISBN != nil && *input.SourceISBN != x.SourceISBN {
		x.SourceISBN = *input.SourceISBN
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}

	if input.InspiredByRecipeID != nil && (x.InspiredByRecipeID == nil || (*input.InspiredByRecipeID != "" && *input.InspiredByRecipeID != *x.InspiredByRecipeID)) {
		x.InspiredByRecipeID = input.InspiredByRecipeID
	}

	if input.MinEstimatedPortions != nil && *input.MinEstimatedPortions != x.MinEstimatedPortions {
		x.MinEstimatedPortions = *input.MinEstimatedPortions
	}

	if input.MaxEstimatedPortions != nil && (x.MaxEstimatedPortions == nil || *input.MaxEstimatedPortions != *x.MaxEstimatedPortions) {
		x.MaxEstimatedPortions = input.MaxEstimatedPortions
	}

	if input.PortionName != nil && *input.PortionName != x.PortionName {
		x.PortionName = *input.PortionName
	}

	if input.PluralPortionName != nil && *input.PluralPortionName != x.PluralPortionName {
		x.PluralPortionName = *input.PluralPortionName
	}

	if input.EligibleForMeals != nil && *input.EligibleForMeals != x.EligibleForMeals {
		x.EligibleForMeals = *input.EligibleForMeals
	}

	if input.YieldsComponentType != nil && *input.YieldsComponentType != x.YieldsComponentType {
		x.YieldsComponentType = *input.YieldsComponentType
	}
}

var _ validation.ValidatableWithContext = (*RecipeCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeCreationRequestInput.
func (x *RecipeCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	errResult := &multierror.Error{}

	if len(x.Steps) < 2 {
		errResult = multierror.Append(fmt.Errorf("recipe must have at least 2 steps"), errResult)
	}

	stepIndicesMentionedInPrepTasks := map[uint32]bool{}
	for i, task := range x.PrepTasks {
		for j, step := range task.RecipeSteps {
			if _, ok := stepIndicesMentionedInPrepTasks[step.BelongsToRecipeStepIndex]; ok {
				errResult = multierror.Append(fmt.Errorf("duplicate step mentioned in step %d for task %d", i+1, j+1), errResult)
			} else {
				stepIndicesMentionedInPrepTasks[step.BelongsToRecipeStepIndex] = true
			}
		}
	}

	if err := errResult.ErrorOrNil(); err != nil {
		return err
	}

	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.MinEstimatedPortions, validation.Required),
		validation.Field(&x.PluralPortionName, validation.Required),
		validation.Field(&x.PortionName, validation.Required),
		validation.Field(&x.Slug, validation.Required),
		validation.Field(&x.Steps, validation.Required),
		validation.Field(&x.YieldsComponentType,
			validation.Required,
			validation.In(
				MealComponentTypesUnspecified,
				MealComponentTypesAmuseBouche,
				MealComponentTypesAppetizer,
				MealComponentTypesSoup,
				MealComponentTypesMain,
				MealComponentTypesSalad,
				MealComponentTypesBeverage,
				MealComponentTypesSide,
				MealComponentTypesDessert,
			),
		),
	)
}

// FindStepByIndex finds a step for a given index.
func (x *RecipeDatabaseCreationInput) FindStepByIndex(index uint32) *RecipeStepDatabaseCreationInput {
	for _, step := range x.Steps {
		if step.Index == index {
			return step
		}
	}

	// we could return an error here, but that would make my life a little harder
	// so if you fuck up and submit a wrong value, and it's your fault here.
	return nil
}

// FindStepByID finds a step for a given ID.
func (x *RecipeDatabaseCreationInput) FindStepByID(id string) *RecipeStepDatabaseCreationInput {
	for _, step := range x.Steps {
		if step.ID == id {
			return step
		}
	}

	// we could return an error here, but that would make my life a little harder
	// so if you fuck up and submit a wrong value, and it's your fault here.
	return nil
}

// GetAllValidIngredientPreparationIDs returns all ValidIngredientPreparationIDs from all steps' ingredients.
func (x *RecipeDatabaseCreationInput) GetAllValidIngredientPreparationIDs() []string {
	var ids []string
	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.ValidIngredientPreparationID != nil && *ingredient.ValidIngredientPreparationID != "" {
				ids = append(ids, *ingredient.ValidIngredientPreparationID)
			}
		}
	}
	return ids
}

// GetAllValidIngredientMeasurementUnitIDs returns all ValidIngredientMeasurementUnitIDs from all steps' ingredients.
func (x *RecipeDatabaseCreationInput) GetAllValidIngredientMeasurementUnitIDs() []string {
	var ids []string
	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.ValidIngredientMeasurementUnitID != nil && *ingredient.ValidIngredientMeasurementUnitID != "" {
				ids = append(ids, *ingredient.ValidIngredientMeasurementUnitID)
			}
		}
	}
	return ids
}

// GetAllValidPreparationInstrumentIDs returns all ValidPreparationInstrumentIDs from all steps' instruments.
func (x *RecipeDatabaseCreationInput) GetAllValidPreparationInstrumentIDs() []string {
	var ids []string
	for _, step := range x.Steps {
		for _, instrument := range step.Instruments {
			if instrument.ValidPreparationInstrumentID != nil && *instrument.ValidPreparationInstrumentID != "" {
				ids = append(ids, *instrument.ValidPreparationInstrumentID)
			}
		}
	}
	return ids
}

// GetAllValidPreparationVesselIDs returns all ValidPreparationVesselIDs from all steps' vessels.
func (x *RecipeDatabaseCreationInput) GetAllValidPreparationVesselIDs() []string {
	var ids []string
	for _, step := range x.Steps {
		for _, vessel := range step.Vessels {
			if vessel.ValidPreparationVesselID != nil && *vessel.ValidPreparationVesselID != "" {
				ids = append(ids, *vessel.ValidPreparationVesselID)
			}
		}
	}
	return ids
}

// GetRelatedRecipeIDs returns all recipe IDs that this recipe references as components.
// It iterates through all steps and their ingredients to find any that reference other recipes
// via RecipeStepProductRecipeID.
func (x *RecipeDatabaseCreationInput) GetRelatedRecipeIDs() []string {
	seen := make(map[string]struct{})
	var relatedIDs []string

	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != "" {
				recipeID := *ingredient.RecipeStepProductRecipeID
				if _, exists := seen[recipeID]; !exists {
					seen[recipeID] = struct{}{}
					relatedIDs = append(relatedIDs, recipeID)
				}
			}
		}
	}

	return relatedIDs
}

var _ validation.ValidatableWithContext = (*RecipeDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeDatabaseCreationInput.
func (x *RecipeDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.CreatedByUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeUpdateRequestInput.
func (x *RecipeUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Source, validation.Required),
		validation.Field(&x.Description, validation.Required),
		validation.Field(&x.MinEstimatedPortions, validation.Required),
	)
}
