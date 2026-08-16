package integration

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countCeremonySessions reports how many ceremony rows are stored under challenge.
//
// The ceremony store is a table rather than a map in the API process, which is the whole point
// of it: a challenge issued by one replica has to be answerable by another. Reading the row from
// outside the server is the closest this single-process suite gets to being a second replica,
// and it is what distinguishes state that is durable from state that merely works here.
func countCeremonySessions(t *testing.T, ctx context.Context, challenge string) int {
	t.Helper()

	var count int
	require.NoError(t, databaseClient.Reader().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM webauthn_sessions WHERE challenge = $1`, challenge).Scan(&count))

	return count
}

// registerPasskeyForTest runs a full registration ceremony and returns the device that now holds
// the user's passkey.
func registerPasskeyForTest(t *testing.T, authedClient client.Client) *virtualAuthenticator {
	t.Helper()

	ctx := t.Context()
	authenticator := newVirtualAuthenticator(t)

	beginRes, err := authedClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, beginRes.Challenge)

	_, err = authedClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
		AttestationResponse: authenticator.register(t, beginRes.Challenge),
		Challenge:           beginRes.Challenge,
	})
	require.NoError(t, err)

	return authenticator
}

func TestAuth_PasskeyCeremony(T *testing.T) {
	T.Parallel()

	T.Run("registers a passkey and logs in with it", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := newVirtualAuthenticator(t)

		beginRegistration, err := testClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, beginRegistration.Challenge)

		// The challenge is in the table, not in the process that issued it.
		assert.Equal(t, 1, countCeremonySessions(t, ctx, beginRegistration.Challenge))

		_, err = testClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			AttestationResponse: authenticator.register(t, beginRegistration.Challenge),
			Challenge:           beginRegistration.Challenge,
		})
		require.NoError(t, err)

		// And it is gone once answered — consumed, not left to expire.
		assert.Equal(t, 0, countCeremonySessions(t, ctx, beginRegistration.Challenge))

		listRes, err := testClient.ListPasskeys(ctx, &authsvc.ListPasskeysRequest{})
		require.NoError(t, err)
		require.Len(t, listRes.Results, 1)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		beginLogin, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
			Username: user.Username,
		})
		require.NoError(t, err)
		require.NotEmpty(t, beginLogin.Challenge)

		tokenRes, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Username:          user.Username,
			AssertionResponse: authenticator.assert(t, beginLogin.Challenge, nil),
			Challenge:         beginLogin.Challenge,
		})
		require.NoError(t, err)
		require.NotNil(t, tokenRes.Result)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)

		// The token a passkey login issues is a token, not a receipt.
		jwtClient, err := buildAuthedGRPCClientWithBearerToken(tokenRes.Result.AccessToken)
		require.NoError(t, err)

		statusRes, err := jwtClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, user.ID, statusRes.UserId)
	})

	// The defect this replaces: the ceremony used to be read and deleted in two calls, and the
	// delete was best-effort on the success path only. An assertion answered twice inside its
	// TTL was accepted twice.
	T.Run("refuses an assertion replayed inside its TTL", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := registerPasskeyForTest(t, testClient)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		beginLogin, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
			Username: user.Username,
		})
		require.NoError(t, err)

		assertion := authenticator.assert(t, beginLogin.Challenge, nil)

		first, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Username:          user.Username,
			AssertionResponse: assertion,
			Challenge:         beginLogin.Challenge,
		})
		require.NoError(t, err)
		require.NotEmpty(t, first.Result.AccessToken)

		second, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Username:          user.Username,
			AssertionResponse: assertion,
			Challenge:         beginLogin.Challenge,
		})
		require.Error(t, err, "the same assertion must not be answerable twice")
		assert.Nil(t, second)
	})

	// A registration attestation is one-shot for the same reason a login assertion is.
	T.Run("refuses an attestation replayed inside its TTL", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		authenticator := newVirtualAuthenticator(t)

		beginRes, err := testClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, beginRes.Challenge)

		attestation := authenticator.register(t, beginRes.Challenge)

		_, err = testClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			AttestationResponse: attestation,
			Challenge:           beginRes.Challenge,
		})
		require.NoError(t, err)

		_, err = testClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			AttestationResponse: attestation,
			Challenge:           beginRes.Challenge,
		})
		require.Error(t, err)

		listRes, err := testClient.ListPasskeys(ctx, &authsvc.ListPasskeysRequest{})
		require.NoError(t, err)
		assert.Len(t, listRes.Results, 1, "a replayed attestation must not register a second credential")
	})

	// Usernameless login: the passkey names the user, and the ceremony that was begun carries
	// no user handle to be answered against.
	T.Run("logs in without a username", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := registerPasskeyForTest(t, testClient)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		beginLogin, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
			Username: "",
		})
		require.NoError(t, err)

		tokenRes, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Username:          "",
			AssertionResponse: authenticator.assert(t, beginLogin.Challenge, webAuthnUserHandle(user)),
			Challenge:         beginLogin.Challenge,
		})
		require.NoError(t, err)
		require.NotNil(t, tokenRes.Result)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})

	// Clone detection is only worth anything if the count is written back. Two logins in a row
	// means the second is compared against the count the first recorded, not against the count
	// the credential was registered with.
	T.Run("records the sign count between logins", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := registerPasskeyForTest(t, testClient)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		var signCounts []uint32

		for range 2 {
			beginLogin, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
				Username: user.Username,
			})
			require.NoError(t, err)

			_, err = unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
				Username:          user.Username,
				AssertionResponse: authenticator.assert(t, beginLogin.Challenge, nil),
				Challenge:         beginLogin.Challenge,
			})
			require.NoError(t, err)

			var stored uint32
			require.NoError(t, databaseClient.Reader().QueryRowContext(ctx,
				`SELECT sign_count FROM webauthn_credentials WHERE belongs_to_user = $1`, user.ID).Scan(&stored))

			signCounts = append(signCounts, stored)
		}

		require.Len(t, signCounts, 2)
		assert.Greater(t, signCounts[1], signCounts[0])
	})

	// An assertion signed for a challenge this server never issued has nothing to be answered
	// against, whether or not the signature itself is good.
	T.Run("refuses a challenge it never issued", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		authenticator := registerPasskeyForTest(t, testClient)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		const forged = "Zm9yZ2VkLWNoYWxsZW5nZS10aGF0LXdhcy1uZXZlci1pc3N1ZWQ"

		res, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Username:          user.Username,
			AssertionResponse: authenticator.assert(t, forged, nil),
			Challenge:         forged,
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})
}

// webAuthnUserHandle is the opaque handle the adapter registers a user under, which is what a
// discoverable assertion echoes back.
func webAuthnUserHandle(user *identity.User) []byte {
	return []byte(user.ID)
}
