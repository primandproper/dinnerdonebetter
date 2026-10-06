package searchindexes

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	syncsource "github.com/primandproper/platform-go/v15/searchsync/source"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	textsearchmock "github.com/primandproper/primitives-go/v2/search/text/mock"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type row struct {
	ID string
}

type document struct {
	ID string `json:"id"`
}

// index is one in-memory index: a source over a fixed set of rows, the text index it feeds, and
// the stamps its Syncer wrote.
type index struct {
	source  *syncsource.Source[row, document]
	target  *textsearchmock.IndexMock[document]
	ids     []string
	stamped []string
	mu      sync.Mutex
}

func buildIndex(t *testing.T) *index {
	t.Helper()

	idx := &index{ids: []string{fake.BuildFakeID(), fake.BuildFakeID()}}

	var err error
	idx.source, err = syncsource.New(fake.BuildFakeID(),
		func(_ context.Context, id string) (*row, error) { return &row{ID: id}, nil },
		func(_ context.Context, after string, _ int) ([]string, error) {
			if after != "" {
				return nil, nil
			}

			return idx.ids, nil
		},
		func(r *row) *document { return &document{ID: r.ID} },
	)
	require.NoError(t, err)

	idx.target = &textsearchmock.IndexMock[document]{
		IndexFunc: func(context.Context, string, any) error { return nil },
	}

	return idx
}

func (idx *index) stamp(_ context.Context, ids []string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.stamped = append(idx.stamped, ids...)

	return nil
}

func (idx *index) registrar() Registrar {
	return func(_ do.Injector, registry *searchsync.Registry) error {
		return RegisterTextIndex(registry, idx.source, idx.target, idx.stamp)
	}
}

func buildInjector() do.Injector {
	i := do.New()
	do.ProvideValue[logging.Logger](i, loggingnoop.NewLogger())
	do.ProvideValue[tracing.Provider](i, tracingnoop.NewTracerProvider())
	do.ProvideValue[metrics.Provider](i, metricsnoop.NewMetricsProvider())

	return i
}

func TestRegister(T *testing.T) {
	T.Parallel()

	T.Run("one pool per index, on the index's own topic, from every registrar", func(t *testing.T) {
		t.Parallel()

		first, second := buildIndex(t), buildIndex(t)
		i := buildInjector()
		Register(t.Context(), i, first.registrar(), second.registrar())

		registry := do.MustInvoke[*Registry](i)
		t.Cleanup(func() { assert.NoError(t, registry.Shutdown(context.WithoutCancel(t.Context()))) })

		var topics []string
		for _, spec := range registry.PoolSpecs() {
			topics = append(topics, spec.Topic)
		}

		assert.ElementsMatch(t, []string{first.source.Name(), second.source.Name()}, topics)
		assert.ElementsMatch(t, []string{first.source.Name(), second.source.Name()}, registry.Names())
	})

	T.Run("an applied event reaches the index, and Shutdown flushes its stamp", func(t *testing.T) {
		t.Parallel()

		idx := buildIndex(t)
		i := buildInjector()
		Register(t.Context(), i, idx.registrar())

		registry := do.MustInvoke[*Registry](i)
		specs := registry.PoolSpecs()
		require.Len(t, specs, 1)

		payload, err := json.Marshal(searchsync.NewEvent(searchsync.OpUpsert, idx.ids[0]))
		require.NoError(t, err)

		require.NoError(t, specs[0].Handler(t.Context(), payload))
		require.Len(t, idx.target.IndexCalls(), 1)
		assert.Equal(t, idx.ids[0], idx.target.IndexCalls()[0].ID)

		// The stamp is buffered behind the Syncer; Shutdown is what the container calls, and
		// it is what writes the buffer out.
		require.NoError(t, registry.Shutdown(t.Context()))

		idx.mu.Lock()
		defer idx.mu.Unlock()
		assert.Equal(t, []string{idx.ids[0]}, idx.stamped)
	})

	T.Run("a rebuild walks every index and stamps nothing", func(t *testing.T) {
		t.Parallel()

		idx := buildIndex(t)
		i := buildInjector()
		Register(t.Context(), i, idx.registrar())

		registry := do.MustInvoke[*Registry](i)

		results, err := registry.ReindexAll(t.Context())
		require.NoError(t, err)
		require.Contains(t, results, idx.source.Name())
		assert.Len(t, idx.target.IndexCalls(), len(idx.ids))

		// A rebuild writes every document there is; stamping it would make the column read as
		// when the last rebuild ran.
		require.NoError(t, registry.Shutdown(t.Context()))

		idx.mu.Lock()
		defer idx.mu.Unlock()
		assert.Empty(t, idx.stamped)
	})

	T.Run("a registrar that fails fails the Registry", func(t *testing.T) {
		t.Parallel()

		expected := errors.New(fake.BuildFakeString())
		i := buildInjector()
		Register(t.Context(), i, buildIndex(t).registrar(), func(do.Injector, *searchsync.Registry) error { return expected })

		_, err := do.Invoke[*Registry](i)
		require.ErrorIs(t, err, expected)
	})

	T.Run("two indexes under one name are refused", func(t *testing.T) {
		t.Parallel()

		idx := buildIndex(t)
		i := buildInjector()
		first, second := idx.registrar(), idx.registrar()
		Register(t.Context(), i, first, second)

		_, err := do.Invoke[*Registry](i)
		require.ErrorIs(t, err, searchsync.ErrDuplicateIndex)
	})
}

func TestRegisterTextIndex(T *testing.T) {
	T.Parallel()

	T.Run("with a nil source", func(t *testing.T) {
		t.Parallel()

		err := RegisterTextIndex[row, document](searchsync.NewRegistry(), nil, &textsearchmock.IndexMock[document]{}, nil)
		require.ErrorIs(t, err, searchsync.ErrNilSource)
	})
}
