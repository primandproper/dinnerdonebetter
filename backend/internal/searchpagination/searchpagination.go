// Package searchpagination adapts the text search index's cursor pagination to the
// filtering.QueryFilter pagination the managers hand back to clients.
//
// The two are not the same model. A QueryFilter cursor is the last row's ID on the
// database path, and an index cursor is an opaque token that only the backend which
// issued it can read. They travel in the same field because a client treats both the
// same way — a string it hands back verbatim — and because which one it is follows
// from the useSearchService flag the client sent alongside it. What that sharing
// costs is spelled out on FilterForDatabaseFallback and CursorRejected.
package searchpagination

import (
	"context"
	"errors"

	"github.com/primandproper/platform-go/v10/filtering"
	textsearch "github.com/primandproper/platform-go/v10/search/text"
	"github.com/primandproper/platform-go/v10/search/text/elasticsearch"
)

// Search runs one page of query against the index, taking the page size and the
// resumption point from the caller's filter.
//
// It exists so that no call site can forget the limit: an unset limit is one page of
// the backend's choosing, and every call site here used to omit it, which is how
// search results came back truncated with no way to reach the rest.
func Search[T any](ctx context.Context, index textsearch.IndexSearcher[T], query string, filter *filtering.QueryFilter) (*textsearch.SearchResults[T], error) {
	return index.Search(ctx, RequestFromFilter(query, filter))
}

// RequestFromFilter builds the index request that a QueryFilter describes. A nil
// filter searches from the beginning at the index's default page size.
func RequestFromFilter(query string, filter *filtering.QueryFilter) textsearch.SearchRequest {
	req := textsearch.SearchRequest{Query: query}

	if filter == nil {
		return req
	}

	if filter.MaxResponseSize != nil {
		req.Limit = int(*filter.MaxResponseSize)
	}

	if Resuming(filter) {
		req.Cursor = textsearch.Cursor(*filter.Cursor)
	}

	return req
}

// Resuming reports whether filter carries a cursor, meaning the caller is part-way
// through a result set rather than asking for the first page.
func Resuming(filter *filtering.QueryFilter) bool {
	return filter != nil && filter.Cursor != nil && *filter.Cursor != ""
}

// NewResult wraps one page of hits, already resolved to domain objects, in the result
// type the managers return.
func NewResult[T any](data []*T, next textsearch.Cursor, filter *filtering.QueryFilter) *filtering.QueryFilteredResult[T] {
	// The total is left at zero, meaning unknown, as it is on the database search
	// path. The index reports whether another page exists but not how many results
	// there are in all, and reporting the page size as the total is what told clients
	// that a truncated page was the entire result set.
	//
	// The ID extractor is a no-op because the cursor it would derive gets overwritten
	// below: paging onward from here means handing the index its own token back, not
	// the last row's ID.
	result := filtering.NewQueryFilteredResult(data, uint64(len(data)), 0, func(*T) string { return "" }, filter)

	// A zero cursor is how the index says the result set is exhausted and an empty
	// Pagination.Cursor is how we say it, so the two line up without translation.
	// Note that this is the end of the results, not a short page — both backends can
	// return fewer hits than asked for and still have more.
	result.Cursor = string(next)

	return result
}

// FilterForDatabaseFallback returns filter with any index cursor dropped, for a
// caller falling back to the database after searching the index.
//
// The cursor cannot come along. The database reads a cursor as the last row's ID and
// would compare an opaque token against the ID column, matching an arbitrary slice
// of the table rather than failing outright. Dropping it restarts at the first page,
// which repeats results the caller has already seen but is at least the results they
// asked for.
func FilterForDatabaseFallback(filter *filtering.QueryFilter) *filtering.QueryFilter {
	if filter == nil || filter.Cursor == nil {
		return filter
	}

	clone := *filter
	clone.Cursor = nil

	return &clone
}

// CursorRejected reports whether err is the index declining the cursor it was given —
// either because it did not issue it, or because it will not page that deep.
//
// It marks the failures that a database fallback cannot cover for. A backend that is
// merely down can be stood in for; a rejected cursor cannot, because the database
// reads a cursor differently and would answer with the first page of its own
// pagination rather than the page that was asked for. These reach the client as
// distinct gRPC statuses instead, so it can stop paging or narrow the query.
func CursorRejected(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, textsearch.ErrInvalidCursor) || errors.Is(err, elasticsearch.ErrResultWindowExceeded)
}
