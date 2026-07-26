// Package mockmanagers provides mock implementations of the managers package's interfaces.
//
// Both the hand-written testify-based types and the moq-generated *Mock types
// live here during the testify -> moq migration. New test code should prefer
// the moq-generated types.
package mockmanagers

// Regenerate the moq mocks via `go generate ./internal/domain/mealplanning/managers/mock/`.

//go:generate go tool github.com/matryer/moq -out mealplanning_manager_mock.go -pkg mockmanagers -rm -fmt goimports .. MealPlanningManager:MealPlanningManagerMock
