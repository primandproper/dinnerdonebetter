package manager

import (
	"context"
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/mock"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildWaitlistManagerForTest builds a manager backed by the given repository mock. A nil repo gets
// an unconfigured mock, which panics if any of its methods are called.
func buildWaitlistManagerForTest(t *testing.T, repo *waitlistmock.RepositoryMock) *waitlistManager {
	t.Helper()

	if repo == nil {
		repo = &waitlistmock.RepositoryMock{}
	}

	ctx := t.Context()

	m, err := NewWaitlistDataManager(ctx, tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)
	require.NoError(t, err)

	manager := m.(*waitlistManager)

	return manager
}

func TestWaitlistDataManager_CreateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleWaitlist := fakes.BuildFakeWaitlist()
		dbInput := converters.ConvertWaitlistToWaitlistDatabaseCreationInput(exampleWaitlist)

		repo := &waitlistmock.RepositoryMock{
			CreateWaitlistFunc: func(_ context.Context, in *types.WaitlistDatabaseCreationInput) (*types.Waitlist, error) {
				assert.Equal(t, dbInput.ID, in.ID)
				assert.Equal(t, dbInput.Name, in.Name)
				assert.Equal(t, dbInput.Description, in.Description)

				return exampleWaitlist, nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		created, err := manager.CreateWaitlist(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, exampleWaitlist.ID, created.ID)
		assert.Len(t, repo.CreateWaitlistCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleWaitlist := fakes.BuildFakeWaitlist()
		dbInput := converters.ConvertWaitlistToWaitlistDatabaseCreationInput(exampleWaitlist)

		repo := &waitlistmock.RepositoryMock{
			CreateWaitlistFunc: func(_ context.Context, _ *types.WaitlistDatabaseCreationInput) (*types.Waitlist, error) {
				return nil, errors.New("db error")
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		created, err := manager.CreateWaitlist(ctx, dbInput)

		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Len(t, repo.CreateWaitlistCalls(), 1)
	})
}

func TestWaitlistDataManager_GetWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := fakes.BuildFakeWaitlist()

		repo := &waitlistmock.RepositoryMock{
			GetWaitlistFunc: func(_ context.Context, waitlistID string) (*types.Waitlist, error) {
				assert.Equal(t, expected.ID, waitlistID)

				return expected, nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		result, err := manager.GetWaitlist(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetWaitlistCalls(), 1)
	})
}

func TestWaitlistDataManager_GetWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeWaitlistsList()

		repo := &waitlistmock.RepositoryMock{
			GetWaitlistsFunc: func(_ context.Context, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Waitlist], error) {
				assert.Equal(t, filter, actualFilter)

				return expected, nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		result, err := manager.GetWaitlists(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetWaitlistsCalls(), 1)
	})
}

func TestWaitlistDataManager_UpdateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		waitlist := fakes.BuildFakeWaitlist()

		repo := &waitlistmock.RepositoryMock{
			UpdateWaitlistFunc: func(_ context.Context, actual *types.Waitlist) error {
				assert.Equal(t, waitlist, actual)

				return nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		err := manager.UpdateWaitlist(ctx, waitlist)

		require.NoError(t, err)
		assert.Len(t, repo.UpdateWaitlistCalls(), 1)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager := buildWaitlistManagerForTest(t, nil)

		err := manager.UpdateWaitlist(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestWaitlistDataManager_ArchiveWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		waitlistID := fakes.BuildFakeID()

		repo := &waitlistmock.RepositoryMock{
			ArchiveWaitlistFunc: func(_ context.Context, actualID string) error {
				assert.Equal(t, waitlistID, actualID)

				return nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		err := manager.ArchiveWaitlist(ctx, waitlistID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveWaitlistCalls(), 1)
	})
}

func TestWaitlistDataManager_CreateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleSignup := fakes.BuildFakeWaitlistSignup()
		dbInput := converters.ConvertWaitlistSignupToWaitlistSignupDatabaseCreationInput(exampleSignup)

		repo := &waitlistmock.RepositoryMock{
			CreateWaitlistSignupFunc: func(_ context.Context, in *types.WaitlistSignupDatabaseCreationInput) (*types.WaitlistSignup, error) {
				assert.Equal(t, dbInput.ID, in.ID)
				assert.Equal(t, dbInput.BelongsToWaitlist, in.BelongsToWaitlist)

				return exampleSignup, nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		created, err := manager.CreateWaitlistSignup(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, exampleSignup.ID, created.ID)
		assert.Len(t, repo.CreateWaitlistSignupCalls(), 1)
	})
}

func TestWaitlistDataManager_UpdateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		signup := fakes.BuildFakeWaitlistSignup()

		repo := &waitlistmock.RepositoryMock{
			UpdateWaitlistSignupFunc: func(_ context.Context, actual *types.WaitlistSignup) error {
				assert.Equal(t, signup, actual)

				return nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		err := manager.UpdateWaitlistSignup(ctx, signup)

		require.NoError(t, err)
		assert.Len(t, repo.UpdateWaitlistSignupCalls(), 1)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager := buildWaitlistManagerForTest(t, nil)

		err := manager.UpdateWaitlistSignup(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestWaitlistDataManager_ArchiveWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		waitlistSignupID := fakes.BuildFakeID()

		repo := &waitlistmock.RepositoryMock{
			ArchiveWaitlistSignupFunc: func(_ context.Context, actualID string) error {
				assert.Equal(t, waitlistSignupID, actualID)

				return nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		err := manager.ArchiveWaitlistSignup(ctx, waitlistSignupID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveWaitlistSignupCalls(), 1)
	})
}

func TestWaitlistDataManager_GetWaitlistSignupByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := fakes.BuildFakeWaitlistSignup()

		repo := &waitlistmock.RepositoryMock{
			GetWaitlistSignupByIDFunc: func(_ context.Context, signupID string) (*types.WaitlistSignup, error) {
				assert.Equal(t, expected.ID, signupID)

				return expected, nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		result, err := manager.GetWaitlistSignupByID(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetWaitlistSignupByIDCalls(), 1)
	})
}

func TestWaitlistDataManager_GetWaitlistSignupsForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := fakes.BuildFakeID()
		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeWaitlistSignupsList()

		repo := &waitlistmock.RepositoryMock{
			GetWaitlistSignupsForUserFunc: func(_ context.Context, actualUserID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.WaitlistSignup], error) {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, filter, actualFilter)

				return expected, nil
			},
		}
		manager := buildWaitlistManagerForTest(t, repo)

		result, err := manager.GetWaitlistSignupsForUser(ctx, userID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetWaitlistSignupsForUserCalls(), 1)
	})
}
