// Package mock provides mock implementations of the auth package's interfaces.
package mock

// Regenerate the moq mocks via `go generate ./internal/domain/auth/mock/`.

//go:generate go tool github.com/matryer/moq -out auth_mock.go -pkg mock -rm -fmt goimports .. PasswordResetTokenDataManager:PasswordResetTokenDataManagerMock UserSessionDataManager:UserSessionDataManagerMock
