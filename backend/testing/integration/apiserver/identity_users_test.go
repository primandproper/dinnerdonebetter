package integration

import (
	"testing"

	identity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/identity/identitypb"

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
			{EventType: "created", ResourceType: identity.ResourceTypeUser, RelevantID: user.ID},
			{EventType: "created", ResourceType: identity.ResourceTypeAccount},
			{EventType: "created", ResourceType: identity.ResourceTypeMembership},
		})
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		input.User.Username = ""

		testClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := testClient.Register(ctx, input)
		assert.Error(t, err)
	})

	// The agreements are required by this application rather than by the directory, which
	// records when somebody last agreed and has no opinion about whether they had to. This
	// is the only place that reading is enforced, so it is the only place it can be checked.
	T.Run("refuses a registration that declined the terms", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		input.Agreements = []identitypb.Agreement{identitypb.Agreement_AGREEMENT_PRIVACY_POLICY}

		testClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := testClient.Register(ctx, input)
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
		AssertAuditLogContainsFuzzyForResource(t, ctx, identity.ResourceTypeUser, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: identity.ResourceTypeUser, RelevantID: user.ID},
			{EventType: "archived", ResourceType: identity.ResourceTypeUser, RelevantID: user.ID},
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
			invitation := inviteForTest(t, owner.ID, accountID, input.GetUser().GetEmailAddress())

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

		AssertAuditLogContainsFuzzyForResource(t, ctx, identity.ResourceTypeAccount, accountID, 10, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: identity.ResourceTypeAccount, RelevantID: accountID},
		})
	})
}
