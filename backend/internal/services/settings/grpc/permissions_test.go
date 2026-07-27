package grpc

import (
	"reflect"
	"testing"

	settingssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v7/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideMethodPermissions_CoverageAndValidity(t *testing.T) {
	t.Parallel()

	t.Run("all permission map keys correspond to valid methods", func(t *testing.T) {
		t.Parallel()

		methodPerms := ProvideMethodPermissions()

		// Get all methods from the interface
		interfaceType := reflect.TypeFor[settingssvc.SettingsServiceServer]()
		validMethods := make(map[string]bool)

		for method := range interfaceType.Methods() {
			m := method
			// Build the full method name format: /settings.SettingsService/MethodName
			fullMethodName := "/settings.SettingsService/" + m.Name
			validMethods[fullMethodName] = true
		}

		// Verify all keys in the permission map are valid methods
		for methodName := range methodPerms {
			assert.True(t, validMethods[methodName], "permission map contains unknown method: %s", methodName)
		}
	})

	t.Run("all permission entries are non-empty", func(t *testing.T) {
		t.Parallel()

		methodPerms := ProvideMethodPermissions()

		for methodName, perms := range methodPerms {
			assert.NotEmpty(t, perms, "method %s has empty permissions", methodName)
		}
	})
}

func TestProvideMethodPermissions_ReflectionBasedVerification(t *testing.T) {
	t.Parallel()

	t.Run("GetMethodName works for service methods", func(t *testing.T) {
		t.Parallel()

		// Create a minimal service instance to test reflection
		s := &serviceImpl{}

		// Verify the reflection utility works with our service methods
		methodName := reflection.GetMethodName(s.CreateServiceSetting)
		require.Equal(t, "CreateServiceSetting", methodName)

		methodName = reflection.GetMethodName(s.GetServiceSetting)
		require.Equal(t, "GetServiceSetting", methodName)

		methodName = reflection.GetMethodName(s.ArchiveServiceSetting)
		require.Equal(t, "ArchiveServiceSetting", methodName)
	})
}
