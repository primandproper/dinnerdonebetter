package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeFinalizeMealPlansRequest builds a faked Webhook.
func BuildFakeFinalizeMealPlansRequest() *mealplanning.FinalizeMealPlansRequest {
	return &mealplanning.FinalizeMealPlansRequest{
		ReturnCount: fake.Bool(),
	}
}

// BuildFakeFinalizeMealPlansResponse builds a faked FinalizeMealPlansResponse.
func BuildFakeFinalizeMealPlansResponse() *mealplanning.FinalizeMealPlansResponse {
	return &mealplanning.FinalizeMealPlansResponse{
		Count: int64(buildFakeNumber()),
	}
}

// BuildFakeCreateMealPlanTasksRequest builds a faked Webhook.
func BuildFakeCreateMealPlanTasksRequest() *mealplanning.CreateMealPlanTasksRequest {
	return &mealplanning.CreateMealPlanTasksRequest{
		AccountID: BuildFakeID(),
	}
}

// BuildFakeCreateMealPlanTasksResponse builds a faked Webhook.
func BuildFakeCreateMealPlanTasksResponse() *mealplanning.CreateMealPlanTasksResponse {
	return &mealplanning.CreateMealPlanTasksResponse{
		Success: true,
	}
}

// BuildFakeInitializeMealPlanGroceryListRequest builds a faked FinalizeMealPlansResponse.
func BuildFakeInitializeMealPlanGroceryListRequest() *mealplanning.InitializeMealPlanGroceryListRequest {
	return &mealplanning.InitializeMealPlanGroceryListRequest{
		AccountID: BuildFakeID(),
	}
}

// BuildFakeInitializeMealPlanGroceryListResponse builds a faked FinalizeMealPlansResponse.
func BuildFakeInitializeMealPlanGroceryListResponse() *mealplanning.InitializeMealPlanGroceryListResponse {
	return &mealplanning.InitializeMealPlanGroceryListResponse{
		Success: true,
	}
}
