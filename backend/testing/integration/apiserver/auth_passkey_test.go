package integration

import (
	"encoding/json"
	"testing"

	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v14/authentication/passkeys/passkeyspb"
	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"
	identity "github.com/primandproper/platform-go/v14/identity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// challengeFromOptions reads the challenge out of the options PasskeysService hands a browser,
// which is what the browser signs.
func challengeFromOptions(t *testing.T, options []byte) string {
	t.Helper()

	var parsed struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
		} `json:"publicKey"`
	}
	require.NoError(t, json.Unmarshal(options, &parsed))
	require.NotEmpty(t, parsed.PublicKey.Challenge)

	return parsed.PublicKey.Challenge
}

// registerPasskeyForTest runs a full registration ceremony through PasskeysService and returns
// the device that now holds the user's passkey.
func registerPasskeyForTest(t *testing.T, authedClient client.Client) *virtualAuthenticator {
	t.Helper()

	ctx := t.Context()
	authenticator := newVirtualAuthenticator(t)

	begin, err := authedClient.BeginRegistration(ctx, &passkeyspb.BeginRegistrationRequest{})
	require.NoError(t, err)

	_, err = authedClient.FinishRegistration(ctx, &passkeyspb.FinishRegistrationRequest{
		Response:     authenticator.register(t, challengeFromOptions(t, begin.GetOptions())),
		FriendlyName: t.Name(),
	})
	require.NoError(t, err)

	return authenticator
}

// webAuthnUserHandle is the handle this application registers a passkey under: the user's ID,
// as it was before PasskeysService was mounted, so a credential registered then is still found.
func webAuthnUserHandle(user *identity.User) []byte {
	return []byte(user.ID)
}

// TestPasskeys_ThisApplicationsWiring pins what this application supplies to platform's
// PasskeysService: which user a handle and a username name, and the door a finished login is
// minted through. The ceremony itself is platform's conformance suite's; see conformance_test.go.
func TestPasskeys_ThisApplicationsWiring(T *testing.T) {
	T.Parallel()

	T.Run("a discoverable login signs in the user the handle names, with a token this application's services accept", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := registerPasskeyForTest(t, testClient)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		begin, err := unauthedClient.BeginLogin(ctx, &passkeyspb.BeginLoginRequest{})
		require.NoError(t, err)

		finished, err := unauthedClient.FinishLogin(ctx, &passkeyspb.FinishLoginRequest{
			Response: authenticator.assert(t, challengeFromOptions(t, begin.GetOptions()), webAuthnUserHandle(user)),
		})
		require.NoError(t, err)
		require.NotEmpty(t, finished.GetToken().GetToken())

		signedIn, err := buildAuthedGRPCClientWithBearerToken(finished.GetToken().GetToken())
		require.NoError(t, err)

		_, err = signedIn.GetValidVessels(ctx, &mealplanningsvc.GetValidVesselsRequest{})
		require.NoError(t, err)

		status, err := signedIn.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, user.ID, status.GetStatus().GetUser().GetId())
	})

	T.Run("a named login finds the user by their username", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := registerPasskeyForTest(t, testClient)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		begin, err := unauthedClient.BeginLogin(ctx, &passkeyspb.BeginLoginRequest{Username: user.Username})
		require.NoError(t, err)

		finished, err := unauthedClient.FinishLogin(ctx, &passkeyspb.FinishLoginRequest{
			Username: user.Username,
			Response: authenticator.assert(t, challengeFromOptions(t, begin.GetOptions()), nil),
		})
		require.NoError(t, err)
		require.NotEmpty(t, finished.GetToken().GetToken())
	})
}
