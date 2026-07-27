package internalops

import (
	"database/sql"
	"testing"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests (validation, no DB required) ---

func TestCreateQueueTestMessage(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.CreateQueueTestMessage(ctx, "", "test-queue")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with empty queue name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.CreateQueueTestMessage(ctx, identifiers.New(), "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestAcknowledgeQueueTestMessage(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.AcknowledgeQueueTestMessage(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestGetQueueTestMessage(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetQueueTestMessage(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestPruneQueueTestMessages(T *testing.T) {
	T.Parallel()

	T.Run("with empty queue name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.PruneQueueTestMessages(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests (require DB container) ---

func TestQuerier_Integration_QueueTestMessages(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	queueName := "test-queue-" + identifiers.New()[:8]
	msgID := identifiers.New()

	// Create
	err := dbc.CreateQueueTestMessage(ctx, msgID, queueName)
	require.NoError(t, err)

	// Get
	msg, err := dbc.GetQueueTestMessage(ctx, msgID)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, msgID, msg.ID)
	assert.Equal(t, queueName, msg.QueueName)
	assert.Nil(t, msg.AcknowledgedAt)

	// Acknowledge
	err = dbc.AcknowledgeQueueTestMessage(ctx, msgID)
	require.NoError(t, err)

	// Get again - should have acknowledged_at
	msg, err = dbc.GetQueueTestMessage(ctx, msgID)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.NotNil(t, msg.AcknowledgedAt)
}

func TestQuerier_Integration_QueueTestMessages_GetNotFound(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	msg, err := dbc.GetQueueTestMessage(ctx, "nonexistent-id-"+identifiers.New())
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestQuerier_Integration_QueueTestMessages_Prune(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	queueName := "prune-test-" + identifiers.New()[:8]

	// Create several messages
	for range 5 {
		require.NoError(t, dbc.CreateQueueTestMessage(ctx, identifiers.New(), queueName))
	}

	// Prune (keeps last 100, so all 5 should remain)
	require.NoError(t, dbc.PruneQueueTestMessages(ctx, queueName))

	// Prune with non-existent queue - should not error
	require.NoError(t, dbc.PruneQueueTestMessages(ctx, "nonexistent-queue-"+identifiers.New()))
}
