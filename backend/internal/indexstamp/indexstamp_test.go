package indexstamp

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/platform-go/v11/batching"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errFromIndex = errors.New("index said no")

// recordingIndex is a textsearch.IndexManager that remembers what it was asked to do, and
// fails the calls a test tells it to fail.
type recordingIndex struct {
	indexErr  error
	deleteErr error
	wipeErr   error

	indexed []string
	deleted []string

	mu    sync.Mutex
	wiped int
}

func (r *recordingIndex) Index(_ context.Context, id string, _ any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.indexErr != nil {
		return r.indexErr
	}

	r.indexed = append(r.indexed, id)

	return nil
}

func (r *recordingIndex) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.deleteErr != nil {
		return r.deleteErr
	}

	r.deleted = append(r.deleted, id)

	return nil
}

func (r *recordingIndex) Wipe(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wipeErr != nil {
		return r.wipeErr
	}

	r.wiped++

	return nil
}

// recordingMark is a MarkFunc that remembers every ID it stamped, and optionally refuses one.
type recordingMark struct {
	refuse  string
	stamped []string

	mu sync.Mutex
}

func (r *recordingMark) mark(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id == r.refuse {
		return errors.New("cannot stamp " + id)
	}

	r.stamped = append(r.stamped, id)

	return nil
}

func (r *recordingMark) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.stamped)
}

// buildStamperForTest builds a Stamper whose buffer only ever flushes when it is closed, so a
// test observes one flush at a moment it chose rather than one the interval chose for it.
func buildStamperForTest(t *testing.T, index *recordingIndex, mark *recordingMark) *Stamper {
	t.Helper()

	stamper, err := New(index, mark.mark, batching.WithFlushInterval(time.Hour))
	require.NoError(t, err)
	require.NotNil(t, stamper)

	return stamper
}

func TestNew(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		stamper, err := New(&recordingIndex{}, (&recordingMark{}).mark)
		require.NoError(t, err)
		require.NotNil(t, stamper)

		assert.NoError(t, stamper.Close(t.Context()))
	})

	T.Run("with nil index", func(t *testing.T) {
		t.Parallel()

		stamper, err := New(nil, (&recordingMark{}).mark)
		assert.Nil(t, stamper)
		assert.ErrorIs(t, err, ErrNilIndex)
	})

	T.Run("with nil mark func", func(t *testing.T) {
		t.Parallel()

		stamper, err := New(&recordingIndex{}, nil)
		assert.Nil(t, stamper)
		assert.ErrorIs(t, err, ErrNilMarkFunc)
	})
}

func TestStamper_Index(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.NoError(t, stamper.Index(ctx, "abc", "whatever"))

		// Nothing is stamped until the buffer flushes, which is the whole point of it.
		assert.Empty(t, mark.all())

		require.NoError(t, stamper.Close(ctx))

		assert.Equal(t, []string{"abc"}, index.indexed)
		assert.Equal(t, []string{"abc"}, mark.all())
	})

	T.Run("collapses repeats of one document into one stamp", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		for range 10 {
			require.NoError(t, stamper.Index(ctx, "abc", "whatever"))
		}

		require.NoError(t, stamper.Close(ctx))

		assert.Len(t, index.indexed, 10)
		assert.Equal(t, []string{"abc"}, mark.all())
	})

	T.Run("stamps every distinct document in one flush", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.NoError(t, stamper.Index(ctx, "ccc", "whatever"))
		require.NoError(t, stamper.Index(ctx, "aaa", "whatever"))
		require.NoError(t, stamper.Index(ctx, "bbb", "whatever"))

		require.NoError(t, stamper.Close(ctx))

		// In ID order, because that is the lock order the buffer was given.
		assert.Equal(t, []string{"aaa", "bbb", "ccc"}, mark.all())
	})

	T.Run("with failing index", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{indexErr: errFromIndex}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.ErrorIs(t, stamper.Index(ctx, "abc", "whatever"), errFromIndex)

		require.NoError(t, stamper.Close(ctx))

		// A document the index refused was not indexed, so its row says what it said before.
		assert.Empty(t, mark.all())
	})

	T.Run("with one row that cannot be stamped", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{refuse: "bbb"}
		stamper := buildStamperForTest(t, index, mark)

		require.NoError(t, stamper.Index(ctx, "aaa", "whatever"))
		require.NoError(t, stamper.Index(ctx, "bbb", "whatever"))
		require.NoError(t, stamper.Index(ctx, "ccc", "whatever"))

		require.Error(t, stamper.Close(ctx))

		// The other two are stamped regardless: one unstampable row says nothing about them.
		assert.Equal(t, []string{"aaa", "ccc"}, mark.all())
	})
}

func TestStamper_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.NoError(t, stamper.Delete(ctx, "abc"))
		require.NoError(t, stamper.Close(ctx))

		assert.Equal(t, []string{"abc"}, index.deleted)
		assert.Empty(t, mark.all())
	})

	T.Run("with failing index", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{deleteErr: errFromIndex}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.ErrorIs(t, stamper.Delete(ctx, "abc"), errFromIndex)
		require.NoError(t, stamper.Close(ctx))

		assert.Empty(t, mark.all())
	})
}

func TestStamper_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.NoError(t, stamper.Wipe(ctx))
		require.NoError(t, stamper.Close(ctx))

		assert.Equal(t, 1, index.wiped)
		assert.Empty(t, mark.all())
	})

	T.Run("with failing index", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{wipeErr: errFromIndex}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.ErrorIs(t, stamper.Wipe(ctx), errFromIndex)
		require.NoError(t, stamper.Close(ctx))

		assert.Empty(t, mark.all())
	})
}

func TestStamper_Close(T *testing.T) {
	T.Parallel()

	T.Run("is safe to call more than once", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}
		stamper := buildStamperForTest(t, index, mark)

		require.NoError(t, stamper.Index(ctx, "abc", "whatever"))
		require.NoError(t, stamper.Close(ctx))
		require.NoError(t, stamper.Close(ctx))

		assert.Equal(t, []string{"abc"}, mark.all())
	})

	T.Run("flushes on the interval without a close", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		index, mark := &recordingIndex{}, &recordingMark{}

		stamper, err := New(index, mark.mark, batching.WithFlushInterval(10*time.Millisecond))
		require.NoError(t, err)

		t.Cleanup(func() { assert.NoError(t, stamper.Close(context.WithoutCancel(ctx))) })

		require.NoError(t, stamper.Index(ctx, "abc", "whatever"))

		assert.Eventually(t, func() bool {
			return slices.Equal([]string{"abc"}, mark.all())
		}, time.Second, 5*time.Millisecond)
	})
}
