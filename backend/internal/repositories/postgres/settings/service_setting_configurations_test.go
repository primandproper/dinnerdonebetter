package settings

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createServiceSettingConfigurationForTest(t *testing.T, ctx context.Context, exampleServiceSettingConfiguration *types.ServiceSettingConfiguration, dbc *Repository) *types.ServiceSettingConfiguration {
	t.Helper()

	// create
	if exampleServiceSettingConfiguration == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		generatedIdentity := generated.New()
		accountID, err := generatedIdentity.GetDefaultAccountIDForUser(ctx, dbc.writeDB, user.ID)
		require.NoError(t, err)

		serviceSetting := createServiceSettingForTest(t, ctx, nil, dbc)
		exampleServiceSettingConfiguration = fakes.BuildFakeServiceSettingConfiguration()
		exampleServiceSettingConfiguration.ServiceSetting = *serviceSetting
		exampleServiceSettingConfiguration.BelongsToUser = user.ID
		exampleServiceSettingConfiguration.BelongsToAccount = accountID
	}
	dbInput := converters.ConvertServiceSettingConfigurationToServiceSettingConfigurationDatabaseCreationInput(exampleServiceSettingConfiguration)

	created, err := dbc.CreateServiceSettingConfiguration(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)
	exampleServiceSettingConfiguration.CreatedAt = created.CreatedAt
	exampleServiceSettingConfiguration.ServiceSetting = created.ServiceSetting
	assert.Equal(t, exampleServiceSettingConfiguration, created)

	serviceSettingConfiguration, err := dbc.GetServiceSettingConfiguration(ctx, created.ID)
	require.NotNil(t, serviceSettingConfiguration)
	require.NoError(t, err)
	exampleServiceSettingConfiguration.CreatedAt = serviceSettingConfiguration.CreatedAt
	exampleServiceSettingConfiguration.ServiceSetting = serviceSettingConfiguration.ServiceSetting

	assert.NoError(t, err)
	assert.Equal(t, serviceSettingConfiguration, exampleServiceSettingConfiguration)

	return created
}

func TestQuerier_Integration_ServiceSettingConfigurations(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	serviceSetting := createServiceSettingForTest(t, ctx, nil, dbc)
	exampleServiceSettingConfiguration := fakes.BuildFakeServiceSettingConfiguration()
	exampleServiceSettingConfiguration.ServiceSetting = *serviceSetting
	exampleServiceSettingConfiguration.BelongsToUser = user.ID
	exampleServiceSettingConfiguration.BelongsToAccount = account.ID
	createdServiceSettingConfigurations := []*types.ServiceSettingConfiguration{}

	// create
	createdServiceSettingConfigurations = append(createdServiceSettingConfigurations, createServiceSettingConfigurationForTest(t, ctx, exampleServiceSettingConfiguration, dbc))

	// Assert audit log entry for create
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeServiceSettingConfigurations, RelevantID: createdServiceSettingConfigurations[0].ID},
	})

	// update
	createdServiceSettingConfigurations[0].Value = "new value"
	require.NoError(t, dbc.UpdateServiceSettingConfiguration(ctx, createdServiceSettingConfigurations[0]))

	// Assert audit log entry for update
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeServiceSettingConfigurations, RelevantID: createdServiceSettingConfigurations[0].ID},
	})

	// delete
	for _, serviceSettingConfiguration := range createdServiceSettingConfigurations {
		require.NoError(t, dbc.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfiguration.ID))

		exists, err := dbc.ServiceSettingConfigurationExists(ctx, serviceSettingConfiguration.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		y, err := dbc.GetServiceSettingConfiguration(ctx, serviceSettingConfiguration.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ServiceSettingConfigurationExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ServiceSettingConfigurationExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetServiceSettingConfiguration(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting configuration ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetServiceSettingConfiguration(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetServiceSettingConfigurationForUserByName(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUserID := fake.BuildFakeID()
		c := buildInertClientForTest(t)

		actual, err := c.GetServiceSettingConfigurationForUserByName(ctx, exampleUserID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetServiceSettingConfigurationForAccountByName(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleAccountID := fake.BuildFakeID()
		c := buildInertClientForTest(t)

		actual, err := c.GetServiceSettingConfigurationForAccountByName(ctx, exampleAccountID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetServiceSettingConfigurationsForUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetServiceSettingConfigurationsForUser(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateServiceSettingConfiguration(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateServiceSettingConfiguration(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateServiceSettingConfiguration(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateServiceSettingConfiguration(ctx, nil))
	})
}

func TestQuerier_ArchiveServiceSettingConfiguration(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveServiceSettingConfiguration(ctx, ""))
	})
}

func TestQuerier_Integration_ServiceSettingConfigurationsForUser_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create a user and account for testing
	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ServiceSettingConfiguration]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "service setting configuration",
		CreateItem: func(ctx context.Context, i int) *types.ServiceSettingConfiguration {
			// Create a unique service setting for each configuration to avoid unique constraint violations
			serviceSetting := fakes.BuildFakeServiceSetting()
			serviceSetting.Name = fmt.Sprintf("Service Setting %02d", i)
			createdServiceSetting := createServiceSettingForTest(t, ctx, serviceSetting, dbc)

			serviceSettingConfiguration := fakes.BuildFakeServiceSettingConfiguration()
			serviceSettingConfiguration.ServiceSetting = *createdServiceSetting
			serviceSettingConfiguration.BelongsToUser = user.ID
			serviceSettingConfiguration.BelongsToAccount = account.ID
			serviceSettingConfiguration.Value = fmt.Sprintf("Value %02d", i)
			return createServiceSettingConfigurationForTest(t, ctx, serviceSettingConfiguration, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSettingConfiguration], error) {
			return dbc.GetServiceSettingConfigurationsForUser(ctx, user.ID, filter)
		},
		GetID: func(serviceSettingConfiguration *types.ServiceSettingConfiguration) string {
			return serviceSettingConfiguration.ID
		},
		CleanupItem: func(ctx context.Context, serviceSettingConfiguration *types.ServiceSettingConfiguration) error {
			return dbc.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfiguration.ID)
		},
	})
}

func TestQuerier_Integration_ServiceSettingConfigurationsForAccount_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create a user and account for testing
	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ServiceSettingConfiguration]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "service setting configuration",
		CreateItem: func(ctx context.Context, i int) *types.ServiceSettingConfiguration {
			// Create a unique service setting for each configuration to avoid unique constraint violations
			serviceSetting := fakes.BuildFakeServiceSetting()
			serviceSetting.Name = fmt.Sprintf("Service Setting %02d", i)
			createdServiceSetting := createServiceSettingForTest(t, ctx, serviceSetting, dbc)

			serviceSettingConfiguration := fakes.BuildFakeServiceSettingConfiguration()
			serviceSettingConfiguration.ServiceSetting = *createdServiceSetting
			serviceSettingConfiguration.BelongsToUser = user.ID
			serviceSettingConfiguration.BelongsToAccount = account.ID
			serviceSettingConfiguration.Value = fmt.Sprintf("Value %02d", i)
			return createServiceSettingConfigurationForTest(t, ctx, serviceSettingConfiguration, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSettingConfiguration], error) {
			return dbc.GetServiceSettingConfigurationsForAccount(ctx, account.ID, filter)
		},
		GetID: func(serviceSettingConfiguration *types.ServiceSettingConfiguration) string {
			return serviceSettingConfiguration.ID
		},
		CleanupItem: func(ctx context.Context, serviceSettingConfiguration *types.ServiceSettingConfiguration) error {
			return dbc.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfiguration.ID)
		},
	})
}
