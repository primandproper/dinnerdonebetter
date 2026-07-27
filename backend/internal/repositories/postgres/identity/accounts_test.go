package identity

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAccountForTest(t *testing.T, ctx context.Context, exampleAccount *identity.Account, dbc identity.Repository) *identity.Account {
	t.Helper()

	// create
	if exampleAccount == nil {
		exampleUser := createUserForTest(t, ctx, nil, dbc)
		exampleAccount = fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
	}
	exampleAccount.PaymentProcessorCustomerID = ""
	exampleAccount.Members = nil
	dbInput := converters.ConvertAccountToAccountDatabaseCreationInput(exampleAccount)

	created, err := dbc.CreateAccount(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)
	exampleAccount.CreatedAt = created.CreatedAt
	exampleAccount.WebhookEncryptionKey = created.WebhookEncryptionKey
	assert.Equal(t, exampleAccount, created)

	account, err := dbc.GetAccount(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, account)

	exampleAccount.CreatedAt = account.CreatedAt
	exampleAccount.Members = account.Members
	exampleAccount.WebhookEncryptionKey = account.WebhookEncryptionKey

	assert.Equal(t, exampleAccount, account)

	return created
}

func TestQuerier_Integration_Accounts(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	exampleUser := createUserForTest(t, ctx, nil, dbc)

	exampleAccount := fakes.BuildFakeAccount()
	exampleAccount.Members = nil
	exampleAccount.BelongsToUser = exampleUser.ID
	exampleAccount.PaymentProcessorCustomerID = ""
	createdAccounts := []*identity.Account{}

	// create
	createdAccounts = append(createdAccounts, createAccountForTest(t, ctx, exampleAccount, dbc))

	accountWithMembers, err := dbc.GetAccount(ctx, createdAccounts[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountWithMembers.Members, "account should have owner as member")

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, createdAccounts[0].ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeAccounts, RelevantID: createdAccounts[0].ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeAccountUserMemberships, RelevantID: accountWithMembers.Members[0].ID},
	})

	// update
	updatedAccount := fakes.BuildFakeAccount()
	updatedAccount.ID = createdAccounts[0].ID
	updatedAccount.BelongsToUser = createdAccounts[0].BelongsToUser
	assert.NoError(t, dbc.UpdateAccount(ctx, updatedAccount))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, createdAccounts[0].ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeAccounts, RelevantID: createdAccounts[0].ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeAccountUserMemberships, RelevantID: accountWithMembers.Members[0].ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeAccounts, RelevantID: createdAccounts[0].ID},
	})

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeAccount()
		input.BelongsToUser = exampleUser.ID
		input.Name = fmt.Sprintf("%s %d", updatedAccount.Name, i)
		createdAccounts = append(createdAccounts, createAccountForTest(t, ctx, input, dbc))
	}

	// fetch as list
	accounts, err := dbc.GetAccounts(ctx, exampleUser.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, accounts.Data)
	assert.GreaterOrEqual(t, len(accounts.Data), len(createdAccounts))

	// delete
	for _, account := range createdAccounts {
		assert.NoError(t, dbc.ArchiveAccount(ctx, account.ID, exampleUser.ID))

		pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
			{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeAccounts, RelevantID: account.ID},
		})

		var y *identity.Account
		y, err = dbc.GetAccount(ctx, account.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}

	assert.NoError(t, dbc.ArchiveUser(ctx, exampleUser.ID))
}

func TestQuerier_GetAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		c := buildInertClientForTest(t)

		actual, err := c.GetAccount(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetAccounts(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetAccounts(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateAccount(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUserID

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateAccount(ctx, nil))
	})
}

func TestQuerier_ArchiveAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveAccount(ctx, "", exampleUserID))
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveAccount(ctx, exampleAccountID, ""))
	})
}
