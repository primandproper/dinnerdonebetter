package payments

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v13/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepository_Integration_RecordAndEmitFailureSurfaces pins that the audit log entry is not
// best-effort: RecordAndEmit returns the recording error rather than swallowing it, so a write
// whose entry the database refused fails loudly instead of leaving a row nothing recorded.
//
// It does not assert that the product rolled back, and that is not an oversight. Recording runs
// in a transaction of its own here, opened after platform's store has already committed the
// catalog row in its own — the store owns its transaction and does not lend it out. So the row
// survives an entry that could not be written about it, where every repository that writes its
// own rows records as two further statements in the transaction that performed the write.
//
// payments is the fifth package with this shape, after comments, issuereports, settings and
// waitlists — exactly the packages whose writes are an adopted platform store. The gap arrived
// with the adoption (#1380), not with RecordAndEmit, and closing it needs platform to accept a
// caller's transaction (platform-go #466). See docs/audit.md, and #1419 for what deletes here
// when it lands.
func TestRepository_Integration_RecordAndEmitFailureSurfaces(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, _ := buildDatabaseClientForTest(t)

	expected := errors.New("the log said no")

	repo, ok := dbc.(*repository)
	require.True(t, ok)

	// The recorder is swapped rather than a field it closed over: it holds its own reference to
	// the audit repository, so reassigning the repository after construction would leave this
	// test asserting nothing. The emitter is nil because the harness builds none, which is what
	// ProvidePaymentsRepository was handed above.
	repo.recorder = recording.NewRecorder(repo.tracer, &auditmock.RepositoryMock{
		RecordFunc: func(context.Context, database.Tx, ...*audit.AuditLogEntry) error {
			return expected
		},
	}, nil)

	product := fakes.BuildFakeProduct()

	created, err := repo.CreateProduct(ctx, ddbpayments.Scope(), product)
	require.Error(t, err)
	require.ErrorIs(t, err, expected)
	assert.Nil(t, created)

	// The catalog row is still there, which is the gap the doc comment describes. Asserted
	// rather than described so that it fails here, loudly, on the day platform lets the store
	// take a caller's transaction and this stops being true. When it does fail, the fix is to
	// delete the workaround in payments.go, not to update this assertion — see #1419.
	repo.recorder = recording.NewRecorder(repo.tracer, auditRepo, nil)

	survived, err := repo.GetProduct(ctx, ddbpayments.Scope(), product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.ID, survived.ID)
}
