package grpc

import (
	"testing"

	webhookmgrmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/manager/mock"
	webhookssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

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
		webhookManager := &webhookmgrmock.WebhookDataManagerMock{}

		service := NewService(logger, tracerProvider, webhookManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*webhookssvc.WebhooksServiceServer)(nil), service)
	})
}
