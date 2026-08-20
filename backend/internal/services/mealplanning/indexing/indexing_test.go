package indexing

import (
	"context"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmocks "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	searchsync "github.com/primandproper/platform-go/v12/search/sync"
	syncsource "github.com/primandproper/platform-go/v12/search/sync/source"
	textsearchmock "github.com/primandproper/platform-go/v12/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingStamper is a searchsync.Stamper that keeps what it was handed.
//
// It is not a batching.Buffer: what is under test here is which ids reach a stamper at all,
// and a real buffer would answer that on its own schedule.
type recordingStamper struct {
	_ struct{} `json:"-"`

	stamped []string
}

func (s *recordingStamper) Add(ids ...string) {
	s.stamped = append(s.stamped, ids...)
}

func TestSyncer_Stamping(T *testing.T) {
	T.Parallel()

	T.Run("stamps the document an upsert indexed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := fakes.BuildFakeValidInstrument()

		repo := &mealplanningmocks.RepositoryMock{
			GetValidInstrumentFunc: func(_ context.Context, id string) (*types.ValidInstrument, error) {
				assert.Equal(t, expected.ID, id)

				return expected, nil
			},
		}

		stamper := &recordingStamper{}
		index := &textsearchmock.IndexMock[ValidInstrumentSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		src, err := NewValidInstrumentSource(repo)
		require.NoError(t, err)

		syncer, err := syncsource.NewSyncer(src, index,
			syncsource.WithSyncerOptions(searchsync.WithSyncerStamper(stamper)))
		require.NoError(t, err)

		require.NoError(t, syncer.Apply(ctx, searchsync.Event{
			OccurredAt: time.Now(),
			DocumentID: expected.ID,
			Op:         searchsync.OpUpsert,
		}))

		assert.Equal(t, []string{expected.ID}, stamper.stamped)
	})

	T.Run("stamps nothing for a delete", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleID := fakes.BuildFakeValidInstrument().ID

		repo := &mealplanningmocks.RepositoryMock{}

		stamper := &recordingStamper{}
		index := &textsearchmock.IndexMock[ValidInstrumentSearchSubset]{
			DeleteFunc: func(_ context.Context, _ string) error { return nil },
		}

		src, err := NewValidInstrumentSource(repo)
		require.NoError(t, err)

		syncer, err := syncsource.NewSyncer(src, index,
			syncsource.WithSyncerOptions(searchsync.WithSyncerStamper(stamper)))
		require.NoError(t, err)

		require.NoError(t, syncer.Apply(ctx, searchsync.Event{
			OccurredAt: time.Now(),
			DocumentID: exampleID,
			Op:         searchsync.OpDelete,
		}))

		// There is no row left to describe as current, so a delete leaves last_indexed_at
		// alone rather than stamping a document the index no longer holds.
		assert.Empty(t, stamper.stamped)
	})
}
