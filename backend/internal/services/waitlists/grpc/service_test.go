package grpc

import (
	"testing"

	waitlistmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/mock"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		waitlistsManager := &waitlistmock.RepositoryMock{}

		service := NewService(logger, tracerProvider, waitlistsManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*waitlistssvc.WaitlistsServiceServer)(nil), service)
	})
}
