package integration

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fulfillmentBudget is how long a test waits for the operations worker to pick a request up and
// finish it.
//
// It is generous because the wait is mostly the queue's poll: nothing wakes the worker when a
// request is enqueued in this deployment, so a submission sits until the next poll comes round.
// That is the production configuration, and running the suite against a shorter one would be
// asserting on timings nothing deploys.
const fulfillmentBudget = 90 * time.Second

// dataPrivacySweepMu orders the data privacy tests against each other.
//
// The sweep is database-wide and destructive: it deletes the artifact of every completed export
// whose expiry has passed, not this test's. A sweeper running on a clock a week ahead —
// which is the only way a test can reach an artifact's window at all — would therefore delete an
// artifact a sibling test is about to fetch, and the sibling would fail for a reason that has
// nothing to do with it.
//
// Only the tests in this file take the lock, and the rest of the suite is unaffected: nothing
// else produces an export.
var dataPrivacySweepMu sync.Mutex

// serializeDataPrivacy holds that order for the whole of a test, from before it submits a
// request until after its last assertion about the artifact.
func serializeDataPrivacy(t *testing.T) {
	t.Helper()

	dataPrivacySweepMu.Lock()
	t.Cleanup(dataPrivacySweepMu.Unlock)
}

// awaitTerminalPrivacyRequest polls a request until it stops moving, and returns it.
//
// It polls the request rather than the operation behind it because that is what a subject can
// see: the row is the answer to "is my export ready", and an operation that finished without
// moving the row would be a request nobody could ever collect.
func awaitTerminalPrivacyRequest(t *testing.T, ctx context.Context, c client.Client, requestID string) *dataprivacygrpc.DataPrivacyRequest {
	t.Helper()

	var request *dataprivacygrpc.DataPrivacyRequest

	require.Eventually(t, func() bool {
		response, err := c.GetDataPrivacyRequest(ctx, &dataprivacygrpc.GetDataPrivacyRequestRequest{
			DataPrivacyRequestId: requestID,
		})
		if err != nil {
			return false
		}

		request = response.GetRequest()

		return platformdataprivacy.Status(request.GetStatus()).Terminal()
	}, fulfillmentBudget, 250*time.Millisecond, "expected the fulfillment worker to reach a terminal state")

	return request
}

// awaitTerminalStoredRequest polls the request row directly until it stops moving.
//
// The API is the right way to read a request and the wrong way to read this one — see the
// erasure test. Everything else about the fulfillment is the same, including the worker that
// wrote the row.
func awaitTerminalStoredRequest(t *testing.T, ctx context.Context, requestID string) *platformdataprivacy.Request {
	t.Helper()

	var request *platformdataprivacy.Request

	require.Eventually(t, func() bool {
		stored, err := dataPrivacyFulfillment.Store.Get(ctx, requestID)
		if err != nil {
			return false
		}

		request = stored

		return request.Status.Terminal()
	}, fulfillmentBudget, 250*time.Millisecond, "expected the fulfillment worker to reach a terminal state")

	return request
}

// submitExport asks for the caller's data and returns the request's ID.
func submitExport(t *testing.T, ctx context.Context, c client.Client) string {
	t.Helper()

	response, err := c.AggregateUserDataReport(ctx, &dataprivacygrpc.AggregateUserDataReportRequest{})
	require.NoError(t, err)

	request := response.GetRequest()
	require.NotEmpty(t, request.GetId())
	assert.Equal(t, string(platformdataprivacy.RequestExport), request.GetRequestType())

	// Recorded, not fulfilled. The whole reason the worker below exists is that this call
	// returns before any of the work has happened.
	assert.False(t, platformdataprivacy.Status(request.GetStatus()).Terminal(),
		"a submission must return before the export has been produced")

	return request.GetId()
}

// fetchExportDocument reads a completed export back through the API and decodes it.
//
// Through the API rather than out of the bucket, because the bucket holds ciphertext: artifacts
// are encrypted at rest, and the reader's compressor and cipher agreeing with the writer's is
// exactly the thing that has no other way of being checked.
func fetchExportDocument(t *testing.T, ctx context.Context, c client.Client, requestID string) (document *platformdataprivacy.Document, artifact []byte) {
	t.Helper()

	response, err := c.FetchUserDataReport(ctx, &dataprivacygrpc.FetchUserDataReportRequest{
		DataPrivacyRequestId: requestID,
	})
	require.NoError(t, err)

	artifact = response.GetArtifact()
	require.NotEmpty(t, artifact)

	document = new(platformdataprivacy.Document)
	require.NoError(t, json.Unmarshal(artifact, document))

	return document, artifact
}

// userRowCount counts a user's row, for the erasure's assertions.
func userRowCount(t *testing.T, ctx context.Context, userID string) int {
	t.Helper()

	var count int
	require.NoError(t, databaseClient.Reader().
		QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count))

	return count
}

// TestDataPrivacy_Export is the assertion an export has never had.
//
// A subject access request is a legal obligation with a shape: a subject asks, and within a
// statutory window they are handed everything the application knows about them, and nothing
// about anybody else. Every part of that spans two processes — the API server records the
// request, the operations worker gathers over eleven collectors and writes an encrypted
// artifact, and the API server reads it back — so no unit test can reach it, and until now
// nothing did.
//
// The strongest assertion here is the empty failure map. Each collector is registered against a
// key and runs against the real schema; a collector whose query no longer matches its tables
// fails on its own and the export is still delivered, with the gap named in the manifest. That
// is the right behavior and it is also why the failure would be silent — an export missing three
// sections still completes.
func TestDataPrivacy_Export(T *testing.T) {
	T.Parallel()

	T.Run("produces an artifact holding the subject's data and nobody else's", func(t *testing.T) {
		t.Parallel()
		serializeDataPrivacy(t)

		ctx := t.Context()

		subject, subjectClient := createUserAndClientForTest(t)

		// A second user, whose data must not appear. Created before the export is submitted,
		// so a collector reading a whole table rather than a subject's rows would pick them up.
		stranger, strangerClient := createUserAndClientForTest(t)
		createWebhookForTest(t, strangerClient)

		// Something of the subject's own beyond the user row, so the export has to reach past
		// the identity collector to be right.
		subjectWebhook := createWebhookForTest(t, subjectClient)

		requestID := submitExport(t, ctx, subjectClient)

		request := awaitTerminalPrivacyRequest(t, ctx, subjectClient, requestID)
		require.Equal(t, string(platformdataprivacy.StatusCompleted), request.GetStatus(),
			"the export did not complete: %v", request.GetFailures())
		assert.Empty(t, request.GetFailures(), "every registered collector must succeed against the real schema")
		assert.NotNil(t, request.GetExpiresAt(), "a completed export's artifact has to have an expiry")

		document, artifact := fetchExportDocument(t, ctx, subjectClient, requestID)

		assert.True(t, document.Complete(), "a document with failures is a partial export")
		assert.Equal(t, subject.ID, document.Manifest.Subject.ID)
		assert.Equal(t, requestID, document.Manifest.RequestID)

		// Every section in the document is one the registry registered. The converse does not
		// hold and must not be asserted: a collector that finds nothing returns no fragment and
		// its section is omitted, which is the difference between an artifact that reads as an
		// answer and one that reads as a form. Which domains are registered at all is asserted
		// against the registry itself, in TestWorkerWiring_Scheduler.
		registered := dataPrivacyFulfillment.Registry.CollectorKeys()
		for _, section := range document.Manifest.Sections {
			assert.Contains(t, registered, section, "the export carries a section nothing registered")
			assert.Contains(t, document.Data, section, "a manifest section with no data behind it")
		}

		// The three this subject certainly has data in: they registered, they own a webhook,
		// and both of those were audited.
		for _, key := range []string{
			ddbdataprivacy.CollectorKeyIdentity,
			ddbdataprivacy.CollectorKeyWebhooks,
			ddbdataprivacy.CollectorKeyAuditLog,
		} {
			assert.Contains(t, document.Data, key, "the export is missing the %q section", key)
			assert.Contains(t, document.Manifest.Sections, key)
		}

		// The subject's own data is in it, from two different collectors.
		assert.Contains(t, string(document.Data[ddbdataprivacy.CollectorKeyIdentity]), subject.Username)
		assert.Contains(t, string(document.Data[ddbdataprivacy.CollectorKeyWebhooks]), subjectWebhook.ID)

		// And nobody else's, anywhere in the artifact. Searched over the whole document rather
		// than per section, because a leak would not announce which collector caused it.
		assert.NotContains(t, string(artifact), stranger.Username,
			"another user's data reached this subject's export")
		assert.NotContains(t, string(artifact), stranger.EmailAddress,
			"another user's data reached this subject's export")
	})

	T.Run("refuses to hand an artifact to anybody but its subject", func(t *testing.T) {
		t.Parallel()
		serializeDataPrivacy(t)

		ctx := t.Context()

		_, subjectClient := createUserAndClientForTest(t)
		_, strangerClient := createUserAndClientForTest(t)

		requestID := submitExport(t, ctx, subjectClient)

		request := awaitTerminalPrivacyRequest(t, ctx, subjectClient, requestID)
		require.Equal(t, string(platformdataprivacy.StatusCompleted), request.GetStatus())

		// NotFound rather than PermissionDenied, in both the missing and the not-yours case: a
		// distinct denial would confirm that a given request ID exists, and whether somebody
		// has asked for their data is itself a fact about them.
		_, err := strangerClient.FetchUserDataReport(ctx, &dataprivacygrpc.FetchUserDataReportRequest{
			DataPrivacyRequestId: requestID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		_, err = strangerClient.GetDataPrivacyRequest(ctx, &dataprivacygrpc.GetDataPrivacyRequestRequest{
			DataPrivacyRequestId: requestID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

// TestDataPrivacy_Erasure is the other half of the obligation, and the one that cannot be undone.
//
// The request path records and returns; the erasure itself happens inside one transaction in the
// worker, over a registry of erasers. What this asserts is that it actually erased — that the
// user row is gone, and with it everything the schema cascades from one — rather than that a row
// somewhere says it did.
func TestDataPrivacy_Erasure(T *testing.T) {
	T.Parallel()

	T.Run("destroys the subject's data", func(t *testing.T) {
		t.Parallel()
		serializeDataPrivacy(t)

		ctx := t.Context()

		subject, subjectClient := createUserAndClientForTest(t)
		createWebhookForTest(t, subjectClient)

		// A bystander, to show the erasure is scoped to its subject rather than to the table.
		bystander, _ := createUserAndClientForTest(t)

		require.Equal(t, 1, userRowCount(t, ctx, subject.ID))

		response, err := subjectClient.DestroyAllUserData(ctx, &dataprivacygrpc.DestroyAllUserDataRequest{})
		require.NoError(t, err)

		request := response.GetRequest()
		require.NotEmpty(t, request.GetId())
		assert.Equal(t, string(platformdataprivacy.RequestErasure), request.GetRequestType())
		// Queued, not deleted. The confirmation window is zero, so nothing further is needed
		// from the subject — but the deletion still happens in the worker's transaction rather
		// than on the request path, where a timeout halfway through would leave a subject in a
		// state no status could describe.
		assert.False(t, platformdataprivacy.Status(request.GetStatus()).Terminal())

		// Read off the row rather than through the API, and that is not a shortcut: an
		// erasure ends by deleting its subject, and the API scopes every read of a privacy
		// request to the subject it belongs to. The one principal entitled to ask how this
		// request ended no longer exists by the time it has. An operator reads the row.
		erasure := awaitTerminalStoredRequest(t, ctx, request.GetId())
		require.Equal(t, platformdataprivacy.StatusCompleted, erasure.Status,
			"the erasure did not complete: %v", erasure.Failures)
		assert.Empty(t, erasure.Failures, "every registered eraser must succeed against the real schema")

		assert.Zero(t, userRowCount(t, ctx, subject.ID), "the subject's user row survived their erasure")
		assert.Equal(t, 1, userRowCount(t, ctx, bystander.ID), "an erasure took somebody else's data with it")

		// An erasure produces no artifact, so there is nothing for the sweep to expire and
		// nothing for anybody to fetch.
		assert.Empty(t, erasure.ArtifactRef)
		assert.True(t, erasure.ExpiresAt.IsZero(), "an erasure has nothing that expires")
	})
}

// TestDataPrivacy_Sweeper covers the chore whose absence is invisible.
//
// An export artifact is everything the application knows about one person, in a single object.
// Without the sweep, every one ever written stays in the bucket forever and nothing about the
// request rows suggests otherwise — the fulfillment worker would look entirely healthy, and the
// only symptom would be a bucket nobody had reason to look in.
//
// The sweep runs on a clock a week ahead rather than against a shortened TTL, because the TTL is
// stamped onto the request row by the worker that completed it: lowering it in configuration
// would expire every artifact this suite produces, not this test's.
func TestDataPrivacy_Sweeper(T *testing.T) {
	T.Parallel()

	T.Run("expires an artifact past its window", func(t *testing.T) {
		t.Parallel()
		serializeDataPrivacy(t)

		ctx := t.Context()

		_, subjectClient := createUserAndClientForTest(t)

		requestID := submitExport(t, ctx, subjectClient)

		request := awaitTerminalPrivacyRequest(t, ctx, subjectClient, requestID)
		require.Equal(t, string(platformdataprivacy.StatusCompleted), request.GetStatus())
		require.NotNil(t, request.GetExpiresAt())

		// Readable before the sweep, or the assertion after it proves nothing.
		fetchExportDocument(t, ctx, subjectClient, requestID)

		// One second past the expiry this artifact was actually stamped with, so the sweep is
		// asked the same question it is asked in production rather than a broader one.
		sweeper, err := dataPrivacyFulfillment.SweeperAt(ctx, request.GetExpiresAt().AsTime().Add(time.Second))
		require.NoError(t, err)

		result, err := sweeper.Sweep(ctx)
		require.NoError(t, err)
		assert.Positive(t, result.ArtifactsExpired, "the sweep deleted no artifacts")

		// The row survives — a subject is entitled to know what was asked in their name — and
		// says the artifact is gone.
		after, err := subjectClient.GetDataPrivacyRequest(ctx, &dataprivacygrpc.GetDataPrivacyRequestRequest{
			DataPrivacyRequestId: requestID,
		})
		require.NoError(t, err)
		assert.Equal(t, string(platformdataprivacy.StatusExpired), after.GetRequest().GetStatus())

		// And the artifact is unavailable rather than an internal error about our storage.
		//
		// FailedPrecondition, not NotFound. The service asks for NotFound and does not get it:
		// platform-go's error mapper recognizes ErrArtifactUnavailable and overrides the
		// default code with its own, on the grounds that the request exists and the caller may
		// see it — it is simply not in a state this call can serve. That is the behavior a
		// client has to handle, so it is the behavior pinned here.
		_, err = subjectClient.FetchUserDataReport(ctx, &dataprivacygrpc.FetchUserDataReportRequest{
			DataPrivacyRequestId: requestID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
		assert.Contains(t, strings.ToLower(status.Convert(err).Message()), "artifact",
			"the refusal should say the artifact is unavailable")
	})
}
