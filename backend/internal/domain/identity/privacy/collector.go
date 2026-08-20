/*
Package privacy is the identity domain's contribution to a subject access
request: the user record, the accounts they appear in, and the invitations they
sent or received.

It is also the domain every other account-scoped collector depends on, because
"which accounts is this person in" is a question only identity can answer. That
lookup is exported as ResolveAccountIDs rather than reached for through the
identity repository directly, so a collector in another domain does not acquire
a dependency on this one.
*/
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformdataprivacy "github.com/primandproper/platform-go/v12/dataprivacy"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const o11yName = "identity_privacy_collector"

// Collector collects identity data about a subject.
type Collector struct {
	repo   identity.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the identity collector.
func NewCollector(repo identity.Repository, logger logging.Logger, tracerProvider tracing.Provider) *Collector {
	return &Collector{
		repo:   repo,
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
//
// A missing user is an error rather than an empty fragment. Every other section
// of an export can be legitimately empty, but a request naming a subject whose
// user record cannot be read has been asked about somebody the application does
// not have, and answering that with a valid document is the one wrong answer
// available.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, subject.ID)

	user, err := c.repo.GetUser(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user")
	}

	accounts, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
		return c.repo.GetAccounts(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching accounts")
	}

	invitations, err := c.invitations(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitations")
	}

	return dataprivacy.Fragment(true, &identity.UserDataCollection{
		User:               *user,
		Accounts:           accounts,
		AccountInvitations: invitations,
	})
}

// invitations returns the invitations the user sent and the ones they received,
// deduplicated.
//
// The two lists overlap whenever somebody invites an address that turns out to be
// their own, which is rare and is exactly the sort of thing a subject notices in
// an export. Deduplication is by ID because that is the only field guaranteed to
// identify the same row on both sides.
func (c *Collector) invitations(ctx context.Context, userID string) ([]identity.AccountInvitation, error) {
	sent, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
		return c.repo.GetPendingAccountInvitationsFromUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, err
	}

	received, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
		return c.repo.GetPendingAccountInvitationsForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, err
	}

	total := len(sent) + len(received)
	seen := make(map[string]struct{}, total)
	out := make([]identity.AccountInvitation, 0, total)

	for _, batch := range [][]identity.AccountInvitation{sent, received} {
		for i := range batch {
			if _, ok := seen[batch[i].ID]; ok {
				continue
			}

			seen[batch[i].ID] = struct{}{}
			out = append(out, batch[i])
		}
	}

	return out, nil
}

// ResolveAccountIDs builds the AccountIDResolver every account-scoped collector
// is handed.
func ResolveAccountIDs(repo identity.Repository) dataprivacy.AccountIDResolver {
	return func(ctx context.Context, userID string) ([]string, error) {
		accounts, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
			return repo.GetAccounts(ctx, userID, filter)
		})
		if err != nil {
			return nil, err
		}

		ids := make([]string, 0, len(accounts))
		for i := range accounts {
			ids = append(ids, accounts[i].ID)
		}

		return ids, nil
	}
}
