# Search pagination

How a client pages through search results, and why the cursor it gets back means two different
things depending on one flag it sent.

## The problem this solves

platform-go v9's `search/text` takes a `SearchRequest` and returns `SearchResults[T]` with an
opaque `NextCursor`. We adopted the signature during the v9 upgrade but not the feature: every
call site passed only `Query`, so every search returned one page of the backend's choosing, and
no call site read `NextCursor`.

The visible symptom was worse than truncation. The managers wrapped the hits in a
`QueryFilteredResult` whose total was `len(data)`, so the pagination metadata handed back to
clients asserted that the truncated page *was* the whole result set. Searching for a common
ingredient returned a page of results and a total agreeing there were only that many.

`internal/searchpagination` is the adapter that fixes it. Every index-backed search goes through
it; see the package doc for the reasoning behind each piece.

## The cursor is the same field either way

`filtering.QueryFilter.Cursor` carries both kinds of cursor:

| `useSearchService` | What the cursor is | Who reads it |
|---|---|---|
| `false` | The last row's ID | The SQL, as `WHERE id > cursor` |
| `true` | An opaque token from the index | Only the backend that issued it |

A client pages the same way in both cases — send back the `cursor` from the last response's
`pagination`, along with the **same `useSearchService` value** — and stops when `cursor` comes
back empty.

Stopping on an empty cursor is the contract, not on a short page. Both Elasticsearch and Algolia
can return fewer hits than asked for and still have more; the cursor is the only thing that says
whether another page exists.

Changing `useSearchService` part-way through a walk invalidates the cursor. Cursors are tagged
with the backend that issued them, so the index refuses a database cursor outright
(`InvalidArgument`, below). The other direction cannot be caught — the database would read an
opaque token as a row ID and answer with an arbitrary slice of the table — which is why
`FilterForDatabaseFallback` strips the cursor whenever a manager falls back from index to
database.

## Totals

`totalCount` is `0` on search responses, meaning unknown. The index reports whether another page
exists but not how many results there are in all, and the database search path has always
reported `0` as well. Page with the cursor, not with a count.

## Statuses a paging client can get

| Status | Cause | What to do |
|---|---|---|
| `OutOfRange` | Paged past Elasticsearch's `index.max_result_window` (10,000 by default) | Stop; narrow the query. Paging further will not work. |
| `InvalidArgument` | A cursor the index did not issue | Start over from the first page. |

Both are mapped in `internal/services/errors/search_grpc_mapper.go`. `OutOfRange` is deliberately
not reported as an empty last page: "there are no more results" and "we will not serve results
this deep" are different facts, and only the second one is worth telling a user about.

## When the index errors

`SearchRecipes` and `SearchMeals` fall back to a database search when the index fails or returns
an empty first page. They do **not** fall back when the index rejects the cursor, because the
database cannot take over mid-walk: it would restart at its own first page under a cursor it
reads differently. The other search managers have no fallback and surface the error directly.

An empty page part-way through a walk is the end of the results, not a reason to fall back —
restarting the database from the top would serve the whole result set over again.

## Page sizes

`filter.MaxResponseSize` becomes `SearchRequest.Limit`. Unset means `DefaultSearchLimit` (25).
Both backends cap a single page at 200, so a request for more comes back short with a cursor
still set, which the rule above already handles.
