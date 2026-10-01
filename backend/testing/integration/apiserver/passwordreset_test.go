package integration

import (
	"testing"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset/passwordresetpb"
	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"
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

		user, testClient := createUserAndClientForTest(t)
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

		// The "your password was reset" mail is queued on the reset's own transaction.
		redeemed, err := outboxPayloads(ctx,
			`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%'`,
			user.ID, ddbidentity.PasswordResetTokenRedeemedEventType)
		require.NoError(t, err)
		assert.Len(t, redeemed, 1)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "password_reset_tokens"},
			{EventType: "updated", ResourceType: "password_reset_tokens"},
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
