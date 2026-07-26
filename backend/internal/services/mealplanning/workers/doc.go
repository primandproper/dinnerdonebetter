// Package workers defines the background worker interfaces for meal planning and their mocks.
package workers

// Regenerate the moq mocks via `go generate ./internal/services/mealplanning/workers/`.

//go:generate go tool github.com/matryer/moq -out worker_mock.go -pkg workers -rm -fmt goimports . Worker:WorkerMock WorkerCounter:WorkerCounterMock
