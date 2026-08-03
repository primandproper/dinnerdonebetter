package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestServiceImpl_CreateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		settingsRepo.CreateServiceSettingConfigurationFunc = func(_ context.Context, input *settings.ServiceSettingConfigurationDatabaseCreationInput) (*settings.ServiceSettingConfiguration, error) {
			assert.True(t, input != nil && input.BelongsToUser == testUserID && input.BelongsToAccount == testAccountID)

			return exampleServiceSettingConfiguration, nil
		}

		actual, err := service.CreateServiceSettingConfiguration(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Created)
		assert.Equal(t, exampleServiceSettingConfiguration.ID, actual.Created.Id)

		assert.Len(t, settingsRepo.CreateServiceSettingConfigurationCalls(), 1)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationCreationRequestInput()

		service, _ := buildTestService(t)

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
	})

	t.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _ := buildTestService(t)

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
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleInput := settingsfakes.BuildFakeServiceSettingConfigurationCreationRequestInput()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: &settingssvc.ServiceSettingConfigurationCreationRequestInput{
				Value:            exampleInput.Value,
				Notes:            exampleInput.Notes,
				ServiceSettingId: exampleInput.ServiceSettingID,
			},
		}

		settingsRepo.CreateServiceSettingConfigurationFunc = func(_ context.Context, input *settings.ServiceSettingConfigurationDatabaseCreationInput) (*settings.ServiceSettingConfiguration, error) {
			assert.True(t, input != nil)

			return nil, errors.New("repository error")
		}

		actual, err := service.CreateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.CreateServiceSettingConfigurationCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSettingConfigurationByName(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingConfigurationByNameRequest{
			ServiceSettingConfigurationName: exampleServiceSettingConfiguration.ServiceSetting.Name,
		}

		settingsRepo.GetServiceSettingConfigurationForAccountByNameFunc = func(_ context.Context, accountID string, serviceSettingConfigurationName string) (*settings.ServiceSettingConfiguration, error) {
			assert.Equal(t, testAccountID, accountID)
			assert.Equal(t, exampleServiceSettingConfiguration.ServiceSetting.Name, serviceSettingConfigurationName)

			return exampleServiceSettingConfiguration, nil
		}

		actual, err := service.GetServiceSettingConfigurationByName(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Result)
		assert.Equal(t, exampleServiceSettingConfiguration.ID, actual.Result.Id)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationForAccountByNameCalls(), 1)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, _ := buildTestService(t)

		request := &settingssvc.GetServiceSettingConfigurationByNameRequest{
			ServiceSettingConfigurationName: exampleServiceSettingConfiguration.ServiceSetting.Name,
		}

		actual, err := service.GetServiceSettingConfigurationByName(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Unauthenticated)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.GetServiceSettingConfigurationByNameRequest{
			ServiceSettingConfigurationName: exampleServiceSettingConfiguration.ServiceSetting.Name,
		}

		settingsRepo.GetServiceSettingConfigurationForAccountByNameFunc = func(_ context.Context, accountID string, serviceSettingConfigurationName string) (*settings.ServiceSettingConfiguration, error) {
			assert.Equal(t, testAccountID, accountID)
			assert.Equal(t, exampleServiceSettingConfiguration.ServiceSetting.Name, serviceSettingConfigurationName)

			return nil, errors.New("repository error")
		}

		actual, err := service.GetServiceSettingConfigurationByName(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationForAccountByNameCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSettingConfigurationsForAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingConfigurationsList := settingsfakes.BuildFakeServiceSettingConfigurationsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingConfigurationsForAccountFunc = func(_ context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
			assert.Equal(t, testAccountID, accountID)
			assert.True(t, filter != nil)

			return exampleServiceSettingConfigurationsList, nil
		}

		actual, err := service.GetServiceSettingConfigurationsForAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingConfigurationsList.Data))

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationsForAccountCalls(), 1)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, _ := buildTestService(t)

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
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingConfigurationsForAccountFunc = func(_ context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
			assert.Equal(t, testAccountID, accountID)
			assert.True(t, filter != nil)

			return nil, errors.New("repository error")
		}

		actual, err := service.GetServiceSettingConfigurationsForAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationsForAccountCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSettingConfigurationsForUser(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingConfigurationsList := settingsfakes.BuildFakeServiceSettingConfigurationsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingConfigurationsForUserFunc = func(_ context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
			assert.Equal(t, testUserID, userID)
			assert.True(t, filter != nil)

			return exampleServiceSettingConfigurationsList, nil
		}

		actual, err := service.GetServiceSettingConfigurationsForUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingConfigurationsList.Data))

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationsForUserCalls(), 1)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		service, _ := buildTestService(t)

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
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingConfigurationsForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingConfigurationsForUserFunc = func(_ context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
			assert.Equal(t, testUserID, userID)
			assert.True(t, filter != nil)

			return nil, errors.New("repository error")
		}

		actual, err := service.GetServiceSettingConfigurationsForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationsForUserCalls(), 1)
	})
}

func TestServiceImpl_UpdateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		settingsRepo.GetServiceSettingConfigurationFunc = func(_ context.Context, serviceSettingConfigurationID string) (*settings.ServiceSettingConfiguration, error) {
			assert.Equal(t, exampleServiceSettingConfiguration.ID, serviceSettingConfigurationID)

			return exampleServiceSettingConfiguration, nil
		}
		settingsRepo.UpdateServiceSettingConfigurationFunc = func(_ context.Context, input *settings.ServiceSettingConfiguration) error {
			assert.True(t, input != nil && input.ID == exampleServiceSettingConfiguration.ID)

			return nil
		}

		actual, err := service.UpdateServiceSettingConfiguration(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Updated)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationCalls(), 1)
		assert.Len(t, settingsRepo.UpdateServiceSettingConfigurationCalls(), 1)
	})

	t.Run("with get repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		settingsRepo.GetServiceSettingConfigurationFunc = func(_ context.Context, serviceSettingConfigurationID string) (*settings.ServiceSettingConfiguration, error) {
			assert.Equal(t, exampleServiceSettingConfiguration.ID, serviceSettingConfigurationID)

			return nil, errors.New("repository error")
		}

		actual, err := service.UpdateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationCalls(), 1)
	})

	t.Run("with update repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		settingsRepo.GetServiceSettingConfigurationFunc = func(_ context.Context, serviceSettingConfigurationID string) (*settings.ServiceSettingConfiguration, error) {
			assert.Equal(t, exampleServiceSettingConfiguration.ID, serviceSettingConfigurationID)

			return exampleServiceSettingConfiguration, nil
		}
		settingsRepo.UpdateServiceSettingConfigurationFunc = func(_ context.Context, input *settings.ServiceSettingConfiguration) error {
			assert.True(t, input != nil && input.ID == exampleServiceSettingConfiguration.ID)

			return errors.New("repository error")
		}

		actual, err := service.UpdateServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingConfigurationCalls(), 1)
		assert.Len(t, settingsRepo.UpdateServiceSettingConfigurationCalls(), 1)
	})
}

func TestServiceImpl_ArchiveServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
		}

		settingsRepo.ArchiveServiceSettingConfigurationFunc = func(_ context.Context, serviceSettingConfigurationID string) error {
			assert.Equal(t, exampleServiceSettingConfiguration.ID, serviceSettingConfigurationID)

			return nil
		}

		actual, err := service.ArchiveServiceSettingConfiguration(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)

		assert.Len(t, settingsRepo.ArchiveServiceSettingConfigurationCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingConfiguration := settingsfakes.BuildFakeServiceSettingConfiguration()

		service, settingsRepo := buildTestService(t)

		request := &settingssvc.ArchiveServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: exampleServiceSettingConfiguration.ID,
		}

		settingsRepo.ArchiveServiceSettingConfigurationFunc = func(_ context.Context, serviceSettingConfigurationID string) error {
			assert.Equal(t, exampleServiceSettingConfiguration.ID, serviceSettingConfigurationID)

			return errors.New("repository error")
		}

		actual, err := service.ArchiveServiceSettingConfiguration(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.ArchiveServiceSettingConfigurationCalls(), 1)
	})
}
