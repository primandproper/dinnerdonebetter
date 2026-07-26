// Package oauthmock provides mock implementations of the oauth package's interfaces.
//
// Both the hand-written testify-based types and the moq-generated *Mock types
// live here during the testify -> moq migration. New test code should prefer
// the moq-generated types.
package oauthmock

// Regenerate the moq mocks via `go generate ./internal/domain/oauth/mock/`.

//go:generate go tool github.com/matryer/moq -out oauth_mock.go -pkg oauthmock -rm -fmt goimports .. Repository:RepositoryMock
