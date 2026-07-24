package grpc

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	settingsfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/fakes"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	settingssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func TestServiceImpl_CreateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationCreationRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: &settingssvc.ServiceSettingConfigurationCreationRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.CreateServiceSettingConfiguration), testutils.ContextMatcher, mock.MatchedBy(func(input *settings.ServiceSettingConfigurationDatabaseCreationInput) bool {
			return input != nil && input.BelongsToUser == "test-user-id" && input.BelongsToAccount == "test-account-id"
		})).Return(exampleServiceSettingConfiguration, nil)

		actual, err := service.CreateServiceSettingConfiguration(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Created)
		assert.Equal(t, exampleServiceSettingConfiguration.ID, actual.Created.Id)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationCreationRequestInput()

		service, settingsRepo := buildTestServiceWithSessionError(t)

		request := &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: &settingssvc.ServiceSettingConfigurationCreationRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		actual, err := service.CreateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Unauthenticated)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, settingsRepo := buildTestService(t)

		request := &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: &settingssvc.ServiceSettingConfigurationCreationRequestInput{
				// Missing required fields to trigger validation error
				Value: "",
			},
		}

		actual, err := service.CreateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.InvalidArgument)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationCreationRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: &settingssvc.ServiceSettingConfigurationCreationRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.CreateServiceSettingConfiguration), testutils.ContextMatcher, mock.MatchedBy(func(input *settings.ServiceSettingConfigurationDatabaseCreationInput) bool {
			return input != nil
		})).Return((*settings.ServiceSettingConfiguration)(nil), errors.New("repository error"))

		actual, err := service.CreateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}

func TestServiceImpl_GetServiceSettingConfigurationByName(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingConfigurationByNameRequest{
			ServiceSettingConfigurationName: exampleServiceSettingConfiguration.ServiceSetting.Name,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfigurationForAccountByName), testutils.ContextMatcher, "test-account-id", exampleServiceSettingConfiguration.ServiceSetting.Name).Return(exampleServiceSettingConfiguration, nil)

		actual, err := service.GetServiceSettingConfigurationByName(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Result)
		assert.Equal(t, exampleServiceSettingConfiguration.ID, actual.Result.Id)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestServiceWithSessionError(t)

		request := &settingssvc.GetServiceSettingConfigurationByNameRequest{
			ServiceSettingConfigurationName: exampleServiceSettingConfiguration.ServiceSetting.Name,
		}

		actual, err := service.GetServiceSettingConfigurationByName(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Unauthenticated)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingConfigurationByNameRequest{
			ServiceSettingConfigurationName: exampleServiceSettingConfiguration.ServiceSetting.Name,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfigurationForAccountByName), testutils.ContextMatcher, "test-account-id", exampleServiceSettingConfiguration.ServiceSetting.Name).Return((*settings.ServiceSettingConfiguration)(nil), errors.New("repository error"))

		actual, err := service.GetServiceSettingConfigurationByName(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}

func TestServiceImpl_GetServiceSettingConfigurationsForAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfigurationsList := settingsfakes.BuildFakeServiceSettingConfigurationsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfigurationsForAccount), testutils.ContextMatcher, "test-account-id", mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil
		})).Return(exampleServiceSettingConfigurationsList, nil)

		actual, err := service.GetServiceSettingConfigurationsForAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingConfigurationsList.Data))

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, settingsRepo := buildTestServiceWithSessionError(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		actual, err := service.GetServiceSettingConfigurationsForAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Unauthenticated)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfigurationsForAccount), testutils.ContextMatcher, "test-account-id", mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil
		})).Return((*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration])(nil), errors.New("repository error"))

		actual, err := service.GetServiceSettingConfigurationsForAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}

func TestServiceImpl_GetServiceSettingConfigurationsForUser(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfigurationsList := settingsfakes.BuildFakeServiceSettingConfigurationsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfigurationsForUser), testutils.ContextMatcher, "test-user-id", mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil
		})).Return(exampleServiceSettingConfigurationsList, nil)

		actual, err := service.GetServiceSettingConfigurationsForUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingConfigurationsList.Data))

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, settingsRepo := buildTestServiceWithSessionError(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		actual, err := service.GetServiceSettingConfigurationsForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Unauthenticated)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfigurationsForUser), testutils.ContextMatcher, "test-user-id", mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil
		})).Return((*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration])(nil), errors.New("repository error"))

		actual, err := service.GetServiceSettingConfigurationsForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}

func TestServiceImpl_UpdateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationUpdateRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.UpdateServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
			Input: &settingssvc.ServiceSettingConfigurationUpdateRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfiguration), testutils.ContextMatcher, exampleServiceSettingConfiguration.ID).Return(exampleServiceSettingConfiguration, nil)
		settingsRepo.On(reflection.GetMethodName(settingsRepo.UpdateServiceSettingConfiguration), testutils.ContextMatcher, mock.MatchedBy(func(input *settings.ServiceSettingConfiguration) bool {
			return input != nil && input.ID == exampleServiceSettingConfiguration.ID
		})).Return(nil)

		actual, err := service.UpdateServiceSettingConfiguration(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Updated)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with get repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationUpdateRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.UpdateServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
			Input: &settingssvc.ServiceSettingConfigurationUpdateRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfiguration), testutils.ContextMatcher, exampleServiceSettingConfiguration.ID).Return((*settings.ServiceSettingConfiguration)(nil), errors.New("repository error"))

		actual, err := service.UpdateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with update repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationUpdateRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.UpdateServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
			Input: &settingssvc.ServiceSettingConfigurationUpdateRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.GetServiceSettingConfiguration), testutils.ContextMatcher, exampleServiceSettingConfiguration.ID).Return(exampleServiceSettingConfiguration, nil)
		settingsRepo.On(reflection.GetMethodName(settingsRepo.UpdateServiceSettingConfiguration), testutils.ContextMatcher, mock.MatchedBy(func(input *settings.ServiceSettingConfiguration) bool {
			return input != nil && input.ID == exampleServiceSettingConfiguration.ID
		})).Return(errors.New("repository error"))

		actual, err := service.UpdateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}

func TestServiceImpl_ArchiveServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.ArchiveServiceSettingConfiguration), testutils.ContextMatcher, exampleServiceSettingConfiguration.ID).Return(nil)

		actual, err := service.ArchiveServiceSettingConfiguration(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
		}

		settingsRepo.On(reflection.GetMethodName(settingsRepo.ArchiveServiceSettingConfiguration), testutils.ContextMatcher, exampleServiceSettingConfiguration.ID).Return(errors.New("repository error"))

		actual, err := service.ArchiveServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		mock.AssertExpectationsForObjects(t, settingsRepo)
	})
}
