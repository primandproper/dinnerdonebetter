package waitlists

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3
)

func createWaitlistForTest(t *testing.T, ctx context.Context, exampleWaitlist *types.Waitlist, dbc *Repository) *types.Waitlist {
	t.Helper()

	if exampleWaitlist == nil {
		exampleWaitlist = fakes.BuildFakeWaitlist()
	}
	dbInput := converters.ConvertWaitlistToWaitlistDatabaseCreationInput(exampleWaitlist)

	created, err := dbc.CreateWaitlist(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleWaitlist.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleWaitlist.ID, created.ID)
	assert.Equal(t, exampleWaitlist.Name, created.Name)
	assert.Equal(t, exampleWaitlist.Description, created.Description)

	waitlist, err := dbc.GetWaitlist(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, waitlist)
	exampleWaitlist.CreatedAt = waitlist.CreatedAt
	assert.Equal(t, waitlist.ID, created.ID)

	return created
}

func TestQuerier_Integration_Waitlists(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	exampleWaitlist := fakes.BuildFakeWaitlist()
	createdWaitlists := []*types.Waitlist{}

	createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, ctx, exampleWaitlist, dbc))

	for i := range exampleQuantity {
		input := fakes.BuildFakeWaitlist()
		input.Name = fmt.Sprintf("%s %d", exampleWaitlist.Name, i)
		createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, ctx, input, dbc))
	}

	// fetch as list
	waitlists, err := dbc.GetWaitlists(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, waitlists.Data)
	assert.Len(t, waitlists.Data, len(createdWaitlists))

	// fetch active waitlists
	activeWaitlists, err := dbc.GetActiveWaitlists(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, activeWaitlists.Data)

	// check not expired
	notExpired, err := dbc.WaitlistIsNotExpired(ctx, createdWaitlists[0].ID)
	require.NoError(t, err)
	assert.True(t, notExpired)

	// Create user and account for signup (enables audit assertions for account-scoped waitlist signups)
	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	signupInput := &types.WaitlistSignupDatabaseCreationInput{
		ID:                fake.BuildFakeID(),
		Notes:             "test signup",
		BelongsToWaitlist: createdWaitlists[0].ID,
		BelongsToUser:     user.ID,
		BelongsToAccount:  account.ID,
	}
	createdSignup, err := dbc.CreateWaitlistSignup(ctx, signupInput)
	require.NoError(t, err)
	require.NotNil(t, createdSignup)

	// Assert audit log entries for waitlist signup create
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlistSignups, RelevantID: createdSignup.ID},
	})

	// fetch signup by ID alone
	fetchedSignup, err := dbc.GetWaitlistSignupByID(ctx, createdSignup.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedSignup)
	assert.Equal(t, createdSignup.ID, fetchedSignup.ID)
	assert.Equal(t, user.ID, fetchedSignup.BelongsToUser)

	// update signup
	createdSignup.Notes = "updated notes"
	require.NoError(t, dbc.UpdateWaitlistSignup(ctx, createdSignup))

	// Assert audit log entry for signup update
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeWaitlistSignups, RelevantID: createdSignup.ID},
	})

	// archive signup before archiving waitlists
	require.NoError(t, dbc.ArchiveWaitlistSignup(ctx, createdSignup.ID))

	// For a minimal integration test we just archive waitlists
	for _, wl := range createdWaitlists {
		require.NoError(t, dbc.ArchiveWaitlist(ctx, wl.ID))

		_, err = dbc.GetWaitlist(ctx, wl.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_WaitlistIsNotExpired(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.WaitlistIsNotExpired(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWaitlist(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_CreateWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateWaitlist(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestQuerier_UpdateWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateWaitlist(ctx, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestQuerier_ArchiveWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveWaitlist(ctx, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist signup ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		exampleWaitlistID := fake.BuildFakeID()

		actual, err := c.GetWaitlistSignup(ctx, "", exampleWaitlistID)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		exampleSignupID := fake.BuildFakeID()

		actual, err := c.GetWaitlistSignup(ctx, exampleSignupID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlistSignupByID(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist signup ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWaitlistSignupByID(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlistSignupsForWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		filter := filtering.DefaultQueryFilter()

		actual, err := c.GetWaitlistSignupsForWaitlist(ctx, "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_CreateWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateWaitlistSignup(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestQuerier_UpdateWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateWaitlistSignup(ctx, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestQuerier_ArchiveWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist signup ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveWaitlistSignup(ctx, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}
