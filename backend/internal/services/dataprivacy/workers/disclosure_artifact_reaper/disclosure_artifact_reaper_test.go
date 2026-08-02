package disclosureartifactreaper

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/reportartifacts"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildReaperForTest(t *testing.T, repo *dataprivacymock.RepositoryMock, artifacts *reportartifacts.StoreMock) *Worker {
	t.Helper()

	w, err := NewDisclosureArtifactReaper(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		repo,
		artifacts,
	)
	require.NoError(t, err)

	return w
}

// expiredDisclosure builds a completed disclosure that has an artifact behind it.
func expiredDisclosure() *dataprivacy.UserDataDisclosure {
	return &dataprivacy.UserDataDisclosure{
		ID:       identifiers.New(),
		ReportID: identifiers.New(),
		Status:   dataprivacy.UserDataDisclosureStatusCompleted,
	}
}

// onceThenEmpty returns the given disclosures on the first call and nothing thereafter, standing
// in for a sweep that reaps everything it found.
func onceThenEmpty(disclosures ...*dataprivacy.UserDataDisclosure) func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
	called := false

	return func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
		if called {
			return nil, nil
		}
		called = true

		return disclosures, nil
	}
}

func TestWorker_Work(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		disclosure := expiredDisclosure()

		var order []string
		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(_ context.Context, reportID string) error {
				assert.Equal(t, disclosure.ReportID, reportID)
				order = append(order, "delete")
				return nil
			},
		}
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: onceThenEmpty(disclosure),
			MarkUserDataDisclosureExpiredFunc: func(_ context.Context, disclosureID string) error {
				assert.Equal(t, disclosure.ID, disclosureID)
				order = append(order, "mark")
				return nil
			},
		}

		w := buildReaperForTest(t, repo, artifacts)

		require.NoError(t, w.Work(ctx))

		// The object must be gone before the row claims it is, or nothing ever comes back for it.
		assert.Equal(t, []string{"delete", "mark"}, order)
	})

	T.Run("with nothing to do", func(t *testing.T) {
		t.Parallel()

		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(_ context.Context, _ string) error {
				t.Error("Delete should not be called when nothing has expired")
				return nil
			},
		}
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
				return nil, nil
			},
		}

		w := buildReaperForTest(t, repo, artifacts)

		require.NoError(t, w.Work(t.Context()))
		assert.Len(t, repo.GetExpiredUserDataDisclosuresCalls(), 1)
	})

	T.Run("with a disclosure that never produced a report", func(t *testing.T) {
		t.Parallel()

		// A pending or failed disclosure has no object to destroy, but its expiry should still
		// be recorded so the sweep stops finding it.
		disclosure := &dataprivacy.UserDataDisclosure{
			ID:     identifiers.New(),
			Status: dataprivacy.UserDataDisclosureStatusFailed,
		}

		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(_ context.Context, _ string) error {
				t.Error("Delete should not be called for a disclosure with no report")
				return nil
			},
		}
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: onceThenEmpty(disclosure),
			MarkUserDataDisclosureExpiredFunc: func(context.Context, string) error { return nil },
		}

		w := buildReaperForTest(t, repo, artifacts)

		require.NoError(t, w.Work(t.Context()))
		assert.Len(t, repo.MarkUserDataDisclosureExpiredCalls(), 1)
	})

	T.Run("leaves the row alone when the object cannot be destroyed", func(t *testing.T) {
		t.Parallel()

		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(context.Context, string) error { return platformerrors.New("blah") },
		}
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: onceThenEmpty(expiredDisclosure()),
			MarkUserDataDisclosureExpiredFunc: func(context.Context, string) error {
				t.Error("a disclosure whose artifact survived must not be marked expired")
				return nil
			},
		}

		w := buildReaperForTest(t, repo, artifacts)

		assert.Error(t, w.Work(t.Context()))
	})

	T.Run("keeps going past one failure", func(t *testing.T) {
		t.Parallel()

		doomed, healthy := expiredDisclosure(), expiredDisclosure()

		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(_ context.Context, reportID string) error {
				if reportID == doomed.ReportID {
					return platformerrors.New("blah")
				}
				return nil
			},
		}
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: onceThenEmpty(doomed, healthy),
			MarkUserDataDisclosureExpiredFunc: func(context.Context, string) error { return nil },
		}

		w := buildReaperForTest(t, repo, artifacts)

		assert.Error(t, w.Work(t.Context()))
		require.Len(t, repo.MarkUserDataDisclosureExpiredCalls(), 1)
		assert.Equal(t, healthy.ID, repo.MarkUserDataDisclosureExpiredCalls()[0].DisclosureID)
	})

	T.Run("with error fetching expired disclosures", func(t *testing.T) {
		t.Parallel()

		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
				return nil, platformerrors.New("blah")
			},
		}

		w := buildReaperForTest(t, repo, &reportartifacts.StoreMock{})

		assert.Error(t, w.Work(t.Context()))
	})

	T.Run("stops instead of spinning on a batch it cannot reap", func(t *testing.T) {
		t.Parallel()

		// A full batch that reaps nothing would otherwise be re-fetched until the job timed out.
		fullBatch := make([]*dataprivacy.UserDataDisclosure, dataprivacy.ExpiredUserDataDisclosureBatchSize)
		for i := range fullBatch {
			fullBatch[i] = expiredDisclosure()
		}

		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(context.Context, string) error { return platformerrors.New("blah") },
		}
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
				return fullBatch, nil
			},
		}

		w := buildReaperForTest(t, repo, artifacts)

		assert.Error(t, w.Work(t.Context()))
		assert.Len(t, repo.GetExpiredUserDataDisclosuresCalls(), 1)
	})

	T.Run("drains a backlog across several batches", func(t *testing.T) {
		t.Parallel()

		// Two full batches then a short one: the run should take all three rather than leaving
		// the rest for the next interval, which is what makes the first-deploy backfill finish.
		fullBatch := make([]*dataprivacy.UserDataDisclosure, dataprivacy.ExpiredUserDataDisclosureBatchSize)
		for i := range fullBatch {
			fullBatch[i] = expiredDisclosure()
		}

		calls := 0
		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
				calls++
				if calls <= 2 {
					return fullBatch, nil
				}
				return []*dataprivacy.UserDataDisclosure{expiredDisclosure()}, nil
			},
			MarkUserDataDisclosureExpiredFunc: func(context.Context, string) error { return nil },
		}
		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(context.Context, string) error { return nil },
		}

		w := buildReaperForTest(t, repo, artifacts)

		require.NoError(t, w.Work(t.Context()))
		assert.Equal(t, 3, calls)
	})

	T.Run("stops at the per-run batch limit", func(t *testing.T) {
		t.Parallel()

		// An endless supply of full batches must not keep one execution running forever.
		fullBatch := make([]*dataprivacy.UserDataDisclosure, dataprivacy.ExpiredUserDataDisclosureBatchSize)
		for i := range fullBatch {
			fullBatch[i] = expiredDisclosure()
		}

		repo := &dataprivacymock.RepositoryMock{
			GetExpiredUserDataDisclosuresFunc: func(context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
				return fullBatch, nil
			},
			MarkUserDataDisclosureExpiredFunc: func(context.Context, string) error { return nil },
		}
		artifacts := &reportartifacts.StoreMock{
			DeleteFunc: func(context.Context, string) error { return nil },
		}

		w := buildReaperForTest(t, repo, artifacts)

		require.NoError(t, w.Work(t.Context()))
		assert.Len(t, repo.GetExpiredUserDataDisclosuresCalls(), maxBatchesPerRun)
	})
}
