package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	analyticspb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"
)

// AnalyticsMethodPermissions is a named type for Wire dependency injection.
type AnalyticsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the analytics service's method permissions.
func ProvideMethodPermissions() AnalyticsMethodPermissions {
	return AnalyticsMethodPermissions{
		analyticspb.AnalyticsService_TrackEvent_FullMethodName: {
			authorization.ReportAnalyticsEventsPermission,
		},
		// TrackAnonymousEvent has no permissions (unauthenticated route)
	}
}
