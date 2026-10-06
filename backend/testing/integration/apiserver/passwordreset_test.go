package integration

import (
	"testing"

	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset/passwordresetpb"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestPasswordReset_ThisApplicationsRules pins what this application supplies to platform's
// PasswordResetService: the mail goes through the outbox, the password answers to this
// application's floor, the reset is announced to the person, and the link's life is audited. The
// flow itself — single use, revocation of the other links, the anti-enumeration answer — is
// platform's conformance suite's; see conformance_test.go.
func TestPasswordReset_ThisApplicationsRules(T *testing.T) {
	T.Parallel()

	T.Run("a reset is mailed through the outbox, completes, and is announced and audited", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := unauthedClient.RequestPasswordReset(ctx, &passwordresetpb.RequestPasswordResetRequest{
			EmailAddress: user.EmailAddress,
		})
		require.NoError(t, err)

		// The secret rides on the event the reset mail is rendered from, and nowhere else.
		secret, err := conformancePasswordResetToken(ctx, tenancy.Global(), user.EmailAddress)
		require.NoError(t, err)

		newPassword := user.HashedPassword + "-reset"

		_, err = unauthedClient.CompletePasswordReset(ctx, &passwordresetpb.CompletePasswordResetRequest{
			Token:       secret,
			NewPassword: newPassword,
		})
		require.NoError(t, err)

		_, err = unauthedClient.LoginForToken(ctx, &signinpb.LoginForTokenRequest{Credentials: &signinpb.Credentials{
			Username: user.Username,
			Password: newPassword,
			TotpCode: generateTOTPCodeForUserForTest(t, user),
		}})
		require.NoError(t, err)

		// The "your password was reset" mail is rendered from the event platform's hooks on the
		// token store emit when the link is spent, on the reset's own transaction.
		redeemed, err := outboxPayloads(ctx,
			`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
			user.ID, passwordreset.EventTokenRedeemed.String())
		require.NoError(t, err)
		assert.Len(t, redeemed, 1)

		// Issued, then spent: the two writes platform records. By resource rather than by actor,
		// because a reset is anonymous — the request that asks for the link and the one that
		// spends it carry no principal — so platform files both entries under the user as their
		// subject and names nobody as the actor. The token's ID is on the mail request beside
		// the secret.
		tokenID := passwordResetTokenIDForTest(t, user.ID)
		AssertAuditLogContainsFuzzyForResource(t, ctx, passwordreset.ResourceTypeToken, tokenID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: passwordreset.ResourceTypeToken, RelevantID: tokenID},
			{EventType: "updated", ResourceType: passwordreset.ResourceTypeToken, RelevantID: tokenID},
		})
	})

	T.Run("a reset password answers to this application's floor, and the link survives the refusal", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := unauthedClient.RequestPasswordReset(ctx, &passwordresetpb.RequestPasswordResetRequest{
			EmailAddress: user.EmailAddress,
		})
		require.NoError(t, err)

		secret, err := conformancePasswordResetToken(ctx, tenancy.Global(), user.EmailAddress)
		require.NoError(t, err)

		_, err = unauthedClient.CompletePasswordReset(ctx, &passwordresetpb.CompletePasswordResetRequest{
			Token:       secret,
			NewPassword: "password",
		})
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		_, err = unauthedClient.CompletePasswordReset(ctx, &passwordresetpb.CompletePasswordResetRequest{
			Token:       secret,
			NewPassword: user.HashedPassword + "-reset",
		})
		require.NoError(t, err)
	})
}

// passwordResetTokenIDForTest reads the ID of the newest reset link mailed to userID off the mail
// request that carries it, which is the only place a test can learn it: the response to the
// request names nothing, deliberately.
func passwordResetTokenIDForTest(t *testing.T, userID string) string {
	t.Helper()

	payloads, err := outboxPayloads(t.Context(),
		`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
		userID, ddbidentity.PasswordResetTokenCreatedEventType)
	require.NoError(t, err)

	for _, payload := range payloads {
		if id := findStringKey(payload, authkeys.PasswordResetTokenIDKey); id != "" {
			return id
		}
	}

	require.FailNow(t, "no reset mail request names a token for the user")

	return ""
}
