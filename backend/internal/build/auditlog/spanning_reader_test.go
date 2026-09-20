package auditlog

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	auditmock "github.com/primandproper/platform-go/v14/audit/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID    = "user_under_test"
	testAccountID = "account_under_test"
)

// sessionFor puts a session on the context, service admin or not.
func sessionFor(ctx context.Context, admin bool) context.Context {
	roleName := authorization.ServiceUserRoleName
	perms := []authorization.Permission{}

	if admin {
		roleName = authorization.ServiceAdminRoleName
		perms = authorization.ServiceAdminPermissions
	}

	return sessions.AttachToContext(ctx, &sessions.ContextData{
		ActiveAccountID: testAccountID,
		Requester: sessions.RequesterInfo{
			UserID:             testUserID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{roleName}, perms),
		},
	})
}

func entriesNamed(ids ...string) []*platformaudit.Entry {
	out := make([]*platformaudit.Entry, 0, len(ids))
	for _, id := range ids {
		out = append(out, &platformaudit.Entry{ID: id})
	}

	return out
}

func idOf(e *platformaudit.Entry) string { return e.ID }

func filterOfSize(size uint16) *filtering.QueryFilter {
	f := filtering.DefaultQueryFilter()
	f.MaxResponseSize = &size

	return f
}

func TestSpanningReader_List(T *testing.T) {
	T.Parallel()

	accountScope := tenancy.Of(testAccountID)

	T.Run("a service admin reads every chain in one unscoped read", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), true)

		var scopes []*tenancy.Scope

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			ListFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				query *platformaudit.Query,
				f *filtering.QueryFilter,
			) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
				scopes = append(scopes, query.Scope)

				return filtering.NewQueryFilteredResult(entriesNamed("a"), 1, 1, idOf, f), nil
			},
		}}

		page, err := reader.List(ctx, nil, &platformaudit.Query{Scope: &accountScope}, filterOfSize(10))
		require.NoError(t, err)
		assert.Equal(t, []string{"a"}, idsOf(page.Data))

		// One read, not two, and unscoped: nil is platform's operator read, so there
		// are no chains left to merge and the store's own counts stand.
		require.Len(t, scopes, 1)
		assert.Nil(t, scopes[0], "an operator's read was narrowed to one chain")
	})

	T.Run("an unscoped query is left alone", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), false)

		var calls int

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			ListFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				_ *platformaudit.Query,
				f *filtering.QueryFilter,
			) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
				calls++

				return filtering.NewQueryFilteredResult(entriesNamed("a"), 1, 1, idOf, f), nil
			},
		}}

		_, err := reader.List(ctx, nil, &platformaudit.Query{}, filterOfSize(10))
		require.NoError(t, err)

		// Nil already spans every chain. Adding the caller's own would narrow it.
		assert.Equal(t, 1, calls)
	})
}

func TestSpanningReader_Get(T *testing.T) {
	T.Parallel()

	accountScope := tenancy.Of(testAccountID)

	T.Run("falls back to the caller's own chain", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), false)

		wanted := &platformaudit.Entry{ID: "entry"}

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			GetFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				scope *tenancy.Scope,
				_ string,
			) (*platformaudit.Entry, error) {
				if scope != nil && *scope == tenancy.Of(testUserID) {
					return wanted, nil
				}

				return nil, platformaudit.ErrEntryNotFound
			},
		}}

		entry, err := reader.Get(ctx, nil, &accountScope, "entry")
		require.NoError(t, err)
		assert.Equal(t, wanted, entry)
	})

	T.Run("an entry in neither chain is not found", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), false)

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			GetFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				scope *tenancy.Scope,
				_ string,
			) (*platformaudit.Entry, error) {
				// Never unscoped for a caller: a nil scope here would be every
				// tenant's log, which is the disclosure the pointer prevents.
				require.NotNil(t, scope)

				return nil, platformaudit.ErrEntryNotFound
			},
		}}

		_, err := reader.Get(ctx, nil, &accountScope, "entry")
		assert.Error(t, err)
	})

	T.Run("a service admin reads any chain", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), true)

		somebodyElses := &platformaudit.Entry{ID: "entry"}

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			GetFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				scope *tenancy.Scope,
				_ string,
			) (*platformaudit.Entry, error) {
				if scope != nil {
					return nil, platformaudit.ErrEntryNotFound
				}

				return somebodyElses, nil
			},
		}}

		entry, err := reader.Get(ctx, nil, &accountScope, "entry")
		require.NoError(t, err)
		assert.Equal(t, somebodyElses, entry)
	})
}

func idsOf(entries []*platformaudit.Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.ID)
	}

	return out
}

func TestCallerChains(T *testing.T) {
	T.Parallel()

	T.Run("a signed-in caller belongs to their account's chain and their own", func(t *testing.T) {
		t.Parallel()

		chains, err := callerChains(sessionFor(t.Context(), false))
		require.NoError(t, err)

		// The account first, because it is the chain the connection resolves to and the
		// one every write goes into; the caller's own second, where a login, a signup
		// and a password reset land because none of those happens inside an account.
		assert.Equal(t, []tenancy.Scope{tenancy.Of(testAccountID), tenancy.Of(testUserID)}, chains)
	})

	T.Run("an account-less session names its one chain once", func(t *testing.T) {
		t.Parallel()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: testUserID},
		})

		chains, err := callerChains(ctx)
		require.NoError(t, err)

		// Not twice. The connection already resolves to this caller's own chain, and
		// naming it again would page one chain as two and report every row of it twice.
		assert.Equal(t, []tenancy.Scope{tenancy.Of(testUserID)}, chains)
	})

	T.Run("a service admin is answered with no chains at all", func(t *testing.T) {
		t.Parallel()

		chains, err := callerChains(sessionFor(t.Context(), true))
		require.NoError(t, err)

		// Their read is every chain in the deployment, which is not a slice anybody can
		// enumerate. Empty means the connection's own scope, and spanningReader widens
		// that one read to the operator's.
		assert.Empty(t, chains)
	})

	T.Run("an anonymous caller names none", func(t *testing.T) {
		t.Parallel()

		chains, err := callerChains(t.Context())
		require.NoError(t, err)
		assert.Empty(t, chains)
	})
}
