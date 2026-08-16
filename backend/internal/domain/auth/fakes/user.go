package fakes

import (
	"fmt"
	"log"
	"time"

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
