// Package mockauthn provides mock implementations of the authentication package's interfaces.
//
// Both the hand-written testify-based types and the moq-generated *Mock types
// live here during the testify -> moq migration. New test code should prefer
// the moq-generated types.
package mockauthn

// Regenerate the moq mocks via `go generate ./internal/authentication/mock/`.

//go:generate go tool github.com/matryer/moq -out authentication_mock.go -pkg mockauthn -rm -fmt goimports .. Manager:ManagerMock Authenticator:AuthenticatorMock
