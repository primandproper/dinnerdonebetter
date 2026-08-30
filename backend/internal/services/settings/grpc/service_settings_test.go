package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/mock"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	"github.com/primandproper/platform-go/v13/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	testAccountID = identifiers.New()
	testUserID    = identifiers.New()
)

func buildTestService(t *testing.T) (*serviceImpl, *settingsmock.RepositoryMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	settingsRepo := &settingsmock.RepositoryMock{}

	service := &serviceImpl{
		tracer:          tracer,
		logger:          logger,
		settingsManager: settingsRepo,
	}

	return service, settingsRepo
}

func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: testAccountID,
		Requester:       sessions.RequesterInfo{UserID: testUserID},
	})
}

func TestServiceImpl_CreateServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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
			assert.NotNil(t, input)

			return exampleServiceSetting, nil
		}

		actual, err := service.CreateServiceSetting(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Created)
		assert.Equal(t, exampleServiceSetting.ID, actual.Created.Id)

		assert.Len(t, settingsRepo.CreateServiceSettingCalls(), 1)
	})

	t.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _ := buildTestService(t)

		request := &settingssvc.CreateServiceSettingRequest{
			Input: &settingssvc.ServiceSettingCreationRequestInput{
				// Missing required fields to trigger validation error
				Name: "",
			},
		}

		actual, err := service.CreateServiceSetting(ctx, request)

		require.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.InvalidArgument)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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
			assert.NotNil(t, input)

			return nil, errors.New("repository error")
		}

		actual, err := service.CreateServiceSetting(ctx, request)

		require.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.CreateServiceSettingCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.NotNil(t, actual.Result)
		assert.Equal(t, exampleServiceSetting.ID, actual.Result.Id)

		assert.Len(t, settingsRepo.GetServiceSettingCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		require.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingCalls(), 1)
	})
}

func TestServiceImpl_GetServiceSettings(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingsList := settingsfakes.BuildFakeServiceSettingsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingsRequest{
			Filter: &filteringpb.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingsFunc = func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			assert.NotNil(t, filter)

			return exampleServiceSettingsList, nil
		}

		actual, err := service.GetServiceSettings(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettingsList.Data))

		assert.Len(t, settingsRepo.GetServiceSettingsCalls(), 1)
	})

	// A page size above the ceiling is the reason this path goes through
	// filtering/grpc rather than a local converter. protobuf carries the value as
	// a uint32 and filtering.QueryFilter holds a uint16, so a converter that
	// narrows before it clamps wraps instead: 70000 becomes 4464, which then
	// clamps to a page size that looks like an answer to a question nobody asked.
	t.Run("with a page size above the ceiling", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		exampleServiceSettingsList := settingsfakes.BuildFakeServiceSettingsList()

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(70_000)
		request := &settingssvc.GetServiceSettingsRequest{
			Filter: &filteringpb.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingsFunc = func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			require.NotNil(t, filter.MaxResponseSize)
			assert.Equal(t, filtering.MaxQueryFilterLimit, *filter.MaxResponseSize)

			return exampleServiceSettingsList, nil
		}

		actual, err := service.GetServiceSettings(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, actual)

		assert.Len(t, settingsRepo.GetServiceSettingsCalls(), 1)
	})

	// A sort direction nobody recognizes is reported rather than silently listed
	// under ascending, so a mistyped filter is not answered with a plausible page.
	t.Run("with an unrecognized sort direction", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		service, settingsRepo := buildTestService(t)

		sortBy := "sideways"
		request := &settingssvc.GetServiceSettingsRequest{
			Filter: &filteringpb.QueryFilter{
				SortBy: &sortBy,
			},
		}

		actual, err := service.GetServiceSettings(ctx, request)

		require.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		assert.Empty(t, settingsRepo.GetServiceSettingsCalls())
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		service, settingsRepo := buildTestService(t)

		pageSize := uint32(50)
		request := &settingssvc.GetServiceSettingsRequest{
			Filter: &filteringpb.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		settingsRepo.GetServiceSettingsFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
			return nil, errors.New("repository error")
		}

		actual, err := service.GetServiceSettings(ctx, request)

		require.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.GetServiceSettingsCalls(), 1)
	})
}

func TestServiceImpl_SearchForServiceSettings(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)
		assert.Len(t, actual.Results, len(exampleServiceSettings.Data))

		assert.Len(t, settingsRepo.SearchForServiceSettingsCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		require.Error(t, err)
		assert.Nil(t, actual)
		assertGRPCErrorHasStatus(t, err, codes.Internal)

		assert.Len(t, settingsRepo.SearchForServiceSettingsCalls(), 1)
	})
}

func TestServiceImpl_ArchiveServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotNil(t, actual.ResponseDetails)

		assert.Len(t, settingsRepo.ArchiveServiceSettingCalls(), 1)
	})

	t.Run("with repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
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

		require.Error(t, err)
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
