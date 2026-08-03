package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanningconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createAccountInstrumentOwnershipForInstrumentForTest creates account instrument ownership for the given instrument.
// Use when the test needs ownership of a specific instrument (e.g. for SearchForRecipesWithInstrumentOwnership).
func createAccountInstrumentOwnershipForInstrumentForTest(t *testing.T, clientToUse client.Client, validInstrument *mealplanning.ValidInstrument) *mealplanning.AccountInstrumentOwnership {
	t.Helper()
	ctx := t.Context()

	exampleAccountInstrumentOwnership := fakes.BuildFakeAccountInstrumentOwnership()
	exampleAccountInstrumentOwnershipInput := converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput(exampleAccountInstrumentOwnership)
	exampleAccountInstrumentOwnershipInput.ValidInstrumentID = validInstrument.ID
	createdAccountInstrumentOwnership, err := clientToUse.CreateAccountInstrumentOwnership(ctx, &settingssvc.CreateAccountInstrumentOwnershipRequest{
		Input: mealplanningconverters.ConvertAccountInstrumentOwnershipCreationRequestInputToGRPCAccountInstrumentOwnershipCreationRequestInput(exampleAccountInstrumentOwnershipInput),
	})
	require.NoError(t, err)
	converted := mealplanningconverters.ConvertGRPCAccountInstrumentOwnershipToAccountInstrumentOwnership(createdAccountInstrumentOwnership.Created)
	return converted
}

func createAccountInstrumentOwnershipForTest(t *testing.T, clientToUse client.Client) *mealplanning.AccountInstrumentOwnership {
	t.Helper()
	ctx := t.Context()

	validInstrument := createValidInstrumentForTest(t)

	exampleAccountInstrumentOwnership := fakes.BuildFakeAccountInstrumentOwnership()
	exampleAccountInstrumentOwnershipInput := converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput(exampleAccountInstrumentOwnership)
	exampleAccountInstrumentOwnershipInput.ValidInstrumentID = validInstrument.ID
	createdAccountInstrumentOwnership, err := clientToUse.CreateAccountInstrumentOwnership(ctx, &settingssvc.CreateAccountInstrumentOwnershipRequest{
		Input: mealplanningconverters.ConvertAccountInstrumentOwnershipCreationRequestInputToGRPCAccountInstrumentOwnershipCreationRequestInput(exampleAccountInstrumentOwnershipInput),
	})
	require.NoError(t, err)
	converted := mealplanningconverters.ConvertGRPCAccountInstrumentOwnershipToAccountInstrumentOwnership(createdAccountInstrumentOwnership.Created)
	assertRoughEquality(t, exampleAccountInstrumentOwnership, converted, defaultIgnoredFields("ID", "BelongsToAccount", "Instrument")...)

	res, err := clientToUse.GetAccountInstrumentOwnership(ctx, &settingssvc.GetAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: createdAccountInstrumentOwnership.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, res)

	serviceSetting := mealplanningconverters.ConvertGRPCAccountInstrumentOwnershipToAccountInstrumentOwnership(res.Result)
	assertRoughEquality(t, converted, serviceSetting, defaultIgnoredFields("ID", "BelongsToAccount", "Instrument")...)

	return serviceSetting
}

func TestAccountInstrumentOwnerships_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createAccountInstrumentOwnershipForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_instrument_ownerships", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeAccountInstrumentOwnershipCreationRequestInput()
		convertedInput := mealplanningconverters.ConvertAccountInstrumentOwnershipCreationRequestInputToGRPCAccountInstrumentOwnershipCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateAccountInstrumentOwnership(ctx, &settingssvc.CreateAccountInstrumentOwnershipRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		creationRequestInput := fakes.BuildFakeAccountInstrumentOwnershipCreationRequestInput()
		convertedInput := mealplanningconverters.ConvertAccountInstrumentOwnershipCreationRequestInputToGRPCAccountInstrumentOwnershipCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.ValidInstrumentId = ""

		created, err := testClient.CreateAccountInstrumentOwnership(ctx, &settingssvc.CreateAccountInstrumentOwnershipRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestAccountInstrumentOwnerships_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createAccountInstrumentOwnershipForTest(t, testClient)

		retrieved, err := testClient.GetAccountInstrumentOwnership(ctx, &settingssvc.GetAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		converted := mealplanningconverters.ConvertGRPCAccountInstrumentOwnershipToAccountInstrumentOwnership(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields("ID", "BelongsToAccount", "Instrument")...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createAccountInstrumentOwnershipForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetAccountInstrumentOwnership(ctx, &settingssvc.GetAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := testClient.GetAccountInstrumentOwnership(ctx, &settingssvc.GetAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestAccountInstrumentOwnerships_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createAccountInstrumentOwnershipForTest(t, testClient)

		_, err := testClient.ArchiveAccountInstrumentOwnership(ctx, &settingssvc.ArchiveAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: created.ID})
		assert.NoError(t, err)

		x, err := testClient.GetAccountInstrumentOwnership(ctx, &settingssvc.GetAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_instrument_ownerships", RelevantID: created.ID},
			{EventType: "archived", ResourceType: "account_instrument_ownerships", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createAccountInstrumentOwnershipForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveAccountInstrumentOwnership(ctx, &settingssvc.ArchiveAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveAccountInstrumentOwnership(ctx, &settingssvc.ArchiveAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestAccountInstrumentOwnerships_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdAccountInstrumentOwnerships := []*mealplanning.AccountInstrumentOwnership{}
	for range exampleQuantity {
		created := createAccountInstrumentOwnershipForTest(T, testClient)
		createdAccountInstrumentOwnerships = append(createdAccountInstrumentOwnerships, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetAccountInstrumentOwnerships(ctx, &settingssvc.GetAccountInstrumentOwnershipsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdAccountInstrumentOwnerships))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetAccountInstrumentOwnerships(ctx, &settingssvc.GetAccountInstrumentOwnershipsRequest{})
		assert.Error(t, err)
	})
}

func TestAccountInstrumentOwnerships_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createAccountInstrumentOwnershipForTest(t, testClient)

		updateInput := fakes.BuildFakeAccountInstrumentOwnershipUpdateRequestInput()
		updateInput.ValidInstrumentID = &created.Instrument.ID
		created.Update(updateInput)

		_, err := testClient.UpdateAccountInstrumentOwnership(ctx, &settingssvc.UpdateAccountInstrumentOwnershipRequest{
			AccountInstrumentOwnershipId: created.ID,
			Input: &settingssvc.AccountInstrumentOwnershipUpdateRequestInput{
				Notes:             updateInput.Notes,
				Quantity:          new(uint32(*updateInput.Quantity)),
				ValidInstrumentId: updateInput.ValidInstrumentID,
			},
		})
		require.NoError(t, err)

		// Re-fetch to verify the update took effect
		res, err := testClient.GetAccountInstrumentOwnership(ctx, &settingssvc.GetAccountInstrumentOwnershipRequest{AccountInstrumentOwnershipId: created.ID})
		require.NoError(t, err)
		updated := mealplanningconverters.ConvertGRPCAccountInstrumentOwnershipToAccountInstrumentOwnership(res.Result)
		require.NotNil(t, updated.LastUpdatedAt)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createAccountInstrumentOwnershipForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateAccountInstrumentOwnership(ctx, &settingssvc.UpdateAccountInstrumentOwnershipRequest{
			AccountInstrumentOwnershipId: created.ID,
			Input: &settingssvc.AccountInstrumentOwnershipUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.UpdateAccountInstrumentOwnership(ctx, &settingssvc.UpdateAccountInstrumentOwnershipRequest{
			AccountInstrumentOwnershipId: nonexistentID,
			Input: &settingssvc.AccountInstrumentOwnershipUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
	})
}

func TestAccountInstrumentOwnerships_SearchForValidInstrumentsNotOwnedByAccount(T *testing.T) {
	T.Parallel()

	T.Run("excludes owned instruments", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		owned := createAccountInstrumentOwnershipForTest(t, testClient)

		res, err := testClient.SearchForValidInstrumentsNotOwnedByAccount(ctx, &settingssvc.SearchForValidInstrumentsNotOwnedByAccountRequest{
			Query: "",
		})
		require.NoError(t, err)
		require.NotNil(t, res)

		ownedInstrumentID := owned.Instrument.ID
		for _, result := range res.Results {
			assert.NotEqual(t, ownedInstrumentID, result.Id, "owned instrument should not appear in search results")
		}
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidInstrumentsNotOwnedByAccount(ctx, &settingssvc.SearchForValidInstrumentsNotOwnedByAccountRequest{})
		assert.Error(t, err)
	})
}
