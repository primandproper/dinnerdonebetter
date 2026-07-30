package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	settingsfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	settingssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v8/filtering"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *settingsmock.RepositoryMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	settingsRepo := &settingsmock.RepositoryMock{}

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: "test-account-id",
				Requester: sessions.RequesterInfo{
					UserID: "test-user-id",
				},
			}, nil
		},
		settingsManager: settingsRepo,
	}

	return service, settingsRepo
}

func buildTestServiceWithSessionError(t *testing.T) (*serviceImpl, *settingsmock.RepositoryMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	settingsRepo := &settingsmock.RepositoryMock{}

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
		settingsManager: settingsRepo,
	}

	return service, settingsRepo
}

func TestServiceImpl_CreateServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()
		exampleInput := settingsfakes.BuildFakeServiceSettingCreationRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.CreateServiceSettingRequest{
			Input: &settingssvc.ServiceSettingCreationRequestInput{
				Name:         exampleInput.Name,
				Type:         exampleInput.Type,
				Description:  exampleInput.Description,
				DefaultValue: exampleInput.DefaultValue,
				Enumeration:  exampleInput.Enumeration,
				AdminsOnly:   exampleInput.AdminsOnly,
			},
		}

		settingsRepo.CreateServiceSettingFunc = func(_ context.Context, input *settings.ServiceSettingDatabaseCreationInput) (*settings.ServiceSetting, error) {
			assert.True(t, input != nil)

			return exampleServiceSetting, nil
		}

		actual, err := service.CreateServiceSetting(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Created)
		assert.Equal(t, exampleServiceSetting.ID, actual.Created.Id)

		assert.Len(t, settingsRepo.CreateServiceSettingCalls(), 1)
	})

	t.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &settingssvc.CreateServiceSettingRequest{
			Input: &settingssvc.ServiceSettingCreationRequestInput{
				// Missing required fields to trigger validation error
				Name: "",
			},
		}

		actual, err := service.CreateServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.InvalidArgument)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleInput := settingsfakes.BuildFakeServiceSettingCreationRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.CreateServiceSettingRequest{
			Input: &settingssvc.ServiceSettingCreationRequestInput{
				Name:         exampleInput.Name,
				Type:         exampleInput.Type,
				Description:  exampleInput.Description,
				DefaultValue: exampleInput.DefaultValue,
				Enumeration:  exampleInput.Enumeration,
				AdminsOnly:   exampleInput.AdminsOnly,
			},
		}

		settingsRepo.CreateServiceSettingFunc = func(_ context.Context, input *settings.ServiceSettingDatabaseCreationInput) (*settings.ServiceSetting, error) {
			assert.True(t, input != nil)

			return nil, errors.New("repository error")
		}

		actual, err := service.CreateServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.CreateServiceSettingCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingRequest{
			ServiceSettingId: exampleServiceSetting.ID,
		}

		settingsRepo.GetServiceSettingFunc = func(_ context.Context, serviceSettingID string) (*settings.ServiceSetting, error) {
			assert.Equal(t, exampleServiceSetting.ID, serviceSettingID)

			return exampleServiceSetting, nil
		}

		actual, err := service.GetServiceSetting(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Result)
		assert.Equal(t, exampleServiceSetting.ID, actual.Result.Id)

		assert.Len(t, settingsRepo.GetServiceSettingCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingRequest{
			ServiceSettingId: exampleServiceSetting.ID,
		}

		settingsRepo.GetServiceSettingFunc = func(_ context.Context, serviceSettingID string) (*settings.ServiceSetting, error) {
			assert.Equal(t, exampleServiceSetting.ID, serviceSettingID)

			return nil, errors.New("repository error")
		}

		actual, err := service.GetServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSettings(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingsList := settingsfakes.BuildFakeServiceSettingsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingsFunc = func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			assert.True(t, filter != nil)

			return exampleServiceSettingsList, nil
		}

		actual, err := service.GetServiceSettings(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingsList.Data))

		assert.Len(t, settingsRepo.GetServiceSettingsCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingsFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			return nil, errors.New("repository error")
		}

		actual, err := service.GetServiceSettings(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingsCalls(), 1)
	})
}

func TestServiceImpl_SearchForServiceSettings(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettings := &filtering.QueryFilteredResult[settings.ServiceSetting]{
			Data: []*settings.ServiceSetting{settingsfakes.BuildFakeServiceSetting()},
		}
		query := "test query"

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.SearchForServiceSettingsRequest{
			Query: query,
		}

		settingsRepo.SearchForServiceSettingsFunc = func(_ context.Context, actualQuery string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			assert.Equal(t, query, actualQuery)

			return exampleServiceSettings, nil
		}

		actual, err := service.SearchForServiceSettings(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettings.Data))

		assert.Len(t, settingsRepo.SearchForServiceSettingsCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		query := "test query"

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.SearchForServiceSettingsRequest{
			Query: query,
		}

		settingsRepo.SearchForServiceSettingsFunc = func(_ context.Context, actualQuery string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			assert.Equal(t, query, actualQuery)

			return nil, errors.New("repository error")
		}

		actual, err := service.SearchForServiceSettings(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.SearchForServiceSettingsCalls(), 1)
	})
}

func TestServiceImpl_ArchiveServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingRequest{
			ServiceSettingId: exampleServiceSetting.ID,
		}

		settingsRepo.ArchiveServiceSettingFunc = func(_ context.Context, serviceSettingID string) error {
			assert.Equal(t, exampleServiceSetting.ID, serviceSettingID)

			return nil
		}

		actual, err := service.ArchiveServiceSetting(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)

		assert.Len(t, settingsRepo.ArchiveServiceSettingCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingRequest{
			ServiceSettingId: exampleServiceSetting.ID,
		}

		settingsRepo.ArchiveServiceSettingFunc = func(_ context.Context, serviceSettingID string) error {
			assert.Equal(t, exampleServiceSetting.ID, serviceSettingID)

			return errors.New("repository error")
		}

		actual, err := service.ArchiveServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.ArchiveServiceSettingCalls(), 1)
	})
}

// Helper function to assert GRPC error codes.
func assertGRPCErrorHasStatus(t *testing.T, err error, expectedCode codes.Code) {
	t.Helper()

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok, "error should be a gRPC status error")
	assert.Equal(t, expectedCode, grpcStatus.Code())
}
