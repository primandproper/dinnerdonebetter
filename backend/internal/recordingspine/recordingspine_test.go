package recordingspine

import (
	"context"
	"database/sql"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformauditmock "github.com/primandproper/platform-go/v15/audit/mock"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(T *testing.T) {
	T.Parallel()

	T.Run("carries the search index rules", func(t *testing.T) {
		t.Parallel()

		// Every write to this outbox owes the index an event, so the rules are registered on
		// the writer rather than passed per call: a recipe update derives a recipes-index
		// event without its caller asking for one.
		var topics []string

		executor := &mockdatabase.SQLQueryExecutorMock{
			ExecContextFunc: func(_ context.Context, _ string, args ...any) (sql.Result, error) {
				for _, arg := range args {
					if topic, ok := arg.(string); ok && topic == mealplanningindexing.IndexTypeRecipes {
						topics = append(topics, topic)
					}
				}

				return nil, nil
			},
		}

		writer, err := NewWriter(&mockdatabase.ClientMock{DialectFunc: func() dialect.Dialect { return dialect.Postgres }}, nil)
		require.NoError(t, err)

		require.NoError(t, writer.EnqueueDerived(t.Context(), database.NewTxForTesting(executor), outbox.Message{
			Payload: &datachanges.Message{
				EventType: types.RecipeUpdatedServiceEventType,
				Context:   map[string]any{mealplanningkeys.RecipeIDKey: fake.BuildFakeID()},
			},
		}))

		assert.Equal(t, []string{mealplanningindexing.IndexTypeRecipes}, topics)
	})
}

func TestNew(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil database client", func(t *testing.T) {
		t.Parallel()

		_, err := New(t.Context(), nil, &platformauditmock.RecorderMock{})
		require.Error(t, err)
	})
}
