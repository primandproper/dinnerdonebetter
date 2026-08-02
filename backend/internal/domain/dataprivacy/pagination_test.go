package dataprivacy

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRow struct {
	ID string
}

func buildTestPage(size int) *filtering.QueryFilteredResult[testRow] {
	data := make([]*testRow, 0, size)
	for range size {
		data = append(data, &testRow{ID: identifiers.New()})
	}

	return filtering.NewQueryFilteredResult(
		data,
		uint64(size),
		uint64(size),
		func(t *testRow) string { return t.ID },
		filtering.DefaultQueryFilter(),
	)
}

func TestCollectAllPages(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := buildTestPage(3)

		fetchCalls := 0
		actual, err := CollectAllPages(t.Context(), func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
			fetchCalls++
			assert.EqualValues(t, filtering.MaxQueryFilterLimit, *filter.MaxResponseSize)
			return expected, nil
		})

		require.NoError(t, err)
		assert.Equal(t, expected.Data, actual)
		assert.Equal(t, 1, fetchCalls, "a short page must terminate the loop")
	})

	T.Run("exhausts multiple pages", func(t *testing.T) {
		t.Parallel()

		fullPage := buildTestPage(filtering.MaxQueryFilterLimit)
		lastPage := buildTestPage(7)

		var cursorsSeen []*string
		fetchCalls := 0
		actual, err := CollectAllPages(t.Context(), func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
			fetchCalls++
			cursorsSeen = append(cursorsSeen, filter.Cursor)
			if fetchCalls == 1 {
				return fullPage, nil
			}
			return lastPage, nil
		})

		require.NoError(t, err)
		assert.Equal(t, 2, fetchCalls)
		assert.Len(t, actual, filtering.MaxQueryFilterLimit+7)

		require.Len(t, cursorsSeen, 2)
		assert.Nil(t, cursorsSeen[0], "first request must not carry a cursor")
		require.NotNil(t, cursorsSeen[1])
		assert.Equal(t, fullPage.Cursor, *cursorsSeen[1], "second request must resume from the first page's cursor")
	})

	T.Run("with error from fetch", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("fetch failed")

		actual, err := CollectAllPages(t.Context(), func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
			return nil, expectedErr
		})

		require.ErrorIs(t, err, expectedErr)
		assert.Nil(t, actual)
	})

	T.Run("halts when the cursor stops advancing", func(t *testing.T) {
		t.Parallel()

		stuckPage := buildTestPage(filtering.MaxQueryFilterLimit)

		fetchCalls := 0
		actual, err := CollectAllPages(t.Context(), func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
			fetchCalls++
			return stuckPage, nil
		})

		require.NoError(t, err)
		assert.Equal(t, 2, fetchCalls, "an unmoving cursor must terminate the loop")
		assert.Len(t, actual, 2*filtering.MaxQueryFilterLimit)
	})
}
