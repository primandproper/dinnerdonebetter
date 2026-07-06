package emaildeliverabilitytest

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/platform-go/v3/email"
	emailmock "github.com/primandproper/platform-go/v3/email/mock"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJob(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		job, err := NewJob(
			&emailmock.EmailerMock{},
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&JobParams{
				RecipientEmailAddress: "test@example.com",
				ServiceEnvironment:    "test",
			},
		)

		require.NoError(t, err)
		assert.NotNil(t, job)
	})

	T.Run("defaults service environment to prod when empty", func(t *testing.T) {
		t.Parallel()

		job, err := NewJob(
			&emailmock.EmailerMock{},
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&JobParams{
				RecipientEmailAddress: "test@example.com",
			},
		)

		require.NoError(t, err)
		assert.NotNil(t, job)
	})

	T.Run("returns error when recipient email is empty", func(t *testing.T) {
		t.Parallel()

		job, err := NewJob(
			&emailmock.EmailerMock{},
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&JobParams{},
		)

		require.Error(t, err)
		assert.Nil(t, job)
	})
}

func TestJob_Do(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		emailer := &emailmock.EmailerMock{
			SendEmailFunc: func(_ context.Context, _ *email.OutboundEmailMessage) error { return nil },
		}

		job, err := NewJob(
			emailer,
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&JobParams{
				RecipientEmailAddress: "test@example.com",
				ServiceEnvironment:    "test",
			},
		)
		require.NoError(t, err)

		err = job.Do(ctx)
		assert.NoError(t, err)
	})

	T.Run("returns error when emailer fails", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expectedErr := errors.New("email send failure")

		emailer := &emailmock.EmailerMock{
			SendEmailFunc: func(_ context.Context, _ *email.OutboundEmailMessage) error { return expectedErr },
		}

		job, err := NewJob(
			emailer,
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&JobParams{
				RecipientEmailAddress: "test@example.com",
				ServiceEnvironment:    "test",
			},
		)
		require.NoError(t, err)

		err = job.Do(ctx)
		assert.Error(t, err)
	})
}
