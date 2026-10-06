package authentication

import (
	"context"
	"database/sql"
	"testing"
	"time"

	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events/eventstest"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformauditmock "github.com/primandproper/platform-go/v15/audit/mock"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mailersHarness is a SignInMailers over a mock executor that keeps every statement it was
// handed, so a test reads the event off the outbox write that would have carried it rather than
// off anything the mailers expose.
type mailersHarness struct {
	mailers  *SignInMailers
	executor *mockdatabase.SQLQueryExecutorMock
}

func buildMailersHarness(t *testing.T) *mailersHarness {
	t.Helper()

	writer, err := outbox.NewWriter(dialect.Postgres)
	require.NoError(t, err)

	executor := &mockdatabase.SQLQueryExecutorMock{
		ExecContextFunc: func(context.Context, string, ...any) (sql.Result, error) { return nil, nil },
	}
	tx := database.NewTxForTesting(executor)

	db := &mockdatabase.ClientMock{
		WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error { return fn(tx) },
	}

	// A mailer emits and never records, so the audit recorder behind the spine is a mock that
	// must not be reached; it accepts rather than fails so that a stray entry would show up in
	// its call count, not as a panic.
	audited := &platformauditmock.RecorderMock{
		RecordFunc: func(context.Context, database.Tx, tenancy.Scope, ...*platformaudit.Entry) error { return nil },
	}

	return &mailersHarness{
		mailers:  NewSignInMailers(loggingnoop.NewLogger(), db, eventstest.New(t, writer, audited)),
		executor: executor,
	}
}

// enqueued is every data change message of eventType the transaction was asked to write, read
// off the statements that would have enqueued them.
//
// The stored body is platform's envelope around this application's message — the event's name
// beside its payload — so each one is read the way the broker consumer reads it, through Decode.
// Any argument that is not an envelope naming eventType is some other column's value.
func (h *mailersHarness) enqueued(t *testing.T, eventType string) []*datachanges.Message {
	t.Helper()

	var out []*datachanges.Message

	calls := h.executor.ExecContextCalls()
	for i := range calls {
		for _, arg := range calls[i].Args {
			raw, ok := arg.([]byte)
			if !ok {
				continue
			}

			msg := &datachanges.Message{}
			if _, matched, err := webhooks.Decode(raw, msg, webhooks.EventType(eventType)); err == nil && matched {
				out = append(out, msg)
			}
		}
	}

	return out
}

func TestSignInMailers_SendPasswordReset(T *testing.T) {
	T.Parallel()

	T.Run("queues the reset mail carrying the link's secret", func(t *testing.T) {
		t.Parallel()

		harness := buildMailersHarness(t)

		user := &identity.User{ID: identifiers.New()}
		issuance := &passwordreset.Issuance{
			Token:  &passwordreset.Token{ID: identifiers.New(), ExpiresAt: time.Now().Add(time.Hour)},
			Secret: identifiers.New(),
		}

		require.NoError(t, harness.mailers.SendPasswordReset(t.Context(), &passwordreset.Mail{User: user, Issuance: issuance}))

		enqueued := harness.enqueued(t, ddbidentity.PasswordResetTokenCreatedEventType)
		require.Len(t, enqueued, 1)
		assert.Equal(t, user.ID, enqueued[0].UserID)
		assert.Equal(t, issuance.Secret, enqueued[0].Context[authkeys.PasswordResetTokenSecretKey])
		assert.Equal(t, issuance.Token.ID, enqueued[0].Context[authkeys.PasswordResetTokenIDKey])
	})

	T.Run("refuses a mail that names nobody", func(t *testing.T) {
		t.Parallel()

		harness := buildMailersHarness(t)

		require.Error(t, harness.mailers.SendPasswordReset(t.Context(), &passwordreset.Mail{}))
		assert.Empty(t, harness.executor.ExecContextCalls())
	})
}

func TestSignInMailers_SendVerification(T *testing.T) {
	T.Parallel()

	T.Run("queues the verification mail carrying the link's token", func(t *testing.T) {
		t.Parallel()

		harness := buildMailersHarness(t)

		user := &identity.User{ID: identifiers.New()}
		token := identifiers.New()

		require.NoError(t, harness.mailers.SendVerification(t.Context(), &signin.VerificationMail{User: user, Token: token}))

		enqueued := harness.enqueued(t, ddbidentity.UserEmailAddressVerificationEmailRequestedEventType)
		require.Len(t, enqueued, 1)
		assert.Equal(t, user.ID, enqueued[0].UserID)
		assert.Equal(t, token, enqueued[0].Context[identitykeys.UserEmailVerificationTokenKey])
	})

	T.Run("refuses a mail that names nobody", func(t *testing.T) {
		t.Parallel()

		harness := buildMailersHarness(t)

		require.Error(t, harness.mailers.SendVerification(t.Context(), &signin.VerificationMail{}))
		assert.Empty(t, harness.executor.ExecContextCalls())
	})
}

func TestSignInMailers_SendHandleReminder(T *testing.T) {
	T.Parallel()

	T.Run("queues the reminder", func(t *testing.T) {
		t.Parallel()

		harness := buildMailersHarness(t)
		user := &identity.User{ID: identifiers.New()}

		require.NoError(t, harness.mailers.SendHandleReminder(t.Context(), &signin.HandleReminderMail{User: user}))

		enqueued := harness.enqueued(t, ddbidentity.UsernameReminderRequestedEventType)
		require.Len(t, enqueued, 1)
		assert.Equal(t, user.ID, enqueued[0].UserID)
	})

	T.Run("refuses a mail that names nobody", func(t *testing.T) {
		t.Parallel()

		harness := buildMailersHarness(t)

		require.Error(t, harness.mailers.SendHandleReminder(t.Context(), &signin.HandleReminderMail{}))
		assert.Empty(t, harness.executor.ExecContextCalls())
	})
}
