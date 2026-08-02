package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
)

// InternalOpsMethodPermissions is a named type for Wire dependency injection.
type InternalOpsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the internal ops service's method permissions.
func ProvideMethodPermissions() InternalOpsMethodPermissions {
	return InternalOpsMethodPermissions{
		internalopssvc.InternalOperations_TestQueueMessage_FullMethodName: {
			authorization.PublishArbitraryQueueMessagePermission,
		},
	}
}
