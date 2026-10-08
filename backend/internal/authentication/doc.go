// Package authentication provides the login, token exchange and session management manager.
package authentication

// AuthenticatorMock is generated in-package, where the tests that fake an Authenticator import it from.

//go:generate go tool github.com/matryer/moq -out authenticator_mock.go -pkg authentication -rm -fmt goimports . Authenticator:AuthenticatorMock
