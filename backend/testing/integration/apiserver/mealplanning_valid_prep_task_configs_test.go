package integration

import (
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanningconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidPrepTaskConfigForTest(t *testing.T) (*types.ValidIngredient, *types.ValidPreparation, *types.ValidPrepTaskConfig) {
	t.Helper()
	ctx := t.Context()

	createdValidIngredient := createValidIngredientForTest(t)
	createdValidPreparation := createValidPreparationForTest(t)

	exampleValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
	exampleValidPrepTaskConfig.Preparation = *createdValidPreparation
	exampleValidPrepTaskConfig.Ingredient = *createdValidIngredient

	exampleValidPrepTaskConfigInput := mealplanningconverters.ConvertValidPrepTaskConfigCreationRequestInputToGRPCValidPrepTaskConfigCreationRequestInput(converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput(exampleValidPrepTaskConfig))
	createdValidPrepTaskConfig, err := adminClient.CreateValidPrepTaskConfig(ctx, &mealplanningsvc.CreateValidPrepTaskConfigRequest{Input: exampleValidPrepTaskConfigInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidPrepTaskConfig)

	validPrepTaskConfigRes, err := adminClient.GetValidPrepTaskConfig(ctx, &mealplanningsvc.GetValidPrepTaskConfigRequest{
		ValidPrepTaskConfigId: createdValidPrepTaskConfig.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepTaskConfigRes.Result)

	return createdValidIngredient, createdValidPreparation, mealplanningconverters.ConvertGRPCValidPrepTaskConfigToValidPrepTaskConfig(validPrepTaskConfigRes.Result)
}

func TestValidPrepTaskConfigs_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidPrepTaskConfigForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
		exampleValidPrepTaskConfigInput := mealplanningconverters.ConvertValidPrepTaskConfigCreationRequestInputToGRPCValidPrepTaskConfigCreationRequestInput(converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput(exampleValidPrepTaskConfig))
		exampleValidPrepTaskConfigInput.ValidPreparationId = ""
		exampleValidPrepTaskConfigInput.ValidIngredientId = ""

		createdValidPrepTaskConfig, err := adminClient.CreateValidPrepTaskConfig(ctx, &mealplanningsvc.CreateValidPrepTaskConfigRequest{Input: exampleValidPrepTaskConfigInput})
		require.Error(t, err)
		require.Nil(t, createdValidPrepTaskConfig)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidPrepTaskConfig(ctx, &mealplanningsvc.CreateValidPrepTaskConfigRequest{})
		assert.Error(t, err)
	})
}

func TestValidPrepTaskConfigs_Listing(T *testing.T) {
	T.Parallel()

	createdValidPrepTaskConfigs := []*types.ValidPrepTaskConfig{}
	validIngredient, validPreparation, created := createValidPrepTaskConfigForTest(T)
	createdValidPrepTaskConfigs = append(createdValidPrepTaskConfigs, created)
	for range exampleQuantity - 1 {
		exampleValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
		exampleValidPrepTaskConfigInput := mealplanningconverters.ConvertValidPrepTaskConfigCreationRequestInputToGRPCValidPrepTaskConfigCreationRequestInput(converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput(exampleValidPrepTaskConfig))
		exampleValidPrepTaskConfigInput.ValidPreparationId = validPreparation.ID
		exampleValidPrepTaskConfigInput.ValidIngredientId = validIngredient.ID

		createdValidPrepTaskConfig, err := adminClient.CreateValidPrepTaskConfig(T.Context(), &mealplanningsvc.CreateValidPrepTaskConfigRequest{Input: exampleValidPrepTaskConfigInput})
		require.NoError(T, err)
		require.NotNil(T, createdValidPrepTaskConfig)

		createdValidPrepTaskConfigs = append(createdValidPrepTaskConfigs, mealplanningconverters.ConvertGRPCValidPrepTaskConfigToValidPrepTaskConfig(createdValidPrepTaskConfig.Result))
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPrepTaskConfigs(ctx, &mealplanningsvc.GetValidPrepTaskConfigsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidPrepTaskConfigs))
	})

	T.Run("by preparation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPrepTaskConfigsByPreparation(ctx, &mealplanningsvc.GetValidPrepTaskConfigsByPreparationRequest{ValidPreparationId: validPreparation.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidPrepTaskConfigs))
	})

	T.Run("by ingredient", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPrepTaskConfigsByIngredient(ctx, &mealplanningsvc.GetValidPrepTaskConfigsByIngredientRequest{ValidIngredientId: validIngredient.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidPrepTaskConfigs))
	})

	T.Run("by ingredient and preparation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPrepTaskConfigsByIngredientAndPreparation(ctx, &mealplanningsvc.GetValidPrepTaskConfigsByIngredientAndPreparationRequest{
			ValidIngredientId:  validIngredient.ID,
			ValidPreparationId: validPreparation.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidPrepTaskConfigs))
	})
}

func TestValidPrepTaskConfigs_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPrepTaskConfigForTest(t)

		result, err := adminClient.GetValidPrepTaskConfig(ctx, &mealplanningsvc.GetValidPrepTaskConfigRequest{ValidPrepTaskConfigId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, created.ID, result.Result.Id)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidPrepTaskConfig(ctx, &mealplanningsvc.GetValidPrepTaskConfigRequest{})
		assert.Error(t, err)
	})
}

func TestValidPrepTaskConfigs_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPrepTaskConfigForTest(t)

		updateInput := mealplanningconverters.ConvertValidPrepTaskConfigUpdateRequestInputToGRPCValidPrepTaskConfigUpdateRequestInput(converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigUpdateRequestInput(created))
		newNotes := "updated notes"
		updateInput.Notes = &newNotes

		updated, err := adminClient.UpdateValidPrepTaskConfig(ctx, &mealplanningsvc.UpdateValidPrepTaskConfigRequest{
			ValidPrepTaskConfigId: created.ID,
			Input:                 updateInput,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, newNotes, updated.Result.Notes)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidPrepTaskConfig(ctx, &mealplanningsvc.UpdateValidPrepTaskConfigRequest{})
		assert.Error(t, err)
	})
}

func TestValidPrepTaskConfigs_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPrepTaskConfigForTest(t)

		_, err := adminClient.ArchiveValidPrepTaskConfig(ctx, &mealplanningsvc.ArchiveValidPrepTaskConfigRequest{ValidPrepTaskConfigId: created.ID})
		require.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidPrepTaskConfig(ctx, &mealplanningsvc.ArchiveValidPrepTaskConfigRequest{})
		assert.Error(t, err)
	})
}
