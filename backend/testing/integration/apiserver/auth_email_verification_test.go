package integration

import (
	"fmt"
	"testing"
	"time"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// These pin, over the wire, the four properties #1385 set out to buy for email verification
// by adopting platform's links package: a link expires, works once, is not stored as itself,
// and dies when the address it was mailed to is left behind.
//
// links was not adopted, because by the time it could be the properties had already arrived
// by another road. The v14 port put verification on platform's identity and signin, which
// store a digest beside a deadline and burn it with a guarded write whose row count is
// checked. A links record beside that column would be a second store for one credential, and
// its Resolve commits by itself, where signin.VerifyEmailAddress spends the link and records
// the proof in one transaction — the objection platform's magiclinks package makes to links
// for the same shape of flow.
//
// What was missing was anything in this repository that would notice losing them. The
// helper the rest of the suite uses to verify an address goes around the link on purpose, so
// until these, nothing here ever redeemed one.

// verificationStoreForTest is the identity store the server verifies against, over the same
// database.
func verificationStoreForTest(t *testing.T) *identity.SQLStore {
	t.Helper()

	store, err := identity.NewSQLStore(databaseClient, identity.WithTablePrefix(ddbidentity.TablePrefix))
	require.NoError(t, err)

	return store
}

// issueVerificationLinkForTest issues a verification link for userID that dies at expiresAt,
// and returns the token it carries.
//
// The token is minted and written here, through the same store call RequestEmailVerificationEmail
// makes, because the RPC mails it and nothing can read it back afterwards — the column holds a
// digest, which is the property under test.
func issueVerificationLinkForTest(t *testing.T, userID string, expiresAt time.Time) string {
	t.Helper()
	ctx := t.Context()

	token, err := random.GenerateBase32EncodedString(ctx, 64)
	require.NoError(t, err)

	store := verificationStoreForTest(t)
	require.NoError(t, databaseClient.WithTransaction(ctx, func(tx database.Tx) error {
		return store.SetUserEmailAddressVerificationToken(ctx, tx, ddbidentity.Scope(), userID, token, expiresAt)
	}))

	return token
}

// storedUserForTest reads the user row as the store holds it.
func storedUserForTest(t *testing.T, userID string) *identity.User {
	t.Helper()

	user, err := verificationStoreForTest(t).GetUser(t.Context(), databaseClient.Reader(), ddbidentity.Scope(), userID)
	require.NoError(t, err)

	return user
}

// requireVerificationRefused asserts that redeeming token is refused the way a wrong password
// is. Expired, spent, never issued and wrong are one answer on purpose — see
// signin.ErrInvalidVerificationToken — so every refusal below asserts the same code.
func requireVerificationRefused(t *testing.T, token string) {
	t.Helper()

	res, err := buildUnauthenticatedGRPCClientForTest(t).VerifyEmailAddress(t.Context(), &authsvc.VerifyEmailAddressRequest{Token: token})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, codes.Unauthenticated, status.Code(err), "a refused verification link reads as invalid credentials, not as a server fault")
}

func TestAuth_VerifyEmailAddress(T *testing.T) {
	T.Parallel()

	T.Run("verifies the address the link was mailed to", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		require.False(t, storedUserForTest(t, user.ID).EmailAddressVerified())

		token := issueVerificationLinkForTest(t, user.ID, time.Now().Add(time.Hour))

		res, err := buildUnauthenticatedGRPCClientForTest(t).VerifyEmailAddress(ctx, &authsvc.VerifyEmailAddressRequest{Token: token})
		require.NoError(t, err)
		assert.True(t, res.GetVerified())

		stored := storedUserForTest(t, user.ID)
		assert.True(t, stored.EmailAddressVerified())
		assert.Empty(t, stored.EmailAddressVerificationTokenDigest, "verifying burns the link")
		assert.Nil(t, stored.EmailAddressVerificationTokenExpiresAt)
	})

	// The defect #1385 was opened over: the guarded write discarded its row count, so the
	// second click matched nothing, returned nil, and answered Verified: true again.
	T.Run("works once", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		token := issueVerificationLinkForTest(t, user.ID, time.Now().Add(time.Hour))

		_, err := buildUnauthenticatedGRPCClientForTest(t).VerifyEmailAddress(ctx, &authsvc.VerifyEmailAddressRequest{Token: token})
		require.NoError(t, err)
		verifiedAt := storedUserForTest(t, user.ID).EmailAddressVerifiedAt
		require.NotNil(t, verifiedAt)

		requireVerificationRefused(t, token)

		stored := storedUserForTest(t, user.ID)
		require.NotNil(t, stored.EmailAddressVerifiedAt)
		assert.True(t, verifiedAt.Equal(*stored.EmailAddressVerifiedAt), "a replay must not re-stamp the proof")
	})

	T.Run("is refused once it has expired", func(t *testing.T) {
		t.Parallel()

		user, _ := createUserAndClientForTest(t)
		token := issueVerificationLinkForTest(t, user.ID, time.Now().Add(-time.Minute))

		requireVerificationRefused(t, token)
		assert.False(t, storedUserForTest(t, user.ID).EmailAddressVerified())
	})

	T.Run("is not stored as itself", func(t *testing.T) {
		t.Parallel()

		user, _ := createUserAndClientForTest(t)
		token := issueVerificationLinkForTest(t, user.ID, time.Now().Add(time.Hour))

		stored := storedUserForTest(t, user.ID)
		require.NotEmpty(t, stored.EmailAddressVerificationTokenDigest, "an issued link is recorded")
		assert.NotEqual(t, token, stored.EmailAddressVerificationTokenDigest)
		assert.NotContains(t, stored.EmailAddressVerificationTokenDigest, token)
		assert.Empty(t, stored.EmailAddressVerificationToken, "nothing reads the secret back out of the row")
	})

	T.Run("is retired when another is issued", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		first := issueVerificationLinkForTest(t, user.ID, time.Now().Add(time.Hour))
		second := issueVerificationLinkForTest(t, user.ID, time.Now().Add(time.Hour))

		requireVerificationRefused(t, first)

		_, err := buildUnauthenticatedGRPCClientForTest(t).VerifyEmailAddress(ctx, &authsvc.VerifyEmailAddressRequest{Token: second})
		require.NoError(t, err)
		assert.True(t, storedUserForTest(t, user.ID).EmailAddressVerified())
	})

	// A link proves control of the inbox it was mailed to and of no other, so moving to a new
	// address has to kill it; otherwise the old inbox verifies the new address.
	T.Run("dies when the address changes", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		token := issueVerificationLinkForTest(t, user.ID, time.Now().Add(time.Hour))

		_, err := testClient.UpdateUserEmailAddress(ctx, &authsvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: fmt.Sprintf("moved_%d@whatever.com", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)

		requireVerificationRefused(t, token)
		assert.False(t, storedUserForTest(t, user.ID).EmailAddressVerified())
	})

	T.Run("refuses a token that was never issued", func(t *testing.T) {
		t.Parallel()

		token, err := random.GenerateBase32EncodedString(t.Context(), 64)
		require.NoError(t, err)

		requireVerificationRefused(t, token)
	})
}
