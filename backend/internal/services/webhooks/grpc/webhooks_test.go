package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	webhookfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	webhookmgrmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/manager/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	webhookssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *webhookmgrmock.WebhookDataManager) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	webhookManager := &webhookmgrmock.WebhookDataManager{}

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: "test-account-id",
				Requester:       sessions.RequesterInfo{UserID: "test-user-id"},
			}, nil
		},
		webhookManager: webhookManager,
	}

	return service, webhookManager
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
		webhookManager: &webhookmgrmock.WebhookDataManager{},
	}

	return service
}

func TestServiceImpl_CreateWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWebhook := webhookfakes.BuildFakeWebhook()
		fakeInput := webhookfakes.BuildFakeWebhookCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWebhook), testutils.ContextMatcher, "test-user-id", "test-account-id", mock.AnythingOfType("*webhooks.WebhookCreationRequestInput")).Return(fakeWebhook, nil)

		request := &webhookssvc.CreateWebhookRequest{
			Input: converters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(fakeInput),
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWebhook.ID, response.Created.Id)
		assert.Equal(t, fakeWebhook.Name, response.Created.Name)
		assert.Equal(t, fakeWebhook.URL, response.Created.Url)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		testEventID := "test_event"
		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        "test webhook",
				Url:         "https://example.com/webhook",
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				Events:      []*webhookssvc.WebhookTriggerEventCreationRequestInput{{Id: &testEventID}},
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		testEventID := "test_event"
		// Invalid request with empty name
		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        "", // Invalid empty name
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				Url:         "https://example.com/webhook",
				Events:      []*webhookssvc.WebhookTriggerEventCreationRequestInput{{Id: &testEventID}},
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeInput := webhookfakes.BuildFakeWebhookCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWebhook), testutils.ContextMatcher, "test-user-id", "test-account-id", mock.AnythingOfType("*webhooks.WebhookCreationRequestInput")).Return(nil, errors.New("repository error"))

		request := &webhookssvc.CreateWebhookRequest{
			Input: converters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(fakeInput),
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_AddWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeConfig := webhookfakes.BuildFakeWebhookTriggerConfig()
		webhookID := "test-webhook-id"
		triggerEventID := fakeConfig.TriggerEventID

		mockRepo.On(reflection.GetMethodName(mockRepo.AddWebhookTriggerConfig), testutils.ContextMatcher, "test-account-id", mock.AnythingOfType("*webhooks.WebhookTriggerConfigCreationRequestInput")).Return(fakeConfig, nil)

		request := &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: webhookID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: webhookID,
				TriggerEventId:   triggerEventID,
			},
		}

		response, err := service.AddWebhookTriggerConfig(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeConfig.ID, response.Created.Id)
		assert.Equal(t, fakeConfig.TriggerEventID, response.Created.TriggerEventId)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: "test-webhook-id",
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: "test-webhook-id",
				TriggerEventId:   "test-event-id",
			},
		}

		response, err := service.AddWebhookTriggerConfig(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.AddWebhookTriggerConfig), testutils.ContextMatcher, "test-account-id", mock.AnythingOfType("*webhooks.WebhookTriggerConfigCreationRequestInput")).Return(nil, errors.New("repository error"))

		request := &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: "test-webhook-id",
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: "test-webhook-id",
				TriggerEventId:   "test_event",
			},
		}

		response, err := service.AddWebhookTriggerConfig(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWebhook := webhookfakes.BuildFakeWebhook()
		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(fakeWebhook, nil)

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWebhook.ID, response.Result.Id)
		assert.Equal(t, fakeWebhook.Name, response.Result.Name)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: "test-webhook-id",
		}

		response, err := service.GetWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(nil, errors.New("repository error"))

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWebhooks(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWebhooks := webhookfakes.BuildFakeWebhooksList()

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhooks), testutils.ContextMatcher, "test-account-id", testutils.QueryFilterMatcher).Return(fakeWebhooks, nil)

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{
				// Add any filter fields as needed
			},
		}

		response, err := service.GetWebhooks(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWebhooks.Data))
		assert.Equal(t, fakeWebhooks.Data[0].ID, response.Results[0].Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWebhooks(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhooks), testutils.ContextMatcher, "test-account-id", testutils.QueryFilterMatcher).Return(nil, errors.New("repository error"))

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWebhooks(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(nil)

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: "test-webhook-id",
		}

		response, err := service.ArchiveWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(errors.New("repository error"))

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"
		configID := "test-config-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhookTriggerConfig), testutils.ContextMatcher, webhookID, configID).Return(nil)

		request := &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              webhookID,
			WebhookTriggerConfigId: configID,
		}

		response, err := service.ArchiveWebhookTriggerConfig(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"
		configID := "test-config-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhookTriggerConfig), testutils.ContextMatcher, webhookID, configID).Return(errors.New("repository error"))

		request := &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              webhookID,
			WebhookTriggerConfigId: configID,
		}

		response, err := service.ArchiveWebhookTriggerConfig(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWebhookTriggerEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		eventID := "test-catalog-event-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhookTriggerEvent), testutils.ContextMatcher, eventID).Return(nil)

		request := &webhookssvc.ArchiveWebhookTriggerEventRequest{
			Id: eventID,
		}

		response, err := service.ArchiveWebhookTriggerEvent(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		eventID := "test-catalog-event-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhookTriggerEvent), testutils.ContextMatcher, eventID).Return(errors.New("repository error"))

		request := &webhookssvc.ArchiveWebhookTriggerEventRequest{
			Id: eventID,
		}

		response, err := service.ArchiveWebhookTriggerEvent(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	t.Run("implements WebhooksServiceServer", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		assert.Implements(t, (*webhookssvc.WebhooksServiceServer)(nil), service)
	})
}
