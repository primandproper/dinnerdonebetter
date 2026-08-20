package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v12/fake"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeSessionContextData builds a faked ContextData.
func BuildFakeSessionContextData() *sessions.ContextData {
	return &sessions.ContextData{
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{},
		Requester: sessions.RequesterInfo{
			ServicePermissions:       nil,
			AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
			AccountStatusExplanation: "fake",
			UserID:                   fake.BuildFakeID(),
			EmailAddress:             gofakeit.Email(),
			Username:                 fake.BuildFakeString(),
		},
		ActiveAccountID: fake.BuildFakeID(),
	}
}

// BuildFakeChangeActiveAccountInput builds a faked ChangeActiveAccountInput.
func BuildFakeChangeActiveAccountInput() *auth.ChangeActiveAccountInput {
	return &auth.ChangeActiveAccountInput{
		AccountID: gofakeit.UUID(),
	}
}

// BuildFakeUserStatusResponse builds a faked UserStatusResponse.
func BuildFakeUserStatusResponse() *auth.UserStatusResponse {
	return &auth.UserStatusResponse{
		UserID:                   fake.BuildFakeID(),
		AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
		AccountStatusExplanation: "",
		ActiveAccount:            fake.BuildFakeID(),
		UserIsAuthenticated:      true,
	}
}

// BuildFakeTokenResponse builds a faked TokenResponse.
func BuildFakeTokenResponse() *auth.TokenResponse {
	return &auth.TokenResponse{
		UserID:      fake.BuildFakeID(),
		AccountID:   fake.BuildFakeID(),
		AccessToken: gofakeit.UUID(),
	}
}

func BuildFakeUserLoginInput() *auth.UserLoginInput {
	return &auth.UserLoginInput{
		Username:  fake.BuildFakeID(),
		Password:  fake.BuildFakePassword(),
		TOTPToken: buildFakeTOTPToken(),
	}
}
