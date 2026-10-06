package signindevices

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/devices"

	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterSignInDevicesRepository registers the sign-in devices repository with the injector, as
// itself — the db-cleaner sweeps it — and as the store the sign-in hooks and annotator read.
func RegisterSignInDevicesRepository(i do.Injector) {
	do.Provide[*Repository](i, func(i do.Injector) (*Repository, error) {
		return ProvideSignInDevicesRepository(
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[database.Client](i),
		), nil
	})

	do.Provide[devices.Store](i, func(i do.Injector) (devices.Store, error) {
		return do.MustInvoke[*Repository](i), nil
	})
}
