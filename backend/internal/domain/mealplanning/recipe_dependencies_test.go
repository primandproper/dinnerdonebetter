package mealplanning_test

import (
	"context"
	"errors"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dependencyGraph answers a RecipeDependencyReader from a fixed graph, and records what it was
// asked.
type dependencyGraph struct {
	edges map[string][]string
	err   error
	asked []string
}

func (g *dependencyGraph) read(_ context.Context, recipeID string) (map[string]bool, error) {
	g.asked = append(g.asked, recipeID)
	if g.err != nil {
		return nil, g.err
	}

	out := map[string]bool{}
	for _, dependency := range g.edges[recipeID] {
		out[dependency] = true
	}

	return out, nil
}

func TestRecipe_CrossRecipeDependencies(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		recipe := fakes.BuildFakeRecipe()
		other := fake.BuildFakeID()
		for _, step := range recipe.Steps {
			for _, ingredient := range step.Ingredients {
				ingredient.RecipeStepProductRecipeID = nil
			}
		}

		// One ingredient draws on another recipe, one on its own recipe, one on nothing.
		ingredients := recipe.Steps[0].Ingredients
		require.GreaterOrEqual(t, len(ingredients), 3)
		ingredients[0].RecipeStepProductRecipeID = pointer.To(other)
		ingredients[1].RecipeStepProductRecipeID = pointer.To(recipe.ID)
		ingredients[2].RecipeStepProductRecipeID = pointer.To("")

		assert.Equal(t, map[string]bool{other: true}, recipe.CrossRecipeDependencies())
	})
}

func TestRecipeDatabaseCreationInput_CrossRecipeDependencies(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		input := &types.RecipeDatabaseCreationInput{ID: fake.BuildFakeID()}
		other := fake.BuildFakeID()
		input.Steps = []*types.RecipeStepDatabaseCreationInput{{
			Ingredients: []*types.RecipeStepIngredientDatabaseCreationInput{
				{RecipeStepProductRecipeID: pointer.To(other)},
				{RecipeStepProductRecipeID: pointer.To(input.ID)},
				{},
			},
		}}

		assert.Equal(t, map[string]bool{other: true}, input.CrossRecipeDependencies())
	})
}

func TestCheckRecipeDependencies(T *testing.T) {
	T.Parallel()

	T.Run("with dependencies that lead nowhere back", func(t *testing.T) {
		t.Parallel()

		recipeID, a, b := fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()
		graph := &dependencyGraph{edges: map[string][]string{a: {b}}}

		require.NoError(t, types.CheckRecipeDependencies(t.Context(), recipeID, map[string]bool{a: true}, graph.read))
		assert.ElementsMatch(t, []string{a, b}, graph.asked)
	})

	T.Run("with a recipe drawing on itself", func(t *testing.T) {
		t.Parallel()

		recipeID := fake.BuildFakeID()
		graph := &dependencyGraph{}

		err := types.CheckRecipeDependencies(t.Context(), recipeID, map[string]bool{recipeID: true}, graph.read)
		require.ErrorIs(t, err, types.ErrInvalidRecipeInput)
		assert.Empty(t, graph.asked)
	})

	T.Run("with a cycle through another recipe", func(t *testing.T) {
		t.Parallel()

		recipeID, a, b := fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()
		graph := &dependencyGraph{edges: map[string][]string{a: {b}, b: {recipeID}}}

		err := types.CheckRecipeDependencies(t.Context(), recipeID, map[string]bool{a: true}, graph.read)
		require.ErrorIs(t, err, types.ErrInvalidRecipeInput)
	})

	T.Run("with a cycle elsewhere that does not pass through the recipe", func(t *testing.T) {
		t.Parallel()

		recipeID, a, b := fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()
		graph := &dependencyGraph{edges: map[string][]string{a: {b}, b: {a}}}

		// A cycle among recipes already stored is still a cycle, and drawing on it is refused.
		err := types.CheckRecipeDependencies(t.Context(), recipeID, map[string]bool{a: true}, graph.read)
		require.ErrorIs(t, err, types.ErrInvalidRecipeInput)
	})

	T.Run("with a dependency that cannot be read", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New(fake.BuildFakeID())
		graph := &dependencyGraph{err: expectedErr}

		err := types.CheckRecipeDependencies(t.Context(), fake.BuildFakeID(), map[string]bool{fake.BuildFakeID(): true}, graph.read)
		require.ErrorIs(t, err, expectedErr)
		assert.NotErrorIs(t, err, types.ErrInvalidRecipeInput)
	})

	T.Run("with no dependencies", func(t *testing.T) {
		t.Parallel()

		graph := &dependencyGraph{}

		require.NoError(t, types.CheckRecipeDependencies(t.Context(), fake.BuildFakeID(), map[string]bool{}, graph.read))
		assert.Empty(t, graph.asked)
	})
}
