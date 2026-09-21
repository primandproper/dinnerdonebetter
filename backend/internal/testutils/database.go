package testutils

import (
	"context"
	"time"

	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
)

// MockDatabaseClient is a database.Client for a test that has mocked the store beneath it.
//
// As of platform-go v14 a store read takes the caller's database.SQLQueryExecutor and a store
// write takes its database.Tx, so anything holding a store also holds a database.Client to get
// them from. A unit test that mocks the store still has to supply one, and this is it: Reader
// and Writer answer with an executor nothing sends statements to, and WithTransaction runs the
// callback on a transaction over that same executor rather than opening anything.
//
// The transaction is a real database.Tx — database.NewTxForTesting exists for exactly this, the
// marker method being unexported — so a store that refuses a nil executor is satisfied. Nothing
// is ever sent on it, because the store above it is mocked. A test that wants to watch a write
// land wants a real database instead.
func MockDatabaseClient() *mockdatabase.ClientMock {
	executor := &mockdatabase.SQLQueryExecutorMock{}

	return &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return executor },
		WriterFunc: func() database.SQLQueryExecutor { return executor },
		WithTransactionFunc: func(ctx context.Context, fn func(querier database.Tx) error) error {
			return fn(database.NewTxForTesting(executor))
		},

		// The wall clock, because this is the clock this application stamps through and
		// a deadline it computes has to be comparable with one platform's stores compute
		// from theirs. A test that needs a known distance between two instants replaces
		// this on its own client rather than moving everybody's.
		CurrentTimeFunc: time.Now,
	}
}
