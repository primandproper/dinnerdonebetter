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
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/filtering"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *settingsmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	settingsRepo := &settingsmock.Repository{}

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

func buildTestServiceWithSessionError(t *testing.T) (*serviceImpl, *settingsmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	settingsRepo := &settingsmock.Repository{}

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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.CreateServiceSetting), testutils.ContextMatcher, mock.MatchedBy(func(input any) bool {
			return input != nil
		})).Return(exampleServiceSetting, nil)

		actual, err := service.CreateServiceSetting(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Created)
		assert.Equal(t, exampleServiceSetting.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, settingsRepo := buildTestService(t)

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

		mock.AssertExpectationsForObjects(t, settingsRepo)
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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.CreateServiceSetting), testutils.ContextMatcher, mock.MatchedBy(func(input *settings.ServiceSettingDatabaseCreationInput) bool {
			return input != nil
		})).Return((*settings.ServiceSetting)(nil), errors.New("repository error"))

		actual, err := service.CreateServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSetting), testutils.ContextMatcher, exampleServiceSetting.ID).Return(exampleServiceSetting, nil)

		actual, err := service.GetServiceSetting(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Result)
		assert.Equal(t, exampleServiceSetting.ID, actual.Result.Id)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingRequest{
			ServiceSettingId: exampleServiceSetting.ID,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSetting), testutils.ContextMatcher, exampleServiceSetting.ID).Return((*settings.ServiceSetting)(nil), errors.New("repository error"))

		actual, err := service.GetServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettings), testutils.ContextMatcher, mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil
		})).Return(exampleServiceSettingsList, nil)

		actual, err := service.GetServiceSettings(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingsList.Data))

		mock.AssertExpectationsForObjects(t, settingsRepo)
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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettings), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[settings.ServiceSetting])(nil), errors.New("repository error"))

		actual, err := service.GetServiceSettings(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.SearchForServiceSettings), testutils.ContextMatcher, query, testutils.QueryFilterMatcher).Return(exampleServiceSettings, nil)

		actual, err := service.SearchForServiceSettings(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettings.Data))

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		query := "test query"

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.SearchForServiceSettingsRequest{
			Query: query,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.SearchForServiceSettings), testutils.ContextMatcher, query, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[settings.ServiceSetting])(nil), errors.New("repository error"))

		actual, err := service.SearchForServiceSettings(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
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

		settingsRepo.On(reflection.GetMethodName(settingsRepo.ArchiveServiceSetting), testutils.ContextMatcher, exampleServiceSetting.ID).Return(nil)

		actual, err := service.ArchiveServiceSetting(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSetting := settingsfakes.BuildFakeServiceSetting()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingRequest{
			ServiceSettingId: exampleServiceSetting.ID,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.ArchiveServiceSetting), testutils.ContextMatcher, exampleServiceSetting.ID).Return(errors.New("repository error"))

		actual, err := service.ArchiveServiceSetting(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}

// Helper function to assert GRPC error codes.
func assertGRPCErrorHasStatus(t *testing.T, err error, expectedCode codes.Code) {
	t.Helper()

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok, "error should be a gRPC status error")
	assert.Equal(t, expectedCode, grpcStatus.Code())
}
