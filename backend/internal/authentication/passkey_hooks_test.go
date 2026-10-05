package authentication

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/authentication/passkeys"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// passkeyHooksHarness is the hooks over a real outbox writer and a transaction whose every
// statement is captured, as signInHooksHarness is.
type passkeyHooksHarness struct {
	audited  *auditmock.RepositoryMock
	hooks    passkeys.Hooks
	tx       database.Tx
	executor *mockdatabase.SQLQueryExecutorMock
}

func buildPasskeyHooksHarness(t *testing.T, execErr error) *passkeyHooksHarness {
	t.Helper()

	writer, err := outbox.NewWriter(dialect.Postgres)
	require.NoError(t, err)

	executor := &mockdatabase.SQLQueryExecutorMock{
		ExecContextFunc: func(context.Context, string, ...any) (sql.Result, error) {
			return nil, execErr
		},
	}

	emitter := events.NewEmitter(writer, t.Name(), nil, nil)
	audited := &auditmock.RepositoryMock{
		RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error { return nil },
	}

	return &passkeyHooksHarness{
		audited:  audited,
		hooks:    NewPasskeyHooks(loggingnoop.NewLogger(), recording.NewRecorder(tracing.NewTracerForTest(t.Name()), audited, emitter)),
		tx:       database.NewTxForTesting(executor),
		executor: executor,
	}
}

func buildFakePasskey() *passkeys.Credential {
	return &passkeys.Credential{
		ID:            identifiers.New(),
		BelongsToUser: identifiers.New(),
		FriendlyName:  identifiers.New(),
		CredentialID:  []byte(identifiers.New()),
		PublicKey:     []byte(identifiers.New()),
	}
}

func TestPasskeyHooks_CredentialWrites(T *testing.T) {
	T.Parallel()

	writes := map[string]struct {
		call           func(ctx context.Context, hooks passkeys.Hooks, tx database.Tx, credential *passkeys.Credential) error
		eventType      string
		auditEventType platformaudit.EventType
	}{
		"a passkey registered": {
			call: func(ctx context.Context, hooks passkeys.Hooks, tx database.Tx, credential *passkeys.Credential) error {
				return hooks.AfterRegisterPasskey(ctx, tx, tenancy.Global(), credential)
			},
			eventType:      ddbidentity.PasskeyRegisteredServiceEventType,
			auditEventType: platformaudit.EventCreated,
		},
		"a passkey archived": {
			call: func(ctx context.Context, hooks passkeys.Hooks, tx database.Tx, credential *passkeys.Credential) error {
				return hooks.AfterArchivePasskey(ctx, tx, tenancy.Global(), credential)
			},
			eventType:      ddbidentity.PasskeyArchivedServiceEventType,
			auditEventType: platformaudit.EventArchived,
		},
	}

	for name, write := range writes {
		T.Run(name, func(t *testing.T) {
			t.Parallel()

			harness := buildPasskeyHooksHarness(t, nil)
			credential := buildFakePasskey()

			require.NoError(t, write.call(t.Context(), harness.hooks, harness.tx, credential))

			enqueued := enqueuedOn(t, harness.executor)
			require.Len(t, enqueued, 1)
			assert.Equal(t, write.eventType, enqueued[0].EventType)
			assert.Equal(t, credential.BelongsToUser, enqueued[0].UserID)
			assert.Equal(t, credential.ID, enqueued[0].Context[identitykeys.PasskeyIDKey])

			recorded := harness.audited.RecordCalls()
			require.Len(t, recorded, 1)
			require.Len(t, recorded[0].Entries, 1)

			entry := recorded[0].Entries[0]
			assert.Equal(t, write.auditEventType, entry.EventType)
			assert.Equal(t, passkeysResourceType, entry.ResourceType)
			assert.Equal(t, credential.ID, entry.ResourceID)
			assert.Equal(t, credential.BelongsToUser, entry.Actor.ID)

			// Named by its ID alone: nothing the owner typed, and no key material.
			assert.Empty(t, entry.Changes)
		})

		T.Run(name+" is refused when the event cannot be enqueued", func(t *testing.T) {
			t.Parallel()

			execErr := errors.New(identifiers.New())
			harness := buildPasskeyHooksHarness(t, execErr)

			require.ErrorIs(t, write.call(t.Context(), harness.hooks, harness.tx, buildFakePasskey()), execErr)
		})

		T.Run(name+" is refused for a passkey with no owner", func(t *testing.T) {
			t.Parallel()

			harness := buildPasskeyHooksHarness(t, nil)
			credential := buildFakePasskey()
			credential.BelongsToUser = ""

			require.Error(t, write.call(t.Context(), harness.hooks, harness.tx, credential))
			assert.Empty(t, harness.audited.RecordCalls())
		})
	}
}

func TestPasskeyHooks_AfterFailedPasskeyLogin(T *testing.T) {
	T.Parallel()

	// As a refused password records nothing: see AfterFailedPasskeyLogin.
	T.Run("records nothing", func(t *testing.T) {
		t.Parallel()

		harness := buildPasskeyHooksHarness(t, nil)

		require.NoError(t, harness.hooks.AfterFailedPasskeyLogin(t.Context(), harness.tx, tenancy.Global(),
			&passkeys.FailedLogin{Cause: passkeys.ErrLoginFailed, UserID: identifiers.New()}))
		assert.Empty(t, enqueuedOn(t, harness.executor))
		assert.Empty(t, harness.audited.RecordCalls())
	})
}
