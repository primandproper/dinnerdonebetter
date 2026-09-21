package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
)

func ConvertGRPCUpdatePasswordRequestToPasswordUpdateInput(request *authsvc.UpdatePasswordRequest) *auth.PasswordUpdateInput {
	return &auth.PasswordUpdateInput{
		NewPassword:     request.NewPassword,
		CurrentPassword: request.CurrentPassword,
		TOTPToken:       request.TotpToken,
	}
}

func ConvertGRPCVerifyTOTPSecretRequestToTOTPSecretVerificationInput(request *authsvc.VerifyTOTPSecretRequest) *auth.TOTPSecretVerificationInput {
	return &auth.TOTPSecretVerificationInput{
		TOTPToken: request.TotpToken,
		UserID:    request.UserId,
	}
}

func ConvertGRPCVerifyEmailAddressRequestToEmailAddressVerificationRequestInput(request *authsvc.VerifyEmailAddressRequest) *auth.EmailAddressVerificationRequestInput {
	return &auth.EmailAddressVerificationRequestInput{
		Token: request.Token,
	}
}

func ConvertGRPCRequestUsernameReminderRequestToUsernameReminderRequestInput(request *authsvc.RequestUsernameReminderRequest) *auth.UsernameReminderRequestInput {
	return &auth.UsernameReminderRequestInput{
		EmailAddress: request.EmailAddress,
	}
}

func ConvertGRPCRequestPasswordResetTokenRequestToPasswordResetTokenCreationRequestInput(request *authsvc.RequestPasswordResetTokenRequest) *auth.PasswordResetTokenCreationRequestInput {
	return &auth.PasswordResetTokenCreationRequestInput{EmailAddress: request.EmailAddress}
}

func ConvertGRPCRefreshTOTPSecretRequestToTOTPSecretRefreshInput(request *authsvc.RefreshTOTPSecretRequest) *auth.TOTPSecretRefreshInput {
	return &auth.TOTPSecretRefreshInput{
		CurrentPassword: request.CurrentPassword,
		TOTPToken:       request.TotpToken,
	}
}

func ConvertTOTPSecretRefreshResponseToGRPCTOTPSecretRefreshResponse(input *auth.TOTPSecretRefreshResponse) *authsvc.TOTPSecretRefreshResponse {
	return &authsvc.TOTPSecretRefreshResponse{
		TwoFactorQrCode: input.TwoFactorQRCode,
		TwoFactorSecret: input.TwoFactorSecret,
	}
}

func ConvertGRPCRedeemPasswordResetTokenRequestToPasswordResetTokenRedemptionRequestInput(request *authsvc.RedeemPasswordResetTokenRequest) *auth.PasswordResetTokenRedemptionRequestInput {
	return &auth.PasswordResetTokenRedemptionRequestInput{
		Token:       request.Token,
		NewPassword: request.NewPassword,
	}
}

func ConvertGRPCCheckPermissionsRequestToUserPermissionsRequestInput(request *authsvc.UserPermissionsRequestInput) *auth.UserPermissionsRequestInput {
	return &auth.UserPermissionsRequestInput{
		Permissions: request.Permissions,
	}
}

func ConvertGRPCUserLoginInputToUserLoginInput(request *authsvc.UserLoginInput) *auth.UserLoginInput {
	if request == nil {
		return &auth.UserLoginInput{}
	}
	return &auth.UserLoginInput{
		Username:         request.GetUsername(),
		Password:         request.GetPassword(),
		TOTPToken:        request.GetTotpToken(),
		DesiredAccountID: request.GetDesiredAccountId(),
	}
}

func ConvertTokenResponseToGRPCTokenResponse(input *auth.TokenResponse) *authsvc.TokenResponse {
	return &authsvc.TokenResponse{
		UserId:       input.UserID,
		AccountId:    input.AccountID,
		AccessToken:  input.AccessToken,
		RefreshToken: input.RefreshToken,
		ExpiresUtc:   grpcconverters.ConvertTimeToPBTimestamp(input.ExpiresUTC),
	}
}

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
