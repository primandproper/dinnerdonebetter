package manager

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/fakes"
	auditmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/mock"

	"github.com/primandproper/platform-go/v9/database"
	mockdatabase "github.com/primandproper/platform-go/v9/database/mock"
	"github.com/primandproper/platform-go/v9/filtering"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

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

func TestAuditDataManager_CreateAuditLogEntry(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleEntry := fakes.BuildFakeAuditLogEntry()
		dbInput := converters.ConvertAuditLogEntryToAuditLogEntryDatabaseCreationInput(exampleEntry)
		querier := &mockdatabase.SQLQueryExecutorMock{}

		repo := &auditmock.RepositoryMock{
			CreateAuditLogEntryFunc: func(_ context.Context, _ database.SQLQueryExecutor, in *types.AuditLogEntryDatabaseCreationInput) (*types.AuditLogEntry, error) {
				assert.Equal(t, dbInput.ID, in.ID)
				assert.Equal(t, dbInput.BelongsToUser, in.BelongsToUser)

				return exampleEntry, nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		created, err := manager.CreateAuditLogEntry(ctx, querier, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, exampleEntry.ID, created.ID)
		assert.Len(t, repo.CreateAuditLogEntryCalls(), 1)
	})
}
