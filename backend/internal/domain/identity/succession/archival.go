package succession

import (
	"context"

	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// ArchivalOutcome is what settling a user's households before archiving them did.
//
// It is Outcome's softer counterpart: a household the subject was alone in is archived
// rather than deleted, because deactivating somebody is reversible and destroying their
// meal plans is not.
type ArchivalOutcome struct {
	_ struct{} `json:"-"`

	// Transferred is every household that went to a remaining member.
	Transferred []Transfer `json:"transferred"`

	// ArchivedAccountIDs is every household the subject was alone in.
	ArchivedAccountIDs []string `json:"archivedAccountIDs"`
}

// SettleForArchival transfers or archives every household the subject owns, so that
// archiving them is a call the directory will accept.
//
// It exists because platform's Store.ArchiveUser refuses an owner outright and says why:
// "an owner archived out from under their accounts leaves them live and answering to a
// user every scoped read now reports as absent", and it asks the caller to "transfer or
// archive the account first". That is a contract rather than a gap — the directory cannot
// know what an application's accounts are for — and this is this application's answer to
// it, in the same shape and for the same reasons as the erasure rule beside it:
//
//   - A household with other members goes to the longest-tenured of them, who keep using
//     it.
//   - A household the subject was alone in is archived. Nobody else can reach it, and the
//     subject may come back.
//
// Unlike the erasure rule this runs as its own calls rather than inside somebody else's
// transaction, because the operations it needs are the service's and each opens one.
// A failure partway leaves some households settled and the user not archived, which is
// the state a retry resolves — where the alternative, archiving the user first, is the
// one platform refuses precisely because nothing resolves it.
func (s *Succession) SettleForArchival(
	ctx context.Context,
	directory *identity.Service,
	reader database.SQLQueryExecutor,
	scope tenancy.Scope,
	userID string,
) (ArchivalOutcome, error) {
	outcome := ArchivalOutcome{}

	owned, err := s.ownedAccounts(ctx, database.NewTxForTesting(reader), scope, userID)
	if err != nil {
		return outcome, err
	}

	for i := range owned {
		accountID := owned[i].ID

		successor, findErr := s.successorFor(ctx, database.NewTxForTesting(reader), scope, accountID, userID)
		if findErr != nil {
			return outcome, findErr
		}

		if successor == "" {
			if _, archiveErr := directory.ArchiveAccount(ctx, scope, accountID); archiveErr != nil {
				return outcome, platformerrors.Wrapf(archiveErr, "archiving account %q", accountID)
			}

			outcome.ArchivedAccountIDs = append(outcome.ArchivedAccountIDs, accountID)

			continue
		}

		if _, transferErr := directory.TransferAccountOwnership(ctx, scope, accountID, successor); transferErr != nil {
			return outcome, platformerrors.Wrapf(transferErr, "transferring account %q", accountID)
		}

		outcome.Transferred = append(outcome.Transferred, Transfer{AccountID: accountID, NewOwnerUserID: successor})
	}

	return outcome, nil
}
