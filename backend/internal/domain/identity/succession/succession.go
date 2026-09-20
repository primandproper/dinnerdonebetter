/*
Package succession decides what becomes of the households a user owned when
that user is erased.

It is this application's answer to a question platform-go deliberately does not
answer. platform's identity.Store.EraseUser destroys the user row and what its
schema hangs off it — the service roles, the memberships, the roles those carry
— and stops there. Accounts the subject owned survive with an owner_user_id that
resolves to nothing, and platform says so in EraseUser's own documentation. The
omission is argued rather than accidental: identity_accounts carries no foreign
key to identity_users, because a cascade would destroy an organization, its
invoices and every other member's work because one member exercised a right to
be forgotten, and RESTRICT would refuse an erasure that has to commit.

That is right for an organization. A household is not an organization, and this
application has to say what happens to one whose owner leaves:

  - A household with other members is transferred to the longest-tenured of
    them. The people still in it keep their meal plans, their recipes and their
    grocery lists, and the household keeps working.
  - A household the owner was alone in is deleted, and everything in it goes
    with it. There is nobody left for it to belong to, and its contents are the
    erased person's data.

# Where this runs

Inside the erasure's transaction, before EraseUser, and the order is not a
preference. A membership cascades from the user row, so after EraseUser there is
nothing left that says which households the subject was in — platform's
EraseUser documentation says to resolve the subject's accounts first for exactly
this reason. Every write here takes the caller's database.Tx, so a failure
anywhere in the erasure takes the transfers and the deletions back with it.

# The one statement that reaches a platform table

Deleting an account is a DELETE this package issues against identity's own
table, because platform's Store offers ArchiveAccount and no delete.

Archiving would not be an erasure. An archived household keeps the row, and the
row keeps the name the erased person chose it; and archiving the account leaves
every one of this application's twelve tables that cascade from it — the meal
plans, the recipes, the webhooks, the subscriptions — untouched, because they
cascade from a deletion rather than from a flag. A solo household's contents are
the erased person's data by definition, so what the subject is owed is the
deletion the foreign keys already perform.

A Store.DeleteAccount would be the tidier home for it and is worth asking
platform for later. It is not worth blocking on: one statement against a table
whose schema platform owns is a small and visible coupling, and it is named here
so that a schema change upstream lands on a comment rather than on a surprise.
*/
package succession

import (
	"context"
	"sort"

	"github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// ErrNilStore indicates a nil identity store.
var ErrNilStore = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil identity store")

// Transfer is one household that changed hands.
type Transfer struct {
	_ struct{} `json:"-"`

	// AccountID is the household.
	AccountID string `json:"accountID"`

	// NewOwnerUserID is the member it went to.
	NewOwnerUserID string `json:"newOwnerUserID"`
}

// Outcome is what an erasure did to the subject's households, reported so the
// caller can put it in the audit entry and in the subject's erasure receipt.
//
// The two are counted separately rather than summed, because they are different
// facts about other people: a transfer is a household that kept going under
// somebody else, and a deletion is one that did not exist for anybody else.
type Outcome struct {
	_ struct{} `json:"-"`

	// Transferred is every household that went to a remaining member.
	Transferred []Transfer `json:"transferred"`

	// DeletedAccountIDs is every household the subject was alone in.
	DeletedAccountIDs []string `json:"deletedAccountIDs"`
}

// Succession applies this application's rule to the households a subject owned.
type Succession struct {
	_ struct{} `json:"-"`

	store identity.Store

	// tableName is the accounts table the solo-household delete runs against.
	// It is derived from the same prefix the store was built with; see the
	// package documentation for why this package issues that statement at all.
	tableName string
}

// New builds a Succession over an identity store.
//
// tablePrefix must be the prefix the store was built with. They are two
// arguments rather than one because platform's Store does not report its own
// prefix, and a Succession that guessed would delete from a table that is not
// the one the transfers were read from — or, on a shared database, somebody
// else's.
func New(store identity.Store, tablePrefix string) (*Succession, error) {
	if store == nil {
		return nil, ErrNilStore
	}

	name := "identity_accounts"
	if tablePrefix != "" {
		name = tablePrefix + "_" + name
	}

	return &Succession{store: store, tableName: name}, nil
}

// Apply transfers or deletes every household the subject owns, and reports what
// it did.
//
// It must run before identity.Store.EraseUser and on the same transaction. See
// the package documentation.
func (s *Succession) Apply(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	userID string,
) (Outcome, error) {
	outcome := Outcome{}

	owned, err := s.ownedAccounts(ctx, tx, scope, userID)
	if err != nil {
		return outcome, err
	}

	for _, account := range owned {
		successor, findErr := s.successorFor(ctx, tx, scope, account.ID, userID)
		if findErr != nil {
			return outcome, findErr
		}

		if successor == "" {
			if deleteErr := s.deleteAccount(ctx, tx, account.ID); deleteErr != nil {
				return outcome, deleteErr
			}

			outcome.DeletedAccountIDs = append(outcome.DeletedAccountIDs, account.ID)

			continue
		}

		if transferErr := s.store.TransferAccountOwnership(ctx, tx, scope, account.ID, successor); transferErr != nil {
			return outcome, platformerrors.Wrapf(transferErr, "transferring account %q", account.ID)
		}

		outcome.Transferred = append(outcome.Transferred, Transfer{AccountID: account.ID, NewOwnerUserID: successor})
	}

	return outcome, nil
}

// ownedAccounts is every live account the subject owns.
//
// The read is of accounts they are a member of, filtered to the ones they own,
// because that is the question platform's reader answers and ownership is a
// column on what it returns. A subject who owns an account they are not a member
// of is not a state this application can reach: registration makes the owner a
// member, and RemoveMembership refuses the last owner.
//
// It runs on the caller's transaction, so it sees that transaction's own earlier
// writes — which matters when an erasure has already moved something.
func (s *Succession) ownedAccounts(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	userID string,
) ([]identity.Account, error) {
	all, err := dataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
		return s.store.ListAccountsForUser(ctx, tx, scope, userID, filter)
	})
	if err != nil {
		return nil, platformerrors.Wrap(err, "listing the subject's accounts")
	}

	var owned []identity.Account

	for i := range all {
		if all[i].OwnerUserID == userID {
			owned = append(owned, all[i])
		}
	}

	return owned, nil
}

// successorFor is the longest-tenured live member other than the subject, or the
// empty string when the subject was alone.
//
// Longest-tenured means the earliest membership. Ties break on the membership
// identifier, which is arbitrary but total: two people added in the same
// transaction have the same CreatedAt, and an erasure that picked between them
// differently on a retry would hand the household to a different person the
// second time.
func (s *Succession) successorFor(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	accountID, subjectUserID string,
) (string, error) {
	roster, err := dataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.MembershipWithUser], error) {
		return s.store.ListAccountMembers(ctx, tx, scope, accountID, filter)
	})
	if err != nil {
		return "", platformerrors.Wrapf(err, "listing members of account %q", accountID)
	}

	var candidates []identity.Membership

	for i := range roster {
		member := roster[i]
		if member.BelongsToUser == subjectUserID {
			continue
		}

		// Archived memberships are somebody who has already left. The store's
		// roster read excludes them; filtering here as well is what keeps that an
		// implementation detail rather than something this rule depends on.
		if member.ArchivedAt != nil {
			continue
		}

		candidates = append(candidates, member.Membership)
	}

	if len(candidates) == 0 {
		return "", nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if !candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
		}

		return candidates[i].ID < candidates[j].ID
	})

	return candidates[0].BelongsToUser, nil
}

// deleteAccount removes the household row, and with it everything this
// application's schema hangs off it by ON DELETE CASCADE.
//
// See the package documentation for why this is a statement rather than a Store
// call.
func (s *Succession) deleteAccount(ctx context.Context, tx database.Tx, accountID string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+s.tableName+` WHERE id = $1`, accountID); err != nil {
		return platformerrors.Wrapf(err, "deleting account %q", accountID)
	}

	return nil
}
