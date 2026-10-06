// Package eventstest builds the recording spine for a unit test: platform's Emitter and
// Recorder over whatever outbox writer and audit recorder the test supplies, with no webhook
// fan-out behind them.
//
// It exists so a test of a hook can read the event off the statement that would have enqueued
// it, and the entry off the recorder it was handed, without building the four platform types by
// hand in every test file. A test that needs the fan-out too uses the database-backed helper in
// internal/repositories/postgres/testing instead.
package eventstest

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/platform-go/v15/webhooks"
	webhooksmock "github.com/primandproper/platform-go/v15/webhooks/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/require"
)

// New builds an Emitter whose events reach writer and whose entries reach recorder.
//
// The dispatcher behind it knows this application's catalog and dispatches nothing: the gate is
// exercised, the fan-out is not, which is the half a unit test over a mock executor can assert.
// Filing follows production — by the subject where an entry names one.
func New(t *testing.T, writer *outbox.Writer, recorder platformaudit.Recorder) *events.Emitter {
	t.Helper()

	dispatcher := &webhooksmock.DispatcherMock{
		CatalogFunc:  catalog.Catalog,
		DispatchFunc: func(context.Context, database.Tx, tenancy.Scope, *webhooks.Delivery) error { return nil },
	}

	emitter, err := webhooks.NewEmitter(writer, dispatcher, queuescfg.DefaultDataChangesTopicName)
	require.NoError(t, err)

	platformRecorder, err := platformrecording.New(recorder, emitter, sessions.PrincipalFromContext,
		platformrecording.WithScopeResolver(bySubject))
	require.NoError(t, err)

	e, err := events.NewEmitter(emitter, platformRecorder, writer)
	require.NoError(t, err)

	return e
}

// bySubject is recordingcfg.FileBySubject's rule, restated because the config package's
// resolver is not exported: the subject's chain where the entry names one, the write's otherwise.
func bySubject(_ context.Context, scope tenancy.Scope, entry *platformrecording.Entry) tenancy.Scope {
	if entry != nil && entry.SubjectID != "" {
		return tenancy.Of(entry.SubjectID)
	}

	return scope
}
