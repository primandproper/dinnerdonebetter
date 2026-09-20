package identity

import (
	"context"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// rosterPageSize is how many members one page of the walk below asks for.
const rosterPageSize uint16 = 50

// AccountRoster is the whole of what this application's meal planning needs from the
// directory: who is in an account.
//
// One method rather than identity.Store, because that is all any of its callers want and a
// worker that could also archive a user is one that could be made to. It is the same
// narrowing platform applies to signin.Directory and identity.SignInReader, and for the
// reason those state: this is the interface a component holding nobody's credentials should
// be able to satisfy.
type AccountRoster interface {
	ListAccountMembers(
		ctx context.Context,
		q database.SQLQueryExecutor,
		scope tenancy.Scope,
		accountID string,
		filter *filtering.QueryFilter,
	) (*filtering.QueryFilteredResult[platformidentity.MembershipWithUser], error)
}

// MembersOfAccount reads the ids of every user in an account, following the cursor to the
// end.
//
// The loop is the point, and it is why this is a function rather than a call each of the
// three callers writes. The read it replaces answered with an Account carrying its whole
// roster; platform's is paged, so a caller that read one page and iterated it would decide a
// meal plan's voting was complete on the strength of the first fifty members and ignore the
// rest. platform makes that hard to get wrong where the set is always small —
// ListMembershipsForUser answers with a slice and says why — and this set is a household,
// which is also small, but the method is paged because an account is not always a household.
//
// A page shorter than the one asked for ends the walk, which is what a cursor means.
func MembersOfAccount(
	ctx context.Context,
	roster AccountRoster,
	q database.SQLQueryExecutor,
	accountID string,
) ([]string, error) {
	var (
		members []string
		cursor  string
	)

	for {
		size := rosterPageSize
		filter := filtering.DefaultQueryFilter()
		filter.MaxResponseSize = &size

		if cursor != "" {
			c := cursor
			filter.Cursor = &c
		}

		page, err := roster.ListAccountMembers(ctx, q, tenancy.Global(), accountID, filter)
		if err != nil {
			return nil, err
		}

		for _, membership := range page.Data {
			if membership != nil && membership.User != nil && membership.User.ID != "" {
				members = append(members, membership.User.ID)
			}
		}

		// A short page is the end of the walk, and a full one with no cursor to follow is
		// also the end: asking again would ask for the same page forever.
		if len(page.Data) < int(size) || page.Cursor == "" {
			return members, nil
		}

		cursor = page.Cursor
	}
}
