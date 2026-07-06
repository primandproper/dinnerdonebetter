package grpc

import (
	"testing"

	managermock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/manager/mock"
	oauthsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"

	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		oauthManager := &managermock.OAuth2Manager{}

		service := NewService(logger, tracerProvider, oauthManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*oauthsvc.OAuthServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, oauthManager, impl.oauthDataManager)
	})
}
