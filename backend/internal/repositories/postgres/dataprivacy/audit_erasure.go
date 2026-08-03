package dataprivacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v9/database/dialect"
	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
	"github.com/primandproper/platform-go/v9/dataprivacy/auditerasure"
	"github.com/primandproper/platform-go/v9/filtering"
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

ScopeFor is what makes that cover most of a departing user's trail. Account
events chain per account and the accounts they owned are theirs to delete;
everything with no account — signup, login, password reset — chains per user, so
that scope is theirs too. What survives is their actions inside accounts they
merely belonged to, which cannot be removed without breaking those accounts'
chains, and which are the entries most likely to be covered by the
legitimate-interest grounds audit logs are normally kept under. They are reported
as retained rather than silently kept.
*/

// eraserForScopes builds an eraser that removes exactly the given audit scopes.
//
// The scopes are fixed at construction rather than resolved when the eraser runs,
// because by then they are unresolvable: the accounts that name them have been
// cascaded away with the user, and asking which accounts a deleted user owned
// returns nothing. They have to be read before the delete and carried across it.
func eraserForScopes(scopes []string) (*auditerasure.Eraser, error) {
	return auditerasure.New(
		dialect.Postgres,
		audit.TablePrefix,
		auditerasure.WithScopeResolver(func(context.Context, platformdataprivacy.Subject) ([]string, error) {
			return scopes, nil
		}),
	)
}

// erasableAuditScopes lists the audit chains that belong to a user outright: the
// accounts they own, and their own user scope.
//
// Owned, not merely joined. Deleting the scope of an account the user was one
// member of would destroy the audit trail of everyone else in it, which is the
// failure the platform's own documentation warns is worth being exact about.
func erasableAuditScopes(ctx context.Context, identityRepo identity.Repository, userID string) ([]string, error) {
	accounts, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
		return identityRepo.GetAccounts(ctx, userID, filter)
	})
	if err != nil {
		return nil, err
	}

	// The user's own scope, which holds every event that happened outside an account
	// — the ones ScopeFor files under the user precisely so that logins do not all
	// serialize on one chain.
	scopes := []string{userID}

	for _, account := range accounts {
		if account.BelongsToUser == userID {
			scopes = append(scopes, account.ID)
		}
	}

	return scopes, nil
}
