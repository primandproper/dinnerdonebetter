// Package mock provides mock implementations of the internalops package's interfaces.
package mock

// Regenerate the moq mocks via `go generate ./internal/domain/internalops/mock/`.

//go:generate go tool github.com/matryer/moq -out internalops_mock.go -pkg mock -rm -fmt goimports .. InternalOpsDataManager:InternalOpsDataManagerMock
