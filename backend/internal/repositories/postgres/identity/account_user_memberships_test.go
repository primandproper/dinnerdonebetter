package identity

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v5/identifiers"

	"github.com/stretchr/testify/assert"
)

func TestQuerier_Integration_AccountUserMemberships(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	exampleUser := createUserForTest(t, ctx, nil, dbc)
	accounts, err := dbc.GetAccounts(ctx, exampleUser.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, accounts.Data, 1)

	account := fakes.BuildFakeAccount()
	account.BelongsToUser = exampleUser.ID
	exampleAccount := createAccountForTest(t, ctx, account, dbc)

	memberUserIDs := []string{exampleUser.ID}

	for range exampleQuantity {
		newMember := createUserForTest(t, ctx, nil, dbc)
		assert.NoError(t, dbc.addUserToAccount(ctx, dbc.writeDB, &identity.AccountUserMembershipDatabaseCreationInput{
			ID:        identifiers.New(),
			Reason:    "testing",
			UserID:    newMember.ID,
			AccountID: exampleAccount.ID,
		}))
		memberUserIDs = append(memberUserIDs, newMember.ID)
	}

	account, err = dbc.GetAccount(ctx, exampleAccount.ID)
	assert.NoError(t, err)

	accountMemberUserIDs := []string{}
	for _, member := range account.Members {
		accountMemberUserIDs = append(accountMemberUserIDs, member.BelongsToUser.ID)
	}

	assert.Subset(t, memberUserIDs, accountMemberUserIDs)

	isMember, err := dbc.UserIsMemberOfAccount(ctx, memberUserIDs[0], exampleAccount.ID)
	assert.NoError(t, err)
	assert.True(t, isMember)

	assert.NoError(t, dbc.MarkAccountAsUserDefault(ctx, memberUserIDs[1], exampleAccount.ID))
	defaultAccountID, err := dbc.GetDefaultAccountIDForUser(ctx, memberUserIDs[1])
	assert.NoError(t, err)
	assert.Equal(t, exampleAccount.ID, defaultAccountID)

	sessionCtxData, err := dbc.BuildSessionContextDataForUser(ctx, memberUserIDs[1], "")
	assert.NoError(t, err)
	assert.NotNil(t, sessionCtxData)
	assert.Equal(t, exampleAccount.ID, sessionCtxData.ActiveAccountID)

	assert.NoError(t, dbc.RemoveUserFromAccount(ctx, memberUserIDs[len(memberUserIDs)-1], exampleAccount.ID))

	assert.NoError(t, dbc.TransferAccountOwnership(ctx, exampleAccount.ID, &identity.AccountOwnershipTransferInput{
		Reason:       "testing",
		CurrentOwner: exampleUser.ID,
		NewOwner:     memberUserIDs[1],
	}))

	assert.NoError(t, dbc.ModifyUserPermissions(ctx, exampleAccount.ID, memberUserIDs[0], &identity.ModifyUserPermissionsInput{
		Reason:  "testing",
		NewRole: "account_admin",
	}))
}

func TestQuerier_BuildSessionContextDataForUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.BuildSessionContextDataForUser(ctx, "", "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetDefaultAccountIDForUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetDefaultAccountIDForUser(ctx, "")
		assert.Error(t, err)
		assert.Zero(t, actual)
	})
}

func TestQuerier_MarkAccountAsUserDefault(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleAccount := fakes.BuildFakeAccount()

		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, "", exampleAccount.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := fakes.BuildFakeUser()

		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkAccountAsUserDefault(ctx, exampleUser.ID, ""))
	})
}

func TestQuerier_UserIsMemberOfAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleAccountID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.UserIsMemberOfAccount(ctx, "", exampleAccountID)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUserID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.UserIsMemberOfAccount(ctx, exampleUserID, "")
		assert.False(t, actual)
		assert.Error(t, err)
	})
}

func TestQuerier_ModifyUserPermissions(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account id", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUserID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		c := buildInertClientForTest(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, "", exampleUserID, exampleInput))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		assert.Error(t, c.ModifyUserPermissions(ctx, exampleAccountID, exampleUserID, nil))
	})
}

func TestSQLQuerier_addUserToAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		err := c.addUserToAccount(ctx, c.writeDB, nil)
		assert.Error(t, err)
	})
}

func TestQuerier_RemoveUserFromAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleAccount := fakes.BuildFakeAccount()

		c := buildInertClientForTest(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, "", exampleAccount.ID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := fakes.BuildFakeUser()

		c := buildInertClientForTest(t)

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUser.ID, ""))
	})

	T.Run("with error creating transaction", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUserID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		c, db := buildMockSQLTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.RemoveUserFromAccount(ctx, exampleUserID, exampleAccountID))
	})
}
