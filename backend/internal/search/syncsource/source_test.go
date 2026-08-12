package syncsource

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exampleRow struct {
	id   string
	name string
}

type exampleDoc struct {
	Name string
}

func convertExample(x *exampleRow) *exampleDoc {
	return &exampleDoc{Name: x.name}
}

// sourceForTest builds a Source over an in-memory table, so the tests exercise this package's
// own contract — omission, ordering, error wrapping — rather than a repository's.
func sourceForTest(rows map[string]*exampleRow, order []string) *Source[exampleRow, exampleDoc] {
	fetch := func(_ context.Context, id string) (*exampleRow, error) {
		row, ok := rows[id]
		if !ok {
			return nil, sql.ErrNoRows
		}

		return row, nil
	}

	scan := func(_ context.Context, after string, limit int) ([]string, error) {
		var page []string
		for _, id := range order {
			if id > after {
				page = append(page, id)
			}
			if len(page) == limit {
				break
			}
		}

		return page, nil
	}

	return New("example", fetch, scan, convertExample)
}

func TestSource_Fetch(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		source := sourceForTest(map[string]*exampleRow{
			"a": {id: "a", name: "apple"},
			"b": {id: "b", name: "banana"},
		}, []string{"a", "b"})

		actual, err := source.Fetch(ctx, "a", "b")
		require.NoError(t, err)

		require.Len(t, actual, 2)
		assert.Equal(t, "a", actual[0].ID)
		assert.Equal(t, "apple", actual[0].Body.Name)
		assert.Equal(t, "b", actual[1].ID)
		assert.Equal(t, "banana", actual[1].Body.Name)
	})

	T.Run("omits rows that no longer exist", func(t *testing.T) {
		t.Parallel()

		// The Syncer relies on this: an omission is how it learns a row was deleted between
		// the event being written and the event being applied, and it removes the document
		// rather than leaving a tombstone. Reporting the miss as an error instead would retry
		// the event until it dead-lettered, with the stale document still in the index.
		ctx := t.Context()
		source := sourceForTest(map[string]*exampleRow{
			"a": {id: "a", name: "apple"},
		}, []string{"a"})

		actual, err := source.Fetch(ctx, "a", "vanished")
		require.NoError(t, err)

		require.Len(t, actual, 1)
		assert.Equal(t, "a", actual[0].ID)
	})

	T.Run("surfaces a real fetch error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := errors.New("database on fire")

		source := New("example",
			func(context.Context, string) (*exampleRow, error) { return nil, expected },
			func(context.Context, string, int) ([]string, error) { return nil, nil },
			convertExample,
		)

		actual, err := source.Fetch(ctx, "a")
		require.Error(t, err)
		require.ErrorIs(t, err, expected)
		assert.Nil(t, actual)
	})
}

func TestSource_Scan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		source := sourceForTest(map[string]*exampleRow{
			"a": {id: "a", name: "apple"},
			"b": {id: "b", name: "banana"},
			"c": {id: "c", name: "cherry"},
		}, []string{"a", "b", "c"})

		actual, err := source.Scan(ctx, "", 10)
		require.NoError(t, err)

		require.Len(t, actual, 3)
		assert.Equal(t, []string{"a", "b", "c"}, []string{actual[0].ID, actual[1].ID, actual[2].ID})
	})

	T.Run("resumes strictly after the cursor", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		source := sourceForTest(map[string]*exampleRow{
			"a": {id: "a", name: "apple"},
			"b": {id: "b", name: "banana"},
			"c": {id: "c", name: "cherry"},
		}, []string{"a", "b", "c"})

		actual, err := source.Scan(ctx, "a", 10)
		require.NoError(t, err)

		require.Len(t, actual, 2)
		assert.Equal(t, "b", actual[0].ID)
		assert.Equal(t, "c", actual[1].ID)
	})

	T.Run("returns documents in ascending byte order despite an unordered fetch", func(t *testing.T) {
		t.Parallel()

		// A reindex merges this stream against the index's own ordered stream to decide what
		// to prune, so an out-of-order page is not cosmetic — it is live documents being
		// deleted. Fetch is free to answer in any order, which is why Scan sorts.
		ctx := t.Context()

		source := New("example",
			func(_ context.Context, id string) (*exampleRow, error) {
				return &exampleRow{id: id, name: id}, nil
			},
			func(context.Context, string, int) ([]string, error) {
				return []string{"c", "a", "b"}, nil
			},
			convertExample,
		)

		actual, err := source.Scan(ctx, "", 10)
		require.NoError(t, err)

		require.Len(t, actual, 3)
		assert.Equal(t, []string{"a", "b", "c"}, []string{actual[0].ID, actual[1].ID, actual[2].ID})
	})

	T.Run("ends the walk on an empty page", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		source := sourceForTest(map[string]*exampleRow{}, nil)

		actual, err := source.Scan(ctx, "", 10)
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	T.Run("surfaces a scan error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := errors.New("database on fire")

		source := New("example",
			func(context.Context, string) (*exampleRow, error) { return nil, nil },
			func(context.Context, string, int) ([]string, error) { return nil, expected },
			convertExample,
		)

		actual, err := source.Scan(ctx, "", 10)
		require.Error(t, err)
		require.ErrorIs(t, err, expected)
		assert.Nil(t, actual)
	})
}
