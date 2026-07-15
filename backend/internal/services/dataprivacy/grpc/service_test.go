package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"

	"github.com/primandproper/platform-go/v4/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v4/observability/logging/noop"
	"github.com/primandproper/platform-go/v4/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v4/observability/tracing/noop"
	"github.com/primandproper/platform-go/v4/uploads"
	mockuploads "github.com/primandproper/platform-go/v4/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) (*serviceImpl, *dataprivacymock.Repository, *mockuploads.UploadManagerMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	mockRepo := &dataprivacymock.Repository{}
	mockUploads := &mockuploads.UploadManagerMock{}

	exampleUserID := identifiers.New()
	sessionFetcher := func(ctx context.Context) (*sessions.ContextData, error) {
		return &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		}, nil
	}

	service := &serviceImpl{
		tracer:                    tracer,
		logger:                    logger,
		sessionContextDataFetcher: sessionFetcher,
		dataPrivacyManager:        mockRepo,
		uploadManager:             mockUploads,
	}

	return service, mockRepo, mockUploads
}

func TestNewDataPrivacyService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		mockRepo := &dataprivacymock.Repository{}
		mockUploads := &mockuploads.UploadManagerMock{}
		sessionFetcher := func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{}, nil
		}

		service := NewDataPrivacyService(logger, tracerProvider, sessionFetcher, mockRepo, mockUploads)

		assert.NotNil(t, service)
		assert.Implements(t, (*dataprivacysvc.DataPrivacyServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
	})
}

func TestServiceImpl_AggregateUserDataReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, mockUploads := buildTestService(t)

		collection := &dataprivacy.UserDataCollection{
			Identity: identity.UserDataCollection{
				User: identity.User{ID: identifiers.New()},
			},
		}

		mockRepo.On("FetchUserDataCollection", mock.Anything, mock.AnythingOfType("string")).Return(collection, nil)
		mockUploads.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotEmpty(t, response.ReportId)

		mockRepo.AssertExpectations(t)
	})
}

func TestServiceImpl_DestroyAllUserData(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.On("DeleteUser", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		request := &dataprivacysvc.DestroyAllUserDataRequest{}

		response, err := service.DestroyAllUserData(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Successful)

		mockRepo.AssertExpectations(t)
	})
}

func TestServiceImpl_FetchUserDataReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, mockUploads := buildTestService(t)

		collection := &dataprivacy.UserDataCollection{
			Identity: identity.UserDataCollection{
				User: identity.User{ID: identifiers.New()},
			},
		}
		collectionBytes, _ := json.Marshal(collection)

		mockUploads.OpenFunc = func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(collectionBytes)), nil
		}

		request := &dataprivacysvc.FetchUserDataReportRequest{
			UserDataAggregationReportId: identifiers.New(),
		}

		response, err := service.FetchUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
	})
}
