package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidInstrumentForTest(t *testing.T, ctx context.Context, exampleValidInstrument *types.ValidInstrument, dbc *repository) *types.ValidInstrument {
	t.Helper()

	// create
	if exampleValidInstrument == nil {
		exampleValidInstrument = fakes.BuildFakeValidInstrument()
	}
	dbInput := converters.ConvertValidInstrumentToValidInstrumentDatabaseCreationInput(exampleValidInstrument)

	created, err := dbc.CreateValidInstrument(ctx, dbInput)
	exampleValidInstrument.CreatedAt = created.CreatedAt
	require.NoError(t, err)
	assert.Equal(t, exampleValidInstrument, created)

	validInstrument, err := dbc.GetValidInstrument(ctx, created.ID)
	exampleValidInstrument.CreatedAt = validInstrument.CreatedAt

	require.NoError(t, err)
	assert.Equal(t, validInstrument, exampleValidInstrument)

	return validInstrument
}

func TestQuerier_Integration_ValidInstruments(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidInstrument := fakes.BuildFakeValidInstrument()
	createdValidInstruments := []*types.ValidInstrument{}

	// create
	createdValidInstruments = append(createdValidInstruments, createValidInstrumentForTest(t, ctx, exampleValidInstrument, dbc))

	// update
	updatedValidInstrument := fakes.BuildFakeValidInstrument()
	updatedValidInstrument.ID = createdValidInstruments[0].ID
	require.NoError(t, dbc.UpdateValidInstrument(ctx, updatedValidInstrument))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeValidInstrument()
		input.Name = fmt.Sprintf("%s %d", updatedValidInstrument.Name, i)
		createdValidInstruments = append(createdValidInstruments, createValidInstrumentForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validInstruments, err := dbc.GetValidInstruments(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, validInstruments.Data)
	assert.Len(t, validInstruments.Data, len(createdValidInstruments))

	// fetch as list of IDs
	validInstrumentIDs := []string{}
	for _, validInstrument := range createdValidInstruments {
		validInstrumentIDs = append(validInstrumentIDs, validInstrument.ID)
	}

	byIDs, err := dbc.GetValidInstrumentsWithIDs(ctx, validInstrumentIDs)
	require.NoError(t, err)
	assert.Equal(t, validInstruments.Data, byIDs)

	// fetch via name search
	byName, err := dbc.SearchForValidInstruments(ctx, updatedValidInstrument.Name, nil)
	require.NoError(t, err)
	assert.Equal(t, validInstruments, byName)

	random, err := dbc.GetRandomValidInstrument(ctx)
	require.NoError(t, err)
	assert.NotNil(t, random)

	needingIndexing, err := dbc.ScanValidInstrumentIDsForReindex(ctx, "", 100)
	require.NoError(t, err)
	assert.NotNil(t, needingIndexing)

	// delete
	for _, validInstrument := range createdValidInstruments {
		assert.NoError(t, dbc.MarkValidInstrumentsAsIndexed(ctx, []string{validInstrument.ID}))
		assert.NoError(t, dbc.ArchiveValidInstrument(ctx, validInstrument.ID))

		var exists bool
		exists, err = dbc.ValidInstrumentExists(ctx, validInstrument.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidInstrument
		y, err = dbc.GetValidInstrument(ctx, validInstrument.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidInstrumentExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid instrument ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidInstrumentExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid instrument ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidInstrument(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid instrument ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.SearchForValidInstruments(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidInstrument(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidInstrument(ctx, nil))
	})
}

func TestQuerier_ArchiveValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid instrument ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidInstrument(ctx, ""))
	})
}

func TestQuerier_MarkValidInstrumentsAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with no ids", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		// The client is inert, so a nil error is the assertion that nothing was
		// executed: an empty flush must not reach the database at all.
		assert.NoError(t, c.MarkValidInstrumentsAsIndexed(ctx, nil))
	})

	T.Run("stamps every id in one flush", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dbc, _ := buildDatabaseClientForTest(t)

		first := createValidInstrumentForTest(t, ctx, fakes.BuildFakeValidInstrument(), dbc)
		second := createValidInstrumentForTest(t, ctx, fakes.BuildFakeValidInstrument(), dbc)
		untouched := createValidInstrumentForTest(t, ctx, fakes.BuildFakeValidInstrument(), dbc)

		// A freshly written row has never been indexed, which is what makes the stamp
		// below observable rather than a value that was already there.
		assert.Nil(t, lastIndexedAtForTest(t, ctx, dbc, first.ID))

		require.NoError(t, dbc.MarkValidInstrumentsAsIndexed(ctx, []string{first.ID, second.ID}))

		assert.NotNil(t, lastIndexedAtForTest(t, ctx, dbc, first.ID))
		assert.NotNil(t, lastIndexedAtForTest(t, ctx, dbc, second.ID))
		assert.Nil(t, lastIndexedAtForTest(t, ctx, dbc, untouched.ID))
	})
}

// lastIndexedAtForTest reads the column directly, because nothing else can: last_indexed_at is
// database-owned, so it is on no domain type and no generated read returns it.
func lastIndexedAtForTest(t *testing.T, ctx context.Context, dbc *repository, id string) *time.Time {
	t.Helper()

	var stampedAt *time.Time
	require.NoError(t, dbc.readDB.QueryRowContext(ctx, "SELECT last_indexed_at FROM valid_instruments WHERE id = $1", id).Scan(&stampedAt))

	return stampedAt
}

func TestQuerier_Integration_ValidInstruments_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidInstrument]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid instrument",
		CreateItem: func(ctx context.Context, i int) *types.ValidInstrument {
			validInstrument := fakes.BuildFakeValidInstrument()
			validInstrument.Name = fmt.Sprintf("Valid Instrument %02d", i)
			return createValidInstrumentForTest(t, ctx, validInstrument, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
			return dbc.GetValidInstruments(ctx, filter)
		},
		GetID: func(validInstrument *types.ValidInstrument) string {
			return validInstrument.ID
		},
		CleanupItem: func(ctx context.Context, validInstrument *types.ValidInstrument) error {
			return dbc.ArchiveValidInstrument(ctx, validInstrument.ID)
		},
	})
}

func TestQuerier_Integration_ValidInstruments_IncludeArchived(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	live := createValidInstrumentForTest(t, ctx, nil, dbc)
	archived := createValidInstrumentForTest(t, ctx, nil, dbc)
	require.NoError(t, dbc.ArchiveValidInstrument(ctx, archived.ID))

	idsOf := func(instruments []*types.ValidInstrument) []string {
		ids := []string{}
		for _, instrument := range instruments {
			ids = append(ids, instrument.ID)
		}

		return ids
	}

	t.Run("absent leaves archived rows out", func(t *testing.T) {
		t.Parallel()

		results, err := dbc.GetValidInstruments(ctx, filtering.DefaultQueryFilter())
		require.NoError(t, err)

		assert.Equal(t, []string{live.ID}, idsOf(results.Data))
		assert.Equal(t, uint64(1), results.FilteredCount)
		assert.Equal(t, uint64(1), results.TotalCount)
	})

	t.Run("set admits them, and both counts agree", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		filter.IncludeArchived = pointer.To(true)

		results, err := dbc.GetValidInstruments(ctx, filter)
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{live.ID, archived.ID}, idsOf(results.Data))
		assert.Equal(t, uint64(2), results.FilteredCount)
		// total_count applies the same toggle, so the subset can never exceed the set.
		assert.Equal(t, uint64(2), results.TotalCount)
	})
}
