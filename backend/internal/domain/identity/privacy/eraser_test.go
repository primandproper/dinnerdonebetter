package privacy

import (
	"context"
	"testing"

	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEraser_Erase(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUserID := identifiers.New()

		// The executor is a sentinel rather than a real transaction: what is being
		// asserted is that the eraser passes through the one it was handed instead of
		// reaching for a handle of its own, because that is what makes the erasure
		// atomic across every domain's eraser.
		executor := database.SQLQueryExecutor(nil)

		repo := &identitymock.RepositoryMock{
			EraseUserFunc: func(_ context.Context, actualExecutor database.SQLQueryExecutor, actualUserID string) (int64, error) {
				assert.Equal(t, executor, actualExecutor)
				assert.Equal(t, exampleUserID, actualUserID)

				return 1, nil
			},
		}

		eraser := NewEraser(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		outcome, err := eraser.Erase(t.Context(), executor, subject(exampleUserID))

		require.NoError(t, err)
		assert.Equal(t, int64(1), outcome.Deleted)
		assert.Zero(t, outcome.Anonymized)
		// Nothing is retained by this eraser, and an empty map here is the honest
		// answer: it is the audit eraser that reports retention, under its own key.
		assert.Empty(t, outcome.Retained)
		assert.Len(t, repo.EraseUserCalls(), 1)
	})

	T.Run("a subject who is already absent is not an error", func(t *testing.T) {
		t.Parallel()

		// An erasure is a statement about the end state, and somebody already gone is in
		// it. Failing here would roll back a transaction with nothing left to do.
		repo := &identitymock.RepositoryMock{
			EraseUserFunc: func(context.Context, database.SQLQueryExecutor, string) (int64, error) {
				return 0, nil
			},
		}

		eraser := NewEraser(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		outcome, err := eraser.Erase(t.Context(), nil, subject(identifiers.New()))

		require.NoError(t, err)
		assert.Zero(t, outcome.Deleted)
	})

	T.Run("with error erasing user", func(t *testing.T) {
		t.Parallel()

		repo := &identitymock.RepositoryMock{
			EraseUserFunc: func(context.Context, database.SQLQueryExecutor, string) (int64, error) {
				return 0, platformerrors.New("blah")
			},
		}

		eraser := NewEraser(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		outcome, err := eraser.Erase(t.Context(), nil, subject(identifiers.New()))

		require.Error(t, err)
		assert.Zero(t, outcome.Deleted)
	})
}
