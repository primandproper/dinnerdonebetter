package testing

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/outbox"
	recordingcfg "github.com/primandproper/platform-go/v15/recording/config"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	"github.com/primandproper/primitives-go/v2/database"

	"github.com/stretchr/testify/require"
)

// NewEmitterForTest builds the recording spine over a real database, the way
// events.RegisterOutboxEmitter does for a process: the outbox writer with the search index rules
// on it, platform's webhooks Emitter with its own fan-out over the webhooks tables, and platform's
// Recorder over the audit recorder handed in.
//
// It is real rather than mocked because the repository tests that use it are the ones that prove
// a write's entry and event land, and a mock would prove the mock.
func NewEmitterForTest(t *testing.T, ctx context.Context, db database.Client, auditRecorder platformaudit.Recorder) *events.Emitter {
	t.Helper()

	effect, err := indexevents.NewSideEffect()
	require.NoError(t, err)

	writer, err := outbox.NewWriter(db.Dialect(), outbox.WithWriterSideEffect(indexevents.SideEffectName, effect))
	require.NoError(t, err)

	emitter, err := webhookscfg.NewEmitter(ctx, &webhookscfg.Config{EmitterTopic: queuescfg.DefaultDataChangesTopicName}, db, writer, catalog.Catalog())
	require.NoError(t, err)

	recorder, err := recordingcfg.NewRecorder(ctx, &recordingcfg.Config{FileBy: recordingcfg.FileBySubject}, auditRecorder, emitter, sessions.PrincipalFromContext)
	require.NoError(t, err)

	e, err := events.NewEmitter(emitter, recorder, writer, effect)
	require.NoError(t, err)

	return e
}
