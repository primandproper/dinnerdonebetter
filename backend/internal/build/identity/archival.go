package identity

import (
	"context"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/succession"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/identity/identitypb"
	"github.com/primandproper/primitives-go/v2/database"
	grpcerrors "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"google.golang.org/grpc/codes"
)

// settlesAccountsOnArchival is platform's identity server with this application's
// household rule in front of ArchiveUser.
//
// platform refuses to archive a user who still owns an account, and is explicit about what
// it wants instead: "Transfer or archive the account first; both are one call away."
// That is right for a directory — it cannot know what an application's accounts are for —
// and it is a problem here, because Register mints every user an account they own alone.
// Left as it stands, no user this application creates could ever be deactivated.
//
// So the rule is stated here, once, on the path an operator actually uses: archiving
// somebody settles the households they own first. A household with other members goes to
// the longest-tenured of them; one they were alone in is archived beside them. It is the
// same rule the erasure path applies, softened — see internal/domain/identity/succession.
//
// A decorator rather than a fork of platform's server: everything else about the surface
// is platform's, including the authorization that has already run by the time this is
// called.
type settlesAccountsOnArchival struct {
	identitypb.IdentityServiceServer

	directory *platformidentity.Service
	rule      *succession.Succession
	db        database.Client
	logger    logging.Logger
	tracer    tracing.Tracer
}

// ArchiveUser settles the caller's households and then archives them.
func (s *settlesAccountsOnArchival) ArchiveUser(
	ctx context.Context,
	request *identitypb.ArchiveUserRequest,
) (*identitypb.ArchiveUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue("user_id", request.GetUserId())

	outcome, err := s.rule.SettleForArchival(ctx, s.directory, s.db.Reader(), ddbidentity.Scope(), request.GetUserId())
	if err != nil {
		return nil, grpcerrors.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal,
			"settling the households of user %q", request.GetUserId())
	}

	logger.WithValue("transferred", len(outcome.Transferred)).
		WithValue("archived_accounts", len(outcome.ArchivedAccountIDs)).
		Info("settled the archived user's households")

	return s.IdentityServiceServer.ArchiveUser(ctx, request)
}
