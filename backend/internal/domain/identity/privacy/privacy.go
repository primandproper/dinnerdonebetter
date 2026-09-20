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

// Eraser is platform's identity eraser with this application's succession rule in
// front of it.
//
// The order is the whole of it, and it is not a preference: a membership cascades
// from the user row, so after EraseUser there is nothing left that says which
// households the subject was in. Both halves run on the caller's transaction, so a
// failure in either takes the other back with it.
type Eraser struct {
	_ struct{} `json:"-"`

	succession *succession.Succession
	inner      platformdataprivacy.Eraser
}

var _ platformdataprivacy.Eraser = (*Eraser)(nil)

// NewEraser builds the identity eraser.
func NewEraser(store identity.Store) (*Eraser, error) {
	rule, err := succession.New(store, ddbidentity.TablePrefix)
	if err != nil {
		return nil, err
	}

	inner, err := identityprivacy.NewEraser(store, Scopes())
	if err != nil {
		return nil, err
	}

	return &Eraser{succession: rule, inner: inner}, nil
}

// Erase settles the subject's households and then destroys the subject.
func (e *Eraser) Erase(
	ctx context.Context,
	tx database.Tx,
	requestScope tenancy.Scope,
	subject platformdataprivacy.Subject,
) (platformdataprivacy.ErasureOutcome, error) {
	if _, err := e.succession.Apply(ctx, tx, ddbidentity.Scope(), subject.ID); err != nil {
		return platformdataprivacy.ErasureOutcome{}, platformerrors.Wrap(err, "settling the subject's households")
	}

	return e.inner.Erase(ctx, tx, requestScope, subject)
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
