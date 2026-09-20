package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"

	"github.com/samber/do/v2"
)

// RegisterIdentityService registers the identity gRPC service with the injector.
func RegisterIdentityService(i do.Injector) {
	do.Provide[IdentityMethodPermissions](i, func(i do.Injector) (IdentityMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[identitysvc.IdentityServiceServer](i, func(i do.Injector) (identitysvc.IdentityServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[manager.IdentityDataManager](i),
			do.MustInvoke[mediaregistry.Store](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
