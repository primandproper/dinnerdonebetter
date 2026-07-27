package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	uploadedmediamanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	"github.com/primandproper/platform-go/v7/uploads"

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
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[func(context.Context) (*sessions.ContextData, error)](i),
			do.MustInvoke[manager.IdentityDataManager](i),
			do.MustInvoke[uploadedmediamanager.UploadedMediaManager](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
