package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	"github.com/primandproper/platform-go/v11/identifiers"
)

func ConvertWaitlistToGRPCWaitlist(waitlist *types.Waitlist) *waitlistssvc.Waitlist {
	return &waitlistssvc.Waitlist{
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(waitlist.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(waitlist.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(waitlist.ArchivedAt),
		Id:            waitlist.ID,
		Name:          waitlist.Name,
		Description:   waitlist.Description,
		ValidUntil:    grpcconverters.ConvertTimeToPBTimestamp(waitlist.ValidUntil),
	}
}

func ConvertGRPCWaitlistToWaitlist(waitlist *waitlistssvc.Waitlist) *types.Waitlist {
	return &types.Waitlist{
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(waitlist.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(waitlist.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(waitlist.ArchivedAt),
		ID:            waitlist.Id,
		Name:          waitlist.Name,
		Description:   waitlist.Description,
		ValidUntil:    grpcconverters.ConvertPBTimestampToTime(waitlist.ValidUntil),
	}
}

func ConvertWaitlistSignupToGRPCWaitlistSignup(signup *types.WaitlistSignup) *waitlistssvc.WaitlistSignup {
	return &waitlistssvc.WaitlistSignup{
		CreatedAt:         grpcconverters.ConvertTimeToPBTimestamp(signup.CreatedAt),
		LastUpdatedAt:     grpcconverters.ConvertTimePointerToPBTimestamp(signup.LastUpdatedAt),
		ArchivedAt:        grpcconverters.ConvertTimePointerToPBTimestamp(signup.ArchivedAt),
		Id:                signup.ID,
		Notes:             signup.Notes,
		BelongsToWaitlist: signup.BelongsToWaitlist,
		BelongsToUser:     signup.BelongsToUser,
		BelongsToAccount:  signup.BelongsToAccount,
	}
}

func ConvertGRPCWaitlistSignupToWaitlistSignup(signup *waitlistssvc.WaitlistSignup) *types.WaitlistSignup {
	return &types.WaitlistSignup{
		CreatedAt:         grpcconverters.ConvertPBTimestampToTime(signup.CreatedAt),
		LastUpdatedAt:     grpcconverters.ConvertPBTimestampToTimePointer(signup.LastUpdatedAt),
		ArchivedAt:        grpcconverters.ConvertPBTimestampToTimePointer(signup.ArchivedAt),
		ID:                signup.Id,
		Notes:             signup.Notes,
		BelongsToWaitlist: signup.BelongsToWaitlist,
		BelongsToUser:     signup.BelongsToUser,
		BelongsToAccount:  signup.BelongsToAccount,
	}
}

func ConvertGRPCWaitlistCreationRequestInputToWaitlistDatabaseCreationInput(input *waitlistssvc.WaitlistCreationRequestInput) *types.WaitlistDatabaseCreationInput {
	return &types.WaitlistDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        input.Name,
		Description: input.Description,
		ValidUntil:  grpcconverters.ConvertPBTimestampToTime(input.ValidUntil),
	}
}

func ConvertGRPCWaitlistSignupCreationRequestInputToWaitlistSignupDatabaseCreationInput(input *waitlistssvc.WaitlistSignupCreationRequestInput) *types.WaitlistSignupDatabaseCreationInput {
	return &types.WaitlistSignupDatabaseCreationInput{
		ID:                identifiers.New(),
		Notes:             input.Notes,
		BelongsToWaitlist: input.BelongsToWaitlist,
		BelongsToUser:     input.BelongsToUser,
		BelongsToAccount:  input.BelongsToAccount,
	}
}

func ConvertGRPCWaitlistUpdateRequestInputToWaitlistUpdateRequestInput(input *waitlistssvc.WaitlistUpdateRequestInput) *types.WaitlistUpdateRequestInput {
	result := &types.WaitlistUpdateRequestInput{}

	if input.Name != nil {
		result.Name = input.Name
	}
	if input.Description != nil {
		result.Description = input.Description
	}
	if input.ValidUntil != nil {
		validUntil := grpcconverters.ConvertPBTimestampToTime(input.ValidUntil)
		result.ValidUntil = &validUntil
	}

	return result
}

func ConvertGRPCWaitlistSignupUpdateRequestInputToWaitlistSignupUpdateRequestInput(input *waitlistssvc.WaitlistSignupUpdateRequestInput) *types.WaitlistSignupUpdateRequestInput {
	result := &types.WaitlistSignupUpdateRequestInput{}

	if input.Notes != nil {
		result.Notes = input.Notes
	}

	return result
}
