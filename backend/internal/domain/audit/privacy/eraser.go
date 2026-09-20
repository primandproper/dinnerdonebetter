package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/platform-go/v14/dataprivacy/auditerasure"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

/*
Erasing a user used to be a single DELETE. Every belongs_to_user foreign key
carried ON DELETE CASCADE, the audit rows carried one too, and they went with
everything else. The platform audit tables have no such key and could not: the
entries are a hash chain, and removing one from the middle of a scope is
indistinguishable — to Verify, and to anyone reading its output — from an
attacker removing it. That is the property the audit package exists to provide,
and an erasure that quietly broke it would trade a real security control for a
checkbox.

So erasure does the one deletion the structure permits: whole scopes, entries and
chain rows together. A scope that disappears entirely leaves no gap in any
surviving chain, because there is nothing left to verify against.

audit.ScopeFor is what makes that cover most of a departing user's trail. Account
events chain per account and the accounts they owned are theirs to delete;
everything with no account — signup, login, password reset — chains per user, so
that scope is theirs too. What survives is their actions inside accounts they
merely belonged to, which cannot be removed without breaking those accounts'
chains, and which are the entries most likely to be covered by the
legitimate-interest grounds audit logs are normally kept under. auditerasure
reports them as retained, with a stated basis, rather than silently keeping them.
*/

// ErasableScopeResolver builds the auditerasure.ScopeResolver this application
// erases audit chains by.
//
// The scopes are resolved when the eraser runs rather than fixed at
// construction, which is only sound because every eraser shares one transaction
// and platform-go runs them in sorted key order: "audit" precedes "identity", so
// the accounts that name these scopes are still there to be read when this asks
// for them. Before the shared transaction existed, this lookup had to happen
// before the delete and be carried across it — and if the deletion order ever
// changes, it has to go back to that.
// The request scope is not consulted. A resolver answers "which chains are this
// subject's", and the chains an owned account holds are the same set whether or
// not the request confined itself to one of them; narrowing here would leave the
// other accounts' chains behind while reporting a complete erasure.
func ErasableScopeResolver(repo identity.Repository) auditerasure.ScopeResolver {
	return func(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject) ([]tenancy.Scope, error) {
		return erasableScopes(ctx, repo, subject.ID)
	}
}

// erasableScopes lists the audit chains that belong to a user outright: the
// accounts they own, and their own user scope.
//
// Owned, not merely joined. Deleting the scope of an account the user was one
// member of would destroy the audit trail of everyone else in it, which is the
// failure platform-go's own documentation warns is worth being exact about.
func erasableScopes(ctx context.Context, repo identity.Repository, userID string) ([]tenancy.Scope, error) {
	accounts, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
		return repo.GetAccounts(ctx, userID, filter)
	})
	if err != nil {
		return nil, err
	}

	// The user's own scope, which holds every event that happened outside an account
	// — the ones ScopeFor files under the user precisely so that logins do not all
	// serialize on one chain.
	scopes := []tenancy.Scope{tenancy.Of(userID)}

	for i := range accounts {
		if accounts[i].BelongsToUser == userID {
			scopes = append(scopes, tenancy.Of(accounts[i].ID))
		}
	}

	return scopes, nil
}
