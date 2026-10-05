package settings

import (
	"context"
	"errors"
	"testing"

	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	settings "github.com/primandproper/platform-go/v15/settings"
	"github.com/primandproper/primitives-go/v2/database"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuerier_Integration_RecordAndEmitFailureSurfaces pins two things about the audit log entry
// that accompanies a catalog write.
//
// The first is that it is not best-effort: RecordAndEmit returns the recording error rather than
// swallowing it, so a write whose entry the database refused fails loudly instead of leaving a
// row nothing recorded.
//
// The second is that the catalog row goes with it. This is the assertion that changed under
// platform-go v14, and it is the point of the whole executor convention: platform's store used
// to own its transaction and commit the definition before recording was even attempted, so the
// row outlived the entry that could not be written about it. v14 hands the store the caller's
// database.Tx, so the write, the entry and the event are one transaction — the definition is
// gone when the entry fails, and there is no window in which a setting changed and nothing
// recorded that it had.
//
// That closes the gap for the whole family at once, not just here: comments (platform-go #457),
// waitlists (#458), settings (#460) and payments (#466) were four filings of one shape, and all
// four were a consequence of an adopted platform store rather than of RecordAndEmit. #1419
// tracked what would delete here when they landed; nothing did, because the fix was a signature
// change upstream rather than a workaround here. See docs/audit.md.
func TestQuerier_Integration_RecordAndEmitFailureSurfaces(t *testing.T) {
	ctx := t.Context()
	_, _, db := buildDatabaseClientForTest(t)

	expected := errors.New("the log said no")

	// The store is built again over the same database with a recorder whose audit repository
	// refuses every entry. The recorder is what the hooks hold, so it is the thing to swap: it
	// holds its own reference to the audit repository, so a repository reassigned after
	// construction would leave this test asserting nothing. The emitter is nil because the
	// harness builds none, which is what ProvideSettingsRepository was handed there too.
	tracerProvider := tracingnoop.NewTracerProvider()
	repo, err := newStore(ctx, loggingnoop.NewLogger(), tracerProvider, metricsnoop.NewMetricsProvider(), db,
		recording.NewRecorder(tracing.NewNamedTracer(tracerProvider, o11yName), &auditmock.RepositoryMock{
			RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error {
				return expected
			},
		}, nil))
	require.NoError(t, err)

	definition := fakes.BuildFakeSettingDefinition()

	created, err := writeT(ctx, db, func(tx database.Tx) (*settings.Definition, error) {
		return repo.CreateDefinition(ctx, tx, tenancy.Global(), definition)
	})
	require.Error(t, err)
	require.ErrorIs(t, err, expected)
	assert.Nil(t, created)

	// The catalog row went with the entry. Read on the database rather than on the rolled-back
	// transaction, so what is being asserted is what committed.
	survived, err := repo.GetDefinition(ctx, db.Reader(), tenancy.Global(), definition.ID)
	require.Error(t, err)
	assert.Nil(t, survived)
}
