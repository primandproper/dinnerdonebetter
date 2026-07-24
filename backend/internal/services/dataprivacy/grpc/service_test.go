package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"

	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/identifiers"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	"github.com/primandproper/platform-go/v6/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v6/uploads/mock"

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
		// An empty config resolves to a noop publisher provider, so Publish succeeds without a real message queue.
		msgConfig:    &msgconfig.Config{},
		queuesConfig: &msgconfig.QueuesConfig{},
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

		service := NewDataPrivacyService(logger, tracerProvider, sessionFetcher, mockRepo, mockUploads, &msgconfig.Config{}, &msgconfig.QueuesConfig{})

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
		service, mockRepo, _ := buildTestService(t)

		disclosure := &dataprivacy.UserDataDisclosure{ID: identifiers.New()}
		mockRepo.On("CreateUserDataDisclosure", mock.Anything, mock.AnythingOfType("*dataprivacy.UserDataDisclosureCreationInput")).Return(disclosure, nil)

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotEmpty(t, response.ReportId)

		// The aggregation is now performed asynchronously; the request path must not gather data inline.
		mockRepo.AssertNotCalled(t, "FetchUserDataCollection", mock.Anything, mock.Anything)
		mockRepo.AssertExpectations(t)
	})

	t.Run("with error creating disclosure", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.On("CreateUserDataDisclosure", mock.Anything, mock.AnythingOfType("*dataprivacy.UserDataDisclosureCreationInput")).Return((*dataprivacy.UserDataDisclosure)(nil), errors.New("blah"))

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

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

		userID := identifiers.New()
		service.sessionContextDataFetcher = func(context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		collection := &dataprivacy.UserDataCollection{
			Identity: identity.UserDataCollection{
				User: identity.User{ID: userID},
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

func TestServiceImpl_GetUserDataDisclosure(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		userID := identifiers.New()
		service.sessionContextDataFetcher = func(context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}, nil
		}

		disclosureID := identifiers.New()
		disclosure := &dataprivacy.UserDataDisclosure{ID: disclosureID, BelongsToUser: userID, Status: dataprivacy.UserDataDisclosureStatusCompleted}
		mockRepo.On("GetUserDataDisclosure", mock.Anything, disclosureID).Return(disclosure, nil)

		request := &dataprivacysvc.GetUserDataDisclosureRequest{UserDataDisclosureId: disclosureID}

		response, err := service.GetUserDataDisclosure(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.UserDataDisclosure)
		assert.Equal(t, disclosureID, response.UserDataDisclosure.Id)

		mockRepo.AssertExpectations(t)
	})

	t.Run("with disclosure belonging to another user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		userID := identifiers.New()
		service.sessionContextDataFetcher = func(context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}, nil
		}

		disclosureID := identifiers.New()
		disclosure := &dataprivacy.UserDataDisclosure{ID: disclosureID, BelongsToUser: identifiers.New()}
		mockRepo.On("GetUserDataDisclosure", mock.Anything, disclosureID).Return(disclosure, nil)

		request := &dataprivacysvc.GetUserDataDisclosureRequest{UserDataDisclosureId: disclosureID}

		response, err := service.GetUserDataDisclosure(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		mockRepo.AssertExpectations(t)
	})
}

func TestServiceImpl_ListUserDataDisclosures(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		userID := identifiers.New()
		service.sessionContextDataFetcher = func(context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}, nil
		}

		result := &filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure]{
			Data: []*dataprivacy.UserDataDisclosure{
				{ID: identifiers.New(), BelongsToUser: userID},
				{ID: identifiers.New(), BelongsToUser: userID},
			},
		}
		mockRepo.On("GetUserDataDisclosuresForUser", mock.Anything, userID, mock.Anything).Return(result, nil)

		request := &dataprivacysvc.ListUserDataDisclosuresRequest{}

		response, err := service.ListUserDataDisclosures(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)

		mockRepo.AssertExpectations(t)
	})
}
