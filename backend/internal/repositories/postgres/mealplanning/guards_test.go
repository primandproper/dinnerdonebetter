package mealplanning

import (
	"testing"

	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace/noop"
)

// recordingSpan is a span that is always recording and keeps every attribute set on it, so a test
// can see exactly what a prologue traced before it returned.
type recordingSpan struct {
	noop.Span
	attributes []attribute.KeyValue
}

func (s *recordingSpan) IsRecording() bool { return true }

func (s *recordingSpan) SetAttributes(kv ...attribute.KeyValue) {
	s.attributes = append(s.attributes, kv...)
}

func TestGuardIDs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mealPlanID, mealPlanEventID := identifiers.New(), identifiers.New()
		span := &recordingSpan{}

		logger, err := guardIDs(loggingnoop.NewLogger(), span,
			idArg{mealplanningkeys.MealPlanIDKey, mealPlanID},
			idArg{mealplanningkeys.MealPlanEventIDKey, mealPlanEventID},
		)
		require.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, []attribute.KeyValue{
			attribute.String(mealplanningkeys.MealPlanIDKey, mealPlanID),
			attribute.String(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID),
		}, span.attributes)
	})

	T.Run("with no IDs", func(t *testing.T) {
		t.Parallel()

		span := &recordingSpan{}

		_, err := guardIDs(loggingnoop.NewLogger(), span)
		require.NoError(t, err)
		assert.Empty(t, span.attributes)
	})

	T.Run("with empty first ID", func(t *testing.T) {
		t.Parallel()

		span := &recordingSpan{}

		_, err := guardIDs(loggingnoop.NewLogger(), span,
			idArg{mealplanningkeys.MealPlanIDKey, ""},
			idArg{mealplanningkeys.MealPlanEventIDKey, identifiers.New()},
		)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Empty(t, span.attributes)
	})

	T.Run("with empty later ID traces the ones before it", func(t *testing.T) {
		t.Parallel()

		mealPlanID := identifiers.New()
		span := &recordingSpan{}

		_, err := guardIDs(loggingnoop.NewLogger(), span,
			idArg{mealplanningkeys.MealPlanIDKey, mealPlanID},
			idArg{mealplanningkeys.MealPlanEventIDKey, ""},
			idArg{mealplanningkeys.MealPlanOptionIDKey, identifiers.New()},
		)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Equal(t, []attribute.KeyValue{
			attribute.String(mealplanningkeys.MealPlanIDKey, mealPlanID),
		}, span.attributes)
	})
}
