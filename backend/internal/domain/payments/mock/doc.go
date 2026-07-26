// Package mock provides mock implementations of the payments package's interfaces.
//
// Both the hand-written testify-based types and the moq-generated *Mock types
// live here during the testify -> moq migration. New test code should prefer
// the moq-generated types.
package mock

// Regenerate the moq mocks via `go generate ./internal/domain/payments/mock/`.

//go:generate go tool github.com/matryer/moq -out payments_mock.go -pkg mock -rm -fmt goimports .. Repository:RepositoryMock
