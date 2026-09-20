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
// Two seams are deliberately left at their defaults:
//
//   - The AuthorAuthorizer. Platform's default, OwnCommentsOnly, is exactly what
//     the deleted service's ownedComment did: a caller may edit and archive their
//     own comments and nobody else's. This application has no moderator role, so
//     naming an authorizer here would only restate the default.
//   - The GrantsExtractor. Its absence clears include_archived on every read,
//     which serves live rows to everybody rather than removed ones to anybody.
//     The deleted service exposed no way to read archived comments either, so
//     this is the behaviour it had. Wiring one is what a moderation queue would
//     need, and it wants the same extractor the authorization interceptor reads.
func RegisterCommentsService(i do.Injector) {
	do.Provide[commentspb.CommentsServiceServer](i, func(i do.Injector) (commentspb.CommentsServiceServer, error) {
		return commentsgrpc.NewServer(
			do.MustInvoke[platformcomments.Store](i),
			do.MustInvoke[database.Client](i),
			sessions.PrincipalFromContext,
			commentsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			commentsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			commentsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
