package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	settingsconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createServiceSettingConfigurationForTest(t *testing.T, clientToUse client.Client) *settings.ServiceSettingConfiguration {
	t.Helper()
	ctx := t.Context()

	serviceSetting := createServiceSettingForTest(t)

	exampleServiceSettingConfiguration := fakes.BuildFakeServiceSettingConfiguration()
	exampleServiceSettingConfiguration.ServiceSetting = *serviceSetting
	exampleServiceSettingConfigurationInput := converters.ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput(exampleServiceSettingConfiguration)
	createdServiceSettingConfiguration, err := clientToUse.CreateServiceSettingConfiguration(ctx, &settingssvc.CreateServiceSettingConfigurationRequest{
		Input: settingsconverters.ConvertServiceSettingConfigurationCreationRequestInputToGRPCServiceSettingConfigurationCreationRequestInput(exampleServiceSettingConfigurationInput),
	})
	require.NoError(t, err)
	converted := settingsconverters.ConvertGRPCServiceSettingConfigurationToServiceSettingConfiguration(createdServiceSettingConfiguration.Created)
	assertRoughEquality(t, exampleServiceSettingConfiguration, converted, defaultIgnoredFields("ServiceSetting", "ID", "BelongsToUser", "BelongsToAccount")...)

	res, err := clientToUse.GetServiceSettingConfigurationByName(ctx, &settingssvc.GetServiceSettingConfigurationByNameRequest{ServiceSettingConfigurationName: createdServiceSettingConfiguration.Created.ServiceSetting.Name})
	require.NoError(t, err)
	require.NotNil(t, res)

	serviceSettingConfiguration := settingsconverters.ConvertGRPCServiceSettingConfigurationToServiceSettingConfiguration(res.Result)
	assertRoughEquality(t, converted, serviceSettingConfiguration, defaultIgnoredFields("ServiceSetting", "ID", "BelongsToUser", "BelongsToAccount")...)

	return serviceSettingConfiguration
}

func TestServiceSettingConfigurations_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "service_setting_configurations", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeServiceSettingConfigurationCreationRequestInput()
		convertedInput := settingsconverters.ConvertServiceSettingConfigurationCreationRequestInputToGRPCServiceSettingConfigurationCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateServiceSettingConfiguration(ctx, &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeServiceSettingConfigurationCreationRequestInput()
		convertedInput := settingsconverters.ConvertServiceSettingConfigurationCreationRequestInputToGRPCServiceSettingConfigurationCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Value = ""

		created, err := adminClient.CreateServiceSettingConfiguration(ctx, &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeServiceSettingConfigurationCreationRequestInput()
		convertedInput := settingsconverters.ConvertServiceSettingConfigurationCreationRequestInputToGRPCServiceSettingConfigurationCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateServiceSettingConfiguration(ctx, &settingssvc.CreateServiceSettingConfigurationRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestServiceSettingConfigurations_Reading_ByName(T *testing.T) {
	T.Parallel()

	T.Run("by name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		retrieved, err := testClient.GetServiceSettingConfigurationByName(ctx, &settingssvc.GetServiceSettingConfigurationByNameRequest{ServiceSettingConfigurationName: created.ServiceSetting.Name})
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		converted := settingsconverters.ConvertGRPCServiceSettingConfigurationToServiceSettingConfiguration(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetServiceSettingConfigurationByName(ctx, &settingssvc.GetServiceSettingConfigurationByNameRequest{ServiceSettingConfigurationName: created.ServiceSetting.Name})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetServiceSettingConfigurationByName(ctx, &settingssvc.GetServiceSettingConfigurationByNameRequest{ServiceSettingConfigurationName: nonexistentID})
		assert.Error(t, err)
	})
}

func TestServiceSettingConfigurations_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		_, err := adminClient.ArchiveServiceSettingConfiguration(ctx, &settingssvc.ArchiveServiceSettingConfigurationRequest{ServiceSettingConfigurationId: created.ID})
		require.NoError(t, err)

		x, err := adminClient.GetServiceSettingConfigurationByName(ctx, &settingssvc.GetServiceSettingConfigurationByNameRequest{ServiceSettingConfigurationName: created.ServiceSetting.Name})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveServiceSettingConfiguration(ctx, &settingssvc.ArchiveServiceSettingConfigurationRequest{ServiceSettingConfigurationId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveServiceSettingConfiguration(ctx, &settingssvc.ArchiveServiceSettingConfigurationRequest{ServiceSettingConfigurationId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestServiceSettingConfigurations_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		updateInput := fakes.BuildFakeServiceSettingConfigurationUpdateRequestInput()
		created.Update(updateInput)

		_, err := adminClient.UpdateServiceSettingConfiguration(ctx, &settingssvc.UpdateServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: created.ID,
			Input: &settingssvc.ServiceSettingConfigurationUpdateRequestInput{
				Value: updateInput.Value,
				Notes: updateInput.Notes,
			},
		})
		require.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createServiceSettingConfigurationForTest(t, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateServiceSettingConfiguration(ctx, &settingssvc.UpdateServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: created.ID,
			Input: &settingssvc.ServiceSettingConfigurationUpdateRequestInput{
				Value: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.UpdateServiceSettingConfiguration(ctx, &settingssvc.UpdateServiceSettingConfigurationRequest{
			ServiceSettingConfigurationId: nonexistentID,
			Input: &settingssvc.ServiceSettingConfigurationUpdateRequestInput{
				Value: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
	})
}

func TestServiceSettingConfigurations_Listing_ForUser(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdServiceSettingConfigurations := []*settings.ServiceSettingConfiguration{}
	for range exampleQuantity {
		created := createServiceSettingConfigurationForTest(T, testClient)
		createdServiceSettingConfigurations = append(createdServiceSettingConfigurations, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetServiceSettingConfigurationsForUser(ctx, &settingssvc.GetServiceSettingConfigurationsForUserRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdServiceSettingConfigurations))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetServiceSettingConfigurationsForUser(ctx, &settingssvc.GetServiceSettingConfigurationsForUserRequest{})
		assert.Error(t, err)
	})
}

func TestServiceSettingConfigurations_Listing_ForAccount(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdServiceSettingConfigurations := []*settings.ServiceSettingConfiguration{}
	for range exampleQuantity {
		created := createServiceSettingConfigurationForTest(T, testClient)
		createdServiceSettingConfigurations = append(createdServiceSettingConfigurations, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetServiceSettingConfigurationsForAccount(ctx, &settingssvc.GetServiceSettingConfigurationsForAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdServiceSettingConfigurations))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetServiceSettingConfigurationsForAccount(ctx, &settingssvc.GetServiceSettingConfigurationsForAccountRequest{})
		assert.Error(t, err)
	})
}
