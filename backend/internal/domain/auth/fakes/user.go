package fakes

import (
	"fmt"
	"log"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/pquerna/otp/totp"
)

// BuildFakeUserLoginInputFromUser builds a faked UserLoginInput.
func BuildFakeUserLoginInputFromUser(user *identity.User) *types.UserLoginInput {
	return &types.UserLoginInput{
		Username:  user.Username,
		Password:  buildFakePassword(),
		TOTPToken: fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakePasswordUpdateInput builds a faked PasswordUpdateInput.
func BuildFakePasswordUpdateInput() *types.PasswordUpdateInput {
	return &types.PasswordUpdateInput{
		NewPassword:     buildFakePassword(),
		CurrentPassword: buildFakePassword(),
		TOTPToken:       fmt.Sprintf("0%s", fake.Zip()),
	}
}

// BuildFakeTOTPSecretRefreshInput builds a faked TOTPSecretRefreshInput.
func BuildFakeTOTPSecretRefreshInput() *types.TOTPSecretRefreshInput {
	return &types.TOTPSecretRefreshInput{
		CurrentPassword: buildFakePassword(),
		TOTPToken:       fmt.Sprintf("0%s", fake.Zip()),
	}
}

func BuildFakeTOTPSecretRefreshResponse() *types.TOTPSecretRefreshResponse {
	return &types.TOTPSecretRefreshResponse{
		TwoFactorQRCode: fake.UUID(),
		TwoFactorSecret: fake.UUID(),
	}
}

// BuildFakeUserPermissionsRequestInput builds a faked UserPermissionsRequestInput.
func BuildFakeUserPermissionsRequestInput() *types.UserPermissionsRequestInput {
	return &types.UserPermissionsRequestInput{
		Permissions: []string{
			buildUniqueString(),
			buildUniqueString(),
			buildUniqueString(),
		},
	}
}

// BuildFakeTOTPSecretVerificationInput builds a faked TOTPSecretVerificationInput for a given user.
func BuildFakeTOTPSecretVerificationInput(user *identity.User) *types.TOTPSecretVerificationInput {
	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		log.Panicf("error generating TOTP token for fakes user: %v", err)
	}

	return &types.TOTPSecretVerificationInput{
		UserID:    user.ID,
		TOTPToken: token,
	}
}

// BuildFakePasswordResetToken builds a faked PasswordResetToken.
func BuildFakePasswordResetToken() *types.PasswordResetToken {
	fakeDate := BuildFakeTime()

	return &types.PasswordResetToken{
		ID:            BuildFakeID(),
		Token:         fake.UUID(),
		BelongsToUser: BuildFakeID(),
		ExpiresAt:     fakeDate.Add(30 * time.Minute),
		CreatedAt:     fakeDate,
	}
}

// BuildFakeUsernameReminderRequestInput builds a faked UsernameReminderRequestInput.
func BuildFakeUsernameReminderRequestInput() *types.UsernameReminderRequestInput {
	return &types.UsernameReminderRequestInput{
		EmailAddress: fake.Email(),
	}
}

// BuildFakePasswordResetTokenCreationRequestInput builds a faked PasswordResetTokenCreationRequestInput.
func BuildFakePasswordResetTokenCreationRequestInput() *types.PasswordResetTokenCreationRequestInput {
	return &types.PasswordResetTokenCreationRequestInput{EmailAddress: fake.Email()}
}

// BuildFakePasswordResetTokenRedemptionRequestInput builds a faked PasswordResetTokenRedemptionRequestInput.
func BuildFakePasswordResetTokenRedemptionRequestInput() *types.PasswordResetTokenRedemptionRequestInput {
	return &types.PasswordResetTokenRedemptionRequestInput{
		Token:       buildUniqueString(),
		NewPassword: buildFakePassword(),
	}
}

// BuildFakeEmailAddressVerificationRequestInput builds a faked EmailAddressVerificationRequestInput.
func BuildFakeEmailAddressVerificationRequestInput() *types.EmailAddressVerificationRequestInput {
	return &types.EmailAddressVerificationRequestInput{
		Token: buildUniqueString(),
	}
}

func BuildFakeUsernameUpdateInput() *types.UsernameUpdateInput {
	return &types.UsernameUpdateInput{
		NewUsername:     buildUniqueString(),
		CurrentPassword: fake.Password(true, true, true, false, false, 32),
		TOTPToken:       "123456",
	}
}

func BuildFakeUserEmailAddressUpdateInput() *types.UserEmailAddressUpdateInput {
	return &types.UserEmailAddressUpdateInput{
		NewEmailAddress: fake.Email(),
		CurrentPassword: fake.Password(true, true, true, false, false, 32),
		TOTPToken:       "123456",
	}
}

func BuildFakePasswordResetResponse() *types.PasswordResetResponse {
	return &types.PasswordResetResponse{Successful: true}
}

func BuildFakeUserPermissionsResponse() *types.UserPermissionsResponse {
	return &types.UserPermissionsResponse{
		Permissions: map[string]bool{
			string(authorization.CreateWebhooksPermission): true,
		},
	}
}
