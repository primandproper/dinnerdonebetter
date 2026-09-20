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

	T.Run("merges the two chains and adds their counts", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), false)

		filter := filterOfSize(10)

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			ListFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				query *platformaudit.Query,
				f *filtering.QueryFilter,
			) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
				if *query.Scope == accountScope {
					return filtering.NewQueryFilteredResult(entriesNamed("b", "d"), 2, 7, idOf, f), nil
				}

				return filtering.NewQueryFilteredResult(entriesNamed("a", "c"), 3, 5, idOf, f), nil
			},
		}}

		page, err := reader.List(ctx, nil, &platformaudit.Query{Scope: &accountScope}, filter)
		require.NoError(t, err)

		// Interleaved by id, which is what makes the merge the same page a single
		// query over both scopes would have cut.
		assert.Equal(t, []string{"a", "b", "c", "d"}, idsOf(page.Data))

		// Summed, because an entry lives in exactly one chain: the two scopes are
		// disjoint, so their counts describe a union with nothing counted twice.
		filtered, total, known := page.Pagination.Counts()
		assert.True(t, known)
		assert.Equal(t, uint64(5), filtered)
		assert.Equal(t, uint64(12), total)
	})

	T.Run("withholds the counts when either chain did not answer", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), false)

		filter := filterOfSize(10)

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			ListFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				query *platformaudit.Query,
				f *filtering.QueryFilter,
			) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
				if *query.Scope == accountScope {
					return filtering.NewQueryFilteredResult(entriesNamed("b"), 1, 4, idOf, f), nil
				}

				// The empty page a querygen store hands back when it has no row to
				// read its counts off. Zero here is "unknown", not "none".
				return filtering.NewQueryFilteredResultWithoutCounts(nil, idOf, f), nil
			},
		}}

		page, err := reader.List(ctx, nil, &platformaudit.Query{Scope: &accountScope}, filter)
		require.NoError(t, err)
		assert.Equal(t, []string{"b"}, idsOf(page.Data))

		// Not 1 and 4. Reporting the account's counts as the union's would be short
		// by however much the other chain holds, and it would look right until
		// somebody paged to the end of one.
		_, _, known := page.Pagination.Counts()
		assert.False(t, known, "a chain that answered no counts was counted as zero")
	})

	T.Run("keeps only the page the filter asked for", func(t *testing.T) {
		t.Parallel()
		ctx := sessionFor(t.Context(), false)

		filter := filterOfSize(3)

		reader := spanningReader{Reader: &auditmock.ReaderMock{
			ListFunc: func(
				_ context.Context,
				_ database.SQLQueryExecutor,
				query *platformaudit.Query,
				f *filtering.QueryFilter,
			) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
				if *query.Scope == accountScope {
					return filtering.NewQueryFilteredResult(entriesNamed("b", "d", "f"), 3, 3, idOf, f), nil
				}

				return filtering.NewQueryFilteredResult(entriesNamed("a", "c", "e"), 3, 3, idOf, f), nil
			},
		}}

		page, err := reader.List(ctx, nil, &platformaudit.Query{Scope: &accountScope}, filter)
		require.NoError(t, err)

		// Two full pages merged are twice as many rows as anybody asked for, and the
		// three smallest of the union are the page a single query would have cut.
		assert.Equal(t, []string{"a", "b", "c"}, idsOf(page.Data))
	})

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
