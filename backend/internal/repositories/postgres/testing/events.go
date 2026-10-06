package testing

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/recordingspine"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"

	"github.com/stretchr/testify/require"
)

// NewSpineForTest builds the recording spine over a real database, the way a process does: the
// outbox writer with the search index rules on it, platform's webhooks Emitter with its own
// fan-out over the webhooks tables, and platform's Recorder over the audit recorder handed in.
//
// It is real rather than mocked because the repository tests that use it are the ones that prove
// a write's entry and event land, and a mock would prove the mock.
func NewSpineForTest(t *testing.T, ctx context.Context, db database.Client, auditRecorder platformaudit.Recorder) *recordingspine.Spine {
	t.Helper()

	spine, err := recordingspine.New(ctx, db, auditRecorder)
	require.NoError(t, err)

	return spine
}

// NewRecorderForTest is the platform Recorder of the same spine, for a store whose
// RecordingHooks are built over it.
func NewRecorderForTest(t *testing.T, ctx context.Context, db database.Client, auditRecorder platformaudit.Recorder) *platformrecording.Recorder {
	t.Helper()

	return NewSpineForTest(t, ctx, db, auditRecorder).Recorder
}

// AsRequester puts a session for userID on the context, which is what the recording spine
// reads the actor off: an entry a store's hook records under this context names userID as the
// one who did it.
func AsRequester(ctx context.Context, userID string) context.Context {
	return sessions.AttachToContext(ctx, &sessions.ContextData{
		Requester: sessions.RequesterInfo{UserID: userID},
	})
}
