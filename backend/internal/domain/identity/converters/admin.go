package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
)

func ConvertGRPCAdminUpdateUserStatusRequestToUserAccountStatusUpdateInput(input *identitysvc.AdminUpdateUserStatusRequest) *identity.UserAccountStatusUpdateInput {
	return &identity.UserAccountStatusUpdateInput{
		NewStatus:    input.NewStatus,
		Reason:       input.Reason,
		TargetUserID: input.TargetUserId,
	}
}
