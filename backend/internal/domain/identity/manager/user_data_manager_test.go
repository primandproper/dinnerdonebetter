package manager

import (
	"context"
	"strings"
	"testing"

	mockauthn "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/mock"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v8/filtering"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"
	"github.com/primandproper/platform-go/v8/qrcodes"
	randommock "github.com/primandproper/platform-go/v8/random/mock"
	mocksearch "github.com/primandproper/platform-go/v8/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildIdentityDataManagerForTest(t *testing.T) *manager {
	t.Helper()

	ctx := t.Context()

	m, err := NewIdentityDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&identitymock.RepositoryMock{},
		&randommock.GeneratorMock{},
		&mockauthn.AuthenticatorMock{},
		&mocksearch.IndexMock[identityindexing.UserSearchSubset]{},
		qrcodes.NewBuilder(qrcodes.Issuer("test"), qrcodes.WithTracerProvider(tracingnoop.NewTracerProvider()), qrcodes.WithLogger(loggingnoop.NewLogger())),
	)
	require.NoError(t, err)

	return m.(*manager)
}

// attachRepositoryToIdentityDataManager wires a configured repository mock and a no-op data changes
// publisher into the manager under test.
func attachRepositoryToIdentityDataManager(m *manager, db *identitymock.RepositoryMock) {
	attachMocksToIdentityDataManager(m, db, nil, nil, nil)
}

// attachMocksToIdentityDataManager wires configured collaborator mocks and a no-op data changes
// publisher into the manager under test. A nil argument gets an unconfigured mock, which panics if
// any of its methods are called.
func attachMocksToIdentityDataManager(
	m *manager,
	db *identitymock.RepositoryMock,
	secretGenerator *randommock.GeneratorMock,
	authenticator *mockauthn.AuthenticatorMock,
	searchIndex *mocksearch.IndexMock[identityindexing.UserSearchSubset],
) {
	if db == nil {
		db = &identitymock.RepositoryMock{}
	}
	m.identityRepo = db

	if secretGenerator == nil {
		secretGenerator = &randommock.GeneratorMock{}
	}
	m.secretGenerator = secretGenerator

	if authenticator == nil {
		authenticator = &mockauthn.AuthenticatorMock{}
	}
	m.authenticator = authenticator

	if searchIndex == nil {
		searchIndex = &mocksearch.IndexMock[identityindexing.UserSearchSubset]{}
	}
	m.userSearchIndex = searchIndex
}

func TestIdentityDataManager_AcceptAccountInvitation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		accountInvitationID := fakes.BuildFakeID()
		input := fakes.BuildFakeAccountInvitationUpdateRequestInput()
		invitation := fakes.BuildFakeAccountInvitation()
		invitation.ID = accountInvitationID
		invitation.Token = input.Token

		db := &identitymock.RepositoryMock{
			GetAccountInvitationByTokenAndIDFunc: func(_ context.Context, token, invitationID string) (*identity.AccountInvitation, error) {
				assert.Equal(t, input.Token, token)
				assert.Equal(t, accountInvitationID, invitationID)
				return invitation, nil
			},
			AcceptAccountInvitationFunc: func(_ context.Context, actualAccountID, invitationID, token, note string) error {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, accountInvitationID, invitationID)
				assert.Equal(t, input.Token, token)
				assert.Equal(t, input.Note, note)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.AcceptAccountInvitation(ctx, accountID, accountInvitationID, input)
		assert.NoError(t, err)

		assert.Len(t, db.GetAccountInvitationByTokenAndIDCalls(), 1)
		assert.Len(t, db.AcceptAccountInvitationCalls(), 1)
	})
}

func TestIdentityDataManager_RejectAccountInvitation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		accountInvitationID := fakes.BuildFakeID()
		input := fakes.BuildFakeAccountInvitationUpdateRequestInput()
		invitation := fakes.BuildFakeAccountInvitation()
		invitation.ID = accountInvitationID
		invitation.Token = input.Token

		db := &identitymock.RepositoryMock{
			GetAccountInvitationByTokenAndIDFunc: func(_ context.Context, token, invitationID string) (*identity.AccountInvitation, error) {
				assert.Equal(t, input.Token, token)
				assert.Equal(t, accountInvitationID, invitationID)
				return invitation, nil
			},
			RejectAccountInvitationFunc: func(_ context.Context, actualAccountID, invitationID, note string) error {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, invitation.ID, invitationID)
				assert.Equal(t, input.Note, note)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.RejectAccountInvitation(ctx, accountID, accountInvitationID, input)
		assert.NoError(t, err)

		assert.Len(t, db.GetAccountInvitationByTokenAndIDCalls(), 1)
		assert.Len(t, db.RejectAccountInvitationCalls(), 1)
	})
}

func TestIdentityDataManager_CancelAccountInvitation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		accountInvitationID := fakes.BuildFakeID()
		note := "test note"

		db := &identitymock.RepositoryMock{
			CancelAccountInvitationFunc: func(_ context.Context, actualAccountID, invitationID, actualNote string) error {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, accountInvitationID, invitationID)
				assert.Equal(t, note, actualNote)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.CancelAccountInvitation(ctx, accountID, accountInvitationID, note)
		assert.NoError(t, err)

		assert.Len(t, db.CancelAccountInvitationCalls(), 1)
	})
}

func TestIdentityDataManager_ArchiveAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		ownerID := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			ArchiveAccountFunc: func(_ context.Context, actualAccountID, userID string) error {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, ownerID, userID)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.ArchiveAccount(ctx, accountID, ownerID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveAccountCalls(), 1)
	})
}

func TestIdentityDataManager_ArchiveUserMembership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		accountID := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			RemoveUserFromAccountFunc: func(_ context.Context, actualUserID, actualAccountID string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, accountID, actualAccountID)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.ArchiveUserMembership(ctx, userID, accountID)
		assert.NoError(t, err)

		assert.Len(t, db.RemoveUserFromAccountCalls(), 1)
	})
}

func TestIdentityDataManager_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			ArchiveUserFunc: func(_ context.Context, actualUserID string) error {
				assert.Equal(t, userID, actualUserID)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.ArchiveUser(ctx, userID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveUserCalls(), 1)
	})
}

func TestIdentityDataManager_CreateAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		input := fakes.BuildFakeAccountCreationRequestInput()
		expected := fakes.BuildFakeAccount()

		db := &identitymock.RepositoryMock{
			CreateAccountFunc: func(_ context.Context, dbInput *identity.AccountDatabaseCreationInput) (*identity.Account, error) {
				assert.NotNil(t, dbInput)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.CreateAccount(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateAccountCalls(), 1)
	})
}

func TestIdentityDataManager_CreateAccountInvitation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		accountID := fakes.BuildFakeID()
		input := fakes.BuildFakeAccountInvitationCreationRequestInput()
		expected := fakes.BuildFakeAccountInvitation()
		token := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			CreateAccountInvitationFunc: func(_ context.Context, dbInput *identity.AccountInvitationDatabaseCreationInput) (*identity.AccountInvitation, error) {
				assert.NotNil(t, dbInput)
				return expected, nil
			},
		}
		secretGenerator := &randommock.GeneratorMock{
			GenerateBase64EncodedStringFunc: func(_ context.Context, _ int) (string, error) {
				return token, nil
			},
		}
		attachMocksToIdentityDataManager(m, db, secretGenerator, nil, nil)

		actual, err := m.CreateAccountInvitation(ctx, userID, accountID, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateAccountInvitationCalls(), 1)
		assert.Len(t, secretGenerator.GenerateBase64EncodedStringCalls(), 1)
	})
}

func TestIdentityDataManager_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		input := fakes.BuildFakeUserRegistrationInput()
		expected := fakes.BuildFakeUser()
		hashedPassword := "hashed-password"
		twoFactorSecret := "two-factor-secret"
		defaultAccountID := fakes.BuildFakeID()
		emailVerificationToken := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			CreateUserFunc: func(_ context.Context, dbInput *identity.UserDatabaseCreationInput) (*identity.User, error) {
				assert.NotNil(t, dbInput)
				return expected, nil
			},
			GetDefaultAccountIDForUserFunc: func(_ context.Context, userID string) (string, error) {
				assert.Equal(t, expected.ID, userID)
				return defaultAccountID, nil
			},
			GetEmailAddressVerificationTokenForUserFunc: func(_ context.Context, userID string) (string, error) {
				assert.Equal(t, expected.ID, userID)
				return emailVerificationToken, nil
			},
		}
		secretGenerator := &randommock.GeneratorMock{
			GenerateBase32EncodedStringFunc: func(_ context.Context, _ int) (string, error) {
				return twoFactorSecret, nil
			},
		}
		authenticator := &mockauthn.AuthenticatorMock{
			HashPasswordFunc: func(_ context.Context, password string) (string, error) {
				assert.NotEmpty(t, password)
				return hashedPassword, nil
			},
		}
		attachMocksToIdentityDataManager(m, db, secretGenerator, authenticator, nil)

		actual, err := m.CreateUser(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Equal(t, expected.ID, actual.CreatedUserID)
		assert.True(t, strings.HasPrefix(actual.TwoFactorQRCode, "data:image/png;base64,"), "two factor QR code should be a PNG data URI")

		assert.Len(t, db.CreateUserCalls(), 1)
		// The default account ID and verification token are no longer re-fetched here: the
		// repository names both from inside the transaction that created them, and emits the
		// signup event there. Two round trips per signup went away with the publish.
		assert.Empty(t, db.GetDefaultAccountIDForUserCalls())
		assert.Empty(t, db.GetEmailAddressVerificationTokenForUserCalls())
		assert.Len(t, secretGenerator.GenerateBase32EncodedStringCalls(), 1)
		assert.Len(t, authenticator.HashPasswordCalls(), 1)
	})
}

func TestIdentityDataManager_GetAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccount()

		db := &identitymock.RepositoryMock{
			GetAccountFunc: func(_ context.Context, actualAccountID string) (*identity.Account, error) {
				assert.Equal(t, accountID, actualAccountID)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetAccount(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetAccountCalls(), 1)
	})
}

func TestIdentityDataManager_GetAccountInvitation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		accountInvitationID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountInvitation()

		db := &identitymock.RepositoryMock{
			GetAccountInvitationByAccountAndIDFunc: func(_ context.Context, actualAccountID, invitationID string) (*identity.AccountInvitation, error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, accountInvitationID, invitationID)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetAccountInvitation(ctx, accountID, accountInvitationID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetAccountInvitationByAccountAndIDCalls(), 1)
	})
}

func TestIdentityDataManager_GetAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountsList()

		db := &identitymock.RepositoryMock{
			GetAccountsFunc: func(_ context.Context, actualUserID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
				assert.Equal(t, userID, actualUserID)
				assert.NotNil(t, filter)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetAccounts(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetAccountsCalls(), 1)
	})
}

func TestIdentityDataManager_GetReceivedAccountInvitations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountInvitationsList()

		db := &identitymock.RepositoryMock{
			GetPendingAccountInvitationsForUserFunc: func(_ context.Context, actualUserID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
				assert.Equal(t, userID, actualUserID)
				assert.NotNil(t, filter)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetReceivedAccountInvitations(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetPendingAccountInvitationsForUserCalls(), 1)
	})
}

func TestIdentityDataManager_GetSentAccountInvitations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountInvitationsList()

		db := &identitymock.RepositoryMock{
			GetPendingAccountInvitationsFromUserFunc: func(_ context.Context, actualUserID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
				assert.Equal(t, userID, actualUserID)
				assert.NotNil(t, filter)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetSentAccountInvitations(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetPendingAccountInvitationsFromUserCalls(), 1)
	})
}

func TestIdentityDataManager_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		expected := fakes.BuildFakeUser()

		db := &identitymock.RepositoryMock{
			GetUserFunc: func(_ context.Context, actualUserID string) (*identity.User, error) {
				assert.Equal(t, userID, actualUserID)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetUser(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetUserCalls(), 1)
	})
}

func TestIdentityDataManager_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		expected := fakes.BuildFakeUsersList()

		db := &identitymock.RepositoryMock{
			GetUsersFunc: func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
				assert.NotNil(t, filter)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetUsers(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetUsersCalls(), 1)
	})
}

func TestIdentityDataManager_GetUsersForAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		expected := fakes.BuildFakeUsersList()

		db := &identitymock.RepositoryMock{
			GetUsersForAccountFunc: func(_ context.Context, actualAccountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.NotNil(t, filter)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.GetUsersForAccount(ctx, accountID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetUsersForAccountCalls(), 1)
	})
}

func TestIdentityDataManager_SearchForUsers(T *testing.T) {
	T.Parallel()

	T.Run("with database search", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		query := "test-query"
		expected := fakes.BuildFakeUsersList()

		db := &identitymock.RepositoryMock{
			SearchForUsersByUsernameFunc: func(_ context.Context, usernameQuery string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
				assert.Equal(t, query, usernameQuery)
				assert.NotNil(t, filter)
				return expected, nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		actual, err := m.SearchForUsers(ctx, query, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForUsersByUsernameCalls(), 1)
	})

	T.Run("with search service", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		query := "test-query"
		searchResults := []*identityindexing.UserSearchSubset{
			{ID: fakes.BuildFakeID()},
			{ID: fakes.BuildFakeID()},
		}
		users := []*identity.User{
			fakes.BuildFakeUser(),
			fakes.BuildFakeUser(),
		}
		users[0].ID = searchResults[0].ID
		users[1].ID = searchResults[1].ID

		db := &identitymock.RepositoryMock{
			GetUsersWithIDsFunc: func(_ context.Context, ids []string) ([]*identity.User, error) {
				assert.Equal(t, []string{searchResults[0].ID, searchResults[1].ID}, ids)
				return users, nil
			},
		}
		searchIndex := &mocksearch.IndexMock[identityindexing.UserSearchSubset]{
			SearchFunc: func(_ context.Context, q string) ([]*identityindexing.UserSearchSubset, error) {
				assert.Equal(t, query, q)
				return searchResults, nil
			},
		}
		attachMocksToIdentityDataManager(m, db, nil, nil, searchIndex)

		actual, err := m.SearchForUsers(ctx, query, true, nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Len(t, actual.Data, 2)

		assert.Len(t, db.GetUsersWithIDsCalls(), 1)
		assert.Len(t, searchIndex.SearchCalls(), 1)
	})

	T.Run("with search service honoring page size and reporting total hits", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		query := "test-query"
		searchResults := []*identityindexing.UserSearchSubset{
			{ID: fakes.BuildFakeID()},
			{ID: fakes.BuildFakeID()},
			{ID: fakes.BuildFakeID()},
		}
		users := []*identity.User{
			fakes.BuildFakeUser(),
			fakes.BuildFakeUser(),
		}
		users[0].ID = searchResults[0].ID
		users[1].ID = searchResults[1].ID

		filter := filtering.DefaultQueryFilter()
		filter.MaxResponseSize = new(uint8(2))

		db := &identitymock.RepositoryMock{
			GetUsersWithIDsFunc: func(_ context.Context, ids []string) ([]*identity.User, error) {
				assert.Equal(t, []string{searchResults[0].ID, searchResults[1].ID}, ids)
				return users, nil
			},
		}
		searchIndex := &mocksearch.IndexMock[identityindexing.UserSearchSubset]{
			SearchFunc: func(_ context.Context, q string) ([]*identityindexing.UserSearchSubset, error) {
				assert.Equal(t, query, q)
				return searchResults, nil
			},
		}
		attachMocksToIdentityDataManager(m, db, nil, nil, searchIndex)

		actual, err := m.SearchForUsers(ctx, query, true, filter)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Len(t, actual.Data, 2)
		assert.Equal(t, uint64(2), actual.FilteredCount)
		assert.Equal(t, uint64(3), actual.TotalCount)

		assert.Len(t, db.GetUsersWithIDsCalls(), 1)
		assert.Len(t, searchIndex.SearchCalls(), 1)
	})
}

func TestIdentityDataManager_SetDefaultAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		accountID := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			MarkAccountAsUserDefaultFunc: func(_ context.Context, actualUserID, actualAccountID string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, accountID, actualAccountID)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.SetDefaultAccount(ctx, userID, accountID)
		assert.NoError(t, err)

		assert.Len(t, db.MarkAccountAsUserDefaultCalls(), 1)
	})
}

func TestIdentityDataManager_TransferAccountOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		input := fakes.BuildFakeAccountOwnershipTransferInput()

		db := &identitymock.RepositoryMock{
			TransferAccountOwnershipFunc: func(_ context.Context, actualAccountID string, actualInput *identity.AccountOwnershipTransferInput) error {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, input, actualInput)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.TransferAccountOwnership(ctx, accountID, input)
		assert.NoError(t, err)

		assert.Len(t, db.TransferAccountOwnershipCalls(), 1)
	})
}

func TestIdentityDataManager_UpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		accountID := fakes.BuildFakeID()
		input := fakes.BuildFakeAccountUpdateRequestInput()
		account := fakes.BuildFakeAccount()
		account.ID = accountID

		db := &identitymock.RepositoryMock{
			GetAccountFunc: func(_ context.Context, actualAccountID string) (*identity.Account, error) {
				assert.Equal(t, accountID, actualAccountID)
				return account, nil
			},
			UpdateAccountFunc: func(_ context.Context, updated *identity.Account) error {
				assert.NotNil(t, updated)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.UpdateAccount(ctx, accountID, input)
		assert.NoError(t, err)

		assert.Len(t, db.GetAccountCalls(), 1)
		assert.Len(t, db.UpdateAccountCalls(), 1)
	})
}

func TestIdentityDataManager_UpdateAccountMemberPermissions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		accountID := fakes.BuildFakeID()
		input := fakes.BuildFakeUserPermissionModificationInput()

		db := &identitymock.RepositoryMock{
			ModifyUserPermissionsFunc: func(_ context.Context, actualAccountID, actualUserID string, actualInput *identity.ModifyUserPermissionsInput) error {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, input, actualInput)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.UpdateAccountMemberPermissions(ctx, accountID, userID, input)
		assert.NoError(t, err)

		assert.Len(t, db.ModifyUserPermissionsCalls(), 1)
	})
}

func TestIdentityDataManager_UpdateUserDetails(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		input := fakes.BuildFakeUserDetailsUpdateRequestInput()

		db := &identitymock.RepositoryMock{
			UpdateUserDetailsFunc: func(_ context.Context, actualUserID string, dbInput *identity.UserDetailsDatabaseUpdateInput) error {
				assert.Equal(t, userID, actualUserID)
				assert.NotNil(t, dbInput)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.UpdateUserDetails(ctx, userID, input)
		assert.NoError(t, err)

		assert.Len(t, db.UpdateUserDetailsCalls(), 1)
	})
}

func TestIdentityDataManager_UpdateUserEmailAddress(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		newEmail := "newemail@example.com"

		db := &identitymock.RepositoryMock{
			UpdateUserEmailAddressFunc: func(_ context.Context, actualUserID, newEmailAddress string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, newEmail, newEmailAddress)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.UpdateUserEmailAddress(ctx, userID, newEmail)
		assert.NoError(t, err)

		assert.Len(t, db.UpdateUserEmailAddressCalls(), 1)
	})
}

func TestIdentityDataManager_UpdateUserUsername(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		newUsername := "newusername"

		db := &identitymock.RepositoryMock{
			UpdateUserUsernameFunc: func(_ context.Context, actualUserID, actualNewUsername string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, newUsername, actualNewUsername)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.UpdateUserUsername(ctx, userID, newUsername)
		assert.NoError(t, err)

		assert.Len(t, db.UpdateUserUsernameCalls(), 1)
	})
}

func TestIdentityDataManager_SetUserAvatar(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()
		uploadedMediaID := fakes.BuildFakeID()

		db := &identitymock.RepositoryMock{
			SetUserAvatarFunc: func(_ context.Context, actualUserID, actualUploadedMediaID string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, uploadedMediaID, actualUploadedMediaID)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.SetUserAvatar(ctx, userID, uploadedMediaID)
		assert.NoError(t, err)

		assert.Len(t, db.SetUserAvatarCalls(), 1)
	})
}

func TestIdentityDataManager_AdminUpdateUserStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		input := &identity.UserAccountStatusUpdateInput{
			TargetUserID: fakes.BuildFakeID(),
			Reason:       "test reason",
		}

		db := &identitymock.RepositoryMock{
			UpdateUserAccountStatusFunc: func(_ context.Context, userID string, actualInput *identity.UserAccountStatusUpdateInput) error {
				assert.Equal(t, input.TargetUserID, userID)
				assert.Equal(t, input, actualInput)
				return nil
			},
		}
		attachRepositoryToIdentityDataManager(m, db)

		err := m.AdminUpdateUserStatus(ctx, input)
		assert.NoError(t, err)

		assert.Len(t, db.UpdateUserAccountStatusCalls(), 1)
	})
}
