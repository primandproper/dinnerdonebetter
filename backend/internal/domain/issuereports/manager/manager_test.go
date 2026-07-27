package manager

import (
	"context"
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	issuereportsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/mock"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v7/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildIssueReportsManagerForTest builds a manager backed by the given repository mock. A nil repo
// gets an unconfigured mock, which panics if any of its methods are called.
func buildIssueReportsManagerForTest(t *testing.T, repo *issuereportsmock.RepositoryMock) *issueReportsManager {
	t.Helper()

	if repo == nil {
		repo = &issuereportsmock.RepositoryMock{}
	}

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: t.Name()}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewIssueReportsDataManager(ctx, tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo, queueCfg, mpp)
	require.NoError(t, err)

	manager := m.(*issueReportsManager)
	manager.dataChangesPublisher = &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	return manager
}

func TestIssueReportsDataManager_CreateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		dbInput := fakes.BuildFakeIssueReportDatabaseCreationInput()
		createdReport := fakes.BuildFakeIssueReport()
		createdReport.ID = dbInput.ID

		repo := &issuereportsmock.RepositoryMock{
			CreateIssueReportFunc: func(_ context.Context, _ *types.IssueReportDatabaseCreationInput) (*types.IssueReport, error) {
				return createdReport, nil
			},
		}
		manager := buildIssueReportsManagerForTest(t, repo)

		created, err := manager.CreateIssueReport(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, dbInput.ID, created.ID)
		assert.Len(t, repo.CreateIssueReportCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		dbInput := fakes.BuildFakeIssueReportDatabaseCreationInput()

		repo := &issuereportsmock.RepositoryMock{
			CreateIssueReportFunc: func(_ context.Context, _ *types.IssueReportDatabaseCreationInput) (*types.IssueReport, error) {
				return nil, errors.New("db error")
			},
		}
		manager := buildIssueReportsManagerForTest(t, repo)

		created, err := manager.CreateIssueReport(ctx, dbInput)

		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Len(t, repo.CreateIssueReportCalls(), 1)
	})
}

func TestIssueReportsDataManager_GetIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := fakes.BuildFakeIssueReport()

		repo := &issuereportsmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*types.IssueReport, error) {
				assert.Equal(t, expected.ID, issueReportID)

				return expected, nil
			},
		}
		manager := buildIssueReportsManagerForTest(t, repo)

		result, err := manager.GetIssueReport(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetIssueReportCalls(), 1)
	})
}

func TestIssueReportsDataManager_GetIssueReports(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		filter := filtering.DefaultQueryFilter()
		report := fakes.BuildFakeIssueReport()
		expected := &filtering.QueryFilteredResult[types.IssueReport]{
			Data: []*types.IssueReport{report},
		}

		repo := &issuereportsmock.RepositoryMock{
			GetIssueReportsFunc: func(_ context.Context, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.IssueReport], error) {
				assert.Equal(t, filter, actualFilter)

				return expected, nil
			},
		}
		manager := buildIssueReportsManagerForTest(t, repo)

		result, err := manager.GetIssueReports(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetIssueReportsCalls(), 1)
	})
}

func TestIssueReportsDataManager_UpdateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		issueReport := fakes.BuildFakeIssueReport()

		repo := &issuereportsmock.RepositoryMock{
			UpdateIssueReportFunc: func(_ context.Context, actual *types.IssueReport) error {
				assert.Equal(t, issueReport, actual)

				return nil
			},
		}
		manager := buildIssueReportsManagerForTest(t, repo)

		err := manager.UpdateIssueReport(ctx, issueReport)

		require.NoError(t, err)
		assert.Len(t, repo.UpdateIssueReportCalls(), 1)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager := buildIssueReportsManagerForTest(t, nil)

		err := manager.UpdateIssueReport(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestIssueReportsDataManager_ArchiveIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		issueReportID := fakes.BuildFakeID()

		repo := &issuereportsmock.RepositoryMock{
			ArchiveIssueReportFunc: func(_ context.Context, actualID string) error {
				assert.Equal(t, issueReportID, actualID)

				return nil
			},
		}
		manager := buildIssueReportsManagerForTest(t, repo)

		err := manager.ArchiveIssueReport(ctx, issueReportID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveIssueReportCalls(), 1)
	})
}
