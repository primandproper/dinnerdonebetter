package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkRecipeMediaSliceEquality(t *testing.T, stepIndex int, expected, actual []*mealplanning.RecipeMedia) {
	t.Helper()

	require.Equal(t, len(expected), len(actual), "expected recipe step %d media length", stepIndex)
	for i := range expected {
		e, a := expected[i], actual[i]
		checkRecipeMediaEquality(t, stepIndex, i, e, a)
	}
}

func checkRecipeMediaEquality(t *testing.T, stepIndex, mediaIndex int, expected, actual *mealplanning.RecipeMedia) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected step %d media %d to have ID", stepIndex, mediaIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected step %d media %d to have CreatedAt", stepIndex, mediaIndex)
	assert.Equal(t, expected.MimeType, actual.MimeType, "expected step %d media %d MimeType", stepIndex, mediaIndex)
	assert.Equal(t, expected.InternalPath, actual.InternalPath, "expected step %d media %d InternalPath", stepIndex, mediaIndex)
	assert.Equal(t, expected.ExternalPath, actual.ExternalPath, "expected step %d media %d ExternalPath", stepIndex, mediaIndex)
	assert.Equal(t, expected.Index, actual.Index, "expected step %d media %d Index", stepIndex, mediaIndex)
	if expected.BelongsToRecipe != nil {
		require.NotNil(t, actual.BelongsToRecipe, "expected step %d media %d BelongsToRecipe non-nil", stepIndex, mediaIndex)
		assert.Equal(t, *expected.BelongsToRecipe, *actual.BelongsToRecipe, "expected step %d media %d BelongsToRecipe", stepIndex, mediaIndex)
	}
	if expected.BelongsToRecipeStep != nil {
		require.NotNil(t, actual.BelongsToRecipeStep, "expected step %d media %d BelongsToRecipeStep non-nil", stepIndex, mediaIndex)
		assert.Equal(t, *expected.BelongsToRecipeStep, *actual.BelongsToRecipeStep, "expected step %d media %d BelongsToRecipeStep", stepIndex, mediaIndex)
	}
}

func checkRecipeLevelMediaSliceEquality(t *testing.T, expected, actual []*mealplanning.RecipeMedia) {
	t.Helper()

	require.Equal(t, len(expected), len(actual), "expected recipe media length")
	for i := range expected {
		checkRecipeLevelMediaEquality(t, i, expected[i], actual[i])
	}
}

func checkRecipeLevelMediaEquality(t *testing.T, mediaIndex int, expected, actual *mealplanning.RecipeMedia) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected recipe media %d to have ID", mediaIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected recipe media %d to have CreatedAt", mediaIndex)
	assert.Equal(t, expected.MimeType, actual.MimeType, "expected recipe media %d MimeType", mediaIndex)
	assert.Equal(t, expected.InternalPath, actual.InternalPath, "expected recipe media %d InternalPath", mediaIndex)
	assert.Equal(t, expected.ExternalPath, actual.ExternalPath, "expected recipe media %d ExternalPath", mediaIndex)
	assert.Equal(t, expected.Index, actual.Index, "expected recipe media %d Index", mediaIndex)
	if expected.BelongsToRecipe != nil {
		require.NotNil(t, actual.BelongsToRecipe, "expected recipe media %d BelongsToRecipe non-nil", mediaIndex)
		assert.Equal(t, *expected.BelongsToRecipe, *actual.BelongsToRecipe, "expected recipe media %d BelongsToRecipe", mediaIndex)
	}
	if expected.BelongsToRecipeStep != nil {
		require.NotNil(t, actual.BelongsToRecipeStep, "expected recipe media %d BelongsToRecipeStep non-nil", mediaIndex)
		assert.Equal(t, *expected.BelongsToRecipeStep, *actual.BelongsToRecipeStep, "expected recipe media %d BelongsToRecipeStep", mediaIndex)
	}
}
