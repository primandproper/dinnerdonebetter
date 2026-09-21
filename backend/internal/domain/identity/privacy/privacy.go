/*
Package privacy is this application's identity contribution to a subject access
request.

Almost none of it is written here. platform-go's identity/privacy ships both
halves — a collector that exports who somebody is to the directory, and an
eraser that destroys them — and what is left for this package is the two things
platform cannot decide: which accounts a subject appears in, which every other
domain's collector needs and only the directory can answer, and what becomes of
the households the subject owned, which is succession's rule and runs inside the
erasure's transaction before the user row goes.
*/
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/succession"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/platform-go/v14/identity"
	identityprivacy "github.com/primandproper/platform-go/v14/identity/privacy"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// Scopes is the scope resolver every identity privacy adapter here is built with.
//
// One directory, so one scope, for every subject. See internal/domain/identity's
// Scope for why this deployment has exactly one.
func Scopes() identityprivacy.ScopeResolver {
	return platformdataprivacy.FixedScopes(ddbidentity.Scope())
}

// NewCollector builds the identity collector over platform's.
func NewCollector(store identity.Store, reader database.SQLQueryExecutor) (platformdataprivacy.Collector, error) {
	return identityprivacy.NewCollector(store, reader, Scopes())
}

// SuccessionStep is this application's rule about the households a departing owner
// leaves behind, in the shape privacyadapters runs ahead of identity's own eraser.
//
// The order is the whole of it, and it is not a preference: a membership cascades from
// the user row, so after the eraser has run there is nothing left that says which
// households the subject was in. It has to happen first or it cannot happen at all.
//
// It is a BeforeErase rather than a wrapper around platform's eraser, and platform
// declines to offer the wrapper on purpose. An ErasureOutcome's Retained carries the
// legal basis for anything kept, and that goes into the request record and in front of a
// regulator; a wrapper holding the real eraser could report counts that eraser never
// produced, and a falsified count is not detectable from outside — the key is registered,
// the roster is satisfied and the artifact is well-formed. So this never holds it.
//
// What it gets in exchange is the guarantee it actually needed: every eraser registered
// for one request runs on one transaction, so the transfer below commits with the erasure
// or not at all.
//
// The outcome is empty by construction. Deleted and Anonymized sum across steps and a
// household that changed hands is neither — it is somebody else's now, and still there.
// Retained would be a legal basis for keeping something, and this keeps nothing that the
// erasure would otherwise have taken.
//
// A failure here takes the whole erasure down and identity's eraser does not run, which
// is the right way round: a precondition that failed is a precondition. It does mean a bug
// in the succession rule blocks erasure for that subject until it is fixed — loud and
// recoverable, where the alternative is a subject told they were erased who was not.
func SuccessionStep(store identity.Store) (platformdataprivacy.Eraser, error) {
	rule, err := succession.New(store, ddbidentity.TablePrefix)
	if err != nil {
		return nil, err
	}

	return &successionStep{rule: rule}, nil
}

type successionStep struct {
	_ struct{} `json:"-"`

	rule *succession.Succession
}

func (e *successionStep) Erase(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	subject platformdataprivacy.Subject,
) (platformdataprivacy.ErasureOutcome, error) {
	if _, err := e.rule.Apply(ctx, tx, ddbidentity.Scope(), subject.ID); err != nil {
		return platformdataprivacy.ErasureOutcome{}, platformerrors.Wrap(err, "settling the subject's households")
	}

	return platformdataprivacy.ErasureOutcome{}, nil
}

// ResolveAccountIDs builds the AccountIDResolver every account-scoped collector is
// handed: which accounts this subject appears in.
//
// It is exported here rather than reached for through the directory directly, so a
// collector in another domain does not acquire a dependency on this one.
func ResolveAccountIDs(store identity.Store, reader database.SQLQueryExecutor) dataprivacy.AccountIDResolver {
	return func(ctx context.Context, userID string) ([]string, error) {
		accounts, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
			return store.ListAccountsForUser(ctx, reader, ddbidentity.Scope(), userID, filter)
		})
		if err != nil {
			return nil, err
		}

		ids := make([]string, 0, len(accounts))
		for _, account := range accounts {
			ids = append(ids, account.ID)
		}

		return ids, nil
	}
}
