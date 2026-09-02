package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v13/fake"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	settingsmock "github.com/primandproper/platform-go/v13/settings/mock"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

// buildTestService builds a service backed by the given store mock. A nil mock gets
// an unconfigured one, which panics if any of its methods are called — so a test that
// reaches the store where it did not arrange to fails loudly rather than on a nil
// result.
func buildTestService(t *testing.T, store *settingsmock.StoreMock) *serviceImpl {
	t.Helper()

	if store == nil {
		store = &settingsmock.StoreMock{}
	}

	return &serviceImpl{
		tracer:   tracing.NewTracerForTest(t.Name()),
		logger:   loggingnoop.NewLogger(),
		settings: store,
	}
}

// requester is who a test is making its requests as.
type requester struct {
	ctx     context.Context
	userID  string
	email   string
	account string
}

// userContextForTest returns an ordinary signed-in user: no service role beyond
// the default one, which is who reads the catalog and answers a setting about
// themselves.
func userContextForTest(t *testing.T) requester {
	t.Helper()

	return contextFor(t, authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil))
}

// adminContextForTest returns a service admin, which is who administers the
// catalog and may see the settings marked admin-only.
func adminContextForTest(t *testing.T) requester {
	t.Helper()

	return contextFor(t, authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceAdminRole.String()}, authorization.ServiceAdminPermissions))
}

func contextFor(t *testing.T, permissions authorization.ServiceRolePermissionChecker) requester {
	t.Helper()

	r := requester{
		userID:  fake.BuildFakeID(),
		email:   gofakeit.Email(),
		account: fake.BuildFakeID(),
	}

	r.ctx = sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: r.account,
		Requester: sessions.RequesterInfo{
			UserID:             r.userID,
			EmailAddress:       r.email,
			ServicePermissions: permissions,
		},
	})

	return r
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&settingsmock.StoreMock{},
		)

		assert.NotNil(t, service)
		assert.Implements(t, (*settingssvc.SettingsServiceServer)(nil), service)
	})
}

func TestProvideMethodPermissions(t *testing.T) {
	t.Parallel()

	t.Run("covers every method the service serves", func(t *testing.T) {
		t.Parallel()

		permissions := ProvideMethodPermissions()

		// A method with no entry is refused by the interceptor, so an RPC added
		// without one is an RPC nobody can call. Listing them here is what makes that
		// a test failure rather than a support ticket.
		for _, method := range []string{
			settingssvc.SettingsService_CreateSettingDefinition_FullMethodName,
			settingssvc.SettingsService_GetSettingDefinition_FullMethodName,
			settingssvc.SettingsService_GetSettingDefinitionByName_FullMethodName,
			settingssvc.SettingsService_GetSettingDefinitions_FullMethodName,
			settingssvc.SettingsService_UpdateSettingDefinition_FullMethodName,
			settingssvc.SettingsService_ArchiveSettingDefinition_FullMethodName,
			settingssvc.SettingsService_SetSettingValue_FullMethodName,
			settingssvc.SettingsService_GetSettingValue_FullMethodName,
			settingssvc.SettingsService_GetSettingValues_FullMethodName,
			settingssvc.SettingsService_GetSettingValuesForDefinition_FullMethodName,
			settingssvc.SettingsService_ClearSettingValue_FullMethodName,
			settingssvc.SettingsService_ResolveSetting_FullMethodName,
			settingssvc.SettingsService_ResolveSettings_FullMethodName,
		} {
			assert.NotEmpty(t, permissions[method], method)
		}
	})
}
