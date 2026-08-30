package dataprivacy

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRow struct {
	ID        string
	AccountID string
}

func buildTestPage(accountID string, size int) *filtering.QueryFilteredResult[testRow] {
	data := make([]*testRow, 0, size)
	for range size {
		data = append(data, &testRow{ID: identifiers.New(), AccountID: accountID})
	}

	return filtering.NewQueryFilteredResult(
		data,
		uint64(size),
		uint64(size),
		func(t *testRow) string { return t.ID },
		filtering.DefaultQueryFilter(),
	)
}

func TestCollectAcrossAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		firstAccountID, secondAccountID := identifiers.New(), identifiers.New()

		var accountsAsked []string
		actual, err := CollectAcrossAccounts(t.Context(), []string{firstAccountID, secondAccountID},
			func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
				accountsAsked = append(accountsAsked, accountID)

				return buildTestPage(accountID, 2), nil
			})

		require.NoError(t, err)
		assert.Equal(t, []string{firstAccountID, secondAccountID}, accountsAsked)

		require.Len(t, actual, 4)
		assert.Equal(t, firstAccountID, actual[0].AccountID)
		assert.Equal(t, firstAccountID, actual[1].AccountID)
		assert.Equal(t, secondAccountID, actual[2].AccountID)
		assert.Equal(t, secondAccountID, actual[3].AccountID)
	})

	T.Run("with no accounts", func(t *testing.T) {
		t.Parallel()

		actual, err := CollectAcrossAccounts(t.Context(), nil,
			func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
				t.Error("fetch must not be called when the subject appears in no accounts")

				return nil, nil
			})

		require.NoError(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error fetching for an account", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("fetch failed")
		goodAccountID, badAccountID := identifiers.New(), identifiers.New()

		actual, err := CollectAcrossAccounts(t.Context(), []string{goodAccountID, badAccountID},
			func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[testRow], error) {
				if accountID == badAccountID {
					return nil, expectedErr
				}

				return buildTestPage(accountID, 2), nil
			})

		// What was collected for the accounts before the failure is discarded rather
		// than returned as though it were the whole answer, because a subject access
		// request that silently omits one account's rows is a compliance defect that
		// looks exactly like a correct export.
		require.ErrorIs(t, err, expectedErr)
		assert.Nil(t, actual)
	})
}
