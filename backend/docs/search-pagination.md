# Search Pagination

Search endpoints are cursor-paginated. A request carries the query, a `useSearchService` flag, and a
query filter; the response carries one page of results and a `pagination` block whose `cursor`
resumes the search. platform-go's `search/pagination` sits between the text search index and the
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

## What a page item carries

Recipe and meal list and search responses return **summaries**, not whole records:
`GetRecipesResponse`, `SearchForRecipesResponse`, `SearchForMealEligibleRecipesResponse` and
`SearchForRecipesWithInstrumentOwnershipResponse` carry `RecipeSummary`; `GetMealsResponse` and
`SearchForMealsResponse` carry `MealSummary`.

A `RecipeSummary` is the recipe's own columns and nothing that hangs off it — no steps, prep tasks,
media, or associated recipes. A `MealSummary` keeps its components, because a meal without them says
almost nothing, but each component carries a `RecipeSummary` rather than a whole `Recipe`.

This is what makes a max-limit page fit in a gRPC message. `server/grpc` bounds both directions at
`DefaultMaxMessageSize` (4 MiB), and a hydrated `Recipe` is much larger than it reads: each
`RecipeStepIngredient` embeds a whole `ValidIngredient` and `ValidMeasurementUnit`, and each
`RecipeStep` a whole `ValidPreparation`. Measured against the fakes, 250 hydrated recipes marshal to
4.67 MiB and 250 hydrated meals to 14.05 MiB; the same pages of summaries are 0.06 MiB and 0.23 MiB.
`Recipe.associated_recipes` is also self-recursive, so a hydrated recipe has no size bound at all and
no page size provably fits — which is why the fix is the projection rather than a larger bound.

`internal/services/mealplanning/grpc/converters/message_size_test.go` pins this: adding a repeated
field to either summary fails there rather than as a `ResourceExhausted` on a client we do not
operate.

**A client that needs the whole record fetches it by ID.** `GetRecipe` and `GetMeal` return the
hydrated object. The iOS meal-plan wizard is the worked example — search hands back a `MealSummary`,
and `CreateMealPlanViewModel.assignMeal(fromSearchResult:to:)` fetches the whole meal before the
option-selection step reads step ingredients.

Note that `GetRecipes` and the database-fallback searches never populated these nested collections in
the first place (`internal/repositories/postgres/mealplanning/recipes.go`), so on those paths the
summary makes an existing shape honest rather than removing anything.

`GetMealPlansForAccountResponse` is **not** covered: a `MealPlan` embeds events, which embed options,
which embed meals, so it has the same problem. It is tracked separately, because trimming it means
reworking the clients that read `mealPlan.events` straight off the list.

## Totals

Index-backed responses report `totalCount` as `0`, meaning unknown: the index reports whether another
page exists, not how many results there are in all. Database-backed responses report a real count,
computed by the search query itself. Page with the cursor, not with a count.

## Statuses a paging client can get

| Status            | Cause                                                                    | What to do                                               |
|-------------------|--------------------------------------------------------------------------|----------------------------------------------------------|
| `OutOfRange`      | Paged past Elasticsearch's `index.max_result_window` (10,000 by default) | Stop and narrow the query; paging further will not work. |
| `InvalidArgument` | A cursor the index did not issue                                         | Start over from the first page.                          |

platform-go maps both sentinels centrally, in `errors/grpc` and `errors/http`, so no application
mapper registers them. `MapToGRPC` consults the platform mapper ahead of every domain mapper, so
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

platform-go's `search/pagination` is the only place that builds a search request or wraps search
hits, so no call site can forget the limit or the cursor:

| Helper                      | Role                                                                              |
|-----------------------------|-----------------------------------------------------------------------------------|
| `Hydrated`                  | Runs the index-then-hydrate loop end to end and wraps the page                    |
| `Search`                    | Runs one page against the index, taking size and resumption point from the filter |
| `RequestFromFilter`         | Builds the `textsearch.SearchRequest` a `QueryFilter` describes                   |
| `Resuming`                  | Reports whether a filter carries a cursor                                         |
| `NewResult`                 | Wraps a page of hits in the `QueryFilteredResult` the managers return             |
| `FilterForDatabaseFallback` | Returns the filter with any index cursor dropped                                  |
| `CursorRejected`            | Reports whether an error is the index declining the cursor it was given           |

Every index-backed search goes through `Hydrated`. The two with a database fallback assemble the
page from it and then decide for themselves what an empty first page means; the other seven hand it
the store's read-many and return what it gives back. An empty page of hits never reaches the store,
so no `GetXWithIDs` is ever asked to read zero IDs.

## Related

- `search/pagination` in platform-go — the adapter described above
- `errors/grpc` and `errors/http` in platform-go — `OutOfRange` and `InvalidArgument` mappings
- `internal/domain/mealplanning/managers/` — index-backed searches with a database fallback
- `internal/domain/identity/manager/user_data_manager.go` — `SearchForUsers`
- `internal/grpc/converters/query_filter.go` — filter and pagination conversion at the gRPC boundary
- `proto/mealplanning/mealplanning_messages.proto` — `RecipeSummary` and `MealSummary`
