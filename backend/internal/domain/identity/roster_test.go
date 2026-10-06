package identity

import (
	"context"
	"testing"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMembersOfAccount(T *testing.T) {
	T.Parallel()

	// The case a one-page read gets wrong: a roster longer than a page. The member on the
	// second page is the one a meal plan's voting would otherwise never wait for.
	T.Run("reads past the first page", func(t *testing.T) {
		t.Parallel()

		accountID := identifiers.New()
		lastMember := identifiers.New()

		var firstPage []*platformidentity.MembershipWithUser
		for range filtering.MaxQueryFilterLimit {
			firstPage = append(firstPage, &platformidentity.MembershipWithUser{User: &platformidentity.User{ID: identifiers.New()}})
		}

		directory := &identitymock.StoreMock{
			ListAccountMembersFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				_ tenancy.Scope,
				id string,
				filter *filtering.QueryFilter,
			) (*filtering.QueryFilteredResult[platformidentity.MembershipWithUser], error) {
				assert.Equal(t, accountID, id)

				if filter.Cursor == nil {
					return &filtering.QueryFilteredResult[platformidentity.MembershipWithUser]{
						Data:   firstPage,
						Cursor: identifiers.New(), MaxResponseSize: filtering.MaxQueryFilterLimit,
					}, nil
				}

				return &filtering.QueryFilteredResult[platformidentity.MembershipWithUser]{
					Data: []*platformidentity.MembershipWithUser{{User: &platformidentity.User{ID: lastMember}}},
				}, nil
			},
		}

		members, err := MembersOfAccount(t.Context(), directory, nil, accountID)
		require.NoError(t, err)

		assert.Len(t, members, int(filtering.MaxQueryFilterLimit)+1)
		assert.Contains(t, members, lastMember)
		assert.Len(t, directory.ListAccountMembersCalls(), 2)
	})
}
