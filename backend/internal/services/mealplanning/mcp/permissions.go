package mcp

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	mealplanningsvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrUndeclaredTool is a tool whose name is not a method of the gRPC surface, and so has no
// entry in its permission table. It is refused rather than admitted: a tool nobody declared
// permissions for is a tool nobody decided was safe to expose, which is the rule the API's
// interceptor applies to a method.
var ErrUndeclaredTool = platformerrors.New("the tool has no gRPC counterpart to take its permissions from")

// methodPermissions is the gRPC surface's permission table, read once.
var methodPermissions = mealplanninggrpc.ProvideMethodPermissions()

// permissionsFor is the grants a tool requires: the ones the gRPC surface declares for the
// method of the same name. Every tool is named for its counterpart, so the table is read
// rather than copied, and a permission changed there changes here.
func permissionsFor(tool *sdkmcp.Tool) ([]authorization.Permission, error) {
	required, ok := methodPermissions["/"+mealplanningsvcpb.MealPlanningService_ServiceDesc.ServiceName+"/"+tool.Name]
	if !ok {
		return nil, platformerrors.Wrap(ErrUndeclaredTool, tool.Name)
	}

	return required, nil
}
