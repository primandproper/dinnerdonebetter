package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
)

// DataPrivacyMethodPermissions is a named type for Wire dependency injection.
// It allows Wire to distinguish between different services' permission maps.
type DataPrivacyMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the data privacy service's method permissions.
func ProvideMethodPermissions() DataPrivacyMethodPermissions {
	return DataPrivacyMethodPermissions{
		dataprivacysvc.DataPrivacyService_AggregateUserDataReport_FullMethodName: {
			authorization.CreateUserDataReportsPermission,
		},
		dataprivacysvc.DataPrivacyService_FetchUserDataReport_FullMethodName: {
			authorization.ReadUserDataReportsPermission,
		},
		dataprivacysvc.DataPrivacyService_DestroyAllUserData_FullMethodName: {
			authorization.DestroyUserDataPermission,
		},
		dataprivacysvc.DataPrivacyService_GetDataPrivacyRequest_FullMethodName: {
			authorization.ReadUserDataReportsPermission,
		},
		dataprivacysvc.DataPrivacyService_ListDataPrivacyRequests_FullMethodName: {
			authorization.ReadUserDataReportsPermission,
		},
	}
}
