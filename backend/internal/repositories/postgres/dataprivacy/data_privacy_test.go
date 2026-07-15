package dataprivacy

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests (validation, no DB required) ---

func TestFetchUserDataCollection(T *testing.T) {
	T.Parallel()

	T.Run("with empty user ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.FetchUserDataCollection(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestDeleteUser(T *testing.T) {
	T.Parallel()

	T.Run("with empty user ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.DeleteUser(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests (require DB container) ---

func TestQuerier_Integration_FetchUserDataCollection(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, identityRepo, container := buildDatabaseClientForTest(t)

	_, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = "dataprivacy_fetch_" + identifiers.New()[:8]
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUser := createUserForTest(t, ctx, exampleUser, identityRepo)

	collection, err := dbc.FetchUserDataCollection(ctx, createdUser.ID)
	require.NoError(t, err)
	require.NotNil(t, collection)
	assert.Equal(t, createdUser.ID, collection.Identity.User.ID)
	assert.Equal(t, createdUser.Username, collection.Identity.User.Username)
	assert.NotNil(t, collection.Webhooks.Data)
}

func TestQuerier_Integration_DeleteUser(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, identityRepo, container := buildDatabaseClientForTest(t)

	_, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = "dataprivacy_del_" + identifiers.New()[:8]
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUser := createUserForTest(t, ctx, exampleUser, identityRepo)

	err = dbc.DeleteUser(ctx, createdUser.ID)
	require.NoError(t, err)

	_, err = identityRepo.GetUser(ctx, createdUser.ID)
	assert.Error(t, err)
}
