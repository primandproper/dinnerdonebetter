package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v14/identity/identitypb"
	webhookspb "github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

func TestAccounts_Inviting(T *testing.T) {
	T.Parallel()

	// What joining an account opens up is this application's: webhooks confine to the caller's
	// active account, so an accepted invitation is what lets the invitee read one the account
	// already had. The invitation mechanics themselves are the identity suite's.
	T.Run("an accepted invitation reaches the account's account-scoped rows", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, testClient)
		createdWebhook := createWebhookForTest(t, testClient)

		input := buildUserRegistrationInputForTest(t)
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		invitation := inviteForTest(t, selfIDForTest(t, testClient), accountID, input.GetUser().GetEmailAddress())

		AssertAuditLogContainsFuzzy(t, ctx, testClient, accountID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "account_invitations", RelevantID: invitation.ID},
		})

		_, err := inviteeClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
			StatusNote:   t.Name(),
		})
		require.NoError(t, err)

		// A new token, because a token says which accounts the caller was in when it was issued.
		inviteeClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, invitee))
		require.NoError(t, err)

		_, err = inviteeClient.IdentityService().SetDefaultAccount(ctx, &identitypb.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		webhook, err := inviteeClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetId()})
		require.NoError(t, err)
		require.NotNil(t, webhook)
	})

	T.Run("a canceled invitation reaches nothing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		input := buildUserRegistrationInputForTest(t)
		_, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		invitation := inviteForTest(t, selfIDForTest(t, testClient), getAccountIDForTest(t, testClient), input.GetUser().GetEmailAddress())

		_, err := testClient.IdentityService().CancelInvitation(ctx, &identitypb.CancelInvitationRequest{
			InvitationId: invitation.ID,
			StatusNote:   t.Name(),
		})
		require.NoError(t, err)

		_, err = inviteeClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
		})
		require.Error(t, err)

		webhook, err := inviteeClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetId()})
		require.Error(t, err)
		assert.Nil(t, webhook)
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

		invitation := inviteForTest(t, selfIDForTest(t, testClient), accountID, input.GetUser().GetEmailAddress())

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

	T.Run("a member's default moves when they are removed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, ownerClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, ownerClient)

		input := buildUserRegistrationInputForTest(t)
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)
		inviteeOwnAccountID := getAccountIDForTest(t, inviteeClient)

		invitation := inviteForTest(t, selfIDForTest(t, ownerClient), accountID, input.GetUser().GetEmailAddress())

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
