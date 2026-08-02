package grpc

import (
	"testing"

	notificationsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/mock"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"

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
		notificationsManager := &notificationsmock.RepositoryMock{}

		service := NewService(logger, tracerProvider, notificationsManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*notificationssvc.UserNotificationsServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, notificationsManager, impl.notificationsManager)
		assert.NotNil(t, impl.sessionContextDataFetcher)
	})
}
