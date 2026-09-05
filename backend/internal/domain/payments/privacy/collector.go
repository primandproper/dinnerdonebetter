/*
Package privacy is the payments domain's contribution to a subject access
request: what every account the subject belongs to was sold, and what it paid.

The collector is platform-go's, over platform-go's store. What this package adds
is the one question platform cannot answer — which accounts a subject's billing is
under — which is this application's tenancy model and is resolved through the
identity repository the way every other account-scoped collector here resolves it.

# There is no eraser here, and there is deliberately none upstream

Payments is the domain whose export and erasure disagree most sharply: a subject
is entitled to see every transaction, and financial records are the ones a
statutory retention generally requires be kept. platform-go's billing/privacy
ships a Collector and no Eraser for exactly that reason, and this package follows
it.

What erases billing rows today is the cascade: every one of the three
account-owned billing tables carries a foreign key to accounts with ON DELETE
CASCADE, re-created by the migration that adopted the store, so the single
identity eraser takes them with the user. That preserves what this schema always
did. Whether it should — whether these rows want a retention rule and an
anonymizing eraser instead — is the decision docs/data-privacy.md has named as
the likeliest first case for one, and it is not taken here.
*/
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/billing"
	billingprivacy "github.com/primandproper/platform-go/v13/billing/privacy"
	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
)

// NewCollector builds the payments collector: every subscription, purchase and
// ledger row in every account the subject belongs to, archived rows included,
// paged to the end and encoded.
func NewCollector(store billing.Store, resolveAccounts dataprivacy.AccountIDResolver) (platformdataprivacy.Collector, error) {
	collector, err := billingprivacy.NewCollector(store, accountResolver(resolveAccounts))
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the billing data privacy collector")
	}

	return collector, nil
}

// accountResolver adapts this application's account resolution to the shape
// platform's collector takes.
//
// Every account is in the one scope this application keeps its billing under, so
// the resolver's only real work is naming the accounts; see payments.Scope.
func accountResolver(resolveAccounts dataprivacy.AccountIDResolver) billingprivacy.AccountResolver {
	return func(ctx context.Context, subject platformdataprivacy.Subject) ([]billingprivacy.Account, error) {
		accountIDs, err := resolveAccounts(ctx, subject.ID)
		if err != nil {
			return nil, err
		}

		accounts := make([]billingprivacy.Account, 0, len(accountIDs))
		for _, accountID := range accountIDs {
			accounts = append(accounts, billingprivacy.Account{ID: accountID, Scope: payments.Scope()})
		}

		return accounts, nil
	}
}
