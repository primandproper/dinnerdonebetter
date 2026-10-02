package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	dataprivacyhttp "github.com/primandproper/platform-go/v14/dataprivacy/http"
	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// privacyCaller is somebody reaching the privacy-request surface over HTTP, as a bearer token.
type privacyCaller string

// privacyCallerForTest signs a fresh user in and returns them with the token their privacy
// requests are made with.
func privacyCallerForTest(t *testing.T) (*identity.User, privacyCaller) {
	t.Helper()

	user := createServiceUserForTest(t, buildUserRegistrationInputForTest(t))

	return user, privacyCaller(fetchLoginTokenForUserForTest(t, user))
}

// do makes one request against the API server's router as this caller.
func (c privacyCaller) do(t *testing.T, ctx context.Context, method, path string, body []byte) (status int, response []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, httpTestServerAddress+path, reader)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+string(c))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { assert.NoError(t, res.Body.Close()) }()

	response, err = io.ReadAll(res.Body)
	require.NoError(t, err)

	return res.StatusCode, response
}

// privacyReceipt is how the surface answers about one request.
type privacyReceipt struct {
	Data struct {
		Request *platformdataprivacy.Request `json:"request"`
	} `json:"data"`
}

// readPrivacyRequest reads one request as the caller, returning the status and, on a 200, the
// request.
func (c privacyCaller) readPrivacyRequest(t *testing.T, ctx context.Context, requestID string) (int, *platformdataprivacy.Request) {
	t.Helper()

	status, body := c.do(t, ctx, http.MethodGet, dataprivacyhttp.BasePath+"/"+requestID, nil)
	if status != http.StatusOK {
		return status, nil
	}

	receipt := &privacyReceipt{}
	require.NoError(t, json.Unmarshal(body, receipt))

	return status, receipt.Data.Request
}

// submitPrivacyRequest asks for a request of kind as the caller and returns it as recorded.
func (c privacyCaller) submitPrivacyRequest(t *testing.T, ctx context.Context, kind platformdataprivacy.RequestType) *platformdataprivacy.Request {
	t.Helper()

	status, body := c.do(t, ctx, http.MethodPost, dataprivacyhttp.BasePath, []byte(`{"type":"`+string(kind)+`"}`))
	require.True(t, status >= 200 && status < 300, "submitting a %s request answered %d: %s", kind, status, body)

	receipt := &privacyReceipt{}
	require.NoError(t, json.Unmarshal(body, receipt))
	require.NotNil(t, receipt.Data.Request)
	require.NotEmpty(t, receipt.Data.Request.ID)

	return receipt.Data.Request
}

// artifactPath is where this server serves a completed export's artifact.
func artifactPath(requestID string) string {
	return dataprivacyhttp.BasePath + "/" + requestID + dataprivacyhttp.ArtifactSuffix
}

// awaitTerminalPrivacyRequest polls a request until it stops moving, and returns it.
//
// It polls the request rather than the operation behind it because that is what a subject can
// see: the row is the answer to "is my export ready", and an operation that finished without
// moving the row would be a request nobody could ever collect.
func awaitTerminalPrivacyRequest(t *testing.T, ctx context.Context, c privacyCaller, requestID string) *platformdataprivacy.Request {
	t.Helper()

	var request *platformdataprivacy.Request

	require.Eventually(t, func() bool {
		status, read := c.readPrivacyRequest(t, ctx, requestID)
		if status != http.StatusOK {
			return false
		}

		request = read

		return request.Status.Terminal()
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
		// A nil scope reads across every one, which is what this assertion wants: the
		// fulfillment worker is checking its own bookkeeping rather than serving a
		// tenant. A request's scope is the subject's, and this test does not know it.
		stored, err := dataPrivacyFulfillment.Store.Get(ctx, databaseClient.Reader(), nil, requestID)
		if err != nil {
			return false
		}

		request = stored

		return request.Status.Terminal()
	}, fulfillmentBudget, 250*time.Millisecond, "expected the fulfillment worker to reach a terminal state")

	return request
}

// submitExport asks for the caller's data and returns the request's ID.
func submitExport(t *testing.T, ctx context.Context, c privacyCaller) string {
	t.Helper()

	request := c.submitPrivacyRequest(t, ctx, platformdataprivacy.RequestExport)
	assert.Equal(t, platformdataprivacy.RequestExport, request.Type)

	// Recorded, not fulfilled. The whole reason the worker below exists is that this call
	// returns before any of the work has happened.
	assert.False(t, request.Status.Terminal(), "a submission must return before the export has been produced")

	return request.ID
}

// fetchExportDocument reads a completed export back through the API and decodes it.
//
// Through the API rather than out of the bucket, because the bucket holds ciphertext: artifacts
// are encrypted at rest, and the reader's compressor and cipher agreeing with the writer's is
// exactly the thing that has no other way of being checked.
func fetchExportDocument(t *testing.T, ctx context.Context, c privacyCaller, requestID string) (document *platformdataprivacy.Document, artifact []byte) {
	t.Helper()

	status, artifact := c.do(t, ctx, http.MethodGet, artifactPath(requestID), nil)
	require.Equal(t, http.StatusOK, status, "fetching the artifact answered %d: %s", status, artifact)
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
		QueryRowContext(ctx, "SELECT COUNT(*) FROM ddb_identity_users WHERE id = $1", userID).Scan(&count))

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

		subject, subjectCaller := privacyCallerForTest(t)

		// A second user, whose data must not appear. Created before the export is submitted,
		// so a collector reading a whole table rather than a subject's rows would pick them up.
		stranger, _ := createUserAndClientForTest(t)
		createUserNotificationForTest(t, stranger.ID)

		// Something of the subject's own beyond the user row, so the export has to reach past
		// the identity collector to be right.
		//
		// A notification rather than a webhook, which is what this used to use: webhooks are
		// no longer collected, because nothing in that domain names a person. An endpoint is
		// an account's delivery configuration. See docs/data-privacy.md.
		subjectNotification := createUserNotificationForTest(t, subject.ID)

		// And a handset. No export named one until the notifications section moved onto
		// platform's adapter: the collector it replaced read the inbox and never the registry.
		subjectDevice := createUserDeviceForTest(t, subject.ID)

		// And a credential, which is the half of an export somebody asks for when they think
		// they have been compromised: which devices can sign in as them. Inserted rather than
		// enrolled, because what is under test is whether the collector is registered at all,
		// not the ceremony that writes the row.
		subjectPasskey := insertWebAuthnCredentialForTest(t, subject.ID, "The Subject's Laptop")

		requestID := submitExport(t, ctx, subjectCaller)

		request := awaitTerminalPrivacyRequest(t, ctx, subjectCaller, requestID)
		require.Equal(t, platformdataprivacy.StatusCompleted, request.Status,
			"the export did not complete: %v", request.Failures)
		assert.Empty(t, request.Failures, "every registered collector must succeed against the real schema")
		assert.False(t, request.ExpiresAt.IsZero(), "a completed export's artifact has to have an expiry")

		document, artifact := fetchExportDocument(t, ctx, subjectCaller, requestID)

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

		// The five this subject certainly has data in: they registered, they were sent a
		// notification, they registered a handset, the registration was audited, and they
		// hold a passkey.
		//
		// The passkey one is here because it was missing. The collector existed in platform
		// over a store this deployment already ran, and nothing registered it, so every export
		// ever produced said nothing about anybody's credentials and said it in a manifest
		// that reported success.
		for _, key := range []string{
			ddbdataprivacy.CollectorKeyIdentity,
			ddbdataprivacy.CollectorKeyNotificationsInbox,
			ddbdataprivacy.CollectorKeyNotificationsDevices,
			ddbdataprivacy.CollectorKeyAuditLog,
			ddbdataprivacy.CollectorKeyPasskeys,
		} {
			assert.Contains(t, document.Data, key, "the export is missing the %q section", key)
			assert.Contains(t, document.Manifest.Sections, key)
		}

		// The subject's own data is in it, from several different collectors.
		assert.Contains(t, string(document.Data[ddbdataprivacy.CollectorKeyIdentity]), subject.Username)
		assert.Contains(t, string(document.Data[ddbdataprivacy.CollectorKeyNotificationsInbox]), subjectNotification.ID)
		assert.Contains(t, string(document.Data[ddbdataprivacy.CollectorKeyNotificationsDevices]), subjectDevice.ID)
		assert.Contains(t, string(document.Data[ddbdataprivacy.CollectorKeyPasskeys]), subjectPasskey,
			"the passkey section does not name the subject's credential")

		// And nobody else's, anywhere in the artifact. Searched over the whole document rather
		// than per section, because a leak would not announce which collector caused it.
		assert.NotContains(t, string(artifact), stranger.Username,
			"another user's data reached this subject's export")
		assert.NotContains(t, string(artifact), stranger.EmailAddress,
			"another user's data reached this subject's export")
	})

	// The request read is platform's surface, and the conformance suite asserts it is absent to
	// a neighbor — but only between two tenants, and a privacy request here names a person
	// rather than an account, so the suite's pair shares one and skips. The artifact is this
	// application's own route. Both are asserted here, between two people.
	T.Run("refuses a request and its artifact to anybody but its subject", func(t *testing.T) {
		t.Parallel()
		serializeDataPrivacy(t)

		ctx := t.Context()

		_, subjectCaller := privacyCallerForTest(t)
		_, strangerCaller := privacyCallerForTest(t)

		requestID := submitExport(t, ctx, subjectCaller)

		request := awaitTerminalPrivacyRequest(t, ctx, subjectCaller, requestID)
		require.Equal(t, platformdataprivacy.StatusCompleted, request.Status)

		// Absent rather than forbidden, in both the missing and the not-yours case: a distinct
		// denial would confirm that a given request ID exists, and whether somebody has asked
		// for their data is itself a fact about them.
		status, _ := strangerCaller.do(t, ctx, http.MethodGet, artifactPath(requestID), nil)
		assert.Equal(t, http.StatusNotFound, status)

		status, _ = strangerCaller.readPrivacyRequest(t, ctx, requestID)
		assert.Equal(t, http.StatusNotFound, status)
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
		subjectCaller := privacyCaller(fetchLoginTokenForUserForTest(t, subject))

		// A bystander, to show the erasure is scoped to its subject rather than to the table.
		bystander, _ := createUserAndClientForTest(t)

		require.Equal(t, 1, userRowCount(t, ctx, subject.ID))

		request := subjectCaller.submitPrivacyRequest(t, ctx, platformdataprivacy.RequestErasure)
		assert.Equal(t, platformdataprivacy.RequestErasure, request.Type)
		// Queued, not deleted. The confirmation window is zero, so nothing further is needed
		// from the subject — but the deletion still happens in the worker's transaction rather
		// than on the request path, where a timeout halfway through would leave a subject in a
		// state no status could describe.
		assert.False(t, request.Status.Terminal())

		// Read off the row rather than through the API, and that is not a shortcut: an
		// erasure ends by deleting its subject, and the API scopes every read of a privacy
		// request to the subject it belongs to. The one principal entitled to ask how this
		// request ended no longer exists by the time it has. An operator reads the row.
		erasure := awaitTerminalStoredRequest(t, ctx, request.ID)
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

		_, subjectCaller := privacyCallerForTest(t)

		requestID := submitExport(t, ctx, subjectCaller)

		request := awaitTerminalPrivacyRequest(t, ctx, subjectCaller, requestID)
		require.Equal(t, platformdataprivacy.StatusCompleted, request.Status)
		require.False(t, request.ExpiresAt.IsZero())

		// Readable before the sweep, or the assertion after it proves nothing.
		fetchExportDocument(t, ctx, subjectCaller, requestID)

		// One second past the expiry this artifact was actually stamped with, so the sweep is
		// asked the same question it is asked in production rather than a broader one.
		sweeper, err := dataPrivacyFulfillment.SweeperAt(ctx, request.ExpiresAt.Add(time.Second))
		require.NoError(t, err)

		result, err := sweeper.Sweep(ctx)
		require.NoError(t, err)
		assert.Positive(t, result.ArtifactsExpired, "the sweep deleted no artifacts")

		// The row survives — a subject is entitled to know what was asked in their name — and
		// says the artifact is gone.
		status, after := subjectCaller.readPrivacyRequest(t, ctx, requestID)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, platformdataprivacy.StatusExpired, after.Status)

		// And the artifact is unavailable rather than an internal error about our storage: a
		// conflict, because the request exists and the caller may see it — it is simply not in
		// a state that has an artifact. That is platform's mapping of ErrArtifactUnavailable.
		status, body := subjectCaller.do(t, ctx, http.MethodGet, artifactPath(requestID), nil)
		assert.Equal(t, http.StatusConflict, status, "fetching an expired artifact answered: %s", body)
	})
}

// insertWebAuthnCredentialForTest inserts a test WebAuthn credential for the given user.
// Returns the credential's internal ID (used by ListPasskeys/ArchivePasskey).
func insertWebAuthnCredentialForTest(t *testing.T, userID, friendlyName string) string {
	t.Helper()

	credID := identifiers.New()
	credentialIDBytes := fmt.Appendf(nil, "test-cred-%s-%d", credID, time.Now().UnixNano())
	publicKeyBytes := []byte("test-public-key-data")

	// platform's table, under this application's prefix: it names its own
	// webauthn_credentials too, and a scope is a column here where the schema this
	// replaced had none.
	//
	// Owner rather than String: the global scope's identifier is the empty string, and
	// "<global>" is only how it reads in a log line. A row written with the prose spelling
	// is one every scoped read passes over.
	_, err := databaseClient.Writer().ExecContext(
		t.Context(),
		`INSERT INTO ddb_webauthn_credentials (id, scope, belongs_to_user, credential_id, public_key, sign_count, transports, friendly_name)
		 VALUES ($1, $2, $3, $4, $5, 0, '[]', $6)`,
		credID, tenancy.Global().Owner(), userID, credentialIDBytes, publicKeyBytes, friendlyName,
	)
	require.NoError(t, err)

	return credID
}
