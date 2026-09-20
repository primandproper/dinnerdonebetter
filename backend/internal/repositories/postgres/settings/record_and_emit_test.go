package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	settings "github.com/primandproper/platform-go/v14/settings"
	"github.com/primandproper/primitives-go/v2/database"

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
	dbc, auditRepo, db := buildDatabaseClientForTest(t)

	expected := errors.New("the log said no")

	repo, ok := dbc.(*repository)
	require.True(t, ok)

	// The recorder is swapped rather than a field it closed over: it holds its own reference to
	// the audit repository, so reassigning the repository after construction would leave this
	// test asserting nothing. The emitter is nil because the harness builds none, which is what
	// ProvideSettingsRepository was handed above.
	repo.recorder = recording.NewRecorder(repo.tracer, &auditmock.RepositoryMock{
		RecordFunc: func(context.Context, database.Tx, ...*audit.AuditLogEntry) error {
			return expected
		},
	}, nil)

	definition := fakes.BuildFakeSettingDefinition()

	created, err := writeT(ctx, db, func(tx database.Tx) (*settings.Definition, error) {
		return repo.CreateDefinition(ctx, tx, ddbsettings.Scope(), definition)
	})
	require.Error(t, err)
	require.ErrorIs(t, err, expected)
	assert.Nil(t, created)

	// The catalog row went with the entry. Read with a working recorder and on the database
	// rather than on the rolled-back transaction, so what is being asserted is what committed.
	repo.recorder = recording.NewRecorder(repo.tracer, auditRepo, nil)

	survived, err := repo.GetDefinition(ctx, db.Reader(), ddbsettings.Scope(), definition.ID)
	require.Error(t, err)
	assert.Nil(t, survived)
}
