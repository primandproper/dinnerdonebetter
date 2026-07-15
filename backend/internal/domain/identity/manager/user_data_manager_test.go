package manager

import (
	"context"
	"testing"

	mockauthn "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	identitymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/mock"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v4/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v4/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v4/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v4/observability/tracing/noop"
	randommock "github.com/primandproper/platform-go/v4/random/mock"
	"github.com/primandproper/platform-go/v4/reflection"
	mocksearch "github.com/primandproper/platform-go/v4/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildIdentityDataManagerForTest(t *testing.T) *manager {
	t.Helper()

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{
		DataChangesTopicName: t.Name(),
	}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{}, nil
		},
	}

	m, err := NewIdentityDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		mpp,
		&identitymock.RepositoryMock{},
		&randommock.GeneratorMock{},
		&mockauthn.Authenticator{},
		&mocksearch.IndexMock[identityindexing.UserSearchSubset]{},
		queueCfg,
	)
	require.NoError(t, err)

	return m.(*manager)
}

func setupExpectationsForIdentityDataManager(
	manager *manager,
	dbSetupFunc func(db *identitymock.RepositoryMock),
	publisherSetupFunc func(mp *mockpublishers.PublisherMock),
	secretGenSetupFunc func(sg *randommock.GeneratorMock),
	authSetupFunc func(auth *mockauthn.Authenticator),
	searchSetupFunc func(us *mocksearch.IndexMock[identityindexing.UserSearchSubset]),
	eventTypeMaps ...map[string][]string,
) []any {
	db := &identitymock.RepositoryMock{}
	if dbSetupFunc != nil {
		dbSetupFunc(db)
	}
	manager.identityRepo = db

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	if publisherSetupFunc != nil {
		publisherSetupFunc(mp)
	}
	manager.dataChangesPublisher = mp

	sg := &randommock.GeneratorMock{}
	if secretGenSetupFunc != nil {
		secretGenSetupFunc(sg)
	}
	manager.secretGenerator = sg

	auth := &mockauthn.Authenticator{}
	if authSetupFunc != nil {
		authSetupFunc(auth)
	}
	manager.authenticator = auth

	us := &mocksearch.IndexMock[identityindexing.UserSearchSubset]{}
	if searchSetupFunc != nil {
		searchSetupFunc(us)
	}
	manager.userSearchIndex = us

	return []any{db, auth}
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetAccountInvitationByTokenAndID), testutils.ContextMatcher, input.Token, accountInvitationID).Return(invitation, nil)
				db.On(reflection.GetMethodName(m.identityRepo.AcceptAccountInvitation), testutils.ContextMatcher, accountID, accountInvitationID, input.Token, input.Note).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountInvitationAcceptedServiceEventType: {identitykeys.AccountInvitationIDKey, "destination_account"},
			},
		)

		err := m.AcceptAccountInvitation(ctx, accountID, accountInvitationID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetAccountInvitationByTokenAndID), testutils.ContextMatcher, input.Token, accountInvitationID).Return(invitation, nil)
				db.On(reflection.GetMethodName(m.identityRepo.RejectAccountInvitation), testutils.ContextMatcher, accountID, invitation.ID, input.Note).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountInvitationRejectedServiceEventType: {identitykeys.AccountInvitationIDKey},
			},
		)

		err := m.RejectAccountInvitation(ctx, accountID, accountInvitationID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.CancelAccountInvitation), testutils.ContextMatcher, accountID, accountInvitationID, note).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountInvitationCanceledServiceEventType: {identitykeys.AccountInvitationIDKey},
			},
		)

		err := m.CancelAccountInvitation(ctx, accountID, accountInvitationID, note)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.ArchiveAccount), testutils.ContextMatcher, accountID, ownerID).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountArchivedServiceEventType: {identitykeys.AccountIDKey, identitykeys.UserIDKey},
			},
		)

		err := m.ArchiveAccount(ctx, accountID, ownerID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.RemoveUserFromAccount), testutils.ContextMatcher, userID, accountID).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountMemberRemovedServiceEventType: {identitykeys.AccountIDKey, identitykeys.UserIDKey},
			},
		)

		err := m.ArchiveUserMembership(ctx, userID, accountID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestIdentityDataManager_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		userID := fakes.BuildFakeID()

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.ArchiveUser), testutils.ContextMatcher, userID).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.UserArchivedServiceEventType: {identitykeys.UserIDKey},
			},
		)

		err := m.ArchiveUser(ctx, userID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.CreateAccount), testutils.ContextMatcher, testutils.MatchType[*identity.AccountDatabaseCreationInput]()).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountCreatedServiceEventType: {identitykeys.AccountIDKey},
			},
		)

		actual, err := m.CreateAccount(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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
		token := "test-token"

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.CreateAccountInvitation), testutils.ContextMatcher, testutils.MatchType[*identity.AccountInvitationDatabaseCreationInput]()).Return(expected, nil)
			},
			nil,
			func(sg *randommock.GeneratorMock) {
				sg.GenerateBase64EncodedStringFunc = func(_ context.Context, _ int) (string, error) {
					return token, nil
				}
			},
			nil,
			nil,
			map[string][]string{
				identity.AccountInvitationCreatedServiceEventType: {identitykeys.AccountInvitationIDKey, identitykeys.UserIDKey, "destination_account"},
			},
		)

		actual, err := m.CreateAccountInvitation(ctx, userID, accountID, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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
		emailVerificationToken := "email-verification-token"

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.CreateUser), testutils.ContextMatcher, testutils.MatchType[*identity.UserDatabaseCreationInput]()).Return(expected, nil)
				db.On(reflection.GetMethodName(m.identityRepo.GetDefaultAccountIDForUser), testutils.ContextMatcher, expected.ID).Return(defaultAccountID, nil)
				db.On(reflection.GetMethodName(m.identityRepo.GetEmailAddressVerificationTokenForUser), testutils.ContextMatcher, expected.ID).Return(emailVerificationToken, nil)
			},
			nil,
			func(sg *randommock.GeneratorMock) {
				sg.GenerateBase32EncodedStringFunc = func(_ context.Context, _ int) (string, error) {
					return twoFactorSecret, nil
				}
			},
			func(auth *mockauthn.Authenticator) {
				auth.On(reflection.GetMethodName(m.authenticator.HashPassword), testutils.ContextMatcher, mock.AnythingOfType("string")).Return(hashedPassword, nil)
			},
			nil,
			map[string][]string{
				identity.UserSignedUpServiceEventType: {identitykeys.AccountIDKey, identitykeys.UserIDKey, identitykeys.UserEmailVerificationTokenKey},
			},
		)

		actual, err := m.CreateUser(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Equal(t, expected.ID, actual.CreatedUserID)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetAccount), testutils.ContextMatcher, accountID).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetAccount(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetAccountInvitationByAccountAndID), testutils.ContextMatcher, accountID, accountInvitationID).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetAccountInvitation(ctx, accountID, accountInvitationID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetAccounts), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetAccounts(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetPendingAccountInvitationsForUser), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetReceivedAccountInvitations(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetPendingAccountInvitationsFromUser), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetSentAccountInvitations(ctx, userID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetUser), testutils.ContextMatcher, userID).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetUser(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestIdentityDataManager_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := buildIdentityDataManagerForTest(t)

		expected := fakes.BuildFakeUsersList()

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetUsers), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetUsers(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetUsersForAccount), testutils.ContextMatcher, accountID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.GetUsersForAccount(ctx, accountID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.SearchForUsersByUsername), testutils.ContextMatcher, query, testutils.QueryFilterMatcher).Return(expected, nil)
			},
			nil,
			nil,
			nil,
			nil,
		)

		actual, err := m.SearchForUsers(ctx, query, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetUsersWithIDs), testutils.ContextMatcher, mock.AnythingOfType("[]string")).Return(users, nil)
			},
			nil,
			nil,
			nil,
			func(us *mocksearch.IndexMock[identityindexing.UserSearchSubset]) {
				us.SearchFunc = func(_ context.Context, q string) ([]*identityindexing.UserSearchSubset, error) {
					return searchResults, nil
				}
			},
		)

		actual, err := m.SearchForUsers(ctx, query, true, nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Len(t, actual.Data, 2)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.MarkAccountAsUserDefault), testutils.ContextMatcher, userID, accountID).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountSetAsDefaultServiceEventType: {identitykeys.AccountIDKey, identitykeys.UserIDKey},
			},
		)

		err := m.SetDefaultAccount(ctx, userID, accountID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.TransferAccountOwnership), testutils.ContextMatcher, accountID, input).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountOwnershipTransferredServiceEventType: {identitykeys.AccountIDKey},
			},
		)

		err := m.TransferAccountOwnership(ctx, accountID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.GetAccount), testutils.ContextMatcher, accountID).Return(account, nil)
				db.On(reflection.GetMethodName(m.identityRepo.UpdateAccount), testutils.ContextMatcher, testutils.MatchType[*identity.Account]()).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountUpdatedServiceEventType: {identitykeys.AccountIDKey},
			},
		)

		err := m.UpdateAccount(ctx, accountID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.ModifyUserPermissions), testutils.ContextMatcher, accountID, userID, input).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.AccountMembershipPermissionsUpdatedServiceEventType: {identitykeys.AccountIDKey},
			},
		)

		err := m.UpdateAccountMemberPermissions(ctx, userID, accountID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.UpdateUserDetails), testutils.ContextMatcher, userID, testutils.MatchType[*identity.UserDetailsDatabaseUpdateInput]()).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.UserDetailsChangedEventType: {identitykeys.UserIDKey},
			},
		)

		err := m.UpdateUserDetails(ctx, userID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.UpdateUserEmailAddress), testutils.ContextMatcher, userID, newEmail).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.EmailAddressChangedEventType: {identitykeys.UserIDKey},
			},
		)

		err := m.UpdateUserEmailAddress(ctx, userID, newEmail)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.UpdateUserUsername), testutils.ContextMatcher, userID, newUsername).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.UsernameChangedEventType: {identitykeys.UserIDKey},
			},
		)

		err := m.UpdateUserUsername(ctx, userID, newUsername)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.SetUserAvatar), testutils.ContextMatcher, userID, uploadedMediaID).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.UserAvatarChangedEventType: {identitykeys.UserIDKey},
			},
		)

		err := m.SetUserAvatar(ctx, userID, uploadedMediaID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForIdentityDataManager(
			m,
			func(db *identitymock.RepositoryMock) {
				db.On(reflection.GetMethodName(m.identityRepo.UpdateUserAccountStatus), testutils.ContextMatcher, input.TargetUserID, input).Return(nil)
			},
			nil,
			nil,
			nil,
			nil,
			map[string][]string{
				identity.UserStatusChangedServiceEventType: {identitykeys.UserIDKey},
			},
		)

		err := m.AdminUpdateUserStatus(ctx, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
