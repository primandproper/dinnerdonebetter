package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexstamp"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"
	searchsync "github.com/primandproper/platform-go/v11/search/sync"
	syncsource "github.com/primandproper/platform-go/v11/search/sync/source"
	textsearchnoop "github.com/primandproper/platform-go/v11/search/text/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lastIndexedAtForTest reads the column directly, because nothing maps it onto a domain type.
// That is the point of the column as it stands: it is written by the sync path and read by
// whoever is asking how current the index is, not by the application.
func lastIndexedAtForTest(t *testing.T, ctx context.Context, dbc *repository, validInstrumentID string) sql.NullTime {
	t.Helper()

	var stamp sql.NullTime
	require.NoError(t, dbc.readDB.QueryRowContext(
		ctx,
		"SELECT last_indexed_at FROM valid_instruments WHERE id = $1",
		validInstrumentID,
	).Scan(&stamp))

	return stamp
}

// buildSyncerForTest wires the real search sync path over a real repository, writing into the
// noop index. What is under test is not the index — it keeps nothing — but the Stamper between
// the Syncer and it, and the row the Syncer's document came from.
func buildSyncerForTest(t *testing.T, dbc *repository) (*searchsync.Syncer[mealplanningindexing.ValidInstrumentSearchSubset], *indexstamp.Stamper) {
	t.Helper()

	stamper, err := indexstamp.New(
		textsearchnoop.NewIndexManager[mealplanningindexing.ValidInstrumentSearchSubset](),
		dbc.MarkValidInstrumentAsIndexed,
	)
	require.NoError(t, err)

	source, err := mealplanningindexing.NewValidInstrumentSource(dbc)
	require.NoError(t, err)

	syncer, err := syncsource.NewSyncer(source, stamper,
		syncsource.WithLogger(loggingnoop.NewLogger()),
		syncsource.WithTracerProvider(tracingnoop.NewTracerProvider()),
	)
	require.NoError(t, err)

	return syncer, stamper
}

func TestQuerier_Integration_SyncedDocumentIsStamped(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	validInstrument := createValidInstrumentForTest(t, ctx, fakes.BuildFakeValidInstrument(), dbc)

	// A row nothing has indexed carries no stamp, which is what every row looked like before
	// anything wrote this column.
	assert.False(t, lastIndexedAtForTest(t, ctx, dbc, validInstrument.ID).Valid)

	syncer, stamper := buildSyncerForTest(t, dbc)

	require.NoError(t, syncer.Apply(ctx, searchsync.NewEvent(searchsync.OpUpsert, validInstrument.ID)))

	// Applying the event does not write the stamp — the buffer holds it — which is what keeps
	// one UPDATE per applied document off the sync path.
	assert.False(t, lastIndexedAtForTest(t, ctx, dbc, validInstrument.ID).Valid)

	require.NoError(t, stamper.Close(ctx))

	stamp := lastIndexedAtForTest(t, ctx, dbc, validInstrument.ID)
	assert.True(t, stamp.Valid)
	assert.False(t, stamp.Time.IsZero())
}

func TestQuerier_Integration_DeletedDocumentIsNotStamped(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	validInstrument := createValidInstrumentForTest(t, ctx, fakes.BuildFakeValidInstrument(), dbc)

	syncer, stamper := buildSyncerForTest(t, dbc)

	require.NoError(t, syncer.Apply(ctx, searchsync.NewEvent(searchsync.OpDelete, validInstrument.ID)))
	require.NoError(t, stamper.Close(ctx))

	// The column says when a document was last written to the index, and a delete wrote none.
	assert.False(t, lastIndexedAtForTest(t, ctx, dbc, validInstrument.ID).Valid)
}
