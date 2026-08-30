package fakes

import (
	"fmt"
	"log"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v13/fake"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/pquerna/otp/totp"
)

// BuildFakeUserLoginInputFromUser builds a faked UserLoginInput.
func BuildFakeUserLoginInputFromUser(user *identity.User) *types.UserLoginInput {
	return &types.UserLoginInput{
		Username:  user.Username,
		Password:  fake.BuildFakePassword(),
		TOTPToken: fmt.Sprintf("0%s", gofakeit.Zip()),
	}
}

// BuildFakePasswordUpdateInput builds a faked PasswordUpdateInput.
func BuildFakePasswordUpdateInput() *types.PasswordUpdateInput {
	return &types.PasswordUpdateInput{
		NewPassword:     fake.BuildFakePassword(),
		CurrentPassword: fake.BuildFakePassword(),
		TOTPToken:       fmt.Sprintf("0%s", gofakeit.Zip()),
	}
}

// BuildFakeTOTPSecretRefreshInput builds a faked TOTPSecretRefreshInput.
func BuildFakeTOTPSecretRefreshInput() *types.TOTPSecretRefreshInput {
	return &types.TOTPSecretRefreshInput{
		CurrentPassword: fake.BuildFakePassword(),
		TOTPToken:       fmt.Sprintf("0%s", gofakeit.Zip()),
	}
}

func BuildFakeTOTPSecretRefreshResponse() *types.TOTPSecretRefreshResponse {
	return &types.TOTPSecretRefreshResponse{
		TwoFactorQRCode: fake.BuildFakeString(),
		TwoFactorSecret: fake.BuildFakeString(),
	}
}

// BuildFakeUserPermissionsRequestInput builds a faked UserPermissionsRequestInput.
func BuildFakeUserPermissionsRequestInput() *types.UserPermissionsRequestInput {
	return &types.UserPermissionsRequestInput{
		Permissions: []string{
			fake.BuildFakeString(),
			fake.BuildFakeString(),
			fake.BuildFakeString(),
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
	token := fake.BuildFakeRecord[types.PasswordResetToken]()

	token.Token = fake.BuildFakeString()

	// A window rather than two arbitrary instants: a token whose expiry faker picked
	// has an even chance of having expired before the test redeems it.
	token.ExpiresAt = token.CreatedAt.Add(30 * time.Minute)

	return token
}

// BuildFakeUsernameReminderRequestInput builds a faked UsernameReminderRequestInput.
func BuildFakeUsernameReminderRequestInput() *types.UsernameReminderRequestInput {
	return &types.UsernameReminderRequestInput{
		EmailAddress: gofakeit.Email(),
	}
}

// BuildFakePasswordResetTokenCreationRequestInput builds a faked PasswordResetTokenCreationRequestInput.
func BuildFakePasswordResetTokenCreationRequestInput() *types.PasswordResetTokenCreationRequestInput {
	return &types.PasswordResetTokenCreationRequestInput{EmailAddress: gofakeit.Email()}
}

// BuildFakePasswordResetTokenRedemptionRequestInput builds a faked PasswordResetTokenRedemptionRequestInput.
func BuildFakePasswordResetTokenRedemptionRequestInput() *types.PasswordResetTokenRedemptionRequestInput {
	return &types.PasswordResetTokenRedemptionRequestInput{
		Token:       fake.BuildFakeString(),
		NewPassword: fake.BuildFakePassword(),
	}
}

// BuildFakeEmailAddressVerificationRequestInput builds a faked EmailAddressVerificationRequestInput.
func BuildFakeEmailAddressVerificationRequestInput() *types.EmailAddressVerificationRequestInput {
	return &types.EmailAddressVerificationRequestInput{
		Token: fake.BuildFakeString(),
	}
}

func BuildFakeUsernameUpdateInput() *types.UsernameUpdateInput {
	return &types.UsernameUpdateInput{
		NewUsername:     fake.BuildFakeString(),
		CurrentPassword: gofakeit.Password(true, true, true, false, false, 32),
		TOTPToken:       "123456",
	}
}

func BuildFakeUserEmailAddressUpdateInput() *types.UserEmailAddressUpdateInput {
	return &types.UserEmailAddressUpdateInput{
		NewEmailAddress: gofakeit.Email(),
		CurrentPassword: gofakeit.Password(true, true, true, false, false, 32),
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
