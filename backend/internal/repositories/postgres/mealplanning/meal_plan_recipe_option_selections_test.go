package mealplanning

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	platformerrors "github.com/primandproper/platform-go/v9/errors"

	"github.com/stretchr/testify/assert"
)

func TestQuerier_GetMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan option ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanRecipeOptionSelection(ctx, "", "recipe_step_id", 0, "ingredient")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanRecipeOptionSelection(ctx, "meal_plan_option_id", "", 0, "ingredient")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Nil(t, actual)
	})

	T.Run("with invalid selection type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanRecipeOptionSelection(ctx, "meal_plan_option_id", "recipe_step_id", 0, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetSelectionsForMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan option ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetSelectionsForMealPlanOption(ctx, "", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetSelectionsForMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetSelectionsForMealPlan(ctx, "", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanRecipeOptionSelection(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateMealPlanRecipeOptionSelection(ctx, "option_id", "step_id", 0, "ingredient", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})

	T.Run("with invalid meal plan option ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		input := fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput()

		err := c.UpdateMealPlanRecipeOptionSelection(ctx, "", "step_id", 0, "ingredient", input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		input := fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput()

		err := c.UpdateMealPlanRecipeOptionSelection(ctx, "option_id", "", 0, "ingredient", input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with invalid selection type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		input := fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput()

		err := c.UpdateMealPlanRecipeOptionSelection(ctx, "option_id", "step_id", 0, "", input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_ArchiveMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan option ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveMealPlanRecipeOptionSelection(ctx, "", "step_id", 0, "ingredient")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveMealPlanRecipeOptionSelection(ctx, "option_id", "", 0, "ingredient")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with invalid selection type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveMealPlanRecipeOptionSelection(ctx, "option_id", "step_id", 0, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}
