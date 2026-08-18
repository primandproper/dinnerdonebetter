package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/qrcodes"
	"github.com/primandproper/platform-go/v11/random"

	"github.com/samber/do/v2"
)

// RegisterIdentityDataManager registers the identity data manager with the injector.
func RegisterIdentityDataManager(i do.Injector) {
	do.Provide[IdentityDataManager](i, func(i do.Injector) (IdentityDataManager, error) {
		return NewIdentityDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[random.Generator](i),
			do.MustInvoke[authentication.Hasher](i),
			do.MustInvoke[identityindexing.UserTextSearcher](i),
			do.MustInvoke[qrcodes.Builder](i),
		)
	})
}
