package comments

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformcomments "github.com/primandproper/platform-go/v14/comments"
	"github.com/primandproper/platform-go/v14/comments/commentspb"
	commentsgrpc "github.com/primandproper/platform-go/v14/comments/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCommentsService registers platform's comment surface with the injector.
//
// There is no service of this application's own any more. The one that used to
// live in internal/services/comments forwarded eight RPCs to the store, checked
// that an edit was the author's own, and converted between two spellings of the
// same four fields — and platform-go v14 ships all three.
//
// What this application still owns is the store it is given: the repository in
// internal/repositories/postgres/comments, which is platform's SQL store with an
// audit entry and a data change event wrapped around every write. That survives
// adoption because the server takes the [platformcomments.Store] interface and
// opens the transaction before calling into it, so those two statements are
// statements of the server's transaction. See
// TestServer_Integration_RecordingRollsBackWithTheWrite, which fails against the
// pre-v14 shape.
//
// The AuthorAuthorizer is left at its default. platform's OwnCommentsOnly is
// exactly what the deleted service's ownedComment did: a caller may edit and
// archive their own comments and nobody else's. This application has no
// moderator role, so naming an authorizer here would only restate the default.
//
// The GrantsExtractor is wired, unlike when this was first adopted. Its absence
// is fail-closed — every read is confined, which is the behaviour the deleted
// service had, since it exposed no archived read at all — but leaving it absent
// makes comments.archive a name rather than a policy, and the three sibling
// surfaces adopted since all carry one. An archivist paging removed comments is
// a small addition over the deleted service and the consistent answer.
func RegisterCommentsService(i do.Injector) {
	do.Provide[commentspb.CommentsServiceServer](i, func(i do.Injector) (commentspb.CommentsServiceServer, error) {
		return commentsgrpc.NewServer(
			do.MustInvoke[platformcomments.Store](i),
			do.MustInvoke[database.Client](i),
			sessions.PrincipalFromContext,
			commentsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			commentsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			commentsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			commentsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
