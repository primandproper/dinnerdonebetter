package integration

import (
	"testing"
	"time"

	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering/filteringpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// archivedListingFilter asks for archived rows created around createdAt, so that the rows other
// tests create while this one runs cannot push the row under test off the page.
func archivedListingFilter(createdAt time.Time, includeArchived bool) *filteringpb.QueryFilter {
	return &filteringpb.QueryFilter{
		CreatedAfter:    timestamppb.New(createdAt.Add(-time.Second)),
		CreatedBefore:   timestamppb.New(createdAt.Add(time.Second)),
		MaxResponseSize: new(uint32(100)),
		IncludeArchived: new(includeArchived),
	}
}

func idsOf[T interface{ GetId() string }](results []T) []string {
	ids := make([]string, 0, len(results))
	for _, result := range results {
		ids = append(ids, result.GetId())
	}

	return ids
}

// TestListSurfaces_IncludeArchived holds one surface of each archive decision to it, against real
// archived rows. internal/services/mealplanning/grpc's TestListRPCArchiveDecisions holds every
// surface to its decision; this is the round trip that shows a decision is what reaches the store.
func TestListSurfaces_IncludeArchived(T *testing.T) {
	T.Parallel()

	T.Run("a member asking for archived catalog entries gets none, and an operator gets them", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, memberClient := createUserAndClientForTest(t)
		created := createValidInstrumentForTest(t)

		_, err := adminClient.ArchiveValidInstrument(ctx, &mealplanninggrpc.ArchiveValidInstrumentRequest{ValidInstrumentId: created.ID})
		require.NoError(t, err)

		asMember, err := memberClient.GetValidInstruments(ctx, &mealplanninggrpc.GetValidInstrumentsRequest{Filter: archivedListingFilter(created.CreatedAt, true)})
		require.NoError(t, err)
		assert.NotContains(t, idsOf(asMember.Results), created.ID)

		asOperator, err := adminClient.GetValidInstruments(ctx, &mealplanninggrpc.GetValidInstrumentsRequest{Filter: archivedListingFilter(created.CreatedAt, true)})
		require.NoError(t, err)
		assert.Contains(t, idsOf(asOperator.Results), created.ID)
	})

	T.Run("a recipe's own author asking for archived recipes gets none", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// adminClient authored the recipe and holds every grant there is, so this is the most
		// privileged caller the surface has.
		_, _, created := createRecipeForTest(t, nil)

		before, err := adminClient.GetRecipes(ctx, &mealplanninggrpc.GetRecipesRequest{Status: created.Status, Filter: archivedListingFilter(created.CreatedAt, false)})
		require.NoError(t, err)
		require.Contains(t, idsOf(before.Results), created.ID, "the recipe must be on the page before it is archived, or its absence after proves nothing")

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: created.ID})
		require.NoError(t, err)

		after, err := adminClient.GetRecipes(ctx, &mealplanninggrpc.GetRecipesRequest{Status: created.Status, Filter: archivedListingFilter(created.CreatedAt, true)})
		require.NoError(t, err)
		assert.NotContains(t, idsOf(after.Results), created.ID)
	})

	T.Run("an account admin asking for the account's archived meal plans gets them", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountAdminClient := createUserAndClientForTest(t)
		created := createMealPlanForTest(t, accountAdminClient, nil)

		_, err := accountAdminClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: created.ID})
		require.NoError(t, err)

		withoutArchived, err := accountAdminClient.GetMealPlansForAccount(ctx, &mealplanninggrpc.GetMealPlansForAccountRequest{Filter: archivedListingFilter(created.CreatedAt, false)})
		require.NoError(t, err)
		assert.NotContains(t, idsOf(withoutArchived.Results), created.ID)

		withArchived, err := accountAdminClient.GetMealPlansForAccount(ctx, &mealplanninggrpc.GetMealPlansForAccountRequest{Filter: archivedListingFilter(created.CreatedAt, true)})
		require.NoError(t, err)
		assert.Contains(t, idsOf(withArchived.Results), created.ID)
	})
}
