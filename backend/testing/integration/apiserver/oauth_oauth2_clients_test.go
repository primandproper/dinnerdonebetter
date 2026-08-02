package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkOAuth2ClientEquality(t *testing.T, expected, actual *oauth.OAuth2Client) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected OAuth2Client to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected OAuth2Client to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected OAuth2Client Name")
	assert.Equal(t, expected.Description, actual.Description, "expected OAuth2Client Description")
}

func createOAuth2ClientForTest(t *testing.T) *oauth.OAuth2Client {
	t.Helper()

	ctx := t.Context()

	exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
	exampleOAuth2ClientInput := converters.ConvertOAuth2ClientToOAuth2ClientCreationInput(exampleOAuth2Client)
	created, err := adminClient.CreateOAuth2Client(ctx, &oauthsvc.CreateOAuth2ClientRequest{
		Input: grpcconverters.ConvertOAuth2ClientCreationRequestInputToGRPCOAuth2ClientCreationRequestInput(exampleOAuth2ClientInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCOAuth2ClientToOAuth2Client(created.Created)
	checkOAuth2ClientEquality(t, exampleOAuth2Client, converted)

	retrieved, err := adminClient.GetOAuth2Client(ctx, &oauthsvc.GetOAuth2ClientRequest{
		Oauth2ClientId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	oauth2Client := grpcconverters.ConvertGRPCOAuth2ClientToOAuth2Client(retrieved.Result)
	checkOAuth2ClientEquality(t, converted, oauth2Client)

	return oauth2Client
}

func TestOAuth2Clients_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createOAuth2ClientForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeOAuth2ClientCreationRequestInput()
		convertedInput := grpcconverters.ConvertOAuth2ClientCreationRequestInputToGRPCOAuth2ClientCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateOAuth2Client(ctx, &oauthsvc.CreateOAuth2ClientRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeOAuth2ClientCreationRequestInput()
		convertedInput := grpcconverters.ConvertOAuth2ClientCreationRequestInputToGRPCOAuth2ClientCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateOAuth2Client(ctx, &oauthsvc.CreateOAuth2ClientRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeOAuth2ClientCreationRequestInput()
		convertedInput := grpcconverters.ConvertOAuth2ClientCreationRequestInputToGRPCOAuth2ClientCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateOAuth2Client(ctx, &oauthsvc.CreateOAuth2ClientRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestOAuth2Clients_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)

		retrieved, err := testClient.GetOAuth2Client(ctx, &oauthsvc.GetOAuth2ClientRequest{Oauth2ClientId: created.ID})
		assert.NoError(t, err)

		converted := grpcconverters.ConvertGRPCOAuth2ClientToOAuth2Client(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetOAuth2Client(ctx, &oauthsvc.GetOAuth2ClientRequest{Oauth2ClientId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetOAuth2Client(ctx, &oauthsvc.GetOAuth2ClientRequest{Oauth2ClientId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestOAuth2Clients_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)

		_, err := adminClient.ArchiveOAuth2Client(ctx, &oauthsvc.ArchiveOAuth2ClientRequest{Oauth2ClientId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetOAuth2Client(ctx, &oauthsvc.GetOAuth2ClientRequest{Oauth2ClientId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveOAuth2Client(ctx, &oauthsvc.ArchiveOAuth2ClientRequest{Oauth2ClientId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveOAuth2Client(ctx, &oauthsvc.ArchiveOAuth2ClientRequest{Oauth2ClientId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveOAuth2Client(ctx, &oauthsvc.ArchiveOAuth2ClientRequest{Oauth2ClientId: created.ID})
		assert.Error(t, err)
	})
}

func TestOAuth2Clients_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdOAuth2Clients := []*oauth.OAuth2Client{}
	for range exampleQuantity {
		created := createOAuth2ClientForTest(T)
		createdOAuth2Clients = append(createdOAuth2Clients, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetOAuth2Clients(ctx, &oauthsvc.GetOAuth2ClientsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.True(t, len(retrieved.Results) >= len(createdOAuth2Clients))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetOAuth2Clients(ctx, &oauthsvc.GetOAuth2ClientsRequest{})
		assert.Error(t, err)
	})
}
