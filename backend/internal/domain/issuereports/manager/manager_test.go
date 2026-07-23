package manager

import (
	"context"
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	issuereportkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	issuereportsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v5/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildIssueReportsManagerForTest(t *testing.T) (*issueReportsManager, *issuereportsmock.Repository) {
	t.Helper()

	ctx := t.Context()
	repo := &issuereportsmock.Repository{}
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

	return m.(*issueReportsManager), repo
}

func setupExpectationsForIssueReportsManager(
	manager *issueReportsManager,
	repoSetupFunc func(repo *issuereportsmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &issuereportsmock.Repository{}
	if repoSetupFunc != nil {
		repoSetupFunc(repo)
	}
	manager.repo = repo

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{repo}
}

func TestIssueReportsDataManager_CreateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildIssueReportsManagerForTest(t)

		dbInput := fakes.BuildFakeIssueReportDatabaseCreationInput()
		createdReport := fakes.BuildFakeIssueReport()
		createdReport.ID = dbInput.ID

		expectations := setupExpectationsForIssueReportsManager(
			manager,
			func(repo *issuereportsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateIssueReport), testutils.ContextMatcher, mock.Anything).Return(createdReport, nil)
			},
			map[string][]string{
				types.IssueReportCreatedServiceEventType: {issuereportkeys.IssueReportIDKey},
			},
		)

		created, err := manager.CreateIssueReport(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, dbInput.ID, created.ID)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildIssueReportsManagerForTest(t)

		dbInput := fakes.BuildFakeIssueReportDatabaseCreationInput()

		expectations := setupExpectationsForIssueReportsManager(
			manager,
			func(repo *issuereportsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateIssueReport), testutils.ContextMatcher, mock.Anything).Return((*types.IssueReport)(nil), errors.New("db error"))
			},
		)

		created, err := manager.CreateIssueReport(ctx, dbInput)

		assert.Error(t, err)
		assert.Nil(t, created)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestIssueReportsDataManager_GetIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildIssueReportsManagerForTest(t)

		expected := fakes.BuildFakeIssueReport()
		repo.On(reflection.GetMethodName(repo.GetIssueReport), testutils.ContextMatcher, expected.ID).Return(expected, nil)

		result, err := manager.GetIssueReport(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestIssueReportsDataManager_GetIssueReports(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildIssueReportsManagerForTest(t)

		filter := filtering.DefaultQueryFilter()
		report := fakes.BuildFakeIssueReport()
		expected := &filtering.QueryFilteredResult[types.IssueReport]{
			Data: []*types.IssueReport{report},
		}
		repo.On(reflection.GetMethodName(repo.GetIssueReports), testutils.ContextMatcher, filter).Return(expected, nil)

		result, err := manager.GetIssueReports(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestIssueReportsDataManager_UpdateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildIssueReportsManagerForTest(t)

		issueReport := fakes.BuildFakeIssueReport()

		expectations := setupExpectationsForIssueReportsManager(
			manager,
			func(repo *issuereportsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.UpdateIssueReport), testutils.ContextMatcher, issueReport).Return(nil)
			},
			map[string][]string{
				types.IssueReportUpdatedServiceEventType: {issuereportkeys.IssueReportIDKey},
			},
		)

		err := manager.UpdateIssueReport(ctx, issueReport)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildIssueReportsManagerForTest(t)

		err := manager.UpdateIssueReport(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestIssueReportsDataManager_ArchiveIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildIssueReportsManagerForTest(t)

		issueReportID := fakes.BuildFakeID()

		expectations := setupExpectationsForIssueReportsManager(
			manager,
			func(repo *issuereportsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveIssueReport), testutils.ContextMatcher, issueReportID).Return(nil)
			},
			map[string][]string{
				types.IssueReportArchivedServiceEventType: {issuereportkeys.IssueReportIDKey},
			},
		)

		err := manager.ArchiveIssueReport(ctx, issueReportID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
