package comments

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	commentsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformcomments "github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/platform-go/v15/comments/commentspb"
	commentsgrpc "github.com/primandproper/platform-go/v15/comments/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file pins the claim the v14 adoption rests on.
//
// platform-go v14 ships a gRPC surface for comments, and adopting it means
// deleting this application's own service, proto and converters — thirteen
// domains' worth, if the claim holds. What it cannot mean is losing the audit
// entry and the data change event every write here owes, because those are
// written by the repository in this package rather than by platform's server.
//
// The claim is that nothing is lost, because platform's server takes the
// comments.Store *interface* and opens the transaction before calling into it:
//
//	s.client.WithTransaction(ctx, func(tx database.Tx) error {
//	    created, createErr := s.store.CreateComment(ctx, tx, req.scope, comment)
//	    ...
//	})
//
// So this package's hooks, which the store calls on the transaction it was
// handed, run inside the server's transaction, and the two statements they add
// belong to that transaction. Before v14 they could not: the store owned its own
// transaction and had already committed by the time anything here ran, which is
// what every `record` helper in these repositories used to say in a comment.
//
// Reading the signatures is not proof. These two tests are.

// failingAuditRepository is the actual audit repository with Record replaced.
//
// Embedded rather than mocked, so every other method is the actual one and the
// only difference from production is the failure being induced.
type failingAuditRecorder struct {
	platformaudit.Recorder

	err error
}

func (f *failingAuditRecorder) Record(context.Context, database.Tx, tenancy.Scope, ...*platformaudit.Entry) error {
	return f.err
}

// commentsFixture is one wired stack: platform's server over this package's
// repository over a actual database, with a actual outbox behind it.
type commentsFixture struct {
	server commentspb.CommentsServiceServer
	store  platformcomments.Store
	db     database.Client
}

// buildFixture wires the stack.
//
// decorate is how a test induces a failure in one of the three statements the
// transaction carries: it is handed the actual audit recorder and returns
// whatever the recording spine should actually be given. Everything else is
// production wiring — the same store constructor, the same emitter, the same
// server.
func buildFixture(t *testing.T, decorate func(platformaudit.Recorder) platformaudit.Recorder) *commentsFixture {
	t.Helper()

	ctx := t.Context()

	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	db, err := postgres.NewDatabaseClient(ctx, config,
		postgres.WithLogger(loggingnoop.NewLogger()),
		postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NoError(t, err)
	require.NotNil(t, db)

	audits, err := auditlogentries.ProvideAuditLog(
		loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), db)
	require.NoError(t, err)

	auditRecorder := audits.Recorder()
	if decorate != nil {
		auditRecorder = decorate(auditRecorder)
	}

	catalog, err := commentsbuild.Catalog()
	require.NoError(t, err)

	// The actual recording spine against the actual tables, because an outbox row that rolls
	// back is the half of the claim a fake emitter could not demonstrate.
	store, err := ProvideCommentsRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		db,
		pgtesting.NewRecorderForTest(t, ctx, db, auditRecorder),
		catalog,
	)
	require.NoError(t, err)

	server, err := commentsgrpc.NewServer(store, db, sessions.PrincipalFromContext,
		commentsgrpc.WithLogger(loggingnoop.NewLogger()),
		commentsgrpc.WithTracerProvider(tracingnoop.NewTracerProvider()),
		commentsgrpc.WithMetricsProvider(metricsnoop.NewMetricsProvider()))
	require.NoError(t, err)

	return &commentsFixture{server: server, store: store, db: db}
}

// callerContext puts a session on the context, which is what
// sessions.PrincipalFromContext reads and platform's server asks it for.
func callerContext(t *testing.T, userID, accountID string) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: accountID,
		Requester:       sessions.RequesterInfo{UserID: userID},
	})
}

// outboxDepth reads the outbox directly, because the point is what is in the
// table rather than what the emitter believes it wrote.
func (f *commentsFixture) outboxDepth(t *testing.T, ctx context.Context) int {
	t.Helper()

	var count int
	require.NoError(t, f.db.Reader().QueryRowContext(ctx, "SELECT COUNT(*) FROM outbox_messages").Scan(&count))

	return count
}

// rootsOn is the comment count on one target, read through the store.
func (f *commentsFixture) rootsOn(t *testing.T, ctx context.Context, targetID string) int {
	t.Helper()

	target := platformcomments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: targetID}

	page, err := f.store.ListRootComments(ctx, f.db.Reader(), tenancy.Global(), target, nil)
	require.NoError(t, err)

	return len(page.Data)
}

func createRequest(targetID, body string) *commentspb.CreateCommentRequest {
	return &commentspb.CreateCommentRequest{
		Comment: &commentspb.CommentInput{
			Body: body,
			Target: &commentspb.CommentTarget{
				Type: string(mealplanning.CommentTargetTypeRecipes),
				Id:   targetID,
			},
		},
	}
}

// TestServer_Integration_RecordingCommitsWithTheWrite is the happy half: one call
// through platform's surface leaves the comment, its audit entry and its outbox
// message behind, all three written by a transaction this package never opened.
func TestServer_Integration_RecordingCommitsWithTheWrite(T *testing.T) {
	T.Parallel()

	T.Run("the comment, the entry and the event all land", func(t *testing.T) {
		t.Parallel()

		fixture := buildFixture(t, nil)

		user := pgtesting.CreateUserForTest(t, nil, fixture.db.Writer())
		ctx := callerContext(t, user.ID, identifiers.New())
		targetID := identifiers.New()

		require.Zero(t, fixture.outboxDepth(t, ctx))

		response, err := fixture.server.CreateComment(ctx, createRequest(targetID, "the transaction is the point"))
		require.NoError(t, err)
		require.NotNil(t, response.GetResult())

		commentID := response.GetResult().GetId()
		require.NotEmpty(t, commentID)

		// The row.
		assert.Equal(t, 1, fixture.rootsOn(t, ctx, targetID))

		// The entry, filed under the author — written by this package's hooks,
		// inside platform's transaction.
		pgtesting.AssertAuditLogContainsForUser(t, ctx, fixture.db, user.ID, []pgtesting.ExpectedAuditEntry{
			{EventType: platformaudit.EventCreated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: commentID},
		})

		// The event.
		assert.Equal(t, 1, fixture.outboxDepth(t, ctx),
			"the data change event should have been enqueued by the server's transaction")
	})
}

// TestServer_Integration_RecordingRollsBackWithTheWrite is the load-bearing half.
//
// The audit write fails. Everything else about the call is production: platform's
// server opened the transaction, this package's repository wrote the comment into
// it, and then the entry that write owes could not be written. The comment must
// not be there afterwards, and neither must the event.
//
// If this fails, thirteen service layers stay where they are — a surface that can
// commit a row whose audit entry failed is not one this application can mount.
func TestServer_Integration_RecordingRollsBackWithTheWrite(T *testing.T) {
	T.Parallel()

	T.Run("a failed audit entry takes the comment and the event with it", func(t *testing.T) {
		t.Parallel()

		errAuditUnavailable := platformerrors.New("audit log is unavailable")

		fixture := buildFixture(t, func(actual platformaudit.Recorder) platformaudit.Recorder {
			return &failingAuditRecorder{Recorder: actual, err: errAuditUnavailable}
		})

		user := pgtesting.CreateUserForTest(t, nil, fixture.db.Writer())
		ctx := callerContext(t, user.ID, identifiers.New())
		targetID := identifiers.New()

		require.Zero(t, fixture.outboxDepth(t, ctx))

		response, err := fixture.server.CreateComment(ctx, createRequest(targetID, "this should not survive"))
		require.Error(t, err)
		assert.Nil(t, response)

		// Nothing about the comment reached the table.
		assert.Zero(t, fixture.rootsOn(t, ctx, targetID),
			"the comment should have rolled back with the audit entry")

		// And the event with it. This is the statement that would be hardest to
		// notice going wrong in production: a published event for a row that does
		// not exist is a webhook subscriber told about a comment nobody can read.
		assert.Zero(t, fixture.outboxDepth(t, ctx),
			"the data change event should have rolled back with the audit entry")
	})
}

// TestServer_Integration_AFailedWriteRecordsNothing is the same guarantee from the
// other side: when the store's own write is what fails, the entry and the event
// that would have described it are not left behind either.
func TestServer_Integration_AFailedWriteRecordsNothing(T *testing.T) {
	T.Parallel()

	T.Run("an unknown target type records nothing", func(t *testing.T) {
		t.Parallel()

		fixture := buildFixture(t, nil)

		user := pgtesting.CreateUserForTest(t, nil, fixture.db.Writer())
		ctx := callerContext(t, user.ID, identifiers.New())

		request := createRequest(identifiers.New(), "a comment on nothing")
		request.Comment.Target.Type = "recipies"

		response, err := fixture.server.CreateComment(ctx, request)
		require.Error(t, err)
		assert.Nil(t, response)

		assert.Empty(t, pgtesting.AuditEntriesForActor(t, ctx, fixture.db, user.ID),
			"a refused write should have recorded no audit entry")

		assert.Zero(t, fixture.outboxDepth(t, ctx), "a refused write should have emitted no event")
	})
}
