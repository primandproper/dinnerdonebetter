package grpc

import (
	"testing"

	uploadedmediamock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v8/uploads/mock"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		uploadedMediaManager := &uploadedmediamock.RepositoryMock{}

		service := NewService(logger, tracerProvider, uploadedMediaManager, &mockuploads.UploadManagerMock{})

		assert.NotNil(t, service)
		assert.Implements(t, (*uploadedmediasvc.UploadedMediaServiceServer)(nil), service)
	})
}
