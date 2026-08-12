package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	webhookmgrmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/manager/mock"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"

	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	testAccountID = identifiers.New()
	testUserID    = identifiers.New()
)

// exampleSecret stands in for a hex-encoded signing secret.
const exampleSecret = "6465616462656566"

func buildTestService(t *testing.T) (*serviceImpl, *webhookmgrmock.WebhookDataManagerMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	webhookManager := &webhookmgrmock.WebhookDataManagerMock{}

	service := &serviceImpl{
		tracer:         tracer,
		logger:         logger,
		webhookManager: webhookManager,
	}

	return service, webhookManager
}

func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: testAccountID,
		Requester:       sessions.RequesterInfo{UserID: testUserID},
	})
}

func TestServiceImpl_CreateWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWebhook := webhookfakes.BuildFakeWebhook()
		fakeInput := webhookfakes.BuildFakeWebhookCreationRequestInput()

		mockRepo.CreateWebhookFunc = func(_ context.Context, userID string, accountID string, _ *webhooks.WebhookCreationRequestInput) (*webhooks.WebhookCreationResponse, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, testAccountID, accountID)

			return &webhooks.WebhookCreationResponse{Webhook: fakeWebhook, Secret: exampleSecret}, nil
		}

		request := &webhookssvc.CreateWebhookRequest{
			Input: converters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(fakeInput),
		}

		response, err := service.CreateWebhook(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWebhook.ID, response.Created.Id)
		// The response is the one place the signing secret is ever produced.
		assert.Equal(t, exampleSecret, response.Secret)
		assert.Equal(t, fakeWebhook.Name, response.Created.Name)
		assert.Equal(t, fakeWebhook.URL, response.Created.Url)

		assert.Len(t, mockRepo.CreateWebhookCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		testEventID := "test_event"
		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        "test webhook",
				Url:         "https://example.com/webhook",
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				EventTypes:  []string{testEventID},
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _ := buildTestService(t)

		testEventID := "test_event"
		// Invalid request with empty name
		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        "", // Invalid empty name
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				Url:         "https://example.com/webhook",
				EventTypes:  []string{testEventID},
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeInput := webhookfakes.BuildFakeWebhookCreationRequestInput()

		mockRepo.CreateWebhookFunc = func(_ context.Context, userID string, accountID string, _ *webhooks.WebhookCreationRequestInput) (*webhooks.WebhookCreationResponse, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, testAccountID, accountID)

			return nil, errors.New("repository error")
		}

		request := &webhookssvc.CreateWebhookRequest{
			Input: converters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(fakeInput),
		}

		response, err := service.CreateWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.CreateWebhookCalls(), 1)
	})
}

func TestServiceImpl_AddWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeConfig := webhookfakes.BuildFakeWebhookTriggerConfig()
		webhookID := identifiers.New()
		triggerEventID := fakeConfig.EventType

		mockRepo.AddWebhookTriggerConfigFunc = func(_ context.Context, accountID string, _ *webhooks.WebhookTriggerConfigCreationRequestInput) (*webhooks.WebhookTriggerConfig, error) {
			assert.Equal(t, testAccountID, accountID)

			return fakeConfig, nil
		}

		request := &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: webhookID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: webhookID,
				EventType:        triggerEventID,
			},
		}

		response, err := service.AddWebhookTriggerConfig(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeConfig.ID, response.Created.Id)
		assert.Equal(t, fakeConfig.EventType, response.Created.EventType)

		assert.Len(t, mockRepo.AddWebhookTriggerConfigCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		webhookID := identifiers.New()
		triggerEventID := identifiers.New()
		request := &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: webhookID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: webhookID,
				EventType:        triggerEventID,
			},
		}

		response, err := service.AddWebhookTriggerConfig(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		mockRepo.AddWebhookTriggerConfigFunc = func(_ context.Context, accountID string, _ *webhooks.WebhookTriggerConfigCreationRequestInput) (*webhooks.WebhookTriggerConfig, error) {
			assert.Equal(t, testAccountID, accountID)

			return nil, errors.New("repository error")
		}

		webhookID := identifiers.New()
		request := &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: webhookID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: webhookID,
				EventType:        webhookfakes.BuildFakeWebhookEventType(),
			},
		}

		response, err := service.AddWebhookTriggerConfig(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.AddWebhookTriggerConfigCalls(), 1)
	})
}

func TestServiceImpl_GetWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWebhook := webhookfakes.BuildFakeWebhook()
		webhookID := identifiers.New()

		mockRepo.GetWebhookFunc = func(_ context.Context, actualWebhookID string, accountID string) (*webhooks.Webhook, error) {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, testAccountID, accountID)

			return fakeWebhook, nil
		}

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWebhook.ID, response.Result.Id)
		assert.Equal(t, fakeWebhook.Name, response.Result.Name)

		assert.Len(t, mockRepo.GetWebhookCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		webhookID := identifiers.New()
		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		webhookID := identifiers.New()

		mockRepo.GetWebhookFunc = func(_ context.Context, actualWebhookID string, accountID string) (*webhooks.Webhook, error) {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, testAccountID, accountID)

			return nil, errors.New("repository error")
		}

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWebhookCalls(), 1)
	})
}

func TestServiceImpl_GetWebhooks(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWebhooks := webhookfakes.BuildFakeWebhooksList()

		mockRepo.GetWebhooksFunc = func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
			assert.Equal(t, testAccountID, accountID)

			return fakeWebhooks, nil
		}

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{
				// Add any filter fields as needed
			},
		}

		response, err := service.GetWebhooks(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWebhooks.Data))
		assert.Equal(t, fakeWebhooks.Data[0].ID, response.Results[0].Id)

		assert.Len(t, mockRepo.GetWebhooksCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWebhooks(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		mockRepo.GetWebhooksFunc = func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
			assert.Equal(t, testAccountID, accountID)

			return nil, errors.New("repository error")
		}

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWebhooks(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWebhooksCalls(), 1)
	})
}

func TestServiceImpl_ArchiveWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		webhookID := identifiers.New()

		mockRepo.ArchiveWebhookFunc = func(_ context.Context, actualWebhookID string, accountID string) error {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, testAccountID, accountID)

			return nil
		}

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockRepo.ArchiveWebhookCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		webhookID := identifiers.New()
		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		webhookID := identifiers.New()

		mockRepo.ArchiveWebhookFunc = func(_ context.Context, actualWebhookID string, accountID string) error {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, testAccountID, accountID)

			return errors.New("repository error")
		}

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.ArchiveWebhookCalls(), 1)
	})
}

func TestServiceImpl_ArchiveWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		webhookID := identifiers.New()
		configID := identifiers.New()

		// the handler first verifies the webhook belongs to the active account.
		fakeWebhook := webhookfakes.BuildFakeWebhook()
		mockRepo.GetWebhookFunc = func(_ context.Context, actualWebhookID string, accountID string) (*webhooks.Webhook, error) {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, testAccountID, accountID)

			return fakeWebhook, nil
		}
		mockRepo.ArchiveWebhookTriggerConfigFunc = func(_ context.Context, actualWebhookID, actualAccountID, actualConfigID string) error {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, configID, actualConfigID)

			return nil
		}

		request := &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              webhookID,
			WebhookTriggerConfigId: configID,
		}

		response, err := service.ArchiveWebhookTriggerConfig(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockRepo.GetWebhookCalls(), 1)
		assert.Len(t, mockRepo.ArchiveWebhookTriggerConfigCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		webhookID := identifiers.New()
		configID := identifiers.New()

		// the handler first verifies the webhook belongs to the active account.
		fakeWebhook := webhookfakes.BuildFakeWebhook()
		mockRepo.GetWebhookFunc = func(_ context.Context, actualWebhookID string, accountID string) (*webhooks.Webhook, error) {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, testAccountID, accountID)

			return fakeWebhook, nil
		}
		mockRepo.ArchiveWebhookTriggerConfigFunc = func(_ context.Context, actualWebhookID, actualAccountID, actualConfigID string) error {
			assert.Equal(t, webhookID, actualWebhookID)
			assert.Equal(t, configID, actualConfigID)

			return errors.New("repository error")
		}

		request := &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              webhookID,
			WebhookTriggerConfigId: configID,
		}

		response, err := service.ArchiveWebhookTriggerConfig(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWebhookCalls(), 1)
		assert.Len(t, mockRepo.ArchiveWebhookTriggerConfigCalls(), 1)
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
