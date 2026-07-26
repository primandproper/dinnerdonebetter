// Package authentication provides the login, token exchange and session management manager.
package authentication

// AuthenticatorMock is generated in-package because the manager's own tests need it and
// internal/authentication/mock imports this package.

//go:generate go tool github.com/matryer/moq -out authenticator_mock.go -pkg authentication -rm -fmt goimports . Authenticator:AuthenticatorMock
