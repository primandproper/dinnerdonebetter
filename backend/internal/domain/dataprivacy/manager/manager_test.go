package manager

import (
	"context"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/mock"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildDataPrivacyManagerForTest builds a manager backed by the given repository mock. A nil repo
// gets an unconfigured mock, which panics if any of its methods are called.
func buildDataPrivacyManagerForTest(t *testing.T, repo *dataprivacymock.RepositoryMock) *dataPrivacyManager {
	t.Helper()

	if repo == nil {
		repo = &dataprivacymock.RepositoryMock{}
	}

	m := NewDataPrivacyManager(tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)

	return m.(*dataPrivacyManager)
}

func TestDataPrivacyManager_FetchUserDataCollection(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := identifiers.New()
		expected := &dataprivacy.UserDataCollection{}

		repo := &dataprivacymock.RepositoryMock{
			FetchUserDataCollectionFunc: func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
				assert.Equal(t, userID, actualUserID)

				return expected, nil
			},
		}
		manager := buildDataPrivacyManagerForTest(t, repo)

		result, err := manager.FetchUserDataCollection(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.FetchUserDataCollectionCalls(), 1)
	})
}

func TestDataPrivacyManager_DeleteUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := identifiers.New()

		repo := &dataprivacymock.RepositoryMock{
			DeleteUserFunc: func(_ context.Context, actualUserID string) error {
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		manager := buildDataPrivacyManagerForTest(t, repo)

		err := manager.DeleteUser(ctx, userID)

		require.NoError(t, err)
		assert.Len(t, repo.DeleteUserCalls(), 1)
	})
}

func TestDataPrivacyManager_CreateUserDataDisclosure(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		disclosureID := identifiers.New()
		userID := identifiers.New()
		input := &dataprivacy.UserDataDisclosureCreationInput{
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			ID:            disclosureID,
			BelongsToUser: userID,
		}

		created := &dataprivacy.UserDataDisclosure{
			ID:            disclosureID,
			BelongsToUser: userID,
		}

		repo := &dataprivacymock.RepositoryMock{
			CreateUserDataDisclosureFunc: func(_ context.Context, in *dataprivacy.UserDataDisclosureCreationInput) (*dataprivacy.UserDataDisclosure, error) {
				assert.NotNil(t, in)

				return created, nil
			},
		}
		manager := buildDataPrivacyManagerForTest(t, repo)

		result, err := manager.CreateUserDataDisclosure(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, disclosureID, result.ID)
		assert.Len(t, repo.CreateUserDataDisclosureCalls(), 1)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &dataprivacymock.RepositoryMock{}
		manager := buildDataPrivacyManagerForTest(t, repo)

		result, err := manager.CreateUserDataDisclosure(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
		assert.Nil(t, result)
		assert.Empty(t, repo.CreateUserDataDisclosureCalls())
	})
}
