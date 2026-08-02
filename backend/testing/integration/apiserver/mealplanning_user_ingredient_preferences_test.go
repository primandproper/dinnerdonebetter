package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	settingsconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserIngredientPreferenceForTest(t *testing.T, clientToUse client.Client) *mealplanning.UserIngredientPreference {
	t.Helper()
	ctx := t.Context()

	validIngredient := createValidIngredientForTest(t)

	exampleUserIngredientPreference := fakes.BuildFakeUserIngredientPreference()
	exampleUserIngredientPreferenceInput := converters.ConvertUserIngredientPreferenceToUserIngredientPreferenceCreationRequestInput(exampleUserIngredientPreference)
	exampleUserIngredientPreferenceInput.ValidIngredientID = validIngredient.ID
	createdUserIngredientPreference, err := clientToUse.CreateUserIngredientPreference(ctx, &settingssvc.CreateUserIngredientPreferenceRequest{
		Input: settingsconverters.ConvertUserIngredientPreferenceCreationRequestInputToGRPCUserIngredientPreferenceCreationRequestInput(exampleUserIngredientPreferenceInput),
	})
	require.NoError(t, err)
	converted := settingsconverters.ConvertGRPCUserIngredientPreferenceToUserIngredientPreference(createdUserIngredientPreference.Created[0])
	assertRoughEquality(t, exampleUserIngredientPreference, converted, defaultIgnoredFields("ID", "CreatedByUser", "Ingredient")...)

	res, err := clientToUse.GetUserIngredientPreference(ctx, &settingssvc.GetUserIngredientPreferenceRequest{UserIngredientPreferenceId: createdUserIngredientPreference.Created[0].Id})
	require.NoError(t, err)
	require.NotNil(t, res)

	serviceSetting := settingsconverters.ConvertGRPCUserIngredientPreferenceToUserIngredientPreference(res.Result)
	assertRoughEquality(t, converted, serviceSetting, defaultIgnoredFields("ID", "CreatedByUser", "Ingredient")...)

	return serviceSetting
}

func TestUserIngredientPreferences_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		createUserIngredientPreferenceForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeUserIngredientPreferenceCreationRequestInput()
		convertedInput := settingsconverters.ConvertUserIngredientPreferenceCreationRequestInputToGRPCUserIngredientPreferenceCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateUserIngredientPreference(ctx, &settingssvc.CreateUserIngredientPreferenceRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		creationRequestInput := fakes.BuildFakeUserIngredientPreferenceCreationRequestInput()
		convertedInput := settingsconverters.ConvertUserIngredientPreferenceCreationRequestInputToGRPCUserIngredientPreferenceCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.ValidIngredientId = ""
		convertedInput.ValidIngredientGroupId = ""

		created, err := testClient.CreateUserIngredientPreference(ctx, &settingssvc.CreateUserIngredientPreferenceRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeUserIngredientPreferenceCreationRequestInput()
		convertedInput := settingsconverters.ConvertUserIngredientPreferenceCreationRequestInputToGRPCUserIngredientPreferenceCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateUserIngredientPreference(ctx, &settingssvc.CreateUserIngredientPreferenceRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestUserIngredientPreferences_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createUserIngredientPreferenceForTest(t, testClient)

		retrieved, err := testClient.GetUserIngredientPreference(ctx, &settingssvc.GetUserIngredientPreferenceRequest{UserIngredientPreferenceId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		converted := settingsconverters.ConvertGRPCUserIngredientPreferenceToUserIngredientPreference(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields("ID", "CreatedByUser", "Ingredient")...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createUserIngredientPreferenceForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetUserIngredientPreference(ctx, &settingssvc.GetUserIngredientPreferenceRequest{UserIngredientPreferenceId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := testClient.GetUserIngredientPreference(ctx, &settingssvc.GetUserIngredientPreferenceRequest{UserIngredientPreferenceId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestUserIngredientPreferences_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createUserIngredientPreferenceForTest(t, testClient)

		_, err := testClient.ArchiveUserIngredientPreference(ctx, &settingssvc.ArchiveUserIngredientPreferenceRequest{UserIngredientPreferenceId: created.ID})
		assert.NoError(t, err)

		x, err := testClient.GetUserIngredientPreference(ctx, &settingssvc.GetUserIngredientPreferenceRequest{UserIngredientPreferenceId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createUserIngredientPreferenceForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveUserIngredientPreference(ctx, &settingssvc.ArchiveUserIngredientPreferenceRequest{UserIngredientPreferenceId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveUserIngredientPreference(ctx, &settingssvc.ArchiveUserIngredientPreferenceRequest{UserIngredientPreferenceId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestUserIngredientPreferences_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createUserIngredientPreferenceForTest(t, testClient)

		updateInput := fakes.BuildFakeUserIngredientPreferenceUpdateRequestInput()
		created.Update(updateInput)

		response, err := testClient.UpdateUserIngredientPreference(ctx, &settingssvc.UpdateUserIngredientPreferenceRequest{
			UserIngredientPreferenceId: created.ID,
			Input: &settingssvc.UserIngredientPreferenceUpdateRequestInput{
				Notes:   updateInput.Notes,
				Rating:  new(int32(*updateInput.Rating)),
				Allergy: updateInput.Allergy,
			},
		})
		assert.NoError(t, err)

		updated := settingsconverters.ConvertGRPCUserIngredientPreferenceToUserIngredientPreference(response.Updated)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields("ID", "CreatedByUser", "Ingredient")...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createUserIngredientPreferenceForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateUserIngredientPreference(ctx, &settingssvc.UpdateUserIngredientPreferenceRequest{
			UserIngredientPreferenceId: created.ID,
			Input: &settingssvc.UserIngredientPreferenceUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.UpdateUserIngredientPreference(ctx, &settingssvc.UpdateUserIngredientPreferenceRequest{
			UserIngredientPreferenceId: nonexistentID,
			Input: &settingssvc.UserIngredientPreferenceUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
	})
}

func TestUserIngredientPreferences_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdUserIngredientPreferences := []*mealplanning.UserIngredientPreference{}
	for range exampleQuantity {
		created := createUserIngredientPreferenceForTest(T, testClient)
		createdUserIngredientPreferences = append(createdUserIngredientPreferences, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetUserIngredientPreferences(ctx, &settingssvc.GetUserIngredientPreferencesRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.True(t, len(retrieved.Results) >= len(createdUserIngredientPreferences))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetUserIngredientPreferences(ctx, &settingssvc.GetUserIngredientPreferencesRequest{})
		assert.Error(t, err)
	})
}
