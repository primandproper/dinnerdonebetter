package integration

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/require"
)

// Finalizing a meal plan no longer builds its prep tasks and grocery list before the call
// returns. It starts a saga, and the saga worker — which init.go runs alongside the API server,
// standing in for the scheduler — advances it a moment later.
//
// So the assertions that used to read straight back poll instead. These are the bounds:
// generous, because the suite runs in parallel against one worker with a batch size and a
// concurrency, and a plan that lands behind a full batch waits for it.
const (
	finalizationPollTimeout  = 30 * time.Second
	finalizationPollInterval = 100 * time.Millisecond
)

// awaitMealPlanFinalized waits for the finalization saga's first step to tally a meal plan's
// ballots and mark it finalized, and returns the plan as it stands afterwards.
func awaitMealPlanFinalized(t *testing.T, ctx context.Context, c client.Client, mealPlanID string) *mealplanning.MealPlan {
	t.Helper()

	var mealPlan *mealplanning.MealPlan

	require.Eventually(t, func() bool {
		res, err := c.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: mealPlanID})
		if err != nil || res == nil || res.Result == nil {
			return false
		}

		converted := converters.ConvertGRPCMealPlanToMealPlan(res.Result)
		if converted.Status != string(mealplanning.MealPlanStatusFinalized) {
			return false
		}

		mealPlan = converted

		return true
	}, finalizationPollTimeout, finalizationPollInterval, "expected the finalization saga to finalize meal plan %s", mealPlanID)

	return mealPlan
}

// awaitMealPlanTasksCreated waits for the finalization saga's task step to finish and returns
// whatever tasks it produced, which may legitimately be none.
//
// It synchronizes on the flag rather than on the tasks, so a test asserting that some particular
// task was *not* generated waits for the step that would have generated it instead of passing
// vacuously against a list the saga had not filled in yet.
func awaitMealPlanTasksCreated(t *testing.T, ctx context.Context, c client.Client, mealPlanID string) []*mealplanninggrpc.MealPlanTask {
	t.Helper()

	require.Eventually(t, func() bool {
		res, err := c.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: mealPlanID})
		if err != nil || res == nil || res.Result == nil {
			return false
		}

		return converters.ConvertGRPCMealPlanToMealPlan(res.Result).TasksCreated
	}, finalizationPollTimeout, finalizationPollInterval, "expected the finalization saga to run the task step for meal plan %s", mealPlanID)

	res, err := c.GetMealPlanTasks(ctx, &mealplanninggrpc.GetMealPlanTasksRequest{MealPlanId: mealPlanID})
	require.NoError(t, err)
	require.NotNil(t, res)

	return res.Results
}

// awaitMealPlanTasks waits for the finalization saga to build a meal plan's prep tasks and
// returns them.
func awaitMealPlanTasks(t *testing.T, ctx context.Context, c client.Client, mealPlanID string) []*mealplanninggrpc.MealPlanTask {
	t.Helper()

	var tasks []*mealplanninggrpc.MealPlanTask

	require.Eventually(t, func() bool {
		res, err := c.GetMealPlanTasks(ctx, &mealplanninggrpc.GetMealPlanTasksRequest{MealPlanId: mealPlanID})
		if err != nil || res == nil || len(res.Results) == 0 {
			return false
		}

		tasks = res.Results

		return true
	}, finalizationPollTimeout, finalizationPollInterval, "expected the finalization saga to create prep tasks for meal plan %s", mealPlanID)

	return tasks
}

// awaitMealPlanGroceryListItems waits for the finalization saga to build a meal plan's grocery
// list and returns its items.
func awaitMealPlanGroceryListItems(t *testing.T, ctx context.Context, c client.Client, mealPlanID string) []*mealplanninggrpc.MealPlanGroceryListItem {
	t.Helper()

	var items []*mealplanninggrpc.MealPlanGroceryListItem

	require.Eventually(t, func() bool {
		res, err := c.GetMealPlanGroceryListItemsForMealPlan(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemsForMealPlanRequest{MealPlanId: mealPlanID})
		if err != nil || res == nil || len(res.Results) == 0 {
			return false
		}

		items = res.Results

		return true
	}, finalizationPollTimeout, finalizationPollInterval, "expected the finalization saga to initialize a grocery list for meal plan %s", mealPlanID)

	return items
}
