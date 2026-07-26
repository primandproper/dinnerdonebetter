// Package identitymock provides mock implementations of the identity package's interfaces.
//
// Both the hand-written testify-based types and the moq-generated *Mock types
// live here during the testify -> moq migration. New test code should prefer
// the moq-generated types.
package identitymock

// Regenerate the moq mocks via `go generate ./internal/domain/identity/mock/`.

//go:generate go tool github.com/matryer/moq -out identity_mock.go -pkg identitymock -rm -fmt goimports .. Repository:RepositoryMock
