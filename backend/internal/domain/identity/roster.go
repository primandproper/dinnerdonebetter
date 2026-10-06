package identity

import (
	"context"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// MembersOfAccount reads the ids of every user in an account.
//
// The walk is platform's: identity.ListAllAccountMembers follows the roster's cursor to the
// end, which is what keeps a meal plan's voting from being decided on the strength of the
// first page of members. What is left here is the projection every caller wants — a user id
// per member — so the six call sites do not each write it.
//
// It takes a DirectoryReader rather than the whole Store because that is the read-only half
// of the directory, and a worker that only ever reads a roster should not be holding a value
// that could also archive a user.
func MembersOfAccount(
	ctx context.Context,
	directory platformidentity.DirectoryReader,
	q database.SQLQueryExecutor,
	accountID string,
) ([]string, error) {
	roster, err := platformidentity.ListAllAccountMembers(ctx, q, tenancy.Global(), directory, accountID)
	if err != nil {
		return nil, err
	}

	var members []string

	for _, membership := range roster {
		if membership != nil && membership.User != nil && membership.User.ID != "" {
			members = append(members, membership.User.ID)
		}
	}

	return members, nil
}
