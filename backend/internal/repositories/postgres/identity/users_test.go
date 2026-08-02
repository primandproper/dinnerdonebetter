package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserForTest(t *testing.T, ctx context.Context, exampleUser *identity.User, dbc identity.Repository) *identity.User {
	t.Helper()

	// create
	if exampleUser == nil {
		exampleUser = fakes.BuildFakeUser()
	}
	dbInput := converters.ConvertUserToUserDatabaseCreationInput(exampleUser)

	exampleUser.TwoFactorSecretVerifiedAt = nil
	created, err := dbc.CreateUser(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleUser.CreatedAt = created.CreatedAt
	exampleUser.TwoFactorSecretVerifiedAt = created.TwoFactorSecretVerifiedAt
	assert.NoError(t, err)
	assert.Equal(t, exampleUser, created)

	user, err := dbc.GetUser(ctx, created.ID)
	exampleUser.CreatedAt = user.CreatedAt
	exampleUser.Birthday = user.Birthday

	assert.NoError(t, err)
	assert.Equal(t, user, exampleUser)

	return created
}

func TestQuerier_Integration_Users(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = fmt.Sprintf("%d", pgtesting.HashStringToNumberForTest(t, exampleUser.Username))
	exampleUser.TwoFactorSecretVerifiedAt = nil
	createdUsers := []*identity.User{}

	// create
	createdUsers = append(createdUsers, createUserForTest(t, ctx, exampleUser, dbc))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeUser()
		input.Username = fmt.Sprintf("%s_%d", exampleUser.Username, i)
		createdUsers = append(createdUsers, createUserForTest(t, ctx, input, dbc))
	}

	// fetch as list
	users, err := dbc.GetUsers(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, users.Data)

	firstUser := createdUsers[0]

	u, err := dbc.GetUserByUsername(ctx, firstUser.Username)
	assert.NoError(t, err)
	assert.Equal(t, firstUser.ID, u.ID)
	firstUser = u

	u, err = dbc.GetUserByEmail(ctx, firstUser.EmailAddress)
	assert.NoError(t, err)
	assert.Equal(t, firstUser, u)

	foundForUsername, err := dbc.SearchForUsersByUsername(ctx, firstUser.Username, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, foundForUsername)

	// update first user's username
	newUsername := fmt.Sprintf("%s_new", firstUser.Username)
	assert.NoError(t, dbc.UpdateUserUsername(ctx, firstUser.ID, newUsername))
	firstUser, err = dbc.GetUser(ctx, firstUser.ID)
	require.NoError(t, err)
	assert.Equal(t, firstUser.Username, newUsername)

	// update first user's details
	newFirstName, newLastName, birthday := "new_first", "new_last", time.Now()
	assert.NoError(t, dbc.UpdateUserDetails(ctx, firstUser.ID, &identity.UserDetailsDatabaseUpdateInput{
		FirstName: newFirstName,
		LastName:  newLastName,
		Birthday:  birthday,
	}))
	firstUser, err = dbc.GetUser(ctx, firstUser.ID)
	require.NoError(t, err)
	assert.Equal(t, firstUser.FirstName, newFirstName)
	assert.Equal(t, firstUser.LastName, newLastName)
	assert.Equal(t, firstUser.Birthday.Round(time.Second), birthday.Round(time.Second))

	// update first user's avatar (create uploaded_media first, then set avatar)
	avatarMediaID := fakes.BuildFakeID()
	_, err = dbc.writeDB.ExecContext(ctx, `INSERT INTO uploaded_media (id, storage_path, mime_type, created_by_user) VALUES ($1, $2, $3, $4)`,
		avatarMediaID, "avatars/"+avatarMediaID+".png", "image/png", firstUser.ID)
	require.NoError(t, err)
	assert.NoError(t, dbc.SetUserAvatar(ctx, firstUser.ID, avatarMediaID))
	firstUser, err = dbc.GetUser(ctx, firstUser.ID)
	require.NoError(t, err)
	require.NotNil(t, firstUser.Avatar)
	assert.Equal(t, firstUser.Avatar.ID, avatarMediaID)

	// update first user's email address
	newEmailAddress := fakes.BuildFakeID()
	assert.NoError(t, dbc.UpdateUserEmailAddress(ctx, firstUser.ID, newEmailAddress))
	firstUser, err = dbc.GetUser(ctx, firstUser.ID)
	require.NoError(t, err)
	assert.Equal(t, firstUser.EmailAddress, newEmailAddress)

	// update first user's password
	newPassword := fakes.BuildFakeID()
	assert.NoError(t, dbc.UpdateUserPassword(ctx, firstUser.ID, newPassword))
	firstUser, err = dbc.GetUser(ctx, firstUser.ID)
	require.NoError(t, err)
	assert.Equal(t, firstUser.HashedPassword, newPassword)

	// update first user's two factor secret
	new2FASecret := fakes.BuildFakeID()
	assert.NoError(t, dbc.UpdateUserTwoFactorSecret(ctx, firstUser.ID, new2FASecret))
	firstUser, err = dbc.GetUser(ctx, firstUser.ID)
	require.NoError(t, err)
	assert.Equal(t, firstUser.TwoFactorSecret, new2FASecret)

	assert.NoError(t, dbc.MarkUserTwoFactorSecretAsVerified(ctx, firstUser.ID))
	assert.NoError(t, dbc.MarkUserTwoFactorSecretAsUnverified(ctx, firstUser.ID, fakes.BuildFakeID()))

	u, err = dbc.GetUserWithUnverifiedTwoFactorSecret(ctx, firstUser.ID)
	assert.NoError(t, err)
	firstUser.LastUpdatedAt = u.LastUpdatedAt                         // we've been changing a bunch of stuff
	firstUser.PasswordLastChangedAt = u.PasswordLastChangedAt         // from the UpdateUserPassword call above
	firstUser.TwoFactorSecretVerifiedAt = u.TwoFactorSecretVerifiedAt // from the two calls above
	firstUser.Birthday = u.Birthday
	firstUser.AccountStatus = u.AccountStatus
	firstUser.TwoFactorSecret = u.TwoFactorSecret
	assert.Equal(t, firstUser, u)

	userIDs, err := dbc.GetUserIDsThatNeedSearchIndexing(ctx)
	assert.NotEmpty(t, userIDs)
	assert.NoError(t, err)

	assert.NoError(t, dbc.MarkUserAsIndexed(ctx, firstUser.ID))

	token, err := dbc.GetEmailAddressVerificationTokenForUser(ctx, firstUser.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	u, err = dbc.GetUserByEmailAddressVerificationToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, firstUser, u)

	assert.NoError(t, dbc.MarkUserEmailAddressAsVerified(ctx, firstUser.ID, token))

	// Promote to service_admin role and verify 2FA.
	_, err = dbc.writeDB.ExecContext(ctx, "UPDATE user_role_assignments SET archived_at = NOW() WHERE user_id = $1 AND account_id IS NULL AND archived_at IS NULL", firstUser.ID)
	assert.NoError(t, err)
	_, err = dbc.writeDB.ExecContext(ctx, "INSERT INTO user_role_assignments (id, user_id, role_id) VALUES ($1, $2, $3)", identifiers.New(), firstUser.ID, authorization.ServiceAdminRoleID)
	assert.NoError(t, err)
	_, err = dbc.writeDB.ExecContext(ctx, "UPDATE users SET two_factor_secret_verified_at = NOW() WHERE id = $1", firstUser.ID)
	assert.NoError(t, err)

	u, err = dbc.GetAdminUserByUsername(ctx, firstUser.Username)
	assert.NoError(t, err)
	firstUser.AccountStatus = u.AccountStatus
	firstUser.AccountStatusExplanation = u.AccountStatusExplanation
	firstUser.EmailAddressVerifiedAt = u.EmailAddressVerifiedAt
	firstUser.TwoFactorSecretVerifiedAt = u.TwoFactorSecretVerifiedAt
	firstUser.LastUpdatedAt = u.LastUpdatedAt
	assert.Equal(t, firstUser, u)

	// archive
	for _, user := range createdUsers {
		assert.NoError(t, dbc.ArchiveUser(ctx, user.ID))

		var y *identity.User
		y, err = dbc.GetUser(ctx, user.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}

	// delete
	for _, user := range createdUsers {
		assert.NoError(t, dbc.DeleteUser(ctx, user.ID))

		var y *filtering.QueryFilteredResult[identity.Account]
		y, err = dbc.GetAccounts(ctx, user.ID, nil)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUser(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserWithUnverifiedTwoFactorSecret(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetUserByEmail(T *testing.T) {
	T.Parallel()

	T.Run("with invalid email", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserByEmail(ctx, "")
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}

func TestQuerier_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserByUsername(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetAdminUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("with empty username", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetAdminUserByUsername(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForUsersByUsername(T *testing.T) {
	T.Parallel()

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.SearchForUsersByUsername(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_MarkUserAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkUserAsIndexed(ctx, ""))
	})
}

func TestQuerier_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateUser(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateUsername(T *testing.T) {
	T.Parallel()

	T.Run("with empty user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserUsername(ctx, "", t.Name()))
	})
}

func TestQuerier_UpdateUserDetails(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserDetails(ctx, exampleUser.ID, nil))
	})
}

func TestQuerier_SetUserAvatar(T *testing.T) {
	T.Parallel()

	T.Run("with empty uploaded media ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.SetUserAvatar(ctx, exampleUser.ID, ""))
	})
}

func TestQuerier_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleHashedPassword := "$argon2i$v=19$m=64,t=10,p=4$RjFtMmRmU2lGYU9CMk1mMw$cuGR9AhTczPR6xDOSAMW+SvEYFyLEIS+7nlRdC9f6ys"

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserPassword(ctx, "", exampleHashedPassword))
	})

	T.Run("with invalid new hash", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserPassword(ctx, exampleUser.ID, ""))
	})
}

func TestQuerier_UpdateUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, "", exampleUser.TwoFactorSecret))
	})

	T.Run("with invalid new secret", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserTwoFactorSecret(ctx, exampleUser.ID, ""))
	})
}

func TestQuerier_MarkUserTwoFactorSecretAsVerified(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkUserTwoFactorSecretAsVerified(ctx, ""))
	})
}

func TestQuerier_MarkUserTwoFactorSecretAsUnverified(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleSecret := fakes.BuildFakeID()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkUserTwoFactorSecretAsUnverified(ctx, "", exampleSecret))
	})

	T.Run("with invalid secret", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUserID := fakes.BuildFakeID()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkUserTwoFactorSecretAsUnverified(ctx, exampleUserID, ""))
	})
}

func TestQuerier_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveUser(ctx, ""))
	})

	T.Run("with error beginning transaction", func(t *testing.T) {
		t.Parallel()

		exampleUserID := fakes.BuildFakeID()

		ctx := t.Context()
		c, db := buildMockSQLTestClient(t)

		db.ExpectBegin().WillReturnError(errors.New("blah"))

		assert.Error(t, c.ArchiveUser(ctx, exampleUserID))
	})
}

func TestQuerier_GetUserByEmailAddressVerificationToken(T *testing.T) {
	T.Parallel()

	T.Run("with missing token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserByEmailAddressVerificationToken(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_MarkUserEmailAddressAsVerified(T *testing.T) {
	T.Parallel()

	T.Run("with missing user ID", func(t *testing.T) {
		t.Parallel()

		exampleInput := authfakes.BuildFakeEmailAddressVerificationRequestInput()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.MarkUserEmailAddressAsVerified(ctx, "", exampleInput.Token)
		assert.Error(t, err)
	})

	T.Run("with missing token", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.MarkUserEmailAddressAsVerified(ctx, exampleUser.ID, "")
		assert.Error(t, err)
	})
}
