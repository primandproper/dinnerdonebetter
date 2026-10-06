package identitystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events/eventstest"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformauditmock "github.com/primandproper/platform-go/v15/audit/mock"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hooksHarness is the wrapper over the recording spine's unit-test shape: platform's Recorder
// writing entries into a slice, and an outbox writer whose statements run on a mock executor
// so the index events can be read back off what it would have inserted.
type hooksHarness struct {
	hooks    *Hooks
	tx       database.Tx
	executor *mockdatabase.SQLQueryExecutorMock
	recorded []*platformaudit.Entry
}

func buildHooksHarness(t *testing.T) *hooksHarness {
	t.Helper()

	h := &hooksHarness{
		executor: &mockdatabase.SQLQueryExecutorMock{
			ExecContextFunc: func(context.Context, string, ...any) (sql.Result, error) { return nil, nil },
		},
	}

	writer, err := outbox.NewWriter(dialect.Postgres, eventstest.IndexRules(t))
	require.NoError(t, err)

	emitter := eventstest.New(t, writer, &platformauditmock.RecorderMock{
		RecordFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, entries ...*platformaudit.Entry) error {
			h.recorded = append(h.recorded, entries...)
			return nil
		},
	})

	h.hooks, err = ProvideHooks(emitter)
	require.NoError(t, err)
	h.tx = database.NewTxForTesting(h.executor)

	return h
}

// indexEvents is every search index event the transaction was asked to write, read off the
// statements that would have inserted them. An index event is the one payload on this outbox
// that decodes to a searchsync.Event naming a document; a data change travels as an envelope
// and decodes to none.
func (h *hooksHarness) indexEvents(t *testing.T) (topics []string, events []searchsync.Event) {
	t.Helper()

	calls := h.executor.ExecContextCalls()
	for i := range calls {
		for _, arg := range calls[i].Args {
			raw, ok := arg.([]byte)
			if !ok {
				continue
			}

			var event searchsync.Event
			if json.Unmarshal(raw, &event) != nil || event.DocumentID == "" {
				continue
			}

			events = append(events, event)
			for _, other := range calls[i].Args {
				if topic, isString := other.(string); isString && topic == identityindexing.IndexTypeUsers {
					topics = append(topics, topic)
				}
			}
		}
	}

	return topics, events
}

func TestProvideHooks(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil emitter", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideHooks(nil)
		require.ErrorIs(t, err, ErrNilEmitter)
	})
}

func TestHooks_AfterRegister(T *testing.T) {
	T.Parallel()

	T.Run("records the registration and indexes the user", func(t *testing.T) {
		t.Parallel()

		h := buildHooksHarness(t)
		user := identityfakes.BuildFakeUser()
		account := identityfakes.BuildFakeAccountForUser(user.ID)
		membership := identityfakes.BuildFakeMembership()

		err := h.hooks.AfterRegister(t.Context(), h.tx, tenancy.Global(), &platformidentity.Registration{
			User: user, Account: account, Membership: membership,
		})
		require.NoError(t, err)

		// platform's three entries, as platform writes them: this wrapper adds none.
		require.Len(t, h.recorded, 3)
		assert.Equal(t, platformidentity.ResourceTypeUser, h.recorded[0].ResourceType)
		assert.Equal(t, user.ID, h.recorded[0].ResourceID)
		assert.Equal(t, platformaudit.EventCreated, h.recorded[0].EventType)

		// And the one index event the recording cannot derive for itself yet.
		topics, indexed := h.indexEvents(t)
		require.Len(t, indexed, 1)
		assert.Equal(t, []string{identityindexing.IndexTypeUsers}, topics)
		assert.Equal(t, user.ID, indexed[0].DocumentID)
		assert.Equal(t, searchsync.OpUpsert, indexed[0].Op)
	})

	T.Run("indexes nothing when the recording is refused", func(t *testing.T) {
		t.Parallel()

		h := buildHooksHarness(t)

		err := h.hooks.AfterRegister(t.Context(), h.tx, tenancy.Global(), nil)
		require.Error(t, err)

		_, indexed := h.indexEvents(t)
		assert.Empty(t, indexed)
	})
}

func TestHooks_AfterRegisterWithInvitation(T *testing.T) {
	T.Parallel()

	T.Run("indexes the user", func(t *testing.T) {
		t.Parallel()

		h := buildHooksHarness(t)
		user := identityfakes.BuildFakeUser()

		err := h.hooks.AfterRegisterWithInvitation(t.Context(), h.tx, tenancy.Global(), &platformidentity.InvitedRegistration{
			User: user, Invitation: identityfakes.BuildFakeInvitation(), Membership: identityfakes.BuildFakeMembership(),
		})
		require.NoError(t, err)

		_, indexed := h.indexEvents(t)
		require.Len(t, indexed, 1)
		assert.Equal(t, user.ID, indexed[0].DocumentID)
		assert.Equal(t, searchsync.OpUpsert, indexed[0].Op)
	})
}

func TestHooks_AfterUpdateProfile(T *testing.T) {
	T.Parallel()

	T.Run("reindexes the user", func(t *testing.T) {
		t.Parallel()

		h := buildHooksHarness(t)
		user := identityfakes.BuildFakeUser()

		require.NoError(t, h.hooks.AfterUpdateProfile(t.Context(), h.tx, tenancy.Global(), user, []string{"username"}))

		_, indexed := h.indexEvents(t)
		require.Len(t, indexed, 1)
		assert.Equal(t, user.ID, indexed[0].DocumentID)
		assert.Equal(t, searchsync.OpUpsert, indexed[0].Op)
	})
}

func TestHooks_AfterArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("removes the user's document", func(t *testing.T) {
		t.Parallel()

		h := buildHooksHarness(t)
		user := identityfakes.BuildFakeUser()

		require.NoError(t, h.hooks.AfterArchiveUser(t.Context(), h.tx, tenancy.Global(), user, nil))

		_, indexed := h.indexEvents(t)
		require.Len(t, indexed, 1)
		assert.Equal(t, user.ID, indexed[0].DocumentID)
		assert.Equal(t, searchsync.OpDelete, indexed[0].Op)
	})
}

func TestHooks_AfterUpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("records through platform and derives no index event", func(t *testing.T) {
		t.Parallel()

		// Not overridden: an account is not indexed, so platform's recording is the whole of
		// what happens, and this is the check that the wrapper adds nothing it should not.
		h := buildHooksHarness(t)
		before := identityfakes.BuildFakeAccount()
		after := identityfakes.BuildFakeAccount()
		after.ID = before.ID

		require.NoError(t, h.hooks.AfterUpdateAccount(t.Context(), h.tx, tenancy.Global(), before, after))

		require.Len(t, h.recorded, 1)
		assert.Equal(t, platformidentity.ResourceTypeAccount, h.recorded[0].ResourceType)

		_, indexed := h.indexEvents(t)
		assert.Empty(t, indexed)
	})
}
