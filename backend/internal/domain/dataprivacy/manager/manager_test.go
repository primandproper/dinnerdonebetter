package manager

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildDataPrivacyManagerForTest(t *testing.T) (*dataPrivacyManager, *dataprivacymock.Repository) {
	t.Helper()

	repo := &dataprivacymock.Repository{}
	m := NewDataPrivacyManager(tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)
	return m.(*dataPrivacyManager), repo
}

func TestDataPrivacyManager_FetchUserDataCollection(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildDataPrivacyManagerForTest(t)

		userID := identifiers.New()
		expected := &dataprivacy.UserDataCollection{}
		repo.On(reflection.GetMethodName(repo.FetchUserDataCollection), testutils.ContextMatcher, userID).Return(expected, nil)

		result, err := manager.FetchUserDataCollection(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestDataPrivacyManager_DeleteUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildDataPrivacyManagerForTest(t)

		userID := identifiers.New()
		repo.On(reflection.GetMethodName(repo.DeleteUser), testutils.ContextMatcher, userID).Return(nil)

		err := manager.DeleteUser(ctx, userID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestDataPrivacyManager_CreateUserDataDisclosure(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildDataPrivacyManagerForTest(t)

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
		repo.On(reflection.GetMethodName(repo.CreateUserDataDisclosure), testutils.ContextMatcher, mock.Anything).Return(created, nil)

		result, err := manager.CreateUserDataDisclosure(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, disclosureID, result.ID)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildDataPrivacyManagerForTest(t)

		result, err := manager.CreateUserDataDisclosure(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
		assert.Nil(t, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}
