package grpc

import (
	analyticspb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"

	"github.com/primandproper/platform-go/v11/analytics/multisource"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAnalyticsService registers the analytics gRPC service with the injector.
func RegisterAnalyticsService(i do.Injector) {
	do.Provide[AnalyticsMethodPermissions](i, func(i do.Injector) (AnalyticsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[analyticspb.AnalyticsServiceServer](i, func(i do.Injector) (analyticspb.AnalyticsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*multisource.MultiSourceEventReporter](i),
		), nil
	})
}
