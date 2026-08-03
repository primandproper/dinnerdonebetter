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

func checkValidIngredientStateEquality(t *testing.T, expected, actual *mealplanning.ValidIngredientState) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected ValidIngredientState to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected ValidIngredientState to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected ValidIngredientState Name")
	assert.Equal(t, expected.Description, actual.Description, "expected ValidIngredientState Description")
	assert.Equal(t, expected.Slug, actual.Slug, "expected ValidIngredientState Slug")
	assert.Equal(t, expected.PastTense, actual.PastTense, "expected ValidIngredientState PastTense")
	assert.Equal(t, expected.IconPath, actual.IconPath, "expected ValidIngredientState IconPath")
	assert.Equal(t, expected.AttributeType, actual.AttributeType, "expected ValidIngredientState AttributeType")
}

func createValidIngredientStateForTest(t *testing.T) *mealplanning.ValidIngredientState {
	t.Helper()

	ctx := t.Context()

	exampleValidIngredientState := fakes.BuildFakeValidIngredientState()
	exampleValidIngredientStateInput := converters.ConvertValidIngredientStateToValidIngredientStateCreationRequestInput(exampleValidIngredientState)
	created, err := adminClient.CreateValidIngredientState(ctx, &mealplanningsvc.CreateValidIngredientStateRequest{
		Input: grpcconverters.ConvertValidIngredientStateCreationRequestInputToGRPCValidIngredientStateCreationRequestInput(exampleValidIngredientStateInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCValidIngredientStateToValidIngredientState(created.Result)
	checkValidIngredientStateEquality(t, exampleValidIngredientState, converted)

	retrieved, err := adminClient.GetValidIngredientState(ctx, &mealplanningsvc.GetValidIngredientStateRequest{
		ValidIngredientStateId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	validIngredientState := grpcconverters.ConvertGRPCValidIngredientStateToValidIngredientState(retrieved.Result)
	checkValidIngredientStateEquality(t, converted, validIngredientState)

	return validIngredientState
}

func TestValidIngredientStates_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidIngredientStateForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidIngredientStateCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidIngredientStateCreationRequestInputToGRPCValidIngredientStateCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateValidIngredientState(ctx, &mealplanningsvc.CreateValidIngredientStateRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidIngredientStateCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidIngredientStateCreationRequestInputToGRPCValidIngredientStateCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateValidIngredientState(ctx, &mealplanningsvc.CreateValidIngredientStateRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeValidIngredientStateCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidIngredientStateCreationRequestInputToGRPCValidIngredientStateCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateValidIngredientState(ctx, &mealplanningsvc.CreateValidIngredientStateRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestValidIngredientStates_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)

		retrieved, err := testClient.GetValidIngredientState(ctx, &mealplanningsvc.GetValidIngredientStateRequest{ValidIngredientStateId: created.ID})
		assert.NoError(t, err)

		converted := grpcconverters.ConvertGRPCValidIngredientStateToValidIngredientState(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidIngredientState(ctx, &mealplanningsvc.GetValidIngredientStateRequest{ValidIngredientStateId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetValidIngredientState(ctx, &mealplanningsvc.GetValidIngredientStateRequest{ValidIngredientStateId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestValidIngredientStates_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)

		updateInput := fakes.BuildFakeValidIngredientStateUpdateRequestInput()
		created.Update(updateInput)

		response, err := adminClient.UpdateValidIngredientState(ctx, &mealplanningsvc.UpdateValidIngredientStateRequest{
			ValidIngredientStateId: created.ID,
			Input:                  grpcconverters.ConvertValidIngredientStateUpdateRequestInputToGRPCValidIngredientStateUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := grpcconverters.ConvertGRPCValidIngredientStateToValidIngredientState(response.Result)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)

		updateInput := fakes.BuildFakeValidIngredientStateUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidIngredientState(ctx, &mealplanningsvc.UpdateValidIngredientStateRequest{
			ValidIngredientStateId: created.ID,
			Input:                  grpcconverters.ConvertValidIngredientStateUpdateRequestInputToGRPCValidIngredientStateUpdateRequestInput(updateInput),
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

		created := createValidIngredientStateForTest(t)

		response, err := testClient.UpdateValidIngredientState(ctx, &mealplanningsvc.UpdateValidIngredientStateRequest{
			ValidIngredientStateId: created.ID,
			Input: &mealplanningsvc.ValidIngredientStateUpdateRequestInput{
				Name: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidIngredientStates_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)

		_, err := adminClient.ArchiveValidIngredientState(ctx, &mealplanningsvc.ArchiveValidIngredientStateRequest{ValidIngredientStateId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidIngredientState(ctx, &mealplanningsvc.GetValidIngredientStateRequest{ValidIngredientStateId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidIngredientState(ctx, &mealplanningsvc.ArchiveValidIngredientStateRequest{ValidIngredientStateId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidIngredientState(ctx, &mealplanningsvc.ArchiveValidIngredientStateRequest{ValidIngredientStateId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidIngredientStateForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidIngredientState(ctx, &mealplanningsvc.ArchiveValidIngredientStateRequest{ValidIngredientStateId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidIngredientStates_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdValidIngredientStates := []*mealplanning.ValidIngredientState{}
	for range exampleQuantity {
		created := createValidIngredientStateForTest(T)
		createdValidIngredientStates = append(createdValidIngredientStates, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetValidIngredientStates(ctx, &mealplanningsvc.GetValidIngredientStatesRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdValidIngredientStates))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidIngredientStates(ctx, &mealplanningsvc.GetValidIngredientStatesRequest{})
		assert.Error(t, err)
	})
}

func TestValidIngredientStates_Searching(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		created := createValidIngredientStateForTest(t)

		retrieved, err := testClient.SearchForValidIngredientStates(ctx, &mealplanningsvc.SearchForValidIngredientStatesRequest{
			Query: created.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidIngredientStates(ctx, &mealplanningsvc.SearchForValidIngredientStatesRequest{})
		assert.Error(t, err)
	})
}
