package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkRecipeStepProductSliceEquality(t *testing.T, stepIndex int, expected, actual []*mealplanning.RecipeStepProduct) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "expected recipe step %d products length", stepIndex)
	for i := range expected {
		checkRecipeStepProductEquality(t, stepIndex, i, expected[i], actual[i])
	}
}

func checkRecipeStepProductEquality(t *testing.T, stepIndex, productIndex int, expected, actual *mealplanning.RecipeStepProduct) {
	t.Helper()
	assert.NotEmpty(t, actual.ID, "expected step %d product %d to have ID", stepIndex, productIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected step %d product %d to have CreatedAt", stepIndex, productIndex)
	assert.NotEmpty(t, actual.BelongsToRecipeStep, "expected step %d product %d to have BelongsToRecipeStep", stepIndex, productIndex)
	assert.Equal(t, expected.Name, actual.Name, "expected step %d product %d Name", stepIndex, productIndex)
	assert.Equal(t, expected.Type, actual.Type, "expected step %d product %d Type", stepIndex, productIndex)
	assert.Equal(t, expected.MinMeasurementQuantity, actual.MinMeasurementQuantity, "expected step %d product %d MinMeasurementQuantity", stepIndex, productIndex)
	assert.Equal(t, expected.MaxMeasurementQuantity, actual.MaxMeasurementQuantity, "expected step %d product %d MaxMeasurementQuantity", stepIndex, productIndex)
	assert.Equal(t, expected.MinItemQuantity, actual.MinItemQuantity, "expected step %d product %d MinItemQuantity", stepIndex, productIndex)
	assert.Equal(t, expected.MaxItemQuantity, actual.MaxItemQuantity, "expected step %d product %d MaxItemQuantity", stepIndex, productIndex)
	assert.Equal(t, expected.QuantityNotes, actual.QuantityNotes, "expected step %d product %d QuantityNotes", stepIndex, productIndex)
	assert.Equal(t, expected.Index, actual.Index, "expected step %d product %d Index", stepIndex, productIndex)
	if expected.MeasurementUnit != nil {
		require.NotNil(t, actual.MeasurementUnit, "expected step %d product %d MeasurementUnit non-nil", stepIndex, productIndex)
		assert.NotEmpty(t, actual.MeasurementUnit.ID, "expected step %d product %d MeasurementUnit.ID", stepIndex, productIndex)
		assert.Equal(t, expected.MeasurementUnit.ID, actual.MeasurementUnit.ID, "expected step %d product %d MeasurementUnit.ID", stepIndex, productIndex)
	}
}

func TestRecipeStepProducts_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should update", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		var (
			createdRecipeStepID string
		)
		for _, step := range createdRecipe.Steps {
			createdRecipeStepID = step.ID
			break
		}

		createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)

		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
		exampleRecipeStepProduct.BelongsToRecipeStep = createdRecipeStepID
		exampleRecipeStepProduct.MeasurementUnit = createdValidMeasurementUnit
		exampleRecipeStepProductInput := mpconverters.ConvertRecipeStepProductToRecipeStepProductCreationRequestInput(exampleRecipeStepProduct)
		createdRecipeStepProductRes, err := adminClient.CreateRecipeStepProduct(ctx, &mealplanninggrpc.CreateRecipeStepProductRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
			Input:        converters.ConvertRecipeStepProductCreationRequestInputToGRPCRecipeStepProductCreationRequestInput(exampleRecipeStepProductInput),
		})
		require.NoError(t, err)

		createdRecipeStepProduct := converters.ConvertGRPCRecipeStepProductToRecipeStepProduct(createdRecipeStepProductRes.Created)
		checkRecipeStepProductEquality(t, -1, -1, exampleRecipeStepProduct, createdRecipeStepProduct)

		retrievedRecipeStepProductRes, err := adminClient.GetRecipeStepProduct(ctx, &mealplanninggrpc.GetRecipeStepProductRequest{
			RecipeId:            createdRecipe.ID,
			RecipeStepId:        createdRecipeStepID,
			RecipeStepProductId: createdRecipeStepProduct.ID,
		})
		require.NoError(t, err)
		require.Equal(t, createdRecipeStepID, createdRecipeStepProduct.BelongsToRecipeStep)

		createdRecipeStepProduct = converters.ConvertGRPCRecipeStepProductToRecipeStepProduct(retrievedRecipeStepProductRes.Result)
		checkRecipeStepProductEquality(t, -1, -1, exampleRecipeStepProduct, createdRecipeStepProduct)

		newRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
		newRecipeStepProduct.MeasurementUnit = createdValidMeasurementUnit
		updateInput := mpconverters.ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput(newRecipeStepProduct)
		createdRecipeStepProduct.Update(updateInput)

		_, err = adminClient.UpdateRecipeStepProduct(ctx, &mealplanninggrpc.UpdateRecipeStepProductRequest{
			RecipeId:            createdRecipe.ID,
			RecipeStepId:        createdRecipeStepID,
			RecipeStepProductId: createdRecipeStepProduct.ID,
			Input:               converters.ConvertRecipeStepProductUpdateRequestInputToGRPCRecipeStepProductUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		retrievedRecipeStepProductRes, err = adminClient.GetRecipeStepProduct(ctx, &mealplanninggrpc.GetRecipeStepProductRequest{
			RecipeId:            createdRecipe.ID,
			RecipeStepId:        createdRecipeStepID,
			RecipeStepProductId: createdRecipeStepProduct.ID,
		})
		require.NoError(t, err)

		actual := converters.ConvertGRPCRecipeStepProductToRecipeStepProduct(retrievedRecipeStepProductRes.Result)

		// assert recipe step product equality
		checkRecipeStepProductEquality(t, -1, -1, newRecipeStepProduct, actual)
		assert.NotNil(t, actual.LastUpdatedAt)

		_, err = adminClient.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
			RecipeId:            createdRecipe.ID,
			RecipeStepId:        createdRecipeStepID,
			RecipeStepProductId: createdRecipeStepProduct.ID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepProducts_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		var (
			createdRecipeStepID string
		)
		for _, step := range createdRecipe.Steps {
			createdRecipeStepID = step.ID
			break
		}

		createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)

		var expected []*mealplanning.RecipeStepProduct
		for range 5 {
			exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
			exampleRecipeStepProduct.BelongsToRecipeStep = createdRecipeStepID
			exampleRecipeStepProduct.MeasurementUnit = createdValidMeasurementUnit
			exampleRecipeStepProductInput := mpconverters.ConvertRecipeStepProductToRecipeStepProductCreationRequestInput(exampleRecipeStepProduct)
			createdRecipeStepProductRes, err := adminClient.CreateRecipeStepProduct(ctx, &mealplanninggrpc.CreateRecipeStepProductRequest{
				RecipeId:     createdRecipe.ID,
				RecipeStepId: createdRecipeStepID,
				Input:        converters.ConvertRecipeStepProductCreationRequestInputToGRPCRecipeStepProductCreationRequestInput(exampleRecipeStepProductInput),
			})
			require.NoError(t, err)

			createdRecipeStepProduct := converters.ConvertGRPCRecipeStepProductToRecipeStepProduct(createdRecipeStepProductRes.Created)
			checkRecipeStepProductEquality(t, -1, -1, exampleRecipeStepProduct, createdRecipeStepProduct)

			retrievedRecipeStepProductRes, err := adminClient.GetRecipeStepProduct(ctx, &mealplanninggrpc.GetRecipeStepProductRequest{
				RecipeId:            createdRecipe.ID,
				RecipeStepId:        createdRecipeStepID,
				RecipeStepProductId: createdRecipeStepProduct.ID,
			})
			require.NoError(t, err)

			createdRecipeStepProduct = converters.ConvertGRPCRecipeStepProductToRecipeStepProduct(retrievedRecipeStepProductRes.Result)
			require.Equal(t, createdRecipeStepID, createdRecipeStepProduct.BelongsToRecipeStep)

			expected = append(expected, createdRecipeStepProduct)
		}

		// assert recipe step product list equality
		actual, err := adminClient.GetRecipeStepProducts(ctx, &mealplanninggrpc.GetRecipeStepProductsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		assert.True(
			t,
			len(expected) <= len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdRecipeStepProduct := range expected {
			_, err = adminClient.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
				RecipeId:            createdRecipe.ID,
				RecipeStepId:        createdRecipeStepID,
				RecipeStepProductId: createdRecipeStepProduct.ID,
			})
			assert.NoError(t, err)
		}

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepProducts_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("non-owner cannot create recipe step product", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
		exampleRecipeStepProductInput := mpconverters.ConvertRecipeStepProductToRecipeStepProductCreationRequestInput(exampleRecipeStepProduct)
		_, err := otherUserClient.CreateRecipeStepProduct(ctx, &mealplanninggrpc.CreateRecipeStepProductRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepProductCreationRequestInputToGRPCRecipeStepProductCreationRequestInput(exampleRecipeStepProductInput),
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot update recipe step product", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]
		createdRecipeStepProduct := createdRecipeStep.Products[0]

		newRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
		updateInput := mpconverters.ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput(newRecipeStepProduct)
		_, err := otherUserClient.UpdateRecipeStepProduct(ctx, &mealplanninggrpc.UpdateRecipeStepProductRequest{
			RecipeId:            createdRecipe.ID,
			RecipeStepId:        createdRecipeStep.ID,
			RecipeStepProductId: createdRecipeStepProduct.ID,
			Input:               converters.ConvertRecipeStepProductUpdateRequestInputToGRPCRecipeStepProductUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot archive recipe step product", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]
		createdRecipeStepProduct := createdRecipeStep.Products[0]

		_, err := otherUserClient.ArchiveRecipeStepProduct(ctx, &mealplanninggrpc.ArchiveRecipeStepProductRequest{
			RecipeId:            createdRecipe.ID,
			RecipeStepId:        createdRecipeStep.ID,
			RecipeStepProductId: createdRecipeStepProduct.ID,
		})
		require.Error(t, err)
	})
}
