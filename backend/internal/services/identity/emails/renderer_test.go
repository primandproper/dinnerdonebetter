package emails

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	waitlistfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	identity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/platform-go/v15/links"
	"github.com/primandproper/platform-go/v15/notifications/mail"
	waitlistsgrpc "github.com/primandproper/platform-go/v15/waitlists/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/retry"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleBaseURL = "https://example.com"

func buildRendererForTest(directory *identitymock.StoreMock) *Renderer {
	return NewRenderer(directory, &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
	}, exampleBaseURL)
}

func verifiedUserForTest() *identity.User {
	user := fakes.BuildFakeUser()
	user.EmailAddressVerifiedAt = new(time.Now())

	return user
}

func TestRenderer_Render(T *testing.T) {
	T.Parallel()

	T.Run("an invitation, in its sender's name, with the token in its link", func(t *testing.T) {
		t.Parallel()

		sender := fakes.BuildFakeUser()
		invitation := fakes.BuildFakeInvitationFromUserToAccount(sender.ID, fake.BuildFakeID())
		// As platform hands it over: the row redacted, the secret beside it.
		invitation.Token = ""
		token := fake.BuildFakeID()

		directory := &identitymock.StoreMock{
			GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, scope tenancy.Scope, userID string) (*identity.User, error) {
				assert.True(t, scope.IsGlobal())
				assert.Equal(t, sender.ID, userID)

				return sender, nil
			},
		}

		actual, err := buildRendererForTest(directory).Render(t.Context(), &mail.Mail{
			Kind:       mail.KindInvitation,
			Invitation: &identity.InvitationMail{Invitation: invitation, Token: token},
		})
		require.NoError(t, err)

		assert.Equal(t, invitation.ToEmail, actual.ToAddress)
		assert.Contains(t, actual.HTMLContent, token)
		assert.Contains(t, actual.HTMLContent, invitation.ID)
		// The redacted row the mailer was handed is not the one the link is built on.
		assert.Empty(t, invitation.Token)
		assert.Len(t, directory.GetUserCalls(), 1)
	})

	T.Run("an invitation whose sender cannot be read is retried", func(t *testing.T) {
		t.Parallel()

		readErr := errors.New(fake.BuildFakeString())
		directory := &identitymock.StoreMock{
			GetUserFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
				return nil, readErr
			},
		}

		_, err := buildRendererForTest(directory).Render(t.Context(), &mail.Mail{
			Kind:       mail.KindInvitation,
			Invitation: &identity.InvitationMail{Invitation: fakes.BuildFakeInvitation(), Token: fake.BuildFakeID()},
		})
		require.ErrorIs(t, err, readErr)
		assert.NotErrorIs(t, err, retry.ErrUnretryable, "a directory that is down now may not be later")
	})

	T.Run("a verification link", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		token := fake.BuildFakeID()

		actual, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), &mail.Mail{
			Kind:         mail.KindVerification,
			Verification: &signin.VerificationMail{User: user, Token: token, ExpiresAt: time.Now().Add(time.Hour)},
		})
		require.NoError(t, err)

		assert.Equal(t, user.EmailAddress, actual.ToAddress)
		assert.Contains(t, actual.HTMLContent, token)
	})

	T.Run("a handle reminder", func(t *testing.T) {
		t.Parallel()

		user := verifiedUserForTest()

		actual, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), &mail.Mail{
			Kind:           mail.KindHandleReminder,
			HandleReminder: &signin.HandleReminderMail{User: user},
		})
		require.NoError(t, err)

		assert.Equal(t, user.EmailAddress, actual.ToAddress)
		assert.Contains(t, strings.ToLower(actual.HTMLContent), strings.ToLower(user.Username))
	})

	T.Run("a password reset, with the issuance's secret in its link", func(t *testing.T) {
		t.Parallel()

		user := verifiedUserForTest()
		secret := fake.BuildFakeID()

		actual, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), &mail.Mail{
			Kind: mail.KindPasswordReset,
			PasswordReset: &passwordreset.Mail{
				User:     user,
				Issuance: &passwordreset.Issuance{Token: &passwordreset.Token{ID: fake.BuildFakeID(), UserID: user.ID}, Secret: secret},
			},
		})
		require.NoError(t, err)

		assert.Equal(t, user.EmailAddress, actual.ToAddress)
		assert.Contains(t, actual.HTMLContent, secret)
	})

	T.Run("a waitlist confirmation", func(t *testing.T) {
		t.Parallel()

		confirmation := &waitlistsgrpc.ConfirmationMail{
			Signup:      waitlistfakes.BuildFakeWaitlistSignup(),
			Confirm:     &links.Link{URL: exampleBaseURL + "/waitlists/confirm?t=" + fake.BuildFakeID()},
			Unsubscribe: &links.Link{URL: exampleBaseURL + "/waitlists/unsubscribe?t=" + fake.BuildFakeID()},
		}

		actual, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), &mail.Mail{
			Kind:                 mail.KindWaitlistConfirmation,
			WaitlistConfirmation: &mail.WaitlistConfirmation{Mail: confirmation, Scope: tenancy.Global()},
		})
		require.NoError(t, err)

		assert.Equal(t, confirmation.Signup.Contact, actual.ToAddress)
		assert.Contains(t, actual.HTMLContent, confirmation.Confirm.URL)
		assert.Contains(t, actual.HTMLContent, confirmation.Unsubscribe.URL)
	})

	T.Run("a mail its builder refuses is not retried", func(t *testing.T) {
		t.Parallel()

		// A reset to an address nobody has proven: the builder refuses it, and it will refuse
		// it exactly the same way on every later attempt.
		user := fakes.BuildFakeUser()

		_, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), &mail.Mail{
			Kind: mail.KindPasswordReset,
			PasswordReset: &passwordreset.Mail{
				User:     user,
				Issuance: &passwordreset.Issuance{Token: &passwordreset.Token{ID: fake.BuildFakeID()}, Secret: fake.BuildFakeID()},
			},
		})
		require.ErrorIs(t, err, ErrUnverifiedEmailRecipient)
		assert.ErrorIs(t, err, retry.ErrUnretryable)
	})

	T.Run("a kind this application sends none of is not retried", func(t *testing.T) {
		t.Parallel()

		_, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), &mail.Mail{
			Kind:      mail.KindMagicLink,
			MagicLink: &signin.MagicLinkMail{User: fakes.BuildFakeUser()},
		})
		require.ErrorIs(t, err, ErrUnrenderedKind)
		assert.ErrorIs(t, err, retry.ErrUnretryable)
	})

	T.Run("with nil mail", func(t *testing.T) {
		t.Parallel()

		_, err := buildRendererForTest(&identitymock.StoreMock{}).Render(t.Context(), nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, retry.ErrUnretryable)
	})
}

func TestBuildWaitlistConfirmationEmail(T *testing.T) {
	T.Parallel()

	T.Run("carries both links to the address that joined", func(t *testing.T) {
		t.Parallel()

		confirmation := &waitlistsgrpc.ConfirmationMail{
			Signup:      waitlistfakes.BuildFakeWaitlistSignup(),
			Confirm:     &links.Link{URL: exampleBaseURL + "/waitlists/confirm?t=" + fake.BuildFakeID()},
			Unsubscribe: &links.Link{URL: exampleBaseURL + "/waitlists/unsubscribe?t=" + fake.BuildFakeID()},
		}

		actual, err := BuildWaitlistConfirmationEmail(confirmation, exampleBaseURL)
		require.NoError(t, err)

		assert.Equal(t, confirmation.Signup.Contact, actual.ToAddress)
		assert.Equal(t, confirmation.Signup.Subject.ID, actual.UserID)
		assert.Contains(t, actual.HTMLContent, confirmation.Confirm.URL)
		assert.Contains(t, actual.HTMLContent, confirmation.Unsubscribe.URL)
	})

	T.Run("with a link missing", func(t *testing.T) {
		t.Parallel()

		actual, err := BuildWaitlistConfirmationEmail(&waitlistsgrpc.ConfirmationMail{
			Signup:  waitlistfakes.BuildFakeWaitlistSignup(),
			Confirm: &links.Link{URL: exampleBaseURL + "/waitlists/confirm?t=" + fake.BuildFakeID()},
		}, exampleBaseURL)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}
