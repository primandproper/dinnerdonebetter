package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v5/filtering"

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
	assert.NoError(t, err)
	assert.Equal(t, exampleValidInstrument, created)

	validInstrument, err := dbc.GetValidInstrument(ctx, created.ID)
	exampleValidInstrument.CreatedAt = validInstrument.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, validInstrument, exampleValidInstrument)

	return validInstrument
}

func TestQuerier_Integration_ValidInstruments(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	exampleValidInstrument := fakes.BuildFakeValidInstrument()
	createdValidInstruments := []*types.ValidInstrument{}

	// create
	createdValidInstruments = append(createdValidInstruments, createValidInstrumentForTest(t, ctx, exampleValidInstrument, dbc))

	// update
	updatedValidInstrument := fakes.BuildFakeValidInstrument()
	updatedValidInstrument.ID = createdValidInstruments[0].ID
	assert.NoError(t, dbc.UpdateValidInstrument(ctx, updatedValidInstrument))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeValidInstrument()
		input.Name = fmt.Sprintf("%s %d", updatedValidInstrument.Name, i)
		createdValidInstruments = append(createdValidInstruments, createValidInstrumentForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validInstruments, err := dbc.GetValidInstruments(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, validInstruments.Data)
	assert.Equal(t, len(createdValidInstruments), len(validInstruments.Data))

	// fetch as list of IDs
	validInstrumentIDs := []string{}
	for _, validInstrument := range createdValidInstruments {
		validInstrumentIDs = append(validInstrumentIDs, validInstrument.ID)
	}

	byIDs, err := dbc.GetValidInstrumentsWithIDs(ctx, validInstrumentIDs)
	assert.NoError(t, err)
	assert.Equal(t, validInstruments.Data, byIDs)

	// fetch via name search
	byName, err := dbc.SearchForValidInstruments(ctx, updatedValidInstrument.Name, nil)
	assert.NoError(t, err)
	assert.Equal(t, validInstruments, byName)

	random, err := dbc.GetRandomValidInstrument(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, random)

	needingIndexing, err := dbc.GetValidInstrumentIDsThatNeedSearchIndexing(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, needingIndexing)

	// delete
	for _, validInstrument := range createdValidInstruments {
		assert.NoError(t, dbc.MarkValidInstrumentAsIndexed(ctx, validInstrument.ID))
		assert.NoError(t, dbc.ArchiveValidInstrument(ctx, validInstrument.ID))

		var exists bool
		exists, err = dbc.ValidInstrumentExists(ctx, validInstrument.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidInstrument
		y, err = dbc.GetValidInstrument(ctx, validInstrument.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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

func TestQuerier_MarkValidInstrumentAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkValidInstrumentAsIndexed(ctx, ""))
	})
}

func TestQuerier_Integration_ValidInstruments_CursorBasedPagination(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

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
