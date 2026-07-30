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

	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	"github.com/primandproper/platform-go/v8/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v8/uploads/mock"

	"github.com/stretchr/testify/assert"
)

// buildTestService builds a service backed by the given mocks. Nil arguments get unconfigured
// mocks, which panic if any of their methods are called.
func buildTestService(t *testing.T, mockRepo *dataprivacymock.RepositoryMock, mockUploads *mockuploads.UploadManagerMock) *serviceImpl {
	t.Helper()

	if mockRepo == nil {
		mockRepo = &dataprivacymock.RepositoryMock{}
	}

	if mockUploads == nil {
		mockUploads = &mockuploads.UploadManagerMock{}
	}

	exampleUserID := identifiers.New()
	sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
		return &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		}, nil
	}

	return &serviceImpl{
		tracer:                    tracing.NewTracerForTest(t.Name()),
		logger:                    loggingnoop.NewLogger(),
		sessionContextDataFetcher: sessionFetcher,
		dataPrivacyManager:        mockRepo,
		uploadManager:             mockUploads,
		// An empty config resolves to a noop publisher provider, so Publish succeeds without a real message queue.
		msgConfig:    &msgconfig.Config{},
		queuesConfig: &msgconfig.QueuesConfig{},
	}
}

// sessionFetcherForUser returns a session context data fetcher that reports the given user.
func sessionFetcherForUser(userID string) func(context.Context) (*sessions.ContextData, error) {
	return func(context.Context) (*sessions.ContextData, error) {
		return &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		}, nil
	}
}

func TestNewDataPrivacyService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		mockRepo := &dataprivacymock.RepositoryMock{}
		mockUploads := &mockuploads.UploadManagerMock{}
		sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
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

		disclosure := &dataprivacy.UserDataDisclosure{ID: identifiers.New()}
		mockRepo := &dataprivacymock.RepositoryMock{
			CreateUserDataDisclosureFunc: func(_ context.Context, _ *dataprivacy.UserDataDisclosureCreationInput) (*dataprivacy.UserDataDisclosure, error) {
				return disclosure, nil
			},
		}
		service := buildTestService(t, mockRepo, nil)

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotEmpty(t, response.ReportId)

		assert.Len(t, mockRepo.CreateUserDataDisclosureCalls(), 1)
		// The aggregation is now performed asynchronously; the request path must not gather data inline.
		assert.Empty(t, mockRepo.FetchUserDataCollectionCalls())
	})

	t.Run("with error creating disclosure", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mockRepo := &dataprivacymock.RepositoryMock{
			CreateUserDataDisclosureFunc: func(_ context.Context, _ *dataprivacy.UserDataDisclosureCreationInput) (*dataprivacy.UserDataDisclosure, error) {
				return nil, errors.New("blah")
			},
		}
		service := buildTestService(t, mockRepo, nil)

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mockRepo.CreateUserDataDisclosureCalls(), 1)
	})
}

func TestServiceImpl_DestroyAllUserData(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mockRepo := &dataprivacymock.RepositoryMock{
			DeleteUserFunc: func(_ context.Context, userID string) error {
				assert.NotEmpty(t, userID)

				return nil
			},
		}
		service := buildTestService(t, mockRepo, nil)

		request := &dataprivacysvc.DestroyAllUserDataRequest{}

		response, err := service.DestroyAllUserData(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Successful)

		assert.Len(t, mockRepo.DeleteUserCalls(), 1)
	})
}

func TestServiceImpl_FetchUserDataReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := identifiers.New()

		collection := &dataprivacy.UserDataCollection{
			Identity: identity.UserDataCollection{
				User: identity.User{ID: userID},
			},
		}
		collectionBytes, err := json.Marshal(collection)
		assert.NoError(t, err)

		mockUploads := &mockuploads.UploadManagerMock{
			OpenFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(collectionBytes)), nil
			},
		}
		service := buildTestService(t, nil, mockUploads)
		service.sessionContextDataFetcher = sessionFetcherForUser(userID)

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

		userID := identifiers.New()
		disclosureID := identifiers.New()
		disclosure := &dataprivacy.UserDataDisclosure{ID: disclosureID, BelongsToUser: userID, Status: dataprivacy.UserDataDisclosureStatusCompleted}

		mockRepo := &dataprivacymock.RepositoryMock{
			GetUserDataDisclosureFunc: func(_ context.Context, actualDisclosureID string) (*dataprivacy.UserDataDisclosure, error) {
				assert.Equal(t, disclosureID, actualDisclosureID)

				return disclosure, nil
			},
		}
		service := buildTestService(t, mockRepo, nil)
		service.sessionContextDataFetcher = sessionFetcherForUser(userID)

		request := &dataprivacysvc.GetUserDataDisclosureRequest{UserDataDisclosureId: disclosureID}

		response, err := service.GetUserDataDisclosure(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.UserDataDisclosure)
		assert.Equal(t, disclosureID, response.UserDataDisclosure.Id)

		assert.Len(t, mockRepo.GetUserDataDisclosureCalls(), 1)
	})

	t.Run("with disclosure belonging to another user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := identifiers.New()
		disclosureID := identifiers.New()
		disclosure := &dataprivacy.UserDataDisclosure{ID: disclosureID, BelongsToUser: identifiers.New()}

		mockRepo := &dataprivacymock.RepositoryMock{
			GetUserDataDisclosureFunc: func(_ context.Context, actualDisclosureID string) (*dataprivacy.UserDataDisclosure, error) {
				assert.Equal(t, disclosureID, actualDisclosureID)

				return disclosure, nil
			},
		}
		service := buildTestService(t, mockRepo, nil)
		service.sessionContextDataFetcher = sessionFetcherForUser(userID)

		request := &dataprivacysvc.GetUserDataDisclosureRequest{UserDataDisclosureId: disclosureID}

		response, err := service.GetUserDataDisclosure(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mockRepo.GetUserDataDisclosureCalls(), 1)
	})
}

func TestServiceImpl_ListUserDataDisclosures(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := identifiers.New()

		result := &filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure]{
			Data: []*dataprivacy.UserDataDisclosure{
				{ID: identifiers.New(), BelongsToUser: userID},
				{ID: identifiers.New(), BelongsToUser: userID},
			},
		}

		mockRepo := &dataprivacymock.RepositoryMock{
			GetUserDataDisclosuresForUserFunc: func(_ context.Context, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure], error) {
				assert.Equal(t, userID, actualUserID)

				return result, nil
			},
		}
		service := buildTestService(t, mockRepo, nil)
		service.sessionContextDataFetcher = sessionFetcherForUser(userID)

		request := &dataprivacysvc.ListUserDataDisclosuresRequest{}

		response, err := service.ListUserDataDisclosures(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)

		assert.Len(t, mockRepo.GetUserDataDisclosuresForUserCalls(), 1)
	})
}
