package integration

import (
	"testing"

	"github.com/primandproper/platform-go/v14/authentication/oauth2clients/oauth2clientspb"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The registry is platform's now, and the two things that changed about it on the wire are
// what most of this file is about.
//
// The secret is gone from every read. It used to ride along on each one — the converter
// copied every field of the local type, and the field held the SHA-256 digest — so a client
// registration's credential material was in the reply to GetOAuth2Client and to the listing.
// platform carries it on IssuedOAuth2Client, which only the creation RPC returns.
//
// And reading is a service admin's, which it was not. See internal/authorization.

func clientInputForTest(t *testing.T) *oauth2clientspb.OAuth2ClientCreationInput {
	t.Helper()

	return &oauth2clientspb.OAuth2ClientCreationInput{
		Name:        t.Name() + "_" + identifiers.New(),
		Description: "integration test client",
		// At least one is required, and it is matched byte for byte at /authorize. Nothing
		// listens here; this registration is never authorized against.
		RedirectUris: []string{"https://192.0.2.1/callback/" + identifiers.New()},
	}
}

func createOAuth2ClientForTest(t *testing.T) *oauth2clientspb.OAuth2Client {
	t.Helper()

	ctx := t.Context()
	input := clientInputForTest(t)

	created, err := adminClient.CreateOAuth2Client(ctx, &oauth2clientspb.CreateOAuth2ClientRequest{Input: input})
	require.NoError(t, err)
	require.NotNil(t, created.GetIssued())

	// The plaintext is here and only here: the row holds a digest, and no read reverses it.
	assert.NotEmpty(t, created.GetIssued().GetClientSecret(), "a registration was issued without a secret")

	client := created.GetIssued().GetClient()
	require.NotNil(t, client)
	assert.NotEmpty(t, client.GetId())
	assert.NotEmpty(t, client.GetClientId(), "a registration was issued without a client_id")
	assert.Equal(t, input.GetName(), client.GetName())
	assert.Equal(t, input.GetDescription(), client.GetDescription())
	assert.Equal(t, input.GetRedirectUris(), client.GetRedirectUris())
	assert.NotNil(t, client.GetCreatedAt())

	retrieved, err := adminClient.GetOAuth2Client(ctx, &oauth2clientspb.GetOAuth2ClientRequest{
		Oauth2ClientId: client.GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())
	assert.Equal(t, client.GetId(), retrieved.GetResult().GetId())
	assert.Equal(t, client.GetClientId(), retrieved.GetResult().GetClientId())

	return retrieved.GetResult()
}

func TestOAuth2Clients_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		created := createOAuth2ClientForTest(t)

		AssertAuditLogContainsFuzzyForResource(t, t.Context(), "oauth2_clients", created.GetId(), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "oauth2_clients", RelevantID: created.GetId()},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		created, err := c.CreateOAuth2Client(ctx, &oauth2clientspb.CreateOAuth2ClientRequest{
			Input: clientInputForTest(t),
		})
		require.Error(t, err)
		assert.Nil(t, created.GetIssued())
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := clientInputForTest(t)
		input.Name = ""

		created, err := adminClient.CreateOAuth2Client(ctx, &oauth2clientspb.CreateOAuth2ClientRequest{Input: input})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, created.GetIssued())
	})

	// A registration with nowhere to send a code can authenticate at /token and still never
	// complete an authorization request, which is a failure that would otherwise surface at
	// first use rather than at registration.
	T.Run("a registration with no redirect URI is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := clientInputForTest(t)
		input.RedirectUris = nil

		_, err := adminClient.CreateOAuth2Client(ctx, &oauth2clientspb.CreateOAuth2ClientRequest{Input: input})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		created, err := testClient.CreateOAuth2Client(ctx, &oauth2clientspb.CreateOAuth2ClientRequest{
			Input: clientInputForTest(t),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.Nil(t, created.GetIssued())
	})
}

func TestOAuth2Clients_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)

		retrieved, err := adminClient.GetOAuth2Client(ctx, &oauth2clientspb.GetOAuth2ClientRequest{
			Oauth2ClientId: created.GetId(),
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved.GetResult())
		assert.Equal(t, created.GetId(), retrieved.GetResult().GetId())
		assert.Equal(t, created.GetName(), retrieved.GetResult().GetName())
		assert.Equal(t, created.GetRedirectUris(), retrieved.GetResult().GetRedirectUris())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetOAuth2Client(ctx, &oauth2clientspb.GetOAuth2ClientRequest{Oauth2ClientId: created.GetId()})
		assert.Error(t, err)
	})

	// The registry is four first-party applications and the MCP server, and there is no
	// console a member reaches it from. Reading it used to be a member's grant, which made
	// every registration in the deployment — names, client_ids, redirect URIs, and the
	// secret digest the converter copied along with them — readable by anybody signed in.
	T.Run("non-admin users are forbidden from reading", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)
		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.GetOAuth2Client(ctx, &oauth2clientspb.GetOAuth2ClientRequest{
			Oauth2ClientId: created.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = testClient.ListOAuth2Clients(ctx, &oauth2clientspb.ListOAuth2ClientsRequest{})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetOAuth2Client(ctx, &oauth2clientspb.GetOAuth2ClientRequest{
			Oauth2ClientId: nonexistentID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestOAuth2Clients_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)

		_, err := adminClient.ArchiveOAuth2Client(ctx, &oauth2clientspb.ArchiveOAuth2ClientRequest{
			Oauth2ClientId: created.GetId(),
		})
		require.NoError(t, err)

		// Withdrawn rather than deleted — the row still answers a client_id resolution, so
		// tokens minted against it can be refused — but gone from every consumer read.
		_, err = adminClient.GetOAuth2Client(ctx, &oauth2clientspb.GetOAuth2ClientRequest{
			Oauth2ClientId: created.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		AssertAuditLogContainsFuzzyForResource(t, ctx, "oauth2_clients", created.GetId(), 15, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "oauth2_clients", RelevantID: created.GetId()},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveOAuth2Client(ctx, &oauth2clientspb.ArchiveOAuth2ClientRequest{
			Oauth2ClientId: created.GetId(),
		})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveOAuth2Client(ctx, &oauth2clientspb.ArchiveOAuth2ClientRequest{
			Oauth2ClientId: nonexistentID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createOAuth2ClientForTest(t)
		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveOAuth2Client(ctx, &oauth2clientspb.ArchiveOAuth2ClientRequest{
			Oauth2ClientId: created.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}

func TestOAuth2Clients_Listing(T *testing.T) {
	T.Parallel()

	created := make([]*oauth2clientspb.OAuth2Client, 0, exampleQuantity)
	for range exampleQuantity {
		created = append(created, createOAuth2ClientForTest(T))
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := adminClient.ListOAuth2Clients(ctx, &oauth2clientspb.ListOAuth2ClientsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.GetResults()), len(created))

		// No secret on any of them, which is the property the type split buys: a digest is
		// a stored credential and a listing is the worst place for one.
		for _, client := range retrieved.GetResults() {
			assert.NotEmpty(t, client.GetClientId(), "a listed registration has no client_id")
		}
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ListOAuth2Clients(ctx, &oauth2clientspb.ListOAuth2ClientsRequest{})
		assert.Error(t, err)
	})
}
