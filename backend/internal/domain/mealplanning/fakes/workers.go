package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/fake"
)

// BuildFakeFinalizeMealPlansRequest builds a faked FinalizeMealPlansRequest.
func BuildFakeFinalizeMealPlansRequest() *mealplanning.FinalizeMealPlansRequest {
	return fake.BuildFakeRecord[mealplanning.FinalizeMealPlansRequest]()
}

// BuildFakeFinalizeMealPlansResponse builds a faked FinalizeMealPlansResponse.
func BuildFakeFinalizeMealPlansResponse() *mealplanning.FinalizeMealPlansResponse {
	return fake.BuildFakeRecord[mealplanning.FinalizeMealPlansResponse]()
}

// BuildFakeCreateMealPlanTasksRequest builds a faked CreateMealPlanTasksRequest.
func BuildFakeCreateMealPlanTasksRequest() *mealplanning.CreateMealPlanTasksRequest {
	return fake.BuildFakeRecord[mealplanning.CreateMealPlanTasksRequest]()
}

// BuildFakeCreateMealPlanTasksResponse builds a faked CreateMealPlanTasksResponse.
func BuildFakeCreateMealPlanTasksResponse() *mealplanning.CreateMealPlanTasksResponse {
	// The response a worker that did its job returns. Faker's coin flip would make half
	// of these the failure case, silently.
	return &mealplanning.CreateMealPlanTasksResponse{
		Success: true,
	}
}

// BuildFakeInitializeMealPlanGroceryListRequest builds a faked InitializeMealPlanGroceryListRequest.
func BuildFakeInitializeMealPlanGroceryListRequest() *mealplanning.InitializeMealPlanGroceryListRequest {
	return fake.BuildFakeRecord[mealplanning.InitializeMealPlanGroceryListRequest]()
}

// BuildFakeInitializeMealPlanGroceryListResponse builds a faked InitializeMealPlanGroceryListResponse.
func BuildFakeInitializeMealPlanGroceryListResponse() *mealplanning.InitializeMealPlanGroceryListResponse {
	return &mealplanning.InitializeMealPlanGroceryListResponse{
		Success: true,
	}
}
