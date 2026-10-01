package integration

import (
	"testing"

	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	authconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/converters"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/identity/identitypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "users", RelevantID: user.ID},
			{EventType: "created", ResourceType: "accounts"},
			{EventType: "created", ResourceType: "account_user_memberships"},
		})
	})

	T.Run("rejects duplicate registration", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.RegisterUser(ctx, &authsvc.RegisterUserRequest{
			Input: authconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input),
		})
		require.NoError(t, err)

		_, err = testClient.RegisterUser(ctx, &authsvc.RegisterUserRequest{
			Input: authconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input),
		})
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		input.Username = ""

		testClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := testClient.RegisterUser(ctx, &authsvc.RegisterUserRequest{
			Input: authconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input),
		})
		assert.Error(t, err)
	})

	// The agreements are required by this application rather than by the directory, which
	// records when somebody last agreed and has no opinion about whether they had to. This
	// is the only place that reading is enforced, so it is the only place it can be checked.
	T.Run("refuses a registration that declined the terms", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		input.AcceptedTOS = false

		testClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := testClient.RegisterUser(ctx, &authsvc.RegisterUserRequest{
			Input: authconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input),
		})
		assert.Error(t, err)
	})
}

func TestUsers_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		_, err := adminClient.IdentityService().ArchiveUser(ctx, &identitypb.ArchiveUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)

		// By resource rather than by actor, because the two entries have different actors:
		// the registration is the user's own act and the archival is the administrator's.
		// Filtering on either one would assert half of what this is checking.
		AssertAuditLogContainsFuzzyForResource(t, ctx, "users", user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "users", RelevantID: user.ID},
			{EventType: "archived", ResourceType: "users", RelevantID: user.ID},
		})
	})

	// platform refuses to archive somebody who still owns an account. This application
	// settles their households first (internal/build/identity/archival.go), because every
	// registrant owns one: a household with other members goes to the longest-tenured of
	// them, and one the owner was alone in is archived beside them.
	T.Run("a household with members passes to the longest-tenured of them", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		owner, ownerClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, ownerClient)

		joinForTest := func() *identity.User {
			input := buildUserRegistrationInputForTest(t)
			member, memberClient := createUserAndClientForTestWithRegistrationInput(t, input)
			invitation := inviteForTest(t, owner.ID, accountID, input.EmailAddress)

			_, err := memberClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
				InvitationId: invitation.ID,
				Token:        invitation.Token,
			})
			require.NoError(t, err)

			return member
		}

		senior := joinForTest()
		_ = joinForTest()

		_, err := adminClient.IdentityService().ArchiveUser(ctx, &identitypb.ArchiveUserRequest{UserId: owner.ID})
		require.NoError(t, err)

		account, err := adminClient.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: accountID})
		require.NoError(t, err)
		assert.Equal(t, senior.ID, account.GetAccount().GetOwnerUserId())
		assert.Nil(t, account.GetAccount().GetArchivedAt())
	})

	T.Run("a household its owner was alone in is archived beside them", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		owner, ownerClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, ownerClient)

		_, err := adminClient.IdentityService().ArchiveUser(ctx, &identitypb.ArchiveUserRequest{UserId: owner.ID})
		require.NoError(t, err)

		_, err = adminClient.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: accountID})
		require.Error(t, err)

		AssertAuditLogContainsFuzzyForResource(t, ctx, "accounts", accountID, 10, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "accounts", RelevantID: accountID},
		})
	})
}

// A registration that answers an invitation joins the inviter's account rather than
// minting one of its own, which is the shape platform's RegisterWithInvitation has and the
// reason it is a second operation rather than a flag.
func TestUsers_RegisteringAgainstAnInvitation(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		inviter, inviterClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, inviterClient)

		registrant := buildUserRegistrationInputForTest(t)

		invitation := inviteForTest(t, selfIDForTest(t, inviterClient), accountID, registrant.EmailAddress)
		require.NotEmpty(t, invitation.Token)

		registrant.InvitationID = invitation.ID
		registrant.InvitationToken = invitation.Token

		created := createServiceUserForTest(t, registrant)

		// The account they landed in is the inviter's, not one of their own.
		accounts, err := adminClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: created.ID,
		})
		require.NoError(t, err)
		require.Len(t, accounts.GetResults(), 1)
		assert.Equal(t, accountID, accounts.GetResults()[0].GetId())
		assert.Equal(t, inviter.ID, accounts.GetResults()[0].GetOwnerUserId())
	})

	// The user and the invitation's answer are one transaction, so an invitation that no
	// longer admits the registrant takes the registration down with it rather than leaving
	// a user committed against a dead link.
	T.Run("a bad token registers nobody", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, inviterClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, inviterClient)

		registrant := buildUserRegistrationInputForTest(t)

		invitation := inviteForTest(t, selfIDForTest(t, inviterClient), accountID, registrant.EmailAddress)

		registrant.InvitationID = invitation.ID
		registrant.InvitationToken = "not the token"

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.RegisterUser(ctx, &authsvc.RegisterUserRequest{
			Input: authconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(registrant),
		})
		require.Error(t, err)

		// And the registrant is not in the directory: the whole transaction rolled back.
		users, err := adminClient.IdentityService().SearchUsersByUsername(ctx, &identitypb.SearchUsersByUsernameRequest{
			Prefix: registrant.Username,
		})
		require.NoError(t, err)
		assert.Empty(t, users.GetResults())
	})
}
