package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/identity/identitypb"
	webhookspb "github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultNumberOfAccountsAssociatedWithUsers = 1
)

// newAccountName is a name for an account a test mints.
//
// Account names are not unique — two unrelated households may both be called "Acme", and
// enforcing otherwise would make registration fail for a reason nobody can act on — so
// this exists for legibility in a failure rather than for uniqueness.
func newAccountName(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("account_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano)))
}

// createAccountForTest mints an account owned by the caller, with them as its administrator.
func createAccountForTest(t *testing.T, c interface {
	IdentityService() identitypb.IdentityServiceClient
},
) *identitypb.Account {
	t.Helper()
	ctx := t.Context()

	created, err := c.IdentityService().CreateAccount(ctx, &identitypb.CreateAccountRequest{
		Name:       newAccountName(t),
		OwnerRoles: []string{authorization.AccountAdminRoleName},
	})
	require.NoError(t, err)
	require.NotNil(t, created.GetAccount())

	return created.GetAccount()
}

func TestAccounts_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdAccount := createAccountForTest(t, testClient)

		AssertAuditLogContainsFuzzyForResource(t, ctx, "accounts", createdAccount.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "accounts", RelevantID: createdAccount.GetId()},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.IdentityService().CreateAccount(ctx, &identitypb.CreateAccountRequest{
			Name:       newAccountName(t),
			OwnerRoles: []string{authorization.AccountAdminRoleName},
		})
		assert.Error(t, err)
	})

	// The owner's roles are required: a membership with none is a member who may do
	// nothing in the account they own, and the store refuses it rather than writing a row
	// whose emptiness surfaces later as an authorization bug.
	T.Run("refuses an owner with no roles", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.IdentityService().CreateAccount(ctx, &identitypb.CreateAccountRequest{
			Name: newAccountName(t),
		})
		assert.Error(t, err)
	})
}

func TestAccounts_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		var createdAccounts []*identitypb.Account
		for range 5 {
			createdAccounts = append(createdAccounts, createAccountForTest(t, testClient))
		}

		accounts, err := testClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, accounts)
		assert.Len(t, accounts.GetResults(), len(createdAccounts)+defaultNumberOfAccountsAssociatedWithUsers)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		user, _ := createUserAndClientForTest(t)
		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
	})

	// One member cannot enumerate another's households. The permission is on the method
	// and the authorizer decides whose rows an allowed call may touch, which for a
	// directory read is the caller and the accounts they are in.
	T.Run("one user cannot list another's accounts", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		subject, _ := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		_, err := otherClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: subject.ID,
		})
		assert.Error(t, err)
	})
}

func TestAccounts_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdAccount := createAccountForTest(t, testClient)

		retrievedAccount, err := testClient.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{
			AccountId: createdAccount.GetId(),
		})
		require.NoError(t, err)
		require.NotNil(t, retrievedAccount.GetAccount())

		assert.Equal(t, createdAccount.GetId(), retrievedAccount.GetAccount().GetId())
		assert.Equal(t, createdAccount.GetName(), retrievedAccount.GetAccount().GetName())
		assert.Equal(t, createdAccount.GetOwnerUserId(), retrievedAccount.GetAccount().GetOwnerUserId())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, testClient := createUserAndClientForTest(t)

		// fetch the account
		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		unauthenticated := buildUnauthenticatedGRPCClientForTest(t)

		_, err = unauthenticated.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: account.Result.GetId()})
		assert.Error(t, err)
	})

	T.Run("for nonexistent account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrievedAccount, err := testClient.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, retrievedAccount)
	})
}

func TestAccounts_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdAccount := createAccountForTest(t, testClient)

		_, err := testClient.IdentityService().SetDefaultAccount(ctx, &identitypb.SetDefaultAccountRequest{
			AccountId: createdAccount.GetId(),
		})
		require.NoError(t, err)

		_, err = testClient.IdentityService().UpdateAccount(ctx, &identitypb.UpdateAccountRequest{
			AccountId: createdAccount.GetId(),
			Input:     &identitypb.AccountUpdateInput{Name: pointer.To("Updated name")},
		})
		require.NoError(t, err)

		updated, err := testClient.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: createdAccount.GetId()})
		require.NoError(t, err)
		assert.Equal(t, "Updated name", updated.GetAccount().GetName())

		AssertAuditLogContainsFuzzy(t, ctx, testClient, createdAccount.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "accounts", RelevantID: createdAccount.GetId()},
			{EventType: "updated", ResourceType: "accounts", RelevantID: createdAccount.GetId()},
		})
	})

	// Neither the billing state nor the owner is on the update, which is what keeps a
	// read-modify-write over a name from losing whatever a processor webhook or an
	// ownership transfer did in between.
	T.Run("leaves unnamed fields alone", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdAccount := createAccountForTest(t, testClient)

		_, err := testClient.IdentityService().UpdateAccount(ctx, &identitypb.UpdateAccountRequest{
			AccountId: createdAccount.GetId(),
			Input:     &identitypb.AccountUpdateInput{TimeZone: pointer.To("America/Chicago")},
		})
		require.NoError(t, err)

		updated, err := testClient.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: createdAccount.GetId()})
		require.NoError(t, err)
		assert.Equal(t, "America/Chicago", updated.GetAccount().GetTimeZone())
		assert.Equal(t, createdAccount.GetName(), updated.GetAccount().GetName())
		assert.Equal(t, createdAccount.GetOwnerUserId(), updated.GetAccount().GetOwnerUserId())
	})

	// A zone name that does not load renders every date on the account wrong, forever,
	// without anything saying so. Failing the write is what keeps that a typo.
	T.Run("refuses a time zone that does not load", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdAccount := createAccountForTest(t, testClient)

		_, err := testClient.IdentityService().UpdateAccount(ctx, &identitypb.UpdateAccountRequest{
			AccountId: createdAccount.GetId(),
			Input:     &identitypb.AccountUpdateInput{TimeZone: pointer.To("America/Chicagoo")},
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, testClient := createUserAndClientForTest(t)

		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		unauthenticated := buildUnauthenticatedGRPCClientForTest(t)

		_, err = unauthenticated.IdentityService().UpdateAccount(ctx, &identitypb.UpdateAccountRequest{
			AccountId: account.Result.GetId(),
			Input:     &identitypb.AccountUpdateInput{Name: pointer.To("nope")},
		})
		assert.Error(t, err)
	})

	T.Run("for nonexistent account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.IdentityService().UpdateAccount(ctx, &identitypb.UpdateAccountRequest{
			AccountId: nonexistentID,
			Input:     &identitypb.AccountUpdateInput{Name: pointer.To("nope")},
		})
		require.Error(t, err)
	})
}

func TestAccounts_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdAccount := createAccountForTest(t, testClient)

		_, err := testClient.IdentityService().ArchiveAccount(ctx, &identitypb.ArchiveAccountRequest{
			AccountId: createdAccount.GetId(),
		})
		require.NoError(t, err)

		// By resource rather than by chain: the account just created is one the caller is
		// not inside, so its entries are not on any chain this session can read.
		AssertAuditLogContainsFuzzyForResource(t, ctx, "accounts", createdAccount.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "accounts", RelevantID: createdAccount.GetId()},
			{EventType: "archived", ResourceType: "accounts", RelevantID: createdAccount.GetId()},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a user so that the account actually exists
		_, testClient := createUserAndClientForTest(t)

		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		unauthenticated := buildUnauthenticatedGRPCClientForTest(t)

		_, err = unauthenticated.IdentityService().ArchiveAccount(ctx, &identitypb.ArchiveAccountRequest{
			AccountId: account.Result.GetId(),
		})
		assert.Error(t, err)
	})

	T.Run("for nonexistent account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.IdentityService().ArchiveAccount(ctx, &identitypb.ArchiveAccountRequest{AccountId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestAccounts_Inviting(T *testing.T) {
	T.Parallel()

	T.Run("invite user via their email address", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account ID to send invites for
		_, testClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, testClient)

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// create a user to invite
		input := buildUserRegistrationInputForTest(t)
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		invitation := inviteForTest(t, selfIDForTest(t, testClient), accountID, input.EmailAddress)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, accountID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_invitations", RelevantID: invitation.ID},
		})

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.GetResults())

		// Proven first: listing what was sent to an address is gated on having proven it,
		// because anybody may claim any address at registration.
		verifyEmailAddressForTest(t, invitee.ID)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.IdentityService().ListInvitationsForEmailAddress(ctx, &identitypb.ListInvitationsForEmailAddressRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.NotEmpty(t, invitations.GetResults())

		// accept the invitation
		_, err = inviteeClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
			StatusNote:   t.Name(),
		})
		require.NoError(t, err)

		// the invited user needs a new token that indicates they're a member of this account
		inviteeClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, invitee))
		require.NoError(t, err)

		// verify that we don't have any sent invitations because they've all been accepted
		sentInvitations, err = testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.GetResults())

		// verify that the invited user can see the account in their accounts list
		accounts, err := inviteeClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: invitee.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, accounts)
		assert.Len(t, accounts.GetResults(), 2)

		var found bool
		for _, account := range accounts.GetResults() {
			if !found {
				found = account.GetId() == accountID
			}
		}
		require.True(t, found)

		// change to the new account
		_, err = inviteeClient.IdentityService().SetDefaultAccount(ctx, &identitypb.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		// validate we can see the webhook created before our user existed
		webhook, err := inviteeClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetId()})
		require.NoError(t, err)
		require.NotNil(t, webhook)
	})

	T.Run("invite user via token and invite ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user and get the account ID to send invites for
		_, testClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, testClient)

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// the registrant, and the invitation addressed to them
		input := buildUserRegistrationInputForTest(t)

		invitation := inviteForTest(t, selfIDForTest(t, testClient), accountID, input.EmailAddress)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, accountID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_invitations", RelevantID: invitation.ID},
		})

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.GetResults())

		// registering against the link answers it in the same transaction that makes the
		// user, so there is no second call to accept it
		input.InvitationID = invitation.ID
		input.InvitationToken = invitation.Token
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// Proven first: listing what was sent to an address is gated on having proven it.
		verifyEmailAddressForTest(t, invitee.ID)

		// nothing outstanding for the invitee, because the registration answered it
		invitations, err := inviteeClient.IdentityService().ListInvitationsForEmailAddress(ctx, &identitypb.ListInvitationsForEmailAddressRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.Empty(t, invitations.GetResults())

		// and nothing outstanding for the sender either
		sentInvitations, err = testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.GetResults())

		// A registration by invitation joins the inviter's account rather than minting
		// one of its own, so this is the only account the registrant holds.
		accounts, err := inviteeClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: invitee.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, accounts)
		require.Len(t, accounts.GetResults(), 1)
		assert.Equal(t, accountID, accounts.GetResults()[0].GetId())

		// validate we can see the webhook created before our user existed
		webhook, err := inviteeClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetId()})
		require.NoError(t, err)
		require.NotNil(t, webhook)
	})

	T.Run("invites can be canceled before acceptance", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user
		_, testClient := createUserAndClientForTest(t)

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// create a user to invite
		input := buildUserRegistrationInputForTest(t)
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		invitation := inviteForTest(t, selfIDForTest(t, testClient), getAccountIDForTest(t, testClient), input.EmailAddress)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_invitations", RelevantID: invitation.ID},
		})

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.GetResults())

		// Proven first: listing what was sent to an address is gated on having proven it,
		// because anybody may claim any address at registration.
		verifyEmailAddressForTest(t, invitee.ID)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.IdentityService().ListInvitationsForEmailAddress(ctx, &identitypb.ListInvitationsForEmailAddressRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.NotEmpty(t, invitations.GetResults())

		_, err = testClient.IdentityService().CancelInvitation(ctx, &identitypb.CancelInvitationRequest{
			InvitationId: invitation.ID,
			StatusNote:   t.Name(),
		})
		require.NoError(t, err)

		// a withdrawn invitation can no longer be answered
		_, err = inviteeClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
			StatusNote:   t.Name(),
		})
		require.Error(t, err)

		// nothing is outstanding any more
		sentInvitations, err = testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.GetResults())

		// and the invitee never got into the account
		webhook, err := inviteeClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetId()})
		require.Error(t, err)
		assert.Nil(t, webhook)
	})

	T.Run("invites can be rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the inviting user
		_, testClient := createUserAndClientForTest(t)

		// create a user to invite
		input := buildUserRegistrationInputForTest(t)
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// create the invitation for the user
		invitation := inviteForTest(t, selfIDForTest(t, testClient), getAccountIDForTest(t, testClient), input.EmailAddress)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_invitations", RelevantID: invitation.ID},
		})

		// verify that we can retrieve the invitation we just created
		sentInvitations, err := testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.NotEmpty(t, sentInvitations.GetResults())

		// Proven first: listing what was sent to an address is gated on having proven it,
		// because anybody may claim any address at registration.
		verifyEmailAddressForTest(t, invitee.ID)

		// verify the invitee can see the invitation as received
		invitations, err := inviteeClient.IdentityService().ListInvitationsForEmailAddress(ctx, &identitypb.ListInvitationsForEmailAddressRequest{})
		require.NoError(t, err)
		require.NotNil(t, invitations)
		assert.NotEmpty(t, invitations.GetResults())

		_, err = inviteeClient.IdentityService().RejectInvitation(ctx, &identitypb.RejectInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
			StatusNote:   t.Name(),
		})
		require.NoError(t, err)

		// nothing is outstanding any more
		sentInvitations, err = testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, sentInvitations)
		assert.Empty(t, sentInvitations.GetResults())
	})

	// An invitation link is a bearer credential for joining somebody else's account, so
	// the token is compared on the row the id names rather than being an index key.
	T.Run("a wrong token answers nothing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		input := buildUserRegistrationInputForTest(t)
		_, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		invitation := inviteForTest(t, selfIDForTest(t, testClient), getAccountIDForTest(t, testClient), input.EmailAddress)

		_, err := inviteeClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        "not the token",
		})
		require.Error(t, err)

		// and it is still outstanding
		sentInvitations, err := testClient.IdentityService().ListInvitationsFromUser(ctx, &identitypb.ListInvitationsFromUserRequest{})
		require.NoError(t, err)
		assert.NotEmpty(t, sentInvitations.GetResults())
	})
}

func TestAccounts_GetInvitation(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		input := buildUserRegistrationInputForTest(t)
		_, _ = createUserAndClientForTestWithRegistrationInput(t, input)

		invitation := inviteForTest(t, selfIDForTest(t, testClient), getAccountIDForTest(t, testClient), input.EmailAddress)
		require.NotNil(t, invitation)

		result, err := testClient.IdentityService().GetInvitation(ctx, &identitypb.GetInvitationRequest{
			InvitationId: invitation.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, invitation.ID, result.GetInvitation().GetId())

		// There is no token on the message at all. Field 15 is reserved for the one it
		// would have held, which is platform saying in the schema what a comment would
		// otherwise have to: an invitation read back carries no secret to replay.
	})

	T.Run("nonexistent invitation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		result, err := testClient.IdentityService().GetInvitation(ctx, &identitypb.GetInvitationRequest{
			InvitationId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, result)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		testClient := buildUnauthenticatedGRPCClientForTest(t)

		result, err := testClient.IdentityService().GetInvitation(ctx, &identitypb.GetInvitationRequest{
			InvitationId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAccounts_ListAccountsForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		// create additional accounts
		for range 3 {
			createAccountForTest(t, testClient)
		}

		// admin fetches accounts for the user
		accounts, err := adminClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		assert.NotNil(t, accounts)
		// 1 default account + 3 created accounts
		assert.Len(t, accounts.GetResults(), 4)
	})

	// A user who does not exist has no accounts, which is an empty page rather than an
	// error: the read is "what does this subject have", and the honest answer is nothing.
	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		accounts, err := adminClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: nonexistentID,
		})
		require.NoError(t, err)
		assert.Empty(t, accounts.GetResults())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		accounts, err := c.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: user.ID,
		})
		require.Error(t, err)
		assert.Nil(t, accounts)
	})
}

func TestAccounts_OwnershipTransfer(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create the transferring user and get the account to transfer
		_, testClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, testClient)

		// create a webhook (to demonstrate access with later)
		createdWebhook := createWebhookForTest(t, testClient)

		// The recipient joins first. An account can only be handed to somebody the caller
		// already shares one with — the authorizer permits the caller themselves and
		// anybody they share a live account with, and a stranger is neither — so the flow
		// a household actually uses is invite, accept, transfer.
		input := buildUserRegistrationInputForTest(t)
		recipient, recipientClient := createUserAndClientForTestWithRegistrationInput(t, input)

		invitation := inviteForTest(t, selfIDForTest(t, testClient), accountID, input.EmailAddress)

		_, err := recipientClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
		})
		require.NoError(t, err)

		_, err = testClient.IdentityService().TransferAccountOwnership(ctx, &identitypb.TransferAccountOwnershipRequest{
			AccountId:      accountID,
			NewOwnerUserId: recipient.ID,
		})
		require.NoError(t, err)

		// A minted membership carries no roles, because ownership is the standing and
		// platform does not know what a role of ours means. Granting them is a separate
		// act, and it is the one that lets the new owner do anything in the account.
		//
		// The token is minted after the grant rather than before. There used to be one on
		// either side and only the second was ever used — a token says what the roles were
		// when it was issued, so the earlier one could not have carried the grant that had
		// not happened yet.
		_, err = testClient.IdentityService().SetMembershipRoles(ctx, &identitypb.SetMembershipRolesRequest{
			AccountId: accountID,
			UserId:    recipient.ID,
			Roles:     []string{authorization.AccountAdminRoleName},
		})
		require.NoError(t, err)

		recipientClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, recipient))
		require.NoError(t, err)

		// change to the new account
		_, err = recipientClient.IdentityService().SetDefaultAccount(ctx, &identitypb.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		recipientClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, recipient))
		require.NoError(t, err)

		// validate we can see the webhook created before our user existed
		webhook, err := recipientClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetId()})
		require.NoError(t, err)
		require.NotNil(t, webhook)

		// The old owner keeps their membership: transferring ownership and ejecting
		// somebody are different acts, and doing both here would make the common case —
		// handing over and staying on — impossible to express.
		AssertAuditLogContainsFuzzy(t, ctx, testClient, accountID, 15, []*ExpectedAuditEntry{
			{EventType: "updated", ResourceType: "accounts", RelevantID: accountID},
		})
	})
}

// Removing a member is refused for the account's owner and permitted for everybody else,
// and a removed member's default moves rather than being left naming an account they are
// no longer in.
//
// It replaces a test that asserted a backup account was created for a user removed from
// their last one. Nothing creates one now, and nothing needs to: an owner cannot be
// removed from the account they own, so the state that test was insuring against — a user
// with memberships nowhere — is one the store refuses to produce.
func TestAccounts_RemovingMembers(T *testing.T) {
	T.Parallel()

	T.Run("the owner cannot be removed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		owner, ownerClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, ownerClient)

		_, err := ownerClient.IdentityService().RemoveMembership(ctx, &identitypb.RemoveMembershipRequest{
			AccountId: accountID,
			UserId:    owner.ID,
		})
		assert.Error(t, err)
	})

	T.Run("a member's default moves when they are removed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, ownerClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, ownerClient)

		input := buildUserRegistrationInputForTest(t)
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)
		inviteeOwnAccountID := getAccountIDForTest(t, inviteeClient)

		invitation := inviteForTest(t, selfIDForTest(t, ownerClient), accountID, input.EmailAddress)

		_, err := inviteeClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
		})
		require.NoError(t, err)

		inviteeClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, invitee))
		require.NoError(t, err)

		_, err = inviteeClient.IdentityService().SetDefaultAccount(ctx, &identitypb.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		// the owner removes them
		_, err = ownerClient.IdentityService().RemoveMembership(ctx, &identitypb.RemoveMembershipRequest{
			AccountId: accountID,
			UserId:    invitee.ID,
		})
		require.NoError(t, err)

		// they land in the account they still hold rather than in one they were removed from
		inviteeClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, invitee))
		require.NoError(t, err)

		assert.Equal(t, inviteeOwnAccountID, getAccountIDForTest(t, inviteeClient))

		AssertAuditLogContainsFuzzy(t, ctx, ownerClient, accountID, 20, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_invitations", RelevantID: invitation.ID},
			{EventType: "archived", ResourceType: "account_user_memberships"},
		})
	})
}
