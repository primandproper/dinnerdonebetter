/*
Package privacy is the audit log's contribution to a subject access request.

The audit log is the section of an export most likely to be misread as an
oversight, so it is worth saying why it is here. An audit entry about a person is
personal data — it says what they did and when — and a subject access request
covers it like anything else. What a subject may not do is have it erased, which
is the asymmetry platform-go models by registering collectors and erasers
separately: this domain exports in full and erases only whole chains it can
remove without making the rest of the log unverifiable. See
internal/domain/audit/privacy/eraser.go for that half.

The collector is platform's audit/privacy, which reads the entries a subject
acted in, the ones they were acted on in, and the ones they recorded while
impersonating somebody else. What this package supplies is which chains to look
in, because which chain an entry lands in is this application's rule: see
audit.ScopeFor.
*/
package privacy

import (
	"context"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// CollectableScopeResolver names every audit chain a subject's entries may be in.
//
// Three sources, because ScopeFor files an entry under the account it happened in
// where there is one and under the acting user otherwise:
//
//   - the subject's own chain, which holds everything they did outside an account;
//   - every account they are a member of now;
//   - every chain holding an entry they acted in, which is the only way to reach an
//     account they have since left. Its chain still holds what they did there, and
//     no membership names it any more.
//
// The request scope is not consulted. An export covers the person, and confining
// it to one account would drop entries that are still about them.
func CollectableScopeResolver(
	store platformidentity.Store,
	log platformaudit.Reader,
	reader database.SQLQueryExecutor,
) platformdataprivacy.ScopeResolver {
	return func(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject) ([]tenancy.Scope, error) {
		seen := map[tenancy.Scope]bool{}
		scopes := []tenancy.Scope{}
		add := func(scope tenancy.Scope) {
			if !seen[scope] {
				seen[scope] = true
				scopes = append(scopes, scope)
			}
		}

		add(tenancy.Of(subject.ID))

		accounts, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformidentity.Account], error) {
			return store.ListAccountsForUser(ctx, reader, tenancy.Global(), subject.ID, filter)
		})
		if err != nil {
			return nil, err
		}

		for i := range accounts {
			add(tenancy.Of(accounts[i].ID))
		}

		acted, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
			return log.ListAcrossScopes(ctx, reader, &platformaudit.Query{ActorID: subject.ID}, filter)
		})
		if err != nil {
			return nil, err
		}

		for i := range acted {
			add(acted[i].Scope)
		}

		return scopes, nil
	}
}
