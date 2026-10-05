package payments

import (
	"context"
	"errors"
	"testing"

	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/billing"
	"github.com/primandproper/primitives-go/v2/database"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepository_Integration_RecordAndEmitFailureSurfaces pins two things about the audit log
// entry that accompanies a catalog write.
//
// The first is that it is not best-effort: RecordAndEmit returns the recording error rather than
// swallowing it, so a write whose entry the database refused fails loudly instead of leaving a
// row nothing recorded.
//
// The second is that the product goes with it. This is the assertion that changed under
// platform-go v14: the billing store used to own its transaction and commit the product before
// recording was even attempted, so the row outlived the entry that could not be written about
// it. v14 hands the store the caller's database.Tx, so the write, the entry and the event are
// one transaction and share one fate.
//
// payments was the fifth filing of that shape, after comments (platform-go #457), waitlists
// (#458), settings (#460) and its own #466 — every one of them a consequence of an adopted
// platform store rather than of RecordAndEmit. #1419 tracked what would delete here when they
// landed; nothing did, because the fix was a signature change upstream rather than a workaround
// here. See docs/audit.md.
func TestRepository_Integration_RecordAndEmitFailureSurfaces(t *testing.T) {
	ctx := t.Context()
	_, _, db := buildDatabaseClientForTest(t)

	expected := errors.New("the log said no")

	// A store whose hooks record through a repository that refuses every entry. The recorder is
	// built over the mock rather than a field swapped afterwards: it holds its own reference to
	// the audit repository, so reassigning the repository after construction would leave this
	// test asserting nothing. The emitter is nil because the harness builds none.
	failing := &auditmock.RepositoryMock{
		RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error {
			return expected
		},
	}

	repo, err := newStore(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), db, &hooks{
		logger:            loggingnoop.NewLogger(),
		auditLogEntryRepo: failing,
		recorder:          recording.NewRecorder(tracing.NewNamedTracer(tracingnoop.NewTracerProvider(), o11yName), failing, nil),
	})
	require.NoError(t, err)

	product := fakes.BuildFakeProduct()

	created, err := writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return repo.CreateProduct(ctx, tx, tenancy.Global(), product)
	})
	require.Error(t, err)
	require.ErrorIs(t, err, expected)
	assert.Nil(t, created)

	// The catalog row went with the entry. Read on the database rather than on the rolled-back
	// transaction, so what is being asserted is what committed.
	survived, err := repo.GetProduct(ctx, db.Reader(), tenancy.Global(), product.ID)
	require.Error(t, err)
	assert.Nil(t, survived)
}
