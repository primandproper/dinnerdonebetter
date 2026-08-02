package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	platformconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
)

// ConvertUserDataDisclosureToGRPCUserDataDisclosure converts a domain UserDataDisclosure to a proto UserDataDisclosure.
func ConvertUserDataDisclosureToGRPCUserDataDisclosure(input *dataprivacy.UserDataDisclosure) *dataprivacysvc.UserDataDisclosure {
	if input == nil {
		return nil
	}

	return &dataprivacysvc.UserDataDisclosure{
		Id:            input.ID,
		BelongsToUser: input.BelongsToUser,
		Status:        string(input.Status),
		ReportId:      input.ReportID,
		CreatedAt:     platformconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		ExpiresAt:     platformconverters.ConvertTimeToPBTimestamp(input.ExpiresAt),
		LastUpdatedAt: platformconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		CompletedAt:   platformconverters.ConvertTimePointerToPBTimestamp(input.CompletedAt),
		ArchivedAt:    platformconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
	}
}
