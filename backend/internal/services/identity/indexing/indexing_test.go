package indexing

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
	mocksearch "github.com/primandproper/platform-go/v9/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleIndexRequest(T *testing.T) {
	T.Parallel()

	T.Run("user index type", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		identityRepo := &identitymock.RepositoryMock{
			GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
				assert.Equal(t, exampleUser.ID, userID)

				return exampleUser, nil
			},
			MarkUserAsIndexedFunc: func(_ context.Context, userID string) error {
				assert.Equal(t, exampleUser.ID, userID)

				return nil
			},
		}

		mim := &mocksearch.IndexMock[UserSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		cdi := NewCoreDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			identityRepo,
			mim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleUser.ID,
			IndexType: IndexTypeUsers,
			Delete:    false,
		}

		require.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))

		assert.Len(t, identityRepo.GetUserCalls(), 1)
		assert.Len(t, identityRepo.MarkUserAsIndexedCalls(), 1)
	})

	T.Run("deleting user index type", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		identityRepo := &identitymock.RepositoryMock{
			GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
				assert.Equal(t, exampleUser.ID, userID)

				return exampleUser, nil
			},
		}

		mim := &mocksearch.IndexMock[UserSearchSubset]{
			DeleteFunc: func(_ context.Context, _ string) error { return nil },
		}

		cdi := NewCoreDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			identityRepo,
			mim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleUser.ID,
			IndexType: IndexTypeUsers,
			Delete:    true,
		}

		require.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))

		assert.Len(t, identityRepo.GetUserCalls(), 1)
		assert.Empty(t, identityRepo.MarkUserAsIndexedCalls())
	})
}
