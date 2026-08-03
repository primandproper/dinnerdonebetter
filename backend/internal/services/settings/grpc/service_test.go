package grpc

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/mock"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		settingsManager := &mock.RepositoryMock{}

		service := NewService(
			logger,
			tracerProvider,
			settingsManager,
		)

		assert.NotNil(t, service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, settingsManager, impl.settingsManager)
	})
}
