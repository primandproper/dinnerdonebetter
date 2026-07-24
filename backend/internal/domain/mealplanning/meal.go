package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v6/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// MealComponentTypesUnspecified represents the unspecified meal component type.
	MealComponentTypesUnspecified = "unspecified"
	// MealComponentTypesAmuseBouche represents the amuse-bouche meal component type.
	MealComponentTypesAmuseBouche = "amuse-bouche"
	// MealComponentTypesAppetizer represents the appetizer meal component type.
	MealComponentTypesAppetizer = "appetizer"
	// MealComponentTypesSoup represents the soup meal component type.
	MealComponentTypesSoup = "soup"
	// MealComponentTypesMain represents the main meal component type.
	MealComponentTypesMain = "main"
	// MealComponentTypesSalad represents the salad meal component type.
	MealComponentTypesSalad = "salad"
	// MealComponentTypesBeverage represents the beverage meal component type.
	MealComponentTypesBeverage = "beverage"
	// MealComponentTypesSide represents the side meal component type.
	MealComponentTypesSide = "side"
	// MealComponentTypesDessert represents the dessert meal component type.
	MealComponentTypesDessert = "dessert"

	// MealCreatedServiceEventType indicates a meal was created.
	MealCreatedServiceEventType = "meal_created"
	// MealUpdatedServiceEventType indicates a meal was updated.
	MealUpdatedServiceEventType = "meal_updated"
	// MealArchivedServiceEventType indicates a meal was archived.
	MealArchivedServiceEventType = "meal_archived"
)

var (
	errOneMainMinimumRequired = errors.New("at least one main required for meal creation")
)

func init() {
	gob.Register(new(Meal))
	gob.Register(new(MealCreationRequestInput))
}

type (
	// Meal represents a meal.
	Meal struct {
		_ struct{} `json:"-"`

		CreatedAt            time.Time        `json:"createdAt"`
		ArchivedAt           *time.Time       `json:"archivedAt"`
		LastUpdatedAt        *time.Time       `json:"lastUpdatedAt"`
		MaxEstimatedPortions *float32         `json:"maxEstimatedPortions,omitempty"`
		ID                   string           `json:"id"`
		Description          string           `json:"description"`
		CreatedByUser        string           `json:"createdByUser"`
		Name                 string           `json:"name"`
		Components           []*MealComponent `json:"components"`
		MinEstimatedPortions float32          `json:"minEstimatedPortions"`
		EligibleForMealPlans bool             `json:"eligibleForMealPlans"`
	}

	// MealComponent is a recipe with some extra data attached to it.
	MealComponent struct {
		_ struct{} `json:"-"`

		ComponentType string  `json:"componentType"`
		Recipe        Recipe  `json:"recipe"`
		RecipeScale   float32 `json:"recipeScale"`
	}

	// MealCreationRequestInput represents what a user could set as input for creating meals.
	MealCreationRequestInput struct {
		_ struct{} `json:"-"`

		MaxEstimatedPortions *float32                             `json:"maxEstimatedPortions,omitempty"`
		Name                 string                               `json:"name"`
		Description          string                               `json:"description"`
		Components           []*MealComponentCreationRequestInput `json:"components"`
		MinEstimatedPortions float32                              `json:"minEstimatedPortions"`
		EligibleForMealPlans bool                                 `json:"eligibleForMealPlans"`
	}

	// MealComponentCreationRequestInput represents what a user could set as input for creating meal recipes.
	MealComponentCreationRequestInput struct {
		_ struct{} `json:"-"`

		RecipeID      string  `json:"recipeID"`
		ComponentType string  `json:"componentType"`
		RecipeScale   float32 `json:"recipeScale"`
	}

	// MealDatabaseCreationInput represents what a user could set as input for creating meals.
	MealDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		MaxEstimatedPortions *float32                              `json:"-"`
		ID                   string                                `json:"-"`
		Name                 string                                `json:"-"`
		Description          string                                `json:"-"`
		CreatedByUser        string                                `json:"-"`
		Components           []*MealComponentDatabaseCreationInput `json:"-"`
		MinEstimatedPortions float32                               `json:"-"`
		EligibleForMealPlans bool                                  `json:"-"`
	}

	// MealComponentDatabaseCreationInput represents what a user could set as input for creating meal recipes.
	MealComponentDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		RecipeID      string  `json:"-"`
		ComponentType string  `json:"-"`
		RecipeScale   float32 `json:"-"`
	}

	// MealDataManager describes a structure capable of storing meals permanently.
	MealDataManager interface {
		MealExists(ctx context.Context, mealID string) (bool, error)
		FindMealWithSameComponents(ctx context.Context, creatorID string, input *MealCreationRequestInput) (*Meal, error)
		GetMeal(ctx context.Context, mealID string) (*Meal, error)
		GetMeals(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Meal], error)
		GetMealsCreatedByUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Meal], error)
		SearchForMeals(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Meal], error)
		CreateMeal(ctx context.Context, input *MealDatabaseCreationInput) (*Meal, error)
		MarkMealAsIndexed(ctx context.Context, mealID string) error
		ArchiveMeal(ctx context.Context, mealID, userID string) error
		GetMealIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error)
		GetMealsWithIDs(ctx context.Context, ids []string) ([]*Meal, error)
		AddMealImage(ctx context.Context, mealID, uploadedMediaID, uploadedByUser string) error
	}
)

var _ validation.ValidatableWithContext = (*MealCreationRequestInput)(nil)

// ValidateWithContext validates a MealCreationRequestInput.
func (x *MealCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	atLeastOneMain := false
	for _, component := range x.Components {
		if component.ComponentType == MealComponentTypesMain {
			atLeastOneMain = true
		}

		if componentValidationErr := component.ValidateWithContext(ctx); componentValidationErr != nil {
			result = multierror.Append(result, componentValidationErr)
		}
	}

	if !atLeastOneMain {
		result = multierror.Append(result, errOneMainMinimumRequired)
	}

	if validationErr := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Components, validation.Required),
	); validationErr != nil {
		result = multierror.Append(result, validationErr)
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*MealCreationRequestInput)(nil)

// ValidateWithContext validates a MealComponentCreationRequestInput.
func (x *MealComponentCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ComponentType,
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

var _ validation.ValidatableWithContext = (*MealDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealDatabaseCreationInput.
func (x *MealDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Components, validation.Required),
		validation.Field(&x.CreatedByUser, validation.Required),
	)
}
