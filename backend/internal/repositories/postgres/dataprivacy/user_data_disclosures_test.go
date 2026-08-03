package dataprivacy

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests (validation, no DB required) ---

func TestCreateUserDataDisclosure(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateUserDataDisclosure(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})

	T.Run("with empty ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		input := &dataprivacy.UserDataDisclosureCreationInput{
			ID:            "",
			BelongsToUser: identifiers.New(),
			ExpiresAt:     time.Now().Add(24 * time.Hour),
		}
		actual, err := c.CreateUserDataDisclosure(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestGetUserDataDisclosure(T *testing.T) {
	T.Parallel()

	T.Run("with empty disclosure ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserDataDisclosure(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestGetUserDataDisclosuresForUser(T *testing.T) {
	T.Parallel()

	T.Run("with empty user ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserDataDisclosuresForUser(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestMarkUserDataDisclosureCompleted(T *testing.T) {
	T.Parallel()

	T.Run("with empty disclosure ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.MarkUserDataDisclosureCompleted(ctx, "", identifiers.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.MarkUserDataDisclosureCompleted(ctx, identifiers.New(), "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestMarkUserDataDisclosureFailed(T *testing.T) {
	T.Parallel()

	T.Run("with empty disclosure ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.MarkUserDataDisclosureFailed(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestArchiveUserDataDisclosure(T *testing.T) {
	T.Parallel()

	T.Run("with empty disclosure ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveUserDataDisclosure(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestMarkUserDataDisclosureExpired(T *testing.T) {
	T.Parallel()

	T.Run("with empty disclosure ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.MarkUserDataDisclosureExpired(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests (require DB container) ---

func TestQuerier_Integration_UserDataDisclosures(t *testing.T) {
	ctx := t.Context()
	dbc, _, identityRepo := buildDatabaseClientForTest(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = "dataprivacy_" + identifiers.New()[:8]
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUser := createUserForTest(t, ctx, exampleUser, identityRepo)

	disclosureID := identifiers.New()
	input := &dataprivacy.UserDataDisclosureCreationInput{
		ID:            disclosureID,
		BelongsToUser: createdUser.ID,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	created, err := dbc.CreateUserDataDisclosure(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, disclosureID, created.ID)
	assert.Equal(t, createdUser.ID, created.BelongsToUser)
	assert.Equal(t, dataprivacy.UserDataDisclosureStatusPending, created.Status)

	fetched, err := dbc.GetUserDataDisclosure(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)

	listResult, err := dbc.GetUserDataDisclosuresForUser(ctx, createdUser.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, listResult)
	assert.Len(t, listResult.Data, 1)
	assert.Equal(t, uint64(1), listResult.TotalCount)

	reportID := identifiers.New()
	err = dbc.MarkUserDataDisclosureCompleted(ctx, created.ID, reportID)
	require.NoError(t, err)

	completed, err := dbc.GetUserDataDisclosure(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, completed)
	assert.Equal(t, dataprivacy.UserDataDisclosureStatusCompleted, completed.Status)
	assert.Equal(t, reportID, completed.ReportID)
	assert.NotNil(t, completed.CompletedAt)
}

func TestQuerier_Integration_UserDataDisclosures_MarkFailed(t *testing.T) {
	ctx := t.Context()
	dbc, _, identityRepo := buildDatabaseClientForTest(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = "dataprivacy_fail_" + identifiers.New()[:8]
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUser := createUserForTest(t, ctx, exampleUser, identityRepo)

	input := &dataprivacy.UserDataDisclosureCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: createdUser.ID,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	created, err := dbc.CreateUserDataDisclosure(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)

	err = dbc.MarkUserDataDisclosureFailed(ctx, created.ID)
	require.NoError(t, err)

	failed, err := dbc.GetUserDataDisclosure(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, failed)
	assert.Equal(t, dataprivacy.UserDataDisclosureStatusFailed, failed.Status)
}

func TestQuerier_Integration_UserDataDisclosures_Expiry(t *testing.T) {
	ctx := t.Context()
	dbc, _, identityRepo := buildDatabaseClientForTest(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = "dataprivacy_exp_" + identifiers.New()[:8]
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUser := createUserForTest(t, ctx, exampleUser, identityRepo)

	// One disclosure whose window has closed and one that is still open. Only the first should
	// ever be handed to the reaper.
	stale, err := dbc.CreateUserDataDisclosure(ctx, &dataprivacy.UserDataDisclosureCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: createdUser.ID,
		ExpiresAt:     time.Now().Add(-24 * time.Hour),
	})
	require.NoError(t, err)

	fresh, err := dbc.CreateUserDataDisclosure(ctx, &dataprivacy.UserDataDisclosureCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: createdUser.ID,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	require.NoError(t, dbc.MarkUserDataDisclosureCompleted(ctx, stale.ID, identifiers.New()))

	expired, err := dbc.GetExpiredUserDataDisclosures(ctx)
	require.NoError(t, err)

	ids := make([]string, 0, len(expired))
	for _, d := range expired {
		ids = append(ids, d.ID)
	}
	assert.Contains(t, ids, stale.ID)
	assert.NotContains(t, ids, fresh.ID)

	require.NoError(t, dbc.MarkUserDataDisclosureExpired(ctx, stale.ID))

	reaped, err := dbc.GetUserDataDisclosure(ctx, stale.ID)
	require.NoError(t, err)
	assert.Equal(t, dataprivacy.UserDataDisclosureStatusExpired, reaped.Status)

	// A reaped disclosure must not come back, or the sweep never converges.
	expired, err = dbc.GetExpiredUserDataDisclosures(ctx)
	require.NoError(t, err)
	for _, d := range expired {
		assert.NotEqual(t, stale.ID, d.ID)
	}
}

func TestQuerier_Integration_UserDataDisclosures_Archive(t *testing.T) {
	ctx := t.Context()
	dbc, _, identityRepo := buildDatabaseClientForTest(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = "dataprivacy_arch_" + identifiers.New()[:8]
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUser := createUserForTest(t, ctx, exampleUser, identityRepo)

	input := &dataprivacy.UserDataDisclosureCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: createdUser.ID,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	created, err := dbc.CreateUserDataDisclosure(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)

	err = dbc.ArchiveUserDataDisclosure(ctx, created.ID)
	require.NoError(t, err)

	_, err = dbc.GetUserDataDisclosure(ctx, created.ID)
	assert.Error(t, err)

	listResult, err := dbc.GetUserDataDisclosuresForUser(ctx, createdUser.ID, nil)
	require.NoError(t, err)
	assert.Len(t, listResult.Data, 0)
}
