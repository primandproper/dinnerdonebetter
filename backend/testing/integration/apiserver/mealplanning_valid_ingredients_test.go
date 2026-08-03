package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkValidIngredientEquality(t *testing.T, expected, actual *mealplanning.ValidIngredient) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected ValidIngredient to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected ValidIngredient to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected ValidIngredient Name")
	assert.Equal(t, expected.Description, actual.Description, "expected ValidIngredient Description")
	assert.Equal(t, expected.Slug, actual.Slug, "expected ValidIngredient Slug")
	assert.Equal(t, expected.PluralName, actual.PluralName, "expected ValidIngredient PluralName")
	assert.Equal(t, expected.IconPath, actual.IconPath, "expected ValidIngredient IconPath")
	assert.Equal(t, expected.Warning, actual.Warning, "expected ValidIngredient Warning")
	assert.Equal(t, expected.StorageInstructions, actual.StorageInstructions, "expected ValidIngredient StorageInstructions")
	assert.Equal(t, expected.ShoppingSuggestions, actual.ShoppingSuggestions, "expected ValidIngredient ShoppingSuggestions")
	assert.Equal(t, expected.ContainsShellfish, actual.ContainsShellfish, "expected ValidIngredient ContainsShellfish")
	assert.Equal(t, expected.IsLiquid, actual.IsLiquid, "expected ValidIngredient IsLiquid")
	assert.Equal(t, expected.ContainsPeanut, actual.ContainsPeanut, "expected ValidIngredient ContainsPeanut")
	assert.Equal(t, expected.ContainsTreeNut, actual.ContainsTreeNut, "expected ValidIngredient ContainsTreeNut")
	assert.Equal(t, expected.ContainsEgg, actual.ContainsEgg, "expected ValidIngredient ContainsEgg")
	assert.Equal(t, expected.ContainsWheat, actual.ContainsWheat, "expected ValidIngredient ContainsWheat")
	assert.Equal(t, expected.ContainsSoy, actual.ContainsSoy, "expected ValidIngredient ContainsSoy")
	assert.Equal(t, expected.AnimalDerived, actual.AnimalDerived, "expected ValidIngredient AnimalDerived")
	assert.Equal(t, expected.RestrictToPreparations, actual.RestrictToPreparations, "expected ValidIngredient RestrictToPreparations")
	assert.Equal(t, expected.ContainsSesame, actual.ContainsSesame, "expected ValidIngredient ContainsSesame")
	assert.Equal(t, expected.ContainsFish, actual.ContainsFish, "expected ValidIngredient ContainsFish")
	assert.Equal(t, expected.MinStorageTemperatureInCelsius, actual.MinStorageTemperatureInCelsius, "expected ValidIngredient MinStorageTemperatureInCelsius")
	assert.Equal(t, expected.MaxStorageTemperatureInCelsius, actual.MaxStorageTemperatureInCelsius, "expected ValidIngredient MaxStorageTemperatureInCelsius")
}

func createValidIngredientForTest(t *testing.T) *mealplanning.ValidIngredient {
	t.Helper()

	ctx := t.Context()

	exampleValidIngredient := fakes.BuildFakeValidIngredient()
	exampleValidIngredientInput := converters.ConvertValidIngredientToValidIngredientCreationRequestInput(exampleValidIngredient)
	created, err := adminClient.CreateValidIngredient(ctx, &mealplanningsvc.CreateValidIngredientRequest{
		Input: grpcconverters.ConvertValidIngredientCreationRequestInputToGRPCValidIngredientCreationRequestInput(exampleValidIngredientInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCValidIngredientToValidIngredient(created.Result)
	checkValidIngredientEquality(t, exampleValidIngredient, converted)

	retrieved, err := adminClient.GetValidIngredient(ctx, &mealplanningsvc.GetValidIngredientRequest{
		ValidIngredientId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	validIngredient := grpcconverters.ConvertGRPCValidIngredientToValidIngredient(retrieved.Result)
	checkValidIngredientEquality(t, converted, validIngredient)

	return validIngredient
}

func TestValidIngredients_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidIngredientForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidIngredientCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidIngredientCreationRequestInputToGRPCValidIngredientCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateValidIngredient(ctx, &mealplanningsvc.CreateValidIngredientRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidIngredientCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidIngredientCreationRequestInputToGRPCValidIngredientCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateValidIngredient(ctx, &mealplanningsvc.CreateValidIngredientRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeValidIngredientCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidIngredientCreationRequestInputToGRPCValidIngredientCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateValidIngredient(ctx, &mealplanningsvc.CreateValidIngredientRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestValidIngredients_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)

		retrieved, err := testClient.GetValidIngredient(ctx, &mealplanningsvc.GetValidIngredientRequest{ValidIngredientId: created.ID})
		assert.NoError(t, err)

		converted := grpcconverters.ConvertGRPCValidIngredientToValidIngredient(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidIngredient(ctx, &mealplanningsvc.GetValidIngredientRequest{ValidIngredientId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetValidIngredient(ctx, &mealplanningsvc.GetValidIngredientRequest{ValidIngredientId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestValidIngredients_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)

		updateInput := fakes.BuildFakeValidIngredientUpdateRequestInput()
		created.Update(updateInput)

		response, err := adminClient.UpdateValidIngredient(ctx, &mealplanningsvc.UpdateValidIngredientRequest{
			ValidIngredientId: created.ID,
			Input:             grpcconverters.ConvertValidIngredientUpdateRequestInputToGRPCValidIngredientUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := grpcconverters.ConvertGRPCValidIngredientToValidIngredient(response.Result)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)

		updateInput := fakes.BuildFakeValidIngredientUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidIngredient(ctx, &mealplanningsvc.UpdateValidIngredientRequest{
			ValidIngredientId: created.ID,
			Input:             grpcconverters.ConvertValidIngredientUpdateRequestInputToGRPCValidIngredientUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		/*
			there's no way to provide invalid input to this method, but
			I want to make it explicit that tests should be written the moment that changes
		*/
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		created := createValidIngredientForTest(t)

		response, err := testClient.UpdateValidIngredient(ctx, &mealplanningsvc.UpdateValidIngredientRequest{
			ValidIngredientId: created.ID,
			Input: &mealplanningsvc.ValidIngredientUpdateRequestInput{
				Name: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidIngredients_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)

		_, err := adminClient.ArchiveValidIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientRequest{ValidIngredientId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidIngredient(ctx, &mealplanningsvc.GetValidIngredientRequest{ValidIngredientId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientRequest{ValidIngredientId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientRequest{ValidIngredientId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientRequest{ValidIngredientId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidIngredients_GetRandom(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// in case we haven't already
		createValidIngredientForTest(t)

		response, err := testClient.GetRandomValidIngredient(ctx, &mealplanningsvc.GetRandomValidIngredientRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		response, err := c.GetRandomValidIngredient(ctx, &mealplanningsvc.GetRandomValidIngredientRequest{})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidIngredients_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdValidIngredients := []*mealplanning.ValidIngredient{}
	for range exampleQuantity {
		created := createValidIngredientForTest(T)
		createdValidIngredients = append(createdValidIngredients, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetValidIngredients(ctx, &mealplanningsvc.GetValidIngredientsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdValidIngredients))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidIngredients(ctx, &mealplanningsvc.GetValidIngredientsRequest{})
		assert.Error(t, err)
	})
}

func TestValidIngredients_Searching(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		created := createValidIngredientForTest(t)

		retrieved, err := testClient.SearchForValidIngredients(ctx, &mealplanningsvc.SearchForValidIngredientsRequest{
			Query: created.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidIngredients(ctx, &mealplanningsvc.SearchForValidIngredientsRequest{})
		assert.Error(t, err)
	})
}
