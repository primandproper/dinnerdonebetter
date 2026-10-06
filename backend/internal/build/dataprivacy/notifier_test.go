package dataprivacy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queuedEmails is every outbound email a mock database was asked to enqueue on topic.
type queuedEmails struct {
	db     *mockdatabase.ClientMock
	topic  string
	emails []*queuemessages.OutboundEmailMessage
}

func buildQueuedEmails(t *testing.T) *queuedEmails {
	t.Helper()

	q := &queuedEmails{topic: fake.BuildFakeID()}

	executor := &mockdatabase.SQLQueryExecutorMock{
		ExecContextFunc: func(_ context.Context, _ string, args ...any) (sql.Result, error) {
			onTopic := false
			var payload []byte

			for _, arg := range args {
				switch v := arg.(type) {
				case string:
					onTopic = onTopic || v == q.topic
				case []byte:
					payload = v
				}
			}

			require.True(t, onTopic, "an email was queued on a topic other than the outbound emails one")

			var msg queuemessages.OutboundEmailMessage
			require.NoError(t, json.Unmarshal(payload, &msg))
			q.emails = append(q.emails, &msg)

			return nil, nil
		},
	}

	q.db = &mockdatabase.ClientMock{
		DialectFunc: func() dialect.Dialect { return dialect.Postgres },
		ReaderFunc:  func() database.SQLQueryExecutor { return executor },
		WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error {
			return fn(database.NewTxForTesting(executor))
		},
	}

	return q
}

func TestCompletionNotifier(T *testing.T) {
	T.Parallel()

	T.Run("queues the completion mail to the subject's own address", func(t *testing.T) {
		t.Parallel()

		q := buildQueuedEmails(t)
		user := fakes.BuildFakeUser()

		directory := &identitymock.StoreMock{
			GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, scope tenancy.Scope, userID string) (*platformidentity.User, error) {
				assert.True(t, scope.IsGlobal())
				assert.Equal(t, user.ID, userID)

				return user, nil
			},
		}

		notifier, err := newCompletionNotifier(q.db, directory, q.topic)
		require.NoError(t, err)

		require.NoError(t, notifier.Notify(t.Context(), &platformdataprivacy.Notification{
			Request: &platformdataprivacy.Request{
				ID:      fake.BuildFakeID(),
				Type:    platformdataprivacy.RequestExport,
				Status:  platformdataprivacy.StatusCompleted,
				Subject: platformdataprivacy.Subject{ID: user.ID, Type: platformdataprivacy.SubjectUser},
			},
		}))

		require.Len(t, q.emails, 1)
		assert.Equal(t, user.EmailAddress, q.emails[0].ToAddress)
		assert.NotEmpty(t, q.emails[0].Subject)
		assert.NotEmpty(t, q.emails[0].HTMLContent)
	})

	T.Run("mails nobody about a subject the directory no longer has", func(t *testing.T) {
		t.Parallel()

		// An erasure that removed the row is the system working, not a failure to report.
		q := buildQueuedEmails(t)
		directory := &identitymock.StoreMock{
			GetUserFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*platformidentity.User, error) {
				return nil, platformidentity.ErrUserNotFound
			},
		}

		notifier, err := newCompletionNotifier(q.db, directory, q.topic)
		require.NoError(t, err)

		require.NoError(t, notifier.Notify(t.Context(), &platformdataprivacy.Notification{
			Request: &platformdataprivacy.Request{
				ID:      fake.BuildFakeID(),
				Type:    platformdataprivacy.RequestErasure,
				Status:  platformdataprivacy.StatusCompleted,
				Subject: platformdataprivacy.Subject{ID: fake.BuildFakeID(), Type: platformdataprivacy.SubjectUser},
			},
		}))

		assert.Empty(t, q.emails)
	})

	T.Run("reports a directory that cannot be read", func(t *testing.T) {
		t.Parallel()

		q := buildQueuedEmails(t)
		readErr := errors.New(fake.BuildFakeString())
		directory := &identitymock.StoreMock{
			GetUserFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*platformidentity.User, error) {
				return nil, readErr
			},
		}

		notifier, err := newCompletionNotifier(q.db, directory, q.topic)
		require.NoError(t, err)

		err = notifier.Notify(t.Context(), &platformdataprivacy.Notification{
			Request: &platformdataprivacy.Request{
				ID:      fake.BuildFakeID(),
				Subject: platformdataprivacy.Subject{ID: fake.BuildFakeID(), Type: platformdataprivacy.SubjectUser},
			},
		})
		require.ErrorIs(t, err, readErr)
		assert.Empty(t, q.emails)
	})

	T.Run("refuses to be built with no topic", func(t *testing.T) {
		t.Parallel()

		_, err := newCompletionNotifier(buildQueuedEmails(t).db, &identitymock.StoreMock{}, "")
		require.Error(t, err)
	})
}
