package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	uploadedmediamanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	"github.com/primandproper/platform-go/v12/uploads"

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
			do.MustInvoke[manager.IdentityDataManager](i),
			do.MustInvoke[uploadedmediamanager.UploadedMediaManager](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
