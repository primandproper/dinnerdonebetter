package queuedmail

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/emails"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	identity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/platform-go/v15/notifications/mail"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/email"
	emailmock "github.com/primandproper/primitives-go/v2/email/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queued is one outbox row the mailer would have inserted: the topic and the bytes the relay
// would publish.
type queued struct {
	topic   string
	payload []byte
}

// transport is the queued mail path end to end, over a mock database: the QueuedMailer New
// builds, the rows its outbox writer would insert, and the Drainer the async message handler
// runs over this application's Renderer, sending into a mock emailer.
type transport struct {
	mailer    *mail.QueuedMailer
	drainer   *mail.Drainer
	directory *identitymock.StoreMock
	topic     string
	sent      []*email.OutboundEmailMessage
	rows      []queued
}

func buildTransport(t *testing.T) *transport {
	t.Helper()

	tr := &transport{topic: fake.BuildFakeID(), directory: &identitymock.StoreMock{}}

	executor := &mockdatabase.SQLQueryExecutorMock{
		ExecContextFunc: func(_ context.Context, _ string, args ...any) (sql.Result, error) {
			row := queued{}
			for _, arg := range args {
				switch v := arg.(type) {
				case string:
					if v == tr.topic {
						row.topic = v
					}
				case []byte:
					row.payload = v
				}
			}
			tr.rows = append(tr.rows, row)

			return nil, nil
		},
	}

	db := &mockdatabase.ClientMock{
		DialectFunc: func() dialect.Dialect { return dialect.Postgres },
		ReaderFunc:  func() database.SQLQueryExecutor { return executor },
		WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error {
			return fn(database.NewTxForTesting(executor))
		},
	}

	var err error
	tr.mailer, err = New(db, tr.topic, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider())
	require.NoError(t, err)

	emailer := &emailmock.EmailerMock{
		SendEmailFunc: func(_ context.Context, msg *email.OutboundEmailMessage) error {
			tr.sent = append(tr.sent, msg)
			return nil
		},
	}

	tr.drainer, err = mail.NewDrainer(emailer, emails.NewRenderer(tr.directory, db, "https://example.com"))
	require.NoError(t, err)

	return tr
}

// drain hands every row queued so far to the Drainer, as the pool on the topic would.
func (tr *transport) drain(t *testing.T) {
	t.Helper()

	for i := range tr.rows {
		require.Equal(t, tr.topic, tr.rows[i].topic, "a mail was queued on a topic other than the mailer's")
		require.NoError(t, tr.drainer.Handle(t.Context(), tr.rows[i].payload))
	}
}

func TestNew(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil database client", func(t *testing.T) {
		t.Parallel()

		_, err := New(nil, fake.BuildFakeID(), loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider())
		require.Error(t, err)
	})

	T.Run("refuses an empty topic", func(t *testing.T) {
		t.Parallel()

		db := &mockdatabase.ClientMock{DialectFunc: func() dialect.Dialect { return dialect.Postgres }}

		_, err := New(db, "", loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider())
		require.Error(t, err)
	})
}

// TestQueuedMail pins the round trip every identity, sign-in, reset and waitlist mail takes: what
// a door hands the mailer is what the recipient's mail carries, secret included, once the
// Drainer has rendered it in this application's words.
func TestQueuedMail(T *testing.T) {
	T.Parallel()

	T.Run("a password reset arrives with its secret in the link", func(t *testing.T) {
		t.Parallel()

		tr := buildTransport(t)
		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = new(time.Now())
		secret := fake.BuildFakeID()

		require.NoError(t, tr.mailer.SendPasswordReset(t.Context(), &passwordreset.Mail{
			User:     user,
			Issuance: &passwordreset.Issuance{Token: &passwordreset.Token{ID: fake.BuildFakeID(), UserID: user.ID}, Secret: secret},
		}))
		tr.drain(t)

		require.Len(t, tr.sent, 1)
		assert.Equal(t, user.EmailAddress, tr.sent[0].ToAddress)
		assert.Contains(t, tr.sent[0].HTMLContent, secret)
	})

	T.Run("an invitation arrives in its sender's name with its token", func(t *testing.T) {
		t.Parallel()

		tr := buildTransport(t)
		sender := fakes.BuildFakeUser()
		invitation := fakes.BuildFakeInvitationFromUserToAccount(sender.ID, fake.BuildFakeID())
		token := fake.BuildFakeID()

		tr.directory.GetUserFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*identity.User, error) {
			assert.Equal(t, sender.ID, userID)

			return sender, nil
		}

		require.NoError(t, tr.mailer.SendInvitation(t.Context(), &identity.InvitationMail{Invitation: invitation, Token: token}))
		tr.drain(t)

		require.Len(t, tr.sent, 1)
		assert.Equal(t, invitation.ToEmail, tr.sent[0].ToAddress)
		assert.Contains(t, tr.sent[0].HTMLContent, token)
	})

	T.Run("a verification link arrives with its token", func(t *testing.T) {
		t.Parallel()

		tr := buildTransport(t)
		user := fakes.BuildFakeUser()
		token := fake.BuildFakeID()

		require.NoError(t, tr.mailer.SendVerification(t.Context(), &signin.VerificationMail{
			User: user, Token: token, ExpiresAt: time.Now().Add(time.Hour),
		}))
		tr.drain(t)

		require.Len(t, tr.sent, 1)
		assert.Equal(t, user.EmailAddress, tr.sent[0].ToAddress)
		assert.Contains(t, tr.sent[0].HTMLContent, token)
	})

	T.Run("the secret is on the queued row and nowhere the topic does not reach", func(t *testing.T) {
		t.Parallel()

		// The mail topic is a credential store for as long as a message sits on it, which is
		// why it has a writer of its own: one row, on the mailer's topic, and no side effect
		// deriving anything else from it.
		tr := buildTransport(t)
		user := fakes.BuildFakeUser()
		token := fake.BuildFakeID()

		require.NoError(t, tr.mailer.SendVerification(t.Context(), &signin.VerificationMail{
			User: user, Token: token, ExpiresAt: time.Now().Add(time.Hour),
		}))

		require.Len(t, tr.rows, 1)
		assert.Equal(t, tr.topic, tr.rows[0].topic)
		assert.Contains(t, string(tr.rows[0].payload), token)
	})
}
