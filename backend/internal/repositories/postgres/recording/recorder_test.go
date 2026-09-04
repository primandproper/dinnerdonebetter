package recording

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/fakes"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/errors"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRecorderForTest(t *testing.T, auditRepo audit.Repository) *Recorder {
	t.Helper()

	return NewRecorder(tracing.NewNamedTracer(tracingnoop.NewTracerProvider(), t.Name()), auditRepo, nil)
}

// TestRecorder_RecordAndEmit covers the half of the pair that has to happen whether or not the
// other one can. The emitter is nil in every case here, which is what a process built without an
// outbox holds: events.NewEmitter returns nil rather than an error, and Emit on nil is inert.
func TestRecorder_RecordAndEmit(T *testing.T) {
	T.Parallel()

	// This is why the recorder is its own type rather than a method on the Emitter. Hanging the
	// audit write off a receiver that is deliberately nil-inert would make a process with no
	// outbox a process with no audit log, silently and in the same line of code.
	T.Run("records even when there is no emitter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		entry := fakes.BuildFakeAuditLogEntry()

		var recorded []*audit.AuditLogEntry
		auditRepo := &auditmock.RepositoryMock{
			RecordFunc: func(_ context.Context, _ database.Tx, entries ...*audit.AuditLogEntry) error {
				recorded = append(recorded, entries...)
				return nil
			},
		}

		err := newRecorderForTest(t, auditRepo).RecordAndEmit(ctx, nil, loggingnoop.NewLogger(), entry, "event.type", "", nil)

		require.NoError(t, err)
		require.Len(t, recorded, 1)
		assert.Equal(t, entry, recorded[0])
	})

	T.Run("returns the recording error rather than swallowing it", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := errors.New("the log said no")

		auditRepo := &auditmock.RepositoryMock{
			RecordFunc: func(context.Context, database.Tx, ...*audit.AuditLogEntry) error {
				return expected
			},
		}

		err := newRecorderForTest(t, auditRepo).RecordAndEmit(ctx, nil, loggingnoop.NewLogger(), fakes.BuildFakeAuditLogEntry(), "event.type", "", nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
	})
}
