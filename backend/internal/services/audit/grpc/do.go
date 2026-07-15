package grpc

import (
	auditmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/manager"
	auditsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/audit"

	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditService registers the audit gRPC service with the injector.
func RegisterAuditService(i do.Injector) {
	do.Provide[AuditMethodPermissions](i, func(i do.Injector) (AuditMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[auditsvc.AuditServiceServer](i, func(i do.Injector) (auditsvc.AuditServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[auditmanager.AuditDataManager](i),
		), nil
	})
}
