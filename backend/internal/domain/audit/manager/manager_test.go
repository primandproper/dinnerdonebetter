package manager

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/fakes"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"

	"github.com/primandproper/platform-go/v10/database"
	mockdatabase "github.com/primandproper/platform-go/v10/database/mock"
	"github.com/primandproper/platform-go/v10/filtering"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildAuditManagerForTest builds a manager backed by the given repository mock. A nil repo gets an
// unconfigured mock, which panics if any of its methods are called.
func buildAuditManagerForTest(t *testing.T, repo *auditmock.RepositoryMock) *auditManager {
	t.Helper()

	if repo == nil {
		repo = &auditmock.RepositoryMock{}
	}

	m := NewAuditDataManager(tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)

	return m.(*auditManager)
}

func TestAuditDataManager_GetAuditLogEntry(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := fakes.BuildFakeAuditLogEntry()

		repo := &auditmock.RepositoryMock{
			GetAuditLogEntryFunc: func(_ context.Context, auditLogID string) (*types.AuditLogEntry, error) {
				assert.Equal(t, expected.ID, auditLogID)

				return expected, nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		result, err := manager.GetAuditLogEntry(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetAuditLogEntryCalls(), 1)
	})
}

func TestAuditDataManager_GetAuditLogEntriesForAccount(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		accountID := fakes.BuildFakeID()
		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeAuditLogEntriesList()

		repo := &auditmock.RepositoryMock{
			GetAuditLogEntriesForAccountFunc: func(_ context.Context, actualAccountID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.AuditLogEntry], error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, filter, actualFilter)

				return expected, nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		result, err := manager.GetAuditLogEntriesForAccount(ctx, accountID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetAuditLogEntriesForAccountCalls(), 1)
	})
}

func TestAuditDataManager_Record(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleEntry := fakes.BuildFakeAuditLogEntry()
		querier := &mockdatabase.SQLQueryExecutorMock{}

		repo := &auditmock.RepositoryMock{
			RecordFunc: func(_ context.Context, _ database.SQLQueryExecutor, entries ...*types.AuditLogEntry) error {
				require.Len(t, entries, 1)
				assert.Equal(t, exampleEntry.ID, entries[0].ID)
				assert.Equal(t, exampleEntry.BelongsToUser, entries[0].BelongsToUser)

				return nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		err := manager.Record(ctx, querier, exampleEntry)

		require.NoError(t, err)
		assert.Len(t, repo.RecordCalls(), 1)
	})

	// The variadic form is the point of Record: a transaction touching three
	// resources should pay one chain-head lookup and one INSERT, not three of each.
	t.Run("passes a batch through as one call", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		first, second := fakes.BuildFakeAuditLogEntry(), fakes.BuildFakeAuditLogEntry()
		querier := &mockdatabase.SQLQueryExecutorMock{}

		repo := &auditmock.RepositoryMock{
			RecordFunc: func(_ context.Context, _ database.SQLQueryExecutor, entries ...*types.AuditLogEntry) error {
				require.Len(t, entries, 2)
				assert.Equal(t, first.ID, entries[0].ID)
				assert.Equal(t, second.ID, entries[1].ID)

				return nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		err := manager.Record(ctx, querier, first, second)

		require.NoError(t, err)
		assert.Len(t, repo.RecordCalls(), 1)
	})
}
