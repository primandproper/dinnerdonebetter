package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	identity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	webhookspb "github.com/primandproper/platform-go/v15/webhooks/webhookspb"
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

		AssertAuditLogContainsFuzzyForResource(t, ctx, identity.ResourceTypeAccount, createdAccount.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: identity.ResourceTypeAccount, RelevantID: createdAccount.GetId()},
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

		AssertAuditLogContainsFuzzy(t, ctx, testClient, createdAccount.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: identity.ResourceTypeAccount, RelevantID: createdAccount.GetId()},
			{EventType: "updated", ResourceType: identity.ResourceTypeAccount, RelevantID: createdAccount.GetId()},
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
		AssertAuditLogContainsFuzzyForResource(t, ctx, identity.ResourceTypeAccount, createdAccount.GetId(), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: identity.ResourceTypeAccount, RelevantID: createdAccount.GetId()},
			{EventType: "archived", ResourceType: identity.ResourceTypeAccount, RelevantID: createdAccount.GetId()},
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
			{EventType: "created", ResourceType: identity.ResourceTypeInvitation, RelevantID: invitation.ID},
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
}

func TestAccounts_ListAccountsForUser(T *testing.T) {
	T.Parallel()

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

// TestAccounts_OwnershipTransfer pins the audit entry this application's identity hooks record
// for a transfer. The transfer itself — the new owner, and them on the roster — is
// conformance/identity's.
func TestAccounts_OwnershipTransfer(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, testClient)

		// The recipient joins first. An account can only be handed to somebody the caller
		// already shares one with, so the flow a household actually uses is invite, accept,
		// transfer.
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

		AssertAuditLogContainsFuzzy(t, ctx, testClient, accountID, 15, []*ExpectedAuditEntry{
			{EventType: "updated", ResourceType: identity.ResourceTypeAccount, RelevantID: accountID},
		})
	})
}

// TestAccounts_RemovingMembers pins the audit entries this application's identity hooks record
// when a member is invited and then removed. Where a removed member lands afterwards is
// conformance/identity's.
func TestAccounts_RemovingMembers(T *testing.T) {
	T.Parallel()

	T.Run("the invitation and the ended membership are recorded", func(t *testing.T) {
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

		_, err = ownerClient.IdentityService().RemoveMembership(ctx, &identitypb.RemoveMembershipRequest{
			AccountId: accountID,
			UserId:    invitee.ID,
		})
		require.NoError(t, err)

		// The invitation is on the account's chain. The ended membership is filed under the
		// member it was about — platform's subject for a membership entry — so it is read as
		// the member, from the chains their session covers.
		AssertAuditLogContainsFuzzy(t, ctx, ownerClient, accountID, 20, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: identity.ResourceTypeInvitation, RelevantID: invitation.ID},
		})
		AssertAuditLogContainsFuzzy(t, ctx, inviteeClient, inviteeOwnAccountID, 20, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: identity.ResourceTypeMembership},
		})
	})
}
