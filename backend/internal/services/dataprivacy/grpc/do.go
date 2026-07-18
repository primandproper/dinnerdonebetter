package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	dataprivacymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/manager"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"

	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/uploads"

	"github.com/samber/do/v2"
)

// RegisterDataPrivacyService registers the data privacy gRPC service with the injector.
func RegisterDataPrivacyService(i do.Injector) {
	do.Provide[dataprivacysvc.DataPrivacyServiceServer](i, func(i do.Injector) (dataprivacysvc.DataPrivacyServiceServer, error) {
		return NewDataPrivacyService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[func(context.Context) (*sessions.ContextData, error)](i),
			do.MustInvoke[dataprivacymanager.DataPrivacyManager](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
