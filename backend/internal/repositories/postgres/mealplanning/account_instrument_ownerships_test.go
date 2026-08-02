package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAccountInstrumentOwnershipForTest(t *testing.T, ctx context.Context, exampleAccountInstrumentOwnership *types.AccountInstrumentOwnership, dbc *repository) *types.AccountInstrumentOwnership {
	t.Helper()

	// create
	if exampleAccountInstrumentOwnership == nil {
		exampleAccountInstrumentOwnership = fakes.BuildFakeAccountInstrumentOwnership()
	}
	dbInput := converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipDatabaseCreationInput(exampleAccountInstrumentOwnership)

	created, err := dbc.CreateAccountInstrumentOwnership(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleAccountInstrumentOwnership.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleAccountInstrumentOwnership.Instrument.ID, created.Instrument.ID)
	exampleAccountInstrumentOwnership.Instrument = created.Instrument
	assert.Equal(t, exampleAccountInstrumentOwnership, created)

	accountInstrumentOwnership, err := dbc.GetAccountInstrumentOwnership(ctx, created.ID, created.BelongsToAccount)
	exampleAccountInstrumentOwnership.CreatedAt = accountInstrumentOwnership.CreatedAt
	assert.Equal(t, exampleAccountInstrumentOwnership.Instrument.ID, accountInstrumentOwnership.Instrument.ID)
	exampleAccountInstrumentOwnership.Instrument = accountInstrumentOwnership.Instrument

	assert.NoError(t, err)
	assert.Equal(t, accountInstrumentOwnership, exampleAccountInstrumentOwnership)

	return created
}

func TestQuerier_Integration_AccountInstrumentOwnerships(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	instrument := createValidInstrumentForTest(t, ctx, nil, dbc)

	exampleAccountInstrumentOwnership := fakes.BuildFakeAccountInstrumentOwnership()
	exampleAccountInstrumentOwnership.BelongsToAccount = account.ID
	exampleAccountInstrumentOwnership.Instrument = *instrument
	createdAccountInstrumentOwnerships := []*types.AccountInstrumentOwnership{}

	// create
	createdAccountInstrumentOwnerships = append(createdAccountInstrumentOwnerships, createAccountInstrumentOwnershipForTest(t, ctx, exampleAccountInstrumentOwnership, dbc))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeAccountInstrumentOwnerships, RelevantID: createdAccountInstrumentOwnerships[0].ID},
	})

	// update
	assert.NoError(t, dbc.UpdateAccountInstrumentOwnership(ctx, createdAccountInstrumentOwnerships[0]))

	// create more
	for range exampleQuantity {
		newInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeAccountInstrumentOwnership()
		input.BelongsToAccount = account.ID
		input.Instrument = *newInstrument
		createdAccountInstrumentOwnerships = append(createdAccountInstrumentOwnerships, createAccountInstrumentOwnershipForTest(t, ctx, input, dbc))
	}

	// fetch as list
	accountInstrumentOwnerships, err := dbc.GetAccountInstrumentOwnerships(ctx, account.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, accountInstrumentOwnerships.Data)
	assert.Equal(t, len(createdAccountInstrumentOwnerships), len(accountInstrumentOwnerships.Data))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeAccountInstrumentOwnerships, RelevantID: createdAccountInstrumentOwnerships[0].ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeAccountInstrumentOwnerships, RelevantID: createdAccountInstrumentOwnerships[0].ID},
	})

	// delete
	for _, accountInstrumentOwnership := range createdAccountInstrumentOwnerships {
		assert.NoError(t, dbc.ArchiveAccountInstrumentOwnership(ctx, accountInstrumentOwnership.ID, account.ID))

		pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
			{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeAccountInstrumentOwnerships, RelevantID: accountInstrumentOwnership.ID},
		})

		var exists bool
		exists, err = dbc.AccountInstrumentOwnershipExists(ctx, accountInstrumentOwnership.ID, account.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.AccountInstrumentOwnership
		y, err = dbc.GetAccountInstrumentOwnership(ctx, accountInstrumentOwnership.ID, account.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_AccountInstrumentOwnershipExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account instrument ownership ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.AccountInstrumentOwnershipExists(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account instrument ownership ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetAccountInstrumentOwnership(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateAccountInstrumentOwnership(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateAccountInstrumentOwnership(ctx, nil))
	})
}

func TestQuerier_ArchiveAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account instrument ownership ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveAccountInstrumentOwnership(ctx, "", exampleAccountID))
	})
}

func TestQuerier_Integration_AccountInstrumentOwnerships_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.AccountInstrumentOwnership]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "account instrument ownership",
		CreateItem: func(ctx context.Context, i int) *types.AccountInstrumentOwnership {
			instrument := createValidInstrumentForTest(t, ctx, nil, dbc)
			accountInstrumentOwnership := fakes.BuildFakeAccountInstrumentOwnership()
			accountInstrumentOwnership.BelongsToAccount = account.ID
			accountInstrumentOwnership.Instrument = *instrument
			return createAccountInstrumentOwnershipForTest(t, ctx, accountInstrumentOwnership, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.AccountInstrumentOwnership], error) {
			return dbc.GetAccountInstrumentOwnerships(ctx, account.ID, filter)
		},
		GetID: func(accountInstrumentOwnership *types.AccountInstrumentOwnership) string {
			return accountInstrumentOwnership.ID
		},
		CleanupItem: func(ctx context.Context, accountInstrumentOwnership *types.AccountInstrumentOwnership) error {
			return dbc.ArchiveAccountInstrumentOwnership(ctx, accountInstrumentOwnership.ID, account.ID)
		},
	})
}
