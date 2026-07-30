package auditlogentries

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v8/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAuditLogEntryForTest(t *testing.T, ctx context.Context, querier database.SQLQueryExecutor, exampleAuditLogEntry *types.AuditLogEntry, user *identity.User, account *identity.Account, dbc *repository) *types.AuditLogEntry {
	t.Helper()

	if user == nil {
		user = pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	}

	if account == nil {
		account = pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)
	}

	// create
	if exampleAuditLogEntry == nil {
		exampleAuditLogEntry = fakes.BuildFakeAuditLogEntry()
	}
	exampleAuditLogEntry.BelongsToUser = user.ID
	exampleAuditLogEntry.BelongsToAccount = &account.ID
	dbInput := converters.ConvertAuditLogEntryToAuditLogEntryDatabaseCreationInput(exampleAuditLogEntry)

	created, err := dbc.CreateAuditLogEntry(ctx, querier, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleAuditLogEntry.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleAuditLogEntry, created)

	auditLogEntry, err := dbc.GetAuditLogEntry(ctx, created.ID)
	exampleAuditLogEntry.CreatedAt = auditLogEntry.CreatedAt
	assert.NoError(t, err)
	assert.Equal(t, auditLogEntry, exampleAuditLogEntry)

	return created
}

func TestQuerier_Integration_AuditLogEntries(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	exampleAccount := identityfakes.BuildFakeAccount()
	exampleAccount.BelongsToUser = user.ID
	account := pgtesting.CreateAccountForTest(t, exampleAccount, user.ID, dbc.writeDB)

	exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()
	exampleAuditLogEntry.BelongsToAccount = &account.ID
	exampleAuditLogEntry.BelongsToUser = user.ID
	createdAuditLogEntries := []*types.AuditLogEntry{}

	// create
	createdAuditLogEntries = append(createdAuditLogEntries, createAuditLogEntryForTest(t, ctx, dbc.writeDB, exampleAuditLogEntry, user, account, dbc))

	// create more
	for range exampleQuantity {
		input := fakes.BuildFakeAuditLogEntry()
		createdAuditLogEntries = append(createdAuditLogEntries, createAuditLogEntryForTest(t, ctx, dbc.writeDB, input, user, account, dbc))
	}

	// fetch as list
	auditLogEntries, err := dbc.GetAuditLogEntriesForUser(ctx, user.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, auditLogEntries.Data)
	assert.Equal(t, len(createdAuditLogEntries), len(auditLogEntries.Data))

	auditLogEntries, err = dbc.GetAuditLogEntriesForAccount(ctx, account.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, auditLogEntries.Data)
	assert.Equal(t, len(createdAuditLogEntries), len(auditLogEntries.Data))
}

func TestQuerier_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("with invalid audit log entry MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetAuditLogEntry(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateAuditLogEntry(ctx, c.writeDB, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
