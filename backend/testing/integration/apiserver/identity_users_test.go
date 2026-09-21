package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	authconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/converters"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/identity/identitypb"
	"github.com/primandproper/primitives-go/v2/pointer"

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

func TestUsers_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		u, _ := createUserAndClientForTest(t)

		user, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: u.ID})
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, u.ID, user.GetUser().GetId())
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, user)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		u, _ := createUserAndClientForTest(t)

		user, err := c.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: u.ID})
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUsers_PermissionChecking(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		response, err := testClient.CheckPermissions(ctx, &authsvc.UserPermissionsRequestInput{Permissions: []string{
			string(authorization.ImpersonateUserPermission),
			string(authorization.ReadWebhookEndpointsPermission), // permission everyone has
		}})
		require.NoError(t, err)
		assert.NotNil(t, response)

		assert.Equal(t, map[string]bool{
			string(authorization.ImpersonateUserPermission):      false,
			string(authorization.ReadWebhookEndpointsPermission): true,
		}, response.Permissions)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		testClient := buildUnauthenticatedGRPCClientForTest(t)

		response, err := testClient.CheckPermissions(ctx, &authsvc.UserPermissionsRequestInput{Permissions: []string{string(authorization.ReadWebhookEndpointsPermission)}})
		require.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestUsers_Searching(T *testing.T) {
	T.Parallel()

	// create some users to search from
	createdUsers := []*identity.User{}
	for range exampleQuantity {
		u, _ := createUserAndClientForTest(T)
		createdUsers = append(createdUsers, u)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.IdentityService().SearchUsersByUsername(ctx, &identitypb.SearchUsersByUsernameRequest{
			Prefix: createdUsers[0].Username[:2],
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), 1)
	})

	T.Run("only admins can do it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.IdentityService().SearchUsersByUsername(ctx, &identitypb.SearchUsersByUsernameRequest{
			Prefix: createdUsers[0].Username[:2],
		})
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestUsers_ListUsers(T *testing.T) {
	T.Parallel()

	// create some users so we have data to list
	for range exampleQuantity {
		createUserAndClientForTest(T)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.IdentityService().ListUsers(ctx, &identitypb.ListUsersRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), exampleQuantity)
	})

	T.Run("only admins can do it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.IdentityService().ListUsers(ctx, &identitypb.ListUsersRequest{})
		require.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		results, err := c.IdentityService().ListUsers(ctx, &identitypb.ListUsersRequest{})
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestUsers_ListAccountMembers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		// get the user's account via admin
		accountsRes, err := adminClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{UserId: user.ID})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(accountsRes.GetResults()), 1)
		accountID := accountsRes.GetResults()[0].GetId()

		results, err := adminClient.IdentityService().ListAccountMembers(ctx, &identitypb.ListAccountMembersRequest{
			AccountId: accountID,
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), 1)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		results, err := c.IdentityService().ListAccountMembers(ctx, &identitypb.ListAccountMembersRequest{
			AccountId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

// TestUsers_UpdateProfile covers what a person may change about themselves without proving
// who they are again: their names.
//
// The handle and the address are not here. They are credential changes — whoever holds the
// address can take the account through a password reset — so they are re-authenticated, on
// the auth service, and tested below.
func TestUsers_UpdateProfile(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		_, err := testClient.IdentityService().UpdateProfile(ctx, &identitypb.UpdateProfileRequest{
			Input: &identitypb.ProfileUpdateInput{
				FirstName: pointer.To("UpdatedFirst"),
				LastName:  pointer.To("UpdatedLast"),
			},
		})
		require.NoError(t, err)

		// verify the update took effect
		updatedUser, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "UpdatedFirst", updatedUser.GetUser().GetFirstName())
		assert.Equal(t, "UpdatedLast", updatedUser.GetUser().GetLastName())
	})

	// A field left unset is not a field set to empty, which is the whole reason the input's
	// fields are pointers: a form sending one value must not blank the others.
	T.Run("leaves unnamed fields alone", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		_, err := testClient.IdentityService().UpdateProfile(ctx, &identitypb.UpdateProfileRequest{
			Input: &identitypb.ProfileUpdateInput{FirstName: pointer.To("OnlyFirst")},
		})
		require.NoError(t, err)

		updatedUser, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.Equal(t, "OnlyFirst", updatedUser.GetUser().GetFirstName())
		assert.Equal(t, user.Username, updatedUser.GetUser().GetUsername())
		assert.Equal(t, user.EmailAddress, updatedUser.GetUser().GetEmailAddress())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.IdentityService().UpdateProfile(ctx, &identitypb.UpdateProfileRequest{
			Input: &identitypb.ProfileUpdateInput{FirstName: pointer.To("UpdatedFirst")},
		})
		assert.Error(t, err)
	})
}

func TestUsers_UpdateUserEmailAddress(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		newEmail := fmt.Sprintf("updated_%d@whatever.com", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano)))

		_, err := testClient.UpdateUserEmailAddress(ctx, &authsvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: newEmail,
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)

		// verify the update took effect
		updatedUser, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newEmail, updatedUser.GetUser().GetEmailAddress())
	})

	// The re-authentication is the point of this RPC existing rather than the change going
	// through the directory's UpdateProfile, so a wrong password has to fail it.
	T.Run("refuses a wrong password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		_, err := testClient.UpdateUserEmailAddress(ctx, &authsvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: fmt.Sprintf("nope_%d@whatever.com", hashStringToNumber(t.Name())),
			CurrentPassword: "not the password",
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateUserEmailAddress(ctx, &authsvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: "new@example.com",
			CurrentPassword: "whatever",
			TotpToken:       "000000",
		})
		assert.Error(t, err)
	})
}

func TestUsers_UpdateUserUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		newUsername := fmt.Sprintf("updated_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano)))

		_, err := testClient.UpdateUserUsername(ctx, &authsvc.UpdateUserUsernameRequest{
			NewUsername:     newUsername,
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)

		// verify the update took effect
		updatedUser, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newUsername, updatedUser.GetUser().GetUsername())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateUserUsername(ctx, &authsvc.UpdateUserUsernameRequest{
			NewUsername:     "newusername",
			CurrentPassword: "whatever",
			TotpToken:       "000000",
		})
		assert.Error(t, err)
	})
}

func TestUsers_RecordAgreement(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.IdentityService().RecordAgreement(ctx, &identitypb.RecordAgreementRequest{
			Agreements: []identitypb.Agreement{identitypb.Agreement_AGREEMENT_TERMS_OF_SERVICE},
		})
		require.NoError(t, err)
		require.NotNil(t, res.GetUser())
		assert.NotNil(t, res.GetUser().GetLastAcceptedTermsOfService())
	})

	// Naming none is a caller who built an empty list and did not notice, which is a
	// refusal rather than a no-op that reports success.
	T.Run("refuses an empty set", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.IdentityService().RecordAgreement(ctx, &identitypb.RecordAgreementRequest{})
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

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.IdentityService().ArchiveUser(ctx, &identitypb.ArchiveUserRequest{
			UserId: nonexistentID,
		})
		assert.Error(t, err)
	})

	T.Run("only admins can archive another user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.IdentityService().ArchiveUser(ctx, &identitypb.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.IdentityService().ArchiveUser(ctx, &identitypb.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
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

		created := createServiceUserForTest(t, true, registrant)

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
