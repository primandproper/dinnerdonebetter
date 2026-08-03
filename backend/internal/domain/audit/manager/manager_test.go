package manager

import (
	"context"
	"errors"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/fakes"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"

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

func TestAuditDataManager_Record(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		querier := &mockdatabase.SQLQueryExecutorMock{}
		entries := []*types.Entry{
			{
				ResourceType: "example",
				ResourceID:   "abc",
				EventType:    types.EventCreated,
				Actor:        types.UserActor("user_123"),
			},
			{
				ResourceType: "example",
				ResourceID:   "def",
				EventType:    types.EventUpdated,
				Actor:        types.UserActor("user_123"),
			},
		}

		repo := &auditmock.RepositoryMock{
			RecordFunc: func(_ context.Context, _ database.SQLQueryExecutor, in ...*types.Entry) error {
				// Passed through variadically rather than one call per entry: the
				// batch is the point, since it costs one chain-head lookup and one
				// INSERT instead of two of each.
				assert.Len(t, in, 2)
				assert.Equal(t, entries[0].ResourceID, in[0].ResourceID)
				assert.Equal(t, entries[1].ResourceID, in[1].ResourceID)

				return nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		require.NoError(t, manager.Record(ctx, querier, entries...))
		assert.Len(t, repo.RecordCalls(), 1)
	})

	t.Run("with error recording", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		repo := &auditmock.RepositoryMock{
			RecordFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ ...*types.Entry) error {
				return errors.New("blah")
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		err := manager.Record(ctx, &mockdatabase.SQLQueryExecutorMock{}, &types.Entry{
			ResourceType: "example",
			EventType:    types.EventCreated,
			Actor:        types.SystemActor(),
		})
		assert.Error(t, err)
	})
}

func TestAuditDataManager_VerifyChain(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := &types.VerificationResult{Scope: "account_123", Checked: 4}
		repo := &auditmock.RepositoryMock{
			VerifyChainFunc: func(_ context.Context, accountID string, _, _ time.Time) (*types.VerificationResult, error) {
				assert.Equal(t, "account_123", accountID)

				return expected, nil
			},
		}
		manager := buildAuditManagerForTest(t, repo)

		actual, err := manager.VerifyChain(ctx, "account_123", time.Time{}, time.Time{})
		require.NoError(t, err)
		assert.True(t, actual.Intact())
		assert.Len(t, repo.VerifyChainCalls(), 1)
	})
}
