# Search Pagination

Search endpoints are cursor-paginated. A request carries the query, a `useSearchService` flag, and a
query filter; the response carries one page of results and a `pagination` block whose `cursor`
resumes the search. `internal/searchpagination` sits between the text search index and the
`filtering.QueryFilter` pagination the managers hand back, so both search paths page the same way.

## Overview

`useSearchService` picks which backend answers the search, and that decides what the cursor is:

| `useSearchService` | Backend                                          | The cursor is                  | Read by                         |
|--------------------|--------------------------------------------------|--------------------------------|---------------------------------|
| `false`            | Postgres (`SearchFor*` repository queries)       | The last row's ID              | The SQL, as `WHERE id > cursor` |
| `true`             | The text search index (Elasticsearch or Algolia) | An opaque token from the index | Only the backend that issued it |

Both kinds travel in `filtering.QueryFilter.Cursor` on the way in and come back in
`pagination.cursor`. A client treats them the same way: a string it hands back verbatim, never one it
constructs, parses, or persists meaning from.

## Paging through a result set

1. Send the first request with no cursor.
2. Read `pagination.cursor` off the response.
3. Send it back as the filter's `cursor`, with the same query and the **same `useSearchService`
   value**.
4. Stop when `pagination.cursor` comes back empty.

An empty cursor is the only signal that the result set is exhausted. A short page is not: both
Elasticsearch and Algolia can return fewer hits than asked for and still have more behind them, so
each issues the next cursor from the total it knows about rather than from the page's length.

Changing `useSearchService` part-way through a walk invalidates the cursor. Cursors are tagged with
the backend that issued them, so the index rejects a database cursor with `InvalidArgument`. The
reverse cannot be caught — the database would compare an opaque token against the ID column and
answer with an arbitrary slice of the table — which is why `searchpagination.FilterForDatabaseFallback`
strips the cursor whenever a manager falls back from index to database.

## Page sizes

The filter's `maxResponseSize` becomes `textsearch.SearchRequest.Limit`.

- A gRPC request that omits `maxResponseSize` gets 50 (`filtering.DefaultQueryFilterLimit`), and
  anything above 250 (`filtering.MaxQueryFilterLimit`) is clamped by
  `ConvertGRPCQueryFilterToQueryFilter`.
- A filter that reaches the index carrying no page size at all uses `textsearch.DefaultSearchLimit`
  (25).
- Both index backends cap a single page at 200, so a request for more comes back short with a cursor
  still set — which the rule above already covers.

## Totals

Index-backed responses report `totalCount` as `0`, meaning unknown: the index reports whether another
page exists, not how many results there are in all. Database-backed responses report a real count,
computed by the search query itself. Page with the cursor, not with a count.

## Statuses a paging client can get

| Status            | Cause                                                                    | What to do                                               |
|-------------------|--------------------------------------------------------------------------|----------------------------------------------------------|
| `OutOfRange`      | Paged past Elasticsearch's `index.max_result_window` (10,000 by default) | Stop and narrow the query; paging further will not work. |
| `InvalidArgument` | A cursor the index did not issue                                         | Start over from the first page.                          |

`internal/services/errors/search_grpc_mapper.go` registers both mappings, and any service package
that blank-imports it picks them up. A registered mapping wins over the handler's default code, so
these reach the client rather than `Internal`. `OutOfRange` is deliberately not reported as an empty
last page: "there are no more results" and "we will not serve results this deep" are different facts,
and only the second is worth telling a user about.

## Database fallback

`SearchRecipes` and `SearchMeals` fall back to a database search when the index errors or returns an
empty first page, in case the index is behind or was never populated. The fallback drops the cursor,
so it restarts from the database's first page.

They do not fall back when the index rejects the cursor (`searchpagination.CursorRejected`), because
the database cannot take over mid-walk: it reads a cursor differently and would answer with the first
page of its own pagination rather than the page that was asked for. Those errors surface as the
statuses above instead.

An empty page part-way through a cursor walk is the end of the results, not a reason to fall back —
restarting the database from the top would serve the whole result set over again — so it comes back
as an empty page carrying the index's cursor.

The remaining index-backed searches (users, valid ingredients, ingredient states, instruments,
measurement units, preparations, vessels) have no fallback and surface index errors directly. Valid
ingredient groups always search the database, whatever `useSearchService` says.

## Helpers

`internal/searchpagination` is the only place that builds a search request or wraps search hits, so
no call site can forget the limit or the cursor:

| Helper                      | Role                                                                              |
|-----------------------------|-----------------------------------------------------------------------------------|
| `Search`                    | Runs one page against the index, taking size and resumption point from the filter |
| `RequestFromFilter`         | Builds the `textsearch.SearchRequest` a `QueryFilter` describes                   |
| `Resuming`                  | Reports whether a filter carries a cursor                                         |
| `NewResult`                 | Wraps a page of hits in the `QueryFilteredResult` the managers return             |
| `FilterForDatabaseFallback` | Returns the filter with any index cursor dropped                                  |
| `CursorRejected`            | Reports whether an error is the index declining the cursor it was given           |

## Related

- `internal/searchpagination/searchpagination.go` — the adapter described above
- `internal/services/errors/search_grpc_mapper.go` — `OutOfRange` and `InvalidArgument` mappings
- `internal/domain/mealplanning/managers/` — index-backed searches with a database fallback
- `internal/domain/identity/manager/user_data_manager.go` — `SearchForUsers`
- `internal/grpc/converters/query_filter.go` — filter and pagination conversion at the gRPC boundary
