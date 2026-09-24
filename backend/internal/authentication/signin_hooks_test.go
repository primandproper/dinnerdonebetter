package authentication

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signInHooksHarness is the hooks over a real outbox writer and a transaction whose every
// statement is captured, so a test reads the event off the statement that would have enqueued
// it rather than off a mock of the thing that writes it.
type signInHooksHarness struct {
	hooks    signin.Hooks
	tx       database.Tx
	executor *mockdatabase.SQLQueryExecutorMock
}

func buildSignInHooksHarness(t *testing.T, execErr error) *signInHooksHarness {
	t.Helper()

	writer, err := outbox.NewWriter(dialect.Postgres)
	require.NoError(t, err)

	executor := &mockdatabase.SQLQueryExecutorMock{
		ExecContextFunc: func(context.Context, string, ...any) (sql.Result, error) {
			return nil, execErr
		},
	}

	return &signInHooksHarness{
		hooks:    NewSignInHooks(loggingnoop.NewLogger(), events.NewEmitter(writer, t.Name(), nil, nil)),
		tx:       database.NewTxForTesting(executor),
		executor: executor,
	}
}

// enqueued is every data change message the transaction was asked to write.
func (h *signInHooksHarness) enqueued(t *testing.T) []*audit.DataChangeMessage {
	t.Helper()

	var out []*audit.DataChangeMessage
	calls := h.executor.ExecContextCalls()
	for i := range calls {
		for _, arg := range calls[i].Args {
			raw, ok := arg.([]byte)
			if !ok {
				continue
			}

			msg := &audit.DataChangeMessage{}
			if json.Unmarshal(raw, msg) == nil && msg.EventType != "" {
				out = append(out, msg)
			}
		}
	}

	return out
}

func TestSignInHooks_AfterAuthenticate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		harness := buildSignInHooksHarness(t, nil)

		user := identityfakes.BuildFakeUser()
		account := identityfakes.BuildFakeAccountForUser(user.ID)

		err := harness.hooks.AfterAuthenticate(ctx, harness.tx, tenancy.Global(), &signin.Authentication{
			Principal: &identity.Principal{User: user, ActiveAccountID: account.ID},
		})
		require.NoError(t, err)

		// On the transaction signin handed the hook, and so committed or rolled back with the
		// sign-in it describes. The context carries no session — nobody has one yet — so the
		// user is named on the event rather than read.
		enqueued := harness.enqueued(t)
		require.Len(t, enqueued, 1)
		assert.Equal(t, ddbidentity.UserLoggedInServiceEventType, enqueued[0].EventType)
		assert.Equal(t, account.ID, enqueued[0].AccountID)
		assert.Equal(t, user.ID, enqueued[0].UserID)
	})

	T.Run("for an administrative sign-in", func(t *testing.T) {
		t.Parallel()

		// The administrative door is a sign-in like any other, and was recorded as one
		// when ProcessLogin published the event itself.
		ctx := t.Context()
		harness := buildSignInHooksHarness(t, nil)

		user := identityfakes.BuildFakeUser()

		err := harness.hooks.AfterAuthenticate(ctx, harness.tx, tenancy.Global(), &signin.Authentication{
			Principal:      &identity.Principal{User: user},
			Administrative: true,
		})
		require.NoError(t, err)

		enqueued := harness.enqueued(t)
		require.Len(t, enqueued, 1)
		assert.Equal(t, user.ID, enqueued[0].UserID)
	})

	T.Run("refuses the sign-in when the event cannot be enqueued", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		execErr := errors.New(identityfakes.BuildFakeUser().ID)
		harness := buildSignInHooksHarness(t, execErr)

		err := harness.hooks.AfterAuthenticate(ctx, harness.tx, tenancy.Global(), &signin.Authentication{
			Principal: &identity.Principal{User: identityfakes.BuildFakeUser()},
		})

		require.ErrorIs(t, err, execErr)
	})

	T.Run("with no emitter", func(t *testing.T) {
		t.Parallel()

		// A process with no data changes topic holds a nil emitter, and signs people in
		// all the same.
		ctx := t.Context()

		err := NewSignInHooks(loggingnoop.NewLogger(), nil).AfterAuthenticate(ctx, nil, tenancy.Global(), &signin.Authentication{
			Principal: &identity.Principal{User: identityfakes.BuildFakeUser()},
		})

		require.NoError(t, err)
	})

	T.Run("with no principal", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		harness := buildSignInHooksHarness(t, nil)

		err := harness.hooks.AfterAuthenticate(ctx, harness.tx, tenancy.Global(), &signin.Authentication{})

		require.Error(t, err)
		assert.Empty(t, harness.executor.ExecContextCalls())
	})
}
