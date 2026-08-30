package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	managermock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/manager/mock"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"

	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildTestService builds a service backed by the given manager mock. A nil manager gets an
// unconfigured mock, which panics if any of its methods are called.
func buildTestService(t *testing.T, oauthManager *managermock.OAuth2ManagerMock) *serviceImpl {
	t.Helper()

	if oauthManager == nil {
		oauthManager = &managermock.OAuth2ManagerMock{}
	}

	return &serviceImpl{
		tracer:           tracing.NewTracerForTest(t.Name()),
		logger:           loggingnoop.NewLogger(),
		oauthDataManager: oauthManager,
	}
}

func TestServiceImpl_CreateOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeClient := oauthfakes.BuildFakeOAuth2Client()
		fakeInput := oauthfakes.BuildFakeOAuth2ClientCreationRequestInput()

		mockManager := &managermock.OAuth2ManagerMock{
			CreateOAuth2ClientFunc: func(_ context.Context, input *oauth.OAuth2ClientCreationRequestInput) (*oauth.OAuth2Client, error) {
				assert.NotNil(t, input)

				return fakeClient, nil
			},
		}
		service := buildTestService(t, mockManager)

		request := &oauthsvc.CreateOAuth2ClientRequest{
			Input: &oauthsvc.OAuth2ClientCreationRequestInput{
				Name: fakeInput.Name,
			},
		}

		response, err := service.CreateOAuth2Client(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeClient.ID, response.Created.Id)
		assert.Equal(t, fakeClient.Name, response.Created.Name)

		assert.Len(t, mockManager.CreateOAuth2ClientCalls(), 1)
	})

	t.Run("manager error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeInput := oauthfakes.BuildFakeOAuth2ClientCreationRequestInput()

		mockManager := &managermock.OAuth2ManagerMock{
			CreateOAuth2ClientFunc: func(_ context.Context, _ *oauth.OAuth2ClientCreationRequestInput) (*oauth.OAuth2Client, error) {
				return nil, errors.New("manager error")
			},
		}
		service := buildTestService(t, mockManager)

		request := &oauthsvc.CreateOAuth2ClientRequest{
			Input: &oauthsvc.OAuth2ClientCreationRequestInput{
				Name: fakeInput.Name,
			},
		}

		response, err := service.CreateOAuth2Client(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockManager.CreateOAuth2ClientCalls(), 1)
	})
}

func TestServiceImpl_GetOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeClient := oauthfakes.BuildFakeOAuth2Client()
		clientID := fakeClient.ID

		mockManager := &managermock.OAuth2ManagerMock{
			GetOAuth2ClientFunc: func(_ context.Context, oauth2ClientID string) (*oauth.OAuth2Client, error) {
				assert.Equal(t, clientID, oauth2ClientID)

				return fakeClient, nil
			},
		}
		service := buildTestService(t, mockManager)

		request := &oauthsvc.GetOAuth2ClientRequest{
			Oauth2ClientId: clientID,
		}

		response, err := service.GetOAuth2Client(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeClient.ID, response.Result.Id)

		assert.Len(t, mockManager.GetOAuth2ClientCalls(), 1)
	})

	t.Run("manager error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		clientID := "nonexistent-client"

		mockManager := &managermock.OAuth2ManagerMock{
			GetOAuth2ClientFunc: func(_ context.Context, oauth2ClientID string) (*oauth.OAuth2Client, error) {
				assert.Equal(t, clientID, oauth2ClientID)

				return nil, errors.New("manager error")
			},
		}
		service := buildTestService(t, mockManager)

		request := &oauthsvc.GetOAuth2ClientRequest{
			Oauth2ClientId: clientID,
		}

		response, err := service.GetOAuth2Client(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockManager.GetOAuth2ClientCalls(), 1)
	})
}

func TestServiceImpl_GetOAuth2Clients(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeClients := oauthfakes.BuildFakeOAuth2ClientsList()
		pageSize := uint16(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockManager := &managermock.OAuth2ManagerMock{
			GetOAuth2ClientsFunc: func(_ context.Context, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[oauth.OAuth2Client], error) {
				assert.NotNil(t, actualFilter)

				return fakeClients, nil
			},
		}
		service := buildTestService(t, mockManager)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &oauthsvc.GetOAuth2ClientsRequest{
			Filter: &filteringpb.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetOAuth2Clients(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeClients.Data))

		assert.Len(t, mockManager.GetOAuth2ClientsCalls(), 1)
	})

	t.Run("manager error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		pageSize := uint16(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockManager := &managermock.OAuth2ManagerMock{
			GetOAuth2ClientsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[oauth.OAuth2Client], error) {
				return nil, errors.New("manager error")
			},
		}
		service := buildTestService(t, mockManager)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &oauthsvc.GetOAuth2ClientsRequest{
			Filter: &filteringpb.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetOAuth2Clients(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockManager.GetOAuth2ClientsCalls(), 1)
	})
}

func TestServiceImpl_ArchiveOAuth2Client(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		clientID := "test-client-id"

		mockManager := &managermock.OAuth2ManagerMock{
			ArchiveOAuth2ClientFunc: func(_ context.Context, oauth2ClientID string) error {
				assert.Equal(t, clientID, oauth2ClientID)

				return nil
			},
		}
		service := buildTestService(t, mockManager)

		request := &oauthsvc.ArchiveOAuth2ClientRequest{
			Oauth2ClientId: clientID,
		}

		response, err := service.ArchiveOAuth2Client(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockManager.ArchiveOAuth2ClientCalls(), 1)
	})

	t.Run("manager error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		clientID := "nonexistent-client"

		mockManager := &managermock.OAuth2ManagerMock{
			ArchiveOAuth2ClientFunc: func(_ context.Context, oauth2ClientID string) error {
				assert.Equal(t, clientID, oauth2ClientID)

				return errors.New("manager error")
			},
		}
		service := buildTestService(t, mockManager)

		request := &oauthsvc.ArchiveOAuth2ClientRequest{
			Oauth2ClientId: clientID,
		}

		response, err := service.ArchiveOAuth2Client(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockManager.ArchiveOAuth2ClientCalls(), 1)
	})
}
