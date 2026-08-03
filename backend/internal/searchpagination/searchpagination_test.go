package searchpagination

import (
	"context"
	"testing"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
	"github.com/primandproper/platform-go/v9/search/text/elasticsearch"
	mocksearch "github.com/primandproper/platform-go/v9/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exampleHit struct {
	ID string
}

func TestSearch(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cursor := "cursor-from-a-previous-page"
		filter := filtering.DefaultQueryFilter()
		filter.MaxResponseSize = new(uint16(17))
		filter.Cursor = &cursor

		expected := &textsearch.SearchResults[exampleHit]{Hits: []*exampleHit{{ID: "hit"}}}
		index := &mocksearch.IndexMock[exampleHit]{
			SearchFunc: func(_ context.Context, req textsearch.SearchRequest) (*textsearch.SearchResults[exampleHit], error) {
				assert.Equal(t, "carrots", req.Query)
				assert.Equal(t, 17, req.Limit)
				assert.Equal(t, textsearch.Cursor(cursor), req.Cursor)
				return expected, nil
			},
		}

		actual, err := Search(ctx, index, "carrots", filter)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, index.SearchCalls(), 1)
	})

	T.Run("with error searching", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := platformerrors.New("blah")
		index := &mocksearch.IndexMock[exampleHit]{
			SearchFunc: func(_ context.Context, _ textsearch.SearchRequest) (*textsearch.SearchResults[exampleHit], error) {
				return nil, expected
			},
		}

		actual, err := Search(ctx, index, "carrots", nil)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, expected)
	})
}

func TestRequestFromFilter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cursor := "cursor-from-a-previous-page"
		filter := filtering.DefaultQueryFilter()
		filter.MaxResponseSize = new(uint16(17))
		filter.Cursor = &cursor

		actual := RequestFromFilter("carrots", filter)
		assert.Equal(t, "carrots", actual.Query)
		assert.Equal(t, 17, actual.Limit)
		assert.Equal(t, textsearch.Cursor(cursor), actual.Cursor)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		actual := RequestFromFilter("carrots", nil)
		assert.Equal(t, "carrots", actual.Query)
		assert.Zero(t, actual.Limit)
		assert.True(t, actual.Cursor.IsZero())
	})

	T.Run("with a cursor pointing at the empty string", func(t *testing.T) {
		t.Parallel()

		// A client that sends cursor="" is asking for the first page, not resuming from
		// a cursor the index would have to make sense of.
		cursor := ""
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		assert.True(t, RequestFromFilter("carrots", filter).Cursor.IsZero())
	})
}

func TestResuming(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cursor := "cursor-from-a-previous-page"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		assert.True(t, Resuming(filter))
	})

	T.Run("without a cursor", func(t *testing.T) {
		t.Parallel()

		assert.False(t, Resuming(filtering.DefaultQueryFilter()))
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		assert.False(t, Resuming(nil))
	})
}

func TestNewResult(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cursor := "cursor-from-a-previous-page"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		data := []*exampleHit{{ID: "first"}, {ID: "second"}}

		actual := NewResult(data, textsearch.Cursor("cursor-for-the-next-page"), filter)
		assert.Equal(t, data, actual.Data)
		assert.Equal(t, "cursor-for-the-next-page", actual.Cursor)
		assert.Equal(t, cursor, actual.PreviousCursor)
		assert.Equal(t, uint64(2), actual.FilteredCount)
	})

	T.Run("carries the index's cursor rather than the last row's ID", func(t *testing.T) {
		t.Parallel()

		// This is the whole point of the type: paging onward means handing the index
		// back the token it issued, which has nothing to do with any row we fetched.
		data := []*exampleHit{{ID: "the-last-row"}}

		actual := NewResult(data, textsearch.Cursor("cursor-for-the-next-page"), filtering.DefaultQueryFilter())
		assert.Equal(t, "cursor-for-the-next-page", actual.Cursor)
	})

	T.Run("with the result set exhausted", func(t *testing.T) {
		t.Parallel()

		// An empty next cursor is how the index says there is no further page, and an
		// empty Pagination.Cursor is how we say it.
		actual := NewResult([]*exampleHit{{ID: "first"}}, "", filtering.DefaultQueryFilter())
		assert.Empty(t, actual.Cursor)
	})

	T.Run("reports an unknown total rather than the page size", func(t *testing.T) {
		t.Parallel()

		// Reporting len(data) as the total is what told clients that a truncated page
		// was the entire result set.
		actual := NewResult([]*exampleHit{{ID: "first"}}, textsearch.Cursor("more"), filtering.DefaultQueryFilter())
		assert.Zero(t, actual.TotalCount)
	})
}

func TestFilterForDatabaseFallback(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cursor := "an-opaque-index-token"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		actual := FilterForDatabaseFallback(filter)
		assert.Nil(t, actual.Cursor)
		assert.Equal(t, filter.MaxResponseSize, actual.MaxResponseSize)

		// The caller's filter is left alone, since it is still describing the search.
		require.NotNil(t, filter.Cursor)
		assert.Equal(t, cursor, *filter.Cursor)
	})

	T.Run("without a cursor", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		assert.Same(t, filter, FilterForDatabaseFallback(filter))
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, FilterForDatabaseFallback(nil))
	})
}

func TestCursorRejected(T *testing.T) {
	T.Parallel()

	T.Run("with a cursor the index did not issue", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CursorRejected(textsearch.ErrInvalidCursor))
	})

	T.Run("with pagination past the result window", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CursorRejected(elasticsearch.ErrResultWindowExceeded))
	})

	T.Run("with a wrapped error", func(t *testing.T) {
		t.Parallel()

		assert.True(t, CursorRejected(platformerrors.Wrap(elasticsearch.ErrResultWindowExceeded, "searching")))
	})

	T.Run("with an unrelated error", func(t *testing.T) {
		t.Parallel()

		// A backend that is merely unreachable can be covered for by the database.
		assert.False(t, CursorRejected(platformerrors.New("connection refused")))
	})

	T.Run("with nil error", func(t *testing.T) {
		t.Parallel()

		assert.False(t, CursorRejected(nil))
	})
}
