package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
)

// ConvertGRPCUserRegistrationInputToUserRegistrationInput renders a sign-up request as the
// input the auth manager takes.
func ConvertGRPCUserRegistrationInputToUserRegistrationInput(input *authsvc.UserRegistrationInput) *auth.UserRegistrationInput {
	if input == nil {
		return nil
	}

	return &auth.UserRegistrationInput{
		Password:              input.Password,
		EmailAddress:          input.EmailAddress,
		InvitationToken:       input.InvitationToken,
		InvitationID:          input.InvitationId,
		Username:              input.Username,
		FirstName:             input.FirstName,
		LastName:              input.LastName,
		AccountName:           input.AccountName,
		AcceptedTOS:           input.AcceptedTos,
		AcceptedPrivacyPolicy: input.AcceptedPrivacyPolicy,
	}
}

// ConvertUserRegistrationInputToGRPCUserRegistrationInput is the other direction, for a
// client assembling the request.
func ConvertUserRegistrationInputToGRPCUserRegistrationInput(input *auth.UserRegistrationInput) *authsvc.UserRegistrationInput {
	if input == nil {
		return nil
	}

	return &authsvc.UserRegistrationInput{
		Password:              input.Password,
		EmailAddress:          input.EmailAddress,
		InvitationToken:       input.InvitationToken,
		InvitationId:          input.InvitationID,
		Username:              input.Username,
		FirstName:             input.FirstName,
		LastName:              input.LastName,
		AccountName:           input.AccountName,
		AcceptedTos:           input.AcceptedTOS,
		AcceptedPrivacyPolicy: input.AcceptedPrivacyPolicy,
	}
}

// ConvertUserCreationResponseToGRPCUserCreationResponse renders what a registration produced.
func ConvertUserCreationResponseToGRPCUserCreationResponse(x *auth.UserCreationResponse) *authsvc.UserCreationResponse {
	if x == nil {
		return nil
	}

	return &authsvc.UserCreationResponse{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(x.CreatedAt),
		Username:         x.Username,
		EmailAddress:     x.EmailAddress,
		TwoFactorQrCode:  x.TwoFactorQRCode,
		CreatedUserId:    x.CreatedUserID,
		CreatedAccountId: x.CreatedAccountID,
		AccountStatus:    x.AccountStatus,
		TwoFactorSecret:  x.TwoFactorSecret,
		FirstName:        x.FirstName,
		LastName:         x.LastName,
	}
}
