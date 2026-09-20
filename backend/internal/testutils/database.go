package testutils

import (
	"context"

	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
)

// MockDatabaseClient is a database.Client for a test that has mocked the store beneath it.
//
// As of platform-go v14 a store read takes the caller's database.SQLQueryExecutor and a store
// write takes its database.Tx, so anything holding a store also holds a database.Client to get
// them from. A unit test that mocks the store still has to supply one, and this is it: Reader
// and Writer answer with an executor nothing sends statements to, and WithTransaction runs the
// callback rather than opening anything.
//
// The transaction the callback receives is nil, and there is no alternative. database.Tx carries
// an unexported method so that only the database package can implement one, and primitives-go
// ships no mock of it, which leaves a real Client.WithTransaction as the only source of a Tx in
// the language. That is fine here precisely because the store is mocked: the executor never
// reaches anything that would send a statement on it. It is not fine for a test that wants to
// see a write land, which is why every such test in this repository runs against a real
// database.
func MockDatabaseClient() *mockdatabase.ClientMock {
	executor := &mockdatabase.SQLQueryExecutorMock{}

	return &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return executor },
		WriterFunc: func() database.SQLQueryExecutor { return executor },
		WithTransactionFunc: func(ctx context.Context, fn func(querier database.Tx) error) error {
			return fn(nil)
		},
	}
}
