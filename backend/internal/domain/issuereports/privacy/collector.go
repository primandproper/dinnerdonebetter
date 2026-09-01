// Package privacy is the issue reports domain's contribution to a subject access
// request.
//
// There is a collector here and no eraser, and the absence is deliberate.
// platform-go ships both halves — the details are free text somebody typed, so a
// report that outlived its reporter would be personal data no erasure reaches —
// but this application re-creates the reporter's foreign key to users when it
// renders the table (see internal/repositories/postgres/migrations), so the row
// goes with the user record the single identity eraser deletes. A second eraser
// would issue a DELETE for rows Postgres has already removed.
//
// If that foreign key ever goes, this package must register
// issuereports/privacy's Eraser under a key of its own. Nothing else covers the
// table.
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	issuereportsprivacy "github.com/primandproper/platform-go/v13/issuereports/privacy"
	"github.com/primandproper/platform-go/v13/tenancy"
)

// NewCollector builds the issue reports collector: every report the subject
// filed, in every account they appear in, paged to the end and encoded.
func NewCollector(store issuereports.Store, resolveAccounts dataprivacy.AccountIDResolver) (platformdataprivacy.Collector, error) {
	return issuereportsprivacy.NewCollector(store, AccountScopes(resolveAccounts))
}

// AccountScopes turns "which accounts does this user appear in" into the scopes
// the platform's collector pages.
//
// It is this application's tenancy model expressed as the mapping platform asks
// for, and it is a mapping rather than a default because there is no default that
// is right twice — a deployment with one tenant would answer the global scope
// here, and answering it in this one would export nothing. See
// ddbissuereports.Scope.
func AccountScopes(resolveAccounts dataprivacy.AccountIDResolver) issuereportsprivacy.ScopeResolver {
	return func(ctx context.Context, subject platformdataprivacy.Subject) ([]tenancy.Scope, error) {
		accountIDs, err := resolveAccounts(ctx, subject.ID)
		if err != nil {
			return nil, platformerrors.Wrap(err, "resolving accounts")
		}

		scopes := make([]tenancy.Scope, 0, len(accountIDs))
		for _, accountID := range accountIDs {
			scopes = append(scopes, ddbissuereports.Scope(accountID))
		}

		return scopes, nil
	}
}
