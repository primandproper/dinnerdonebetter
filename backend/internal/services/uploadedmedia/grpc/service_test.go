package grpc

import (
	"testing"

	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	meteringmock "github.com/primandproper/platform-go/v13/metering/mock"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v13/uploads/mock"
	registrymock "github.com/primandproper/platform-go/v13/uploads/registry/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		uploadsRegistry := &registrymock.StoreMock{}
		uploadManager := &mockuploads.UploadManagerMock{}
		usageRecorder := &meteringmock.RecorderMock{}

		service := NewService(logger, tracerProvider, uploadsRegistry, uploadManager, usageRecorder)

		assert.NotNil(t, service)
		assert.Implements(t, (*uploadedmediasvc.UploadedMediaServiceServer)(nil), service)

		impl, ok := service.(*serviceImpl)
		require.True(t, ok)
		assert.Equal(t, uploadsRegistry, impl.registry)
		assert.Equal(t, uploadManager, impl.uploadManager)
		assert.Equal(t, usageRecorder, impl.usageRecorder)
	})
}
