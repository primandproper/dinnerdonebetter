package grpcconverters

import (
	"testing"

	fakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"
	platformgrpc "github.com/primandproper/platform-go/v13/server/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// maxPage is the largest page any list endpoint can be made to answer with.
// QueryFilter clamps to MaxQueryFilterLimit rather than rejecting, so a client
// asking for more gets exactly this many, and every assertion below is about
// the worst page the server can be talked into producing.
func maxPage() int { return int(filtering.MaxQueryFilterLimit) }

// The tests below are the reason RecipeSummary and MealSummary exist. Whoever
// adds a repeated field to either message will fail here rather than in a
// ResourceExhausted on a client we do not operate.
//
// The fakes are the point: BuildFakeRecipe populates steps, prep tasks, media
// and associated recipes, so the recipe going into the converter is the heavy
// one the wire used to carry.

func TestRecipeSummary_MaxLimitPageFitsInAGRPCMessage(T *testing.T) {
	T.Parallel()

	T.Run("a full page of summaries fits", func(t *testing.T) {
		t.Parallel()

		response := &mealplanningsvc.GetRecipesResponse{}
		for range maxPage() {
			response.Results = append(response.Results, ConvertRecipeToGRPCRecipeSummary(fakes.BuildFakeRecipe()))
		}

		require.Len(t, response.Results, maxPage())
		assert.Less(t, proto.Size(response), platformgrpc.DefaultMaxMessageSize)
	})

	T.Run("the same page of whole recipes does not", func(t *testing.T) {
		t.Parallel()

		var results []*mealplanningsvc.Recipe
		for range maxPage() {
			results = append(results, ConvertRecipeToGRPCRecipe(fakes.BuildFakeRecipe()))
		}

		// Not an invariant anyone should rely on — it is the measurement that
		// makes the summary worth having. A hydrated Recipe embeds a whole
		// ValidPreparation per step and a whole ValidIngredient and
		// ValidMeasurementUnit per ingredient, and associated_recipes is
		// self-recursive, so there is no page size that provably fits.
		size := 0
		for _, r := range results {
			size += proto.Size(r)
		}
		assert.Greater(t, size, platformgrpc.DefaultMaxMessageSize)
	})
}

func TestMealSummary_MaxLimitPageFitsInAGRPCMessage(T *testing.T) {
	T.Parallel()

	T.Run("a full page of summaries fits", func(t *testing.T) {
		t.Parallel()

		response := &mealplanningsvc.GetMealsResponse{}
		for range maxPage() {
			response.Results = append(response.Results, ConvertMealToGRPCMealSummary(fakes.BuildFakeMeal()))
		}

		require.Len(t, response.Results, maxPage())
		assert.Less(t, proto.Size(response), platformgrpc.DefaultMaxMessageSize)
	})

	T.Run("the same page of whole meals does not", func(t *testing.T) {
		t.Parallel()

		var results []*mealplanningsvc.Meal
		for range maxPage() {
			results = append(results, ConvertMealToGRPCMeal(fakes.BuildFakeMeal()))
		}

		// A MealComponent embeds a whole Recipe, so a Meal is strictly worse
		// than the recipe above by however many components it has.
		size := 0
		for _, m := range results {
			size += proto.Size(m)
		}
		assert.Greater(t, size, platformgrpc.DefaultMaxMessageSize)
	})
}

func TestConvertRecipeToGRPCRecipeSummary(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		input := fakes.BuildFakeRecipe()
		require.NotEmpty(t, input.Steps)

		result := ConvertRecipeToGRPCRecipeSummary(input)

		require.NotNil(t, result)
		assert.Equal(t, input.ID, result.Id)
		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Slug, result.Slug)
		assert.Equal(t, input.Status, result.Status)
		assert.Equal(t, input.EligibleForMeals, result.EligibleForMeals)
	})

	T.Run("round trips back to a recipe with nothing hanging off it", func(t *testing.T) {
		t.Parallel()

		input := fakes.BuildFakeRecipe()

		result := ConvertGRPCRecipeSummaryToRecipe(ConvertRecipeToGRPCRecipeSummary(input))

		require.NotNil(t, result)
		assert.Equal(t, input.ID, result.ID)
		assert.Equal(t, input.Name, result.Name)
		assert.Empty(t, result.Steps)
		assert.Empty(t, result.PrepTasks)
		assert.Empty(t, result.Media)
		assert.Empty(t, result.AssociatedRecipes)
	})
}

func TestConvertMealToGRPCMealSummary(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		input := fakes.BuildFakeMeal()
		require.NotEmpty(t, input.Components)

		result := ConvertMealToGRPCMealSummary(input)

		require.NotNil(t, result)
		assert.Equal(t, input.ID, result.Id)
		assert.Equal(t, input.Name, result.Name)
		assert.Len(t, result.Components, len(input.Components))

		// The components survive; the recipes inside them are trimmed.
		for i, component := range result.Components {
			require.NotNil(t, component.Recipe)
			assert.Equal(t, input.Components[i].Recipe.ID, component.Recipe.Id)
			assert.Equal(t, input.Components[i].Recipe.Name, component.Recipe.Name)
		}
	})

	T.Run("round trips back to a meal whose recipes carry nothing", func(t *testing.T) {
		t.Parallel()

		input := fakes.BuildFakeMeal()

		result := ConvertGRPCMealSummaryToMeal(ConvertMealToGRPCMealSummary(input))

		require.NotNil(t, result)
		assert.Equal(t, input.ID, result.ID)
		assert.Len(t, result.Components, len(input.Components))
		for _, component := range result.Components {
			assert.Empty(t, component.Recipe.Steps)
			assert.Empty(t, component.Recipe.AssociatedRecipes)
		}
	})
}
