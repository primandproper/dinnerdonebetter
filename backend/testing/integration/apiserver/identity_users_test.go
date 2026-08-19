package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	identityconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc/converters"

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

		_, err := testClient.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: identityconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input)})
		require.NoError(t, err)

		_, err = testClient.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: identityconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input)})
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := buildUserRegistrationInputForTest(t)
		input.Username = ""

		testClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := testClient.CreateUser(ctx, &identitysvc.CreateUserRequest{Input: identityconverters.ConvertUserRegistrationInputToGRPCUserRegistrationInput(input)})
		assert.Error(t, err)
	})
}

func TestUsers_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		u, _ := createUserAndClientForTest(t)

		user, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: u.ID})
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, u.ID, user.Result.Id)
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, user)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		u, _ := createUserAndClientForTest(t)

		user, err := c.GetUser(ctx, &identitysvc.GetUserRequest{UserId: u.ID})
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
			string(authorization.ReadWebhooksPermission), // permission everyone has
		}})
		require.NoError(t, err)
		assert.NotNil(t, response)

		assert.Equal(t, map[string]bool{
			string(authorization.ImpersonateUserPermission): false,
			string(authorization.ReadWebhooksPermission):    true,
		}, response.Permissions)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		testClient := buildUnauthenticatedGRPCClientForTest(t)

		response, err := testClient.CheckPermissions(ctx, &authsvc.UserPermissionsRequestInput{Permissions: []string{string(authorization.ReadWebhooksPermission)}})
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

		results, err := adminClient.SearchForUsers(ctx, &identitysvc.SearchForUsersRequest{
			Query: createdUsers[0].Username[:2],
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1)
	})

	T.Run("only admins can do it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.SearchForUsers(ctx, &identitysvc.SearchForUsersRequest{
			Query: createdUsers[0].Username[:2],
		})
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestUsers_GetUsers(T *testing.T) {
	T.Parallel()

	// create some users so we have data to list
	for range exampleQuantity {
		createUserAndClientForTest(T)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetUsers(ctx, &identitysvc.GetUsersRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), exampleQuantity)
	})

	T.Run("only admins can do it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.GetUsers(ctx, &identitysvc.GetUsersRequest{})
		require.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		results, err := c.GetUsers(ctx, &identitysvc.GetUsersRequest{})
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestUsers_GetUsersForAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		// get the user's account via admin
		accountsRes, err := adminClient.GetAccountsForUser(ctx, &identitysvc.GetAccountsForUserRequest{UserId: user.ID})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(accountsRes.Results), 1)
		accountID := accountsRes.Results[0].Id

		results, err := adminClient.GetUsersForAccount(ctx, &identitysvc.GetUsersForAccountRequest{
			AccountId: accountID,
		})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		results, err := c.GetUsersForAccount(ctx, &identitysvc.GetUsersForAccountRequest{
			AccountId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestUsers_UpdateUserDetails(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		_, err := testClient.UpdateUserDetails(ctx, &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName:       "UpdatedFirst",
				LastName:        "UpdatedLast",
				CurrentPassword: user.HashedPassword,
				TotpToken:       generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)

		// verify the update took effect
		updatedUser, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "UpdatedFirst", updatedUser.Result.FirstName)
		assert.Equal(t, "UpdatedLast", updatedUser.Result.LastName)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateUserDetails(ctx, &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName: "UpdatedFirst",
				LastName:  "UpdatedLast",
			},
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

		_, err := testClient.UpdateUserEmailAddress(ctx, &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: newEmail,
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)

		// verify the update took effect
		updatedUser, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newEmail, updatedUser.Result.EmailAddress)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateUserEmailAddress(ctx, &identitysvc.UpdateUserEmailAddressRequest{
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

		_, err := testClient.UpdateUserUsername(ctx, &identitysvc.UpdateUserUsernameRequest{
			NewUsername: newUsername,
		})
		require.NoError(t, err)

		// verify the update took effect
		updatedUser, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newUsername, updatedUser.Result.Username)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateUserUsername(ctx, &identitysvc.UpdateUserUsernameRequest{
			NewUsername: "newusername",
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

		_, err := adminClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, adminClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "users", RelevantID: user.ID},
			{EventType: "archived", ResourceType: "users", RelevantID: user.ID},
		})
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: nonexistentID,
		})
		assert.Error(t, err)
	})

	T.Run("only admins can archive another user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		testClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := testClient.ArchiveUser(ctx, &identitysvc.ArchiveUserRequest{
			UserId: user.ID,
		})
		assert.Error(t, err)
	})
}

const userAvatarUploadChunkSize = 32 * 1024

func TestUsers_UploadUserAvatar(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		fileData := []byte("fake image data for integration test")
		filename := "avatar.jpg"
		contentType := uploadedmedia.MimeTypeImageJPEG

		stream, err := testClient.UploadUserAvatar(ctx)
		require.NoError(t, err)

		// First message: metadata
		err = stream.Send(&uploadedmediagrpc.UploadRequest{
			Payload: &uploadedmediagrpc.UploadRequest_Metadata{
				Metadata: &uploadedmediagrpc.UploadMetadata{
					ObjectName:  filename,
					ContentType: contentType,
				},
			},
		})
		require.NoError(t, err)

		// Stream chunks
		for offset := 0; offset < len(fileData); offset += userAvatarUploadChunkSize {
			end := min(offset+userAvatarUploadChunkSize, len(fileData))
			chunk := fileData[offset:end]
			err = stream.Send(&uploadedmediagrpc.UploadRequest{
				Payload: &uploadedmediagrpc.UploadRequest_Chunk{Chunk: chunk},
			})
			require.NoError(t, err)
		}

		resp, err := stream.CloseAndRecv()
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Created)
		assert.NotEmpty(t, resp.Created.Id)

		// Verify user avatar is set when reading the user
		retrieved, err := adminClient.GetUser(ctx, &identitysvc.GetUserRequest{UserId: user.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.NotNil(t, retrieved.Result.Avatar)
		assert.Equal(t, resp.Created.Id, retrieved.Result.Avatar.Id)
		assert.Equal(t, uploadedmediagrpc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_JPEG, retrieved.Result.Avatar.MimeType)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		stream, err := c.UploadUserAvatar(ctx)
		require.NoError(t, err)

		err = stream.Send(&uploadedmediagrpc.UploadRequest{
			Payload: &uploadedmediagrpc.UploadRequest_Metadata{
				Metadata: &uploadedmediagrpc.UploadMetadata{
					ObjectName:  "test.jpg",
					ContentType: uploadedmedia.MimeTypeImageJPEG,
				},
			},
		})
		requireStreamSend(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})
}
