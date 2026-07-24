package manager

import (
	"context"
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/keys"
	waitlistmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v6/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildWaitlistManagerForTest(t *testing.T) (*waitlistManager, *waitlistmock.Repository) {
	t.Helper()

	ctx := t.Context()
	repo := &waitlistmock.Repository{}
	queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: t.Name()}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewWaitlistDataManager(ctx, tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo, queueCfg, mpp)
	require.NoError(t, err)

	return m.(*waitlistManager), repo
}

func setupExpectationsForWaitlistManager(
	manager *waitlistManager,
	repoSetupFunc func(repo *waitlistmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &waitlistmock.Repository{}
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

func TestWaitlistDataManager_CreateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		exampleWaitlist := fakes.BuildFakeWaitlist()
		dbInput := converters.ConvertWaitlistToWaitlistDatabaseCreationInput(exampleWaitlist)

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateWaitlist), testutils.ContextMatcher, mock.MatchedBy(func(in *types.WaitlistDatabaseCreationInput) bool {
					return in.ID == dbInput.ID && in.Name == dbInput.Name && in.Description == dbInput.Description
				})).Return(exampleWaitlist, nil)
			},
			map[string][]string{
				types.WaitlistCreatedServiceEventType: {waitlistkeys.WaitlistIDKey},
			},
		)

		created, err := manager.CreateWaitlist(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, exampleWaitlist.ID, created.ID)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		exampleWaitlist := fakes.BuildFakeWaitlist()
		dbInput := converters.ConvertWaitlistToWaitlistDatabaseCreationInput(exampleWaitlist)

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateWaitlist), testutils.ContextMatcher, mock.Anything).Return((*types.Waitlist)(nil), errors.New("db error"))
			},
		)

		created, err := manager.CreateWaitlist(ctx, dbInput)

		assert.Error(t, err)
		assert.Nil(t, created)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWaitlistDataManager_GetWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWaitlistManagerForTest(t)

		expected := fakes.BuildFakeWaitlist()
		repo.On(reflection.GetMethodName(repo.GetWaitlist), testutils.ContextMatcher, expected.ID).Return(expected, nil)

		result, err := manager.GetWaitlist(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestWaitlistDataManager_GetWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWaitlistManagerForTest(t)

		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeWaitlistsList()
		repo.On(reflection.GetMethodName(repo.GetWaitlists), testutils.ContextMatcher, filter).Return(expected, nil)

		result, err := manager.GetWaitlists(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestWaitlistDataManager_UpdateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		waitlist := fakes.BuildFakeWaitlist()

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.UpdateWaitlist), testutils.ContextMatcher, waitlist).Return(nil)
			},
			map[string][]string{
				types.WaitlistUpdatedServiceEventType: {waitlistkeys.WaitlistIDKey},
			},
		)

		err := manager.UpdateWaitlist(ctx, waitlist)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		err := manager.UpdateWaitlist(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestWaitlistDataManager_ArchiveWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		waitlistID := fakes.BuildFakeID()

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveWaitlist), testutils.ContextMatcher, waitlistID).Return(nil)
			},
			map[string][]string{
				types.WaitlistArchivedServiceEventType: {waitlistkeys.WaitlistIDKey},
			},
		)

		err := manager.ArchiveWaitlist(ctx, waitlistID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWaitlistDataManager_CreateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		exampleSignup := fakes.BuildFakeWaitlistSignup()
		dbInput := converters.ConvertWaitlistSignupToWaitlistSignupDatabaseCreationInput(exampleSignup)

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateWaitlistSignup), testutils.ContextMatcher, mock.MatchedBy(func(in *types.WaitlistSignupDatabaseCreationInput) bool {
					return in.ID == dbInput.ID && in.BelongsToWaitlist == dbInput.BelongsToWaitlist
				})).Return(exampleSignup, nil)
			},
			map[string][]string{
				types.WaitlistSignupCreatedServiceEventType: {waitlistkeys.WaitlistSignupIDKey, waitlistkeys.WaitlistIDKey},
			},
		)

		created, err := manager.CreateWaitlistSignup(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, exampleSignup.ID, created.ID)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWaitlistDataManager_UpdateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		signup := fakes.BuildFakeWaitlistSignup()

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.UpdateWaitlistSignup), testutils.ContextMatcher, signup).Return(nil)
			},
			map[string][]string{
				types.WaitlistSignupUpdatedServiceEventType: {waitlistkeys.WaitlistSignupIDKey, waitlistkeys.WaitlistIDKey},
			},
		)

		err := manager.UpdateWaitlistSignup(ctx, signup)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		err := manager.UpdateWaitlistSignup(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestWaitlistDataManager_ArchiveWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWaitlistManagerForTest(t)

		waitlistSignupID := fakes.BuildFakeID()

		expectations := setupExpectationsForWaitlistManager(
			manager,
			func(repo *waitlistmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveWaitlistSignup), testutils.ContextMatcher, waitlistSignupID).Return(nil)
			},
			map[string][]string{
				types.WaitlistSignupArchivedServiceEventType: {waitlistkeys.WaitlistSignupIDKey},
			},
		)

		err := manager.ArchiveWaitlistSignup(ctx, waitlistSignupID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWaitlistDataManager_GetWaitlistSignupByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWaitlistManagerForTest(t)

		expected := fakes.BuildFakeWaitlistSignup()
		repo.On(reflection.GetMethodName(repo.GetWaitlistSignupByID), testutils.ContextMatcher, expected.ID).Return(expected, nil)

		result, err := manager.GetWaitlistSignupByID(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestWaitlistDataManager_GetWaitlistSignupsForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWaitlistManagerForTest(t)

		userID := fakes.BuildFakeID()
		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeWaitlistSignupsList()
		repo.On(reflection.GetMethodName(repo.GetWaitlistSignupsForUser), testutils.ContextMatcher, userID, filter).Return(expected, nil)

		result, err := manager.GetWaitlistSignupsForUser(ctx, userID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}
