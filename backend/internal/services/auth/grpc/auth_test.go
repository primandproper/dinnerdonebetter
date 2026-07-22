package grpc

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v5/featureflags"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func buildFakeSessionContextData() *sessions.ContextData {
	return &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   identityfakes.BuildFakeID(),
			AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
			AccountStatusExplanation: "",
			ServicePermissions:       authorization.NewServiceRolePermissionChecker([]string{"service_user"}, nil),
		},
		ActiveAccountID: identityfakes.BuildFakeID(),
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			identityfakes.BuildFakeID(): authorization.NewAccountRolePermissionChecker(nil),
		},
	}
}

func buildContextWithSessionData(t *testing.T) context.Context {
	t.Helper()
	sessionData := buildFakeSessionContextData()
	sessionData.AccountPermissions[sessionData.ActiveAccountID] = authorization.NewAccountRolePermissionChecker(nil)
	return context.WithValue(t.Context(), sessions.SessionContextDataKey, sessionData)
}

func TestServiceImpl_GetAuthStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		identityDataManager.On("UserRequiresPasswordChange", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

		request := &authsvc.GetAuthStatusRequest{}

		response, err := service.GetAuthStatus(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotEmpty(t, response.UserId)
		assert.NotEmpty(t, response.ActiveAccount)
		assert.Equal(t, identity.GoodStandingUserAccountStatus.String(), response.AccountStatus)
		assert.False(t, response.RequiresPasswordChange)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.GetAuthStatusRequest{}

		response, err := service.GetAuthStatus(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})
}

func TestServiceImpl_EvaluateBooleanFeatureFlag(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, featureFlagManager := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		featureFlagManager.CanUseFeatureFunc = func(_ context.Context, _ string, _ featureflags.EvaluationContext) (bool, error) {
			return true, nil
		}

		response, err := service.EvaluateBooleanFeatureFlag(ctx, &authsvc.EvaluateBooleanFeatureFlagRequest{
			FeatureFlag: "test-flag",
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Enabled)
	})

	t.Run("error unauthenticated", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context()

		response, err := service.EvaluateBooleanFeatureFlag(ctx, &authsvc.EvaluateBooleanFeatureFlagRequest{
			FeatureFlag: "test-flag",
		})

		assert.Error(t, err)
		assert.Nil(t, response)
		grpcErr, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error feature_flag required", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		response, err := service.EvaluateBooleanFeatureFlag(ctx, &authsvc.EvaluateBooleanFeatureFlagRequest{
			FeatureFlag: "",
		})

		assert.Error(t, err)
		assert.Nil(t, response)
		grpcErr, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, grpcErr.Code())
	})
}

func TestServiceImpl_EvaluateStringFeatureFlag(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, featureFlagManager := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		featureFlagManager.GetStringValueFunc = func(_ context.Context, _ string, _ string, _ featureflags.EvaluationContext) (string, error) {
			return "variant-a", nil
		}

		response, err := service.EvaluateStringFeatureFlag(ctx, &authsvc.EvaluateStringFeatureFlagRequest{
			FeatureFlag: "test-flag",
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "variant-a", response.Value)
	})
}

func TestServiceImpl_EvaluateInt64FeatureFlag(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, featureFlagManager := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		featureFlagManager.GetInt64ValueFunc = func(_ context.Context, _ string, _ int64, _ featureflags.EvaluationContext) (int64, error) {
			return int64(42), nil
		}

		response, err := service.EvaluateInt64FeatureFlag(ctx, &authsvc.EvaluateInt64FeatureFlagRequest{
			FeatureFlag: "test-flag",
		})

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, int64(42), response.Value)
	})
}

func TestServiceImpl_ExchangeToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeTokenResponse := &auth.TokenResponse{
			UserID:       identityfakes.BuildFakeID(),
			AccountID:    identityfakes.BuildFakeID(),
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresUTC:   time.Now().Add(time.Hour),
		}

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ExchangeTokenForUser), mock.Anything, "refresh-token", mock.Anything).Return(fakeTokenResponse, nil)

		request := &authsvc.ExchangeTokenRequest{
			RefreshToken: "refresh-token",
		}

		response, err := service.ExchangeToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.Equal(t, fakeTokenResponse.AccessToken, response.AccessToken)
		assert.Equal(t, fakeTokenResponse.RefreshToken, response.RefreshToken)
		assert.NotEmpty(t, response.UserId)
		assert.NotEmpty(t, response.AccountId)

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.ExchangeTokenRequest{
			RefreshToken: "refresh-token",
		}

		response, err := service.ExchangeToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error exchanging token", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ExchangeTokenForUser), mock.Anything, "refresh-token", mock.Anything).Return((*auth.TokenResponse)(nil), errors.New("exchange failed"))

		request := &authsvc.ExchangeTokenRequest{
			RefreshToken: "refresh-token",
		}

		response, err := service.ExchangeToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})
}

func TestServiceImpl_LoginForToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager, _ := buildTestService(t)
		ctx := t.Context()

		fakeTokenResponse := &auth.TokenResponse{
			UserID:       identityfakes.BuildFakeID(),
			AccountID:    identityfakes.BuildFakeID(),
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresUTC:   time.Now().Add(time.Hour),
		}

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, false, mock.AnythingOfType("*auth.UserLoginInput"), mock.AnythingOfType("*authentication.LoginMetadata")).Return(fakeTokenResponse, nil)

		request := &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "testuser",
				Password:  "password123",
				TotpToken: "123456",
			},
		}

		response, err := service.LoginForToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeTokenResponse.AccessToken, response.Result.AccessToken)
		assert.Equal(t, fakeTokenResponse.RefreshToken, response.Result.RefreshToken)

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})

	t.Run("error processing login", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager, _ := buildTestService(t)
		ctx := t.Context()

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, false, mock.AnythingOfType("*auth.UserLoginInput"), mock.AnythingOfType("*authentication.LoginMetadata")).Return((*auth.TokenResponse)(nil), errors.New("login failed"))

		request := &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "testuser",
				Password:  "wrongpassword",
				TotpToken: "123456",
			},
		}

		response, err := service.LoginForToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		// Login failures return a single generic Unauthenticated status so distinct outcomes
		// (unknown user, wrong password, banned account) can't be used to enumerate accounts.
		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})
}

func TestServiceImpl_AdminLoginForToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager, _ := buildTestService(t)
		ctx := t.Context()

		fakeTokenResponse := &auth.TokenResponse{
			UserID:       identityfakes.BuildFakeID(),
			AccountID:    identityfakes.BuildFakeID(),
			AccessToken:  "admin-access-token",
			RefreshToken: "admin-refresh-token",
			ExpiresUTC:   time.Now().Add(time.Hour),
		}

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, true, mock.AnythingOfType("*auth.UserLoginInput"), mock.AnythingOfType("*authentication.LoginMetadata")).Return(fakeTokenResponse, nil)

		request := &authsvc.AdminLoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "adminuser",
				Password:  "adminpassword123",
				TotpToken: "123456",
			},
		}

		response, err := service.AdminLoginForToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeTokenResponse.AccessToken, response.Result.AccessToken)
		assert.Equal(t, fakeTokenResponse.RefreshToken, response.Result.RefreshToken)

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})

	t.Run("error processing admin login", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager, _ := buildTestService(t)
		ctx := t.Context()

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, true, mock.AnythingOfType("*auth.UserLoginInput"), mock.AnythingOfType("*authentication.LoginMetadata")).Return((*auth.TokenResponse)(nil), errors.New("admin login failed"))

		request := &authsvc.AdminLoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "adminuser",
				Password:  "wrongpassword",
				TotpToken: "123456",
			},
		}

		response, err := service.AdminLoginForToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		// Login failures return a single generic Unauthenticated status so distinct outcomes
		// (unknown user, wrong password, banned account) can't be used to enumerate accounts.
		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})
}

func TestServiceImpl_CheckPermissions(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakePermissionsResponse := &auth.UserPermissionsResponse{
			Permissions: map[string]bool{
				"read_users":   true,
				"create_users": false,
			},
		}

		authManager.On(reflection.GetMethodName(authManager.CheckUserPermissions), mock.Anything, mock.AnythingOfType("*auth.UserPermissionsRequestInput")).Return(fakePermissionsResponse, nil)

		request := &authsvc.UserPermissionsRequestInput{
			Permissions: []string{"read_users", "create_users"},
		}

		response, err := service.CheckPermissions(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.Equal(t, fakePermissionsResponse.Permissions, response.Permissions)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.UserPermissionsRequestInput{
			Permissions: []string{"read_users"},
		}

		response, err := service.CheckPermissions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error checking permissions", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.CheckUserPermissions), mock.Anything, mock.AnythingOfType("*auth.UserPermissionsRequestInput")).Return((*auth.UserPermissionsResponse)(nil), errors.New("permission check failed"))

		request := &authsvc.UserPermissionsRequestInput{
			Permissions: []string{"read_users"},
		}

		response, err := service.CheckPermissions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_GetActiveAccount(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeAccount := identityfakes.BuildFakeAccount()

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, sessionData.GetActiveAccountID()).Return(fakeAccount, nil)

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeAccount.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("account not found", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, sessionData.GetActiveAccountID()).Return((*identity.Account)(nil), sql.ErrNoRows)

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, sessionData.GetActiveAccountID()).Return((*identity.Account)(nil), errors.New("database error"))

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})
}

func TestServiceImpl_GetSelf(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeUser := identityfakes.BuildFakeUser()

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), mock.Anything, sessionData.GetUserID()).Return(fakeUser, nil)

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeUser.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), mock.Anything, sessionData.GetUserID()).Return((*identity.User)(nil), sql.ErrNoRows)

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), mock.Anything, sessionData.GetUserID()).Return((*identity.User)(nil), errors.New("database error"))

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})
}

func TestServiceImpl_RedeemPasswordResetToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.PasswordResetTokenRedemption), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenRedemptionRequestInput")).Return(nil)

		request := &authsvc.RedeemPasswordResetTokenRequest{
			Token:       "reset-token",
			NewPassword: "newpassword123",
		}

		response, err := service.RedeemPasswordResetToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("success without session", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data - unauthenticated flow

		authManager.On(reflection.GetMethodName(authManager.PasswordResetTokenRedemption), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenRedemptionRequestInput")).Return(nil)

		request := &authsvc.RedeemPasswordResetTokenRequest{
			Token:       "reset-token",
			NewPassword: "newpassword123",
		}

		response, err := service.RedeemPasswordResetToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error redeeming token", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.PasswordResetTokenRedemption), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenRedemptionRequestInput")).Return(errors.New("redemption failed"))

		request := &authsvc.RedeemPasswordResetTokenRequest{
			Token:       "invalid-token",
			NewPassword: "newpassword123",
		}

		response, err := service.RedeemPasswordResetToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}
func TestServiceImpl_RefreshTOTPSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeTOTPResponse := &auth.TOTPSecretRefreshResponse{
			TwoFactorSecret: "new-secret",
			TwoFactorQRCode: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
		}

		authManager.On(reflection.GetMethodName(authManager.NewTOTPSecret), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretRefreshInput")).Return(fakeTOTPResponse, nil)

		request := &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: "password123",
			TotpToken:       "123456",
		}

		response, err := service.RefreshTOTPSecret(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeTOTPResponse.TwoFactorSecret, response.Result.TwoFactorSecret)
		assert.Equal(t, fakeTOTPResponse.TwoFactorQRCode, response.Result.TwoFactorQrCode)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: "password123",
			TotpToken:       "123456",
		}

		response, err := service.RefreshTOTPSecret(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error refreshing TOTP secret", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.NewTOTPSecret), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretRefreshInput")).Return((*auth.TOTPSecretRefreshResponse)(nil), errors.New("refresh failed"))

		request := &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: "wrongpassword",
			TotpToken:       "123456",
		}

		response, err := service.RefreshTOTPSecret(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RequestEmailVerificationEmail(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestEmailVerificationEmail), mock.Anything).Return(nil)

		request := &authsvc.RequestEmailVerificationEmailRequest{}

		response, err := service.RequestEmailVerificationEmail(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RequestEmailVerificationEmailRequest{}

		response, err := service.RequestEmailVerificationEmail(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error requesting email verification", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestEmailVerificationEmail), mock.Anything).Return(errors.New("email request failed"))

		request := &authsvc.RequestEmailVerificationEmailRequest{}

		response, err := service.RequestEmailVerificationEmail(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RequestPasswordResetToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.CreatePasswordResetToken), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenCreationRequestInput")).Return(nil)

		request := &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestPasswordResetToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("success without session", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data - unauthenticated flow

		authManager.On(reflection.GetMethodName(authManager.CreatePasswordResetToken), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenCreationRequestInput")).Return(nil)

		request := &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestPasswordResetToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error creating password reset token", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.CreatePasswordResetToken), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenCreationRequestInput")).Return(errors.New("token creation failed"))

		request := &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: "invalid@example.com",
		}

		response, err := service.RequestPasswordResetToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RequestUsernameReminder(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestUsernameReminder), mock.Anything, mock.AnythingOfType("*auth.UsernameReminderRequestInput")).Return(nil)

		request := &authsvc.RequestUsernameReminderRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestUsernameReminder(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("without session context", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		// The session is optional for this unauthenticated endpoint; a missing session must not error.
		authManager.On(reflection.GetMethodName(authManager.RequestUsernameReminder), mock.Anything, mock.AnythingOfType("*auth.UsernameReminderRequestInput")).Return(nil)

		request := &authsvc.RequestUsernameReminderRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestUsernameReminder(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error requesting username reminder", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestUsernameReminder), mock.Anything, mock.AnythingOfType("*auth.UsernameReminderRequestInput")).Return(errors.New("reminder request failed"))

		request := &authsvc.RequestUsernameReminderRequest{
			EmailAddress: "invalid@example.com",
		}

		response, err := service.RequestUsernameReminder(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_VerifyEmailAddress(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context()

		authManager.On(reflection.GetMethodName(authManager.VerifyUserEmailAddressByToken), mock.Anything, "verification-token").Return(nil)

		request := &authsvc.VerifyEmailAddressRequest{
			Token: "verification-token",
		}

		response, err := service.VerifyEmailAddress(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Verified)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error verifying email address", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context()

		authManager.On(reflection.GetMethodName(authManager.VerifyUserEmailAddressByToken), mock.Anything, "invalid-token").Return(errors.New("verification failed"))

		request := &authsvc.VerifyEmailAddressRequest{
			Token: "invalid-token",
		}

		response, err := service.VerifyEmailAddress(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_VerifyTOTPSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context()

		authManager.On(reflection.GetMethodName(authManager.TOTPSecretVerification), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretVerificationInput")).Return(nil)

		request := &authsvc.VerifyTOTPSecretRequest{
			UserId:    identityfakes.BuildFakeID(),
			TotpToken: "123456",
		}

		response, err := service.VerifyTOTPSecret(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Accepted)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error verifying TOTP secret", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := t.Context()

		authManager.On(reflection.GetMethodName(authManager.TOTPSecretVerification), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretVerificationInput")).Return(errors.New("verification failed"))

		request := &authsvc.VerifyTOTPSecretRequest{
			UserId:    identityfakes.BuildFakeID(),
			TotpToken: "invalid",
		}

		response, err := service.VerifyTOTPSecret(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_UpdatePassword(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.UpdatePassword), mock.Anything, mock.AnythingOfType("*auth.PasswordUpdateInput")).Return(nil)

		request := &authsvc.UpdatePasswordRequest{
			NewPassword:     "newpassword123",
			CurrentPassword: "oldpassword123",
			TotpToken:       "123456",
		}

		response, err := service.UpdatePassword(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.UpdatePasswordRequest{
			NewPassword:     "newpassword123",
			CurrentPassword: "oldpassword123",
			TotpToken:       "123456",
		}

		response, err := service.UpdatePassword(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error updating password", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.UpdatePassword), mock.Anything, mock.AnythingOfType("*auth.PasswordUpdateInput")).Return(errors.New("password update failed"))

		request := &authsvc.UpdatePasswordRequest{
			NewPassword:     "newpassword123",
			CurrentPassword: "wrongpassword",
			TotpToken:       "123456",
		}

		response, err := service.UpdatePassword(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func buildContextWithSessionDataAndSessionID(t *testing.T) (context.Context, *sessions.ContextData) {
	t.Helper()
	sessionData := &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   identityfakes.BuildFakeID(),
			AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
			AccountStatusExplanation: "",
			ServicePermissions:       authorization.NewServiceRolePermissionChecker([]string{"service_user"}, nil),
		},
		ActiveAccountID: identityfakes.BuildFakeID(),
		SessionID:       identityfakes.BuildFakeID(),
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			identityfakes.BuildFakeID(): authorization.NewAccountRolePermissionChecker(nil),
		},
	}
	sessionData.AccountPermissions[sessionData.ActiveAccountID] = authorization.NewAccountRolePermissionChecker(nil)
	ctx := context.WithValue(t.Context(), sessions.SessionContextDataKey, sessionData)
	return ctx, sessionData
}

func Test_extractLoginMetadata(t *testing.T) {
	t.Parallel()

	t.Run("extracts user-agent from gRPC metadata", func(t *testing.T) {
		t.Parallel()

		md := metadata.New(map[string]string{
			"user-agent": "TestBrowser/1.0",
		})
		ctx := metadata.NewIncomingContext(t.Context(), md)

		meta := extractLoginMetadata(ctx)

		assert.Equal(t, "TestBrowser/1.0", meta.UserAgent)
	})

	t.Run("extracts x-forwarded-for for client IP", func(t *testing.T) {
		t.Parallel()

		md := metadata.New(map[string]string{
			"x-forwarded-for": "203.0.113.50",
			"user-agent":      "TestBrowser/1.0",
		})
		ctx := metadata.NewIncomingContext(t.Context(), md)

		meta := extractLoginMetadata(ctx)

		assert.Equal(t, "203.0.113.50", meta.ClientIP)
		assert.Equal(t, "TestBrowser/1.0", meta.UserAgent)
	})

	t.Run("falls back to peer address when no x-forwarded-for", func(t *testing.T) {
		t.Parallel()

		md := metadata.New(map[string]string{
			"user-agent": "TestBrowser/1.0",
		})
		ctx := metadata.NewIncomingContext(t.Context(), md)

		addr, _ := net.ResolveTCPAddr("tcp", "192.168.1.100:12345")
		p := &peer.Peer{Addr: addr}
		ctx = peer.NewContext(ctx, p)

		meta := extractLoginMetadata(ctx)

		assert.Equal(t, "192.168.1.100:12345", meta.ClientIP)
		assert.Equal(t, "TestBrowser/1.0", meta.UserAgent)
	})

	t.Run("returns empty metadata when no context metadata exists", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		meta := extractLoginMetadata(ctx)

		assert.NotNil(t, meta)
		assert.Equal(t, &authentication.LoginMetadata{}, meta)
	})
}

func TestServiceImpl_ListActiveSessions(t *testing.T) {
	t.Parallel()

	t.Run("success with is_current flag", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx, sessionData := buildContextWithSessionDataAndSessionID(t)

		currentSessionID := sessionData.GetSessionID()
		otherSessionID := identityfakes.BuildFakeID()

		now := time.Now()
		fakeSessions := &filtering.QueryFilteredResult[auth.UserSession]{
			Data: []*auth.UserSession{
				{
					ID:           currentSessionID,
					ClientIP:     "192.168.1.1",
					UserAgent:    "TestBrowser/1.0",
					DeviceName:   "My Laptop",
					LoginMethod:  auth.LoginMethodPassword,
					CreatedAt:    now.Add(-time.Hour),
					LastActiveAt: now,
					ExpiresAt:    now.Add(time.Hour),
				},
				{
					ID:           otherSessionID,
					ClientIP:     "10.0.0.1",
					UserAgent:    "OtherBrowser/2.0",
					DeviceName:   "My Phone",
					LoginMethod:  auth.LoginMethodPasskey,
					CreatedAt:    now.Add(-2 * time.Hour),
					LastActiveAt: now.Add(-30 * time.Minute),
					ExpiresAt:    now.Add(30 * time.Minute),
				},
			},
			Pagination: filtering.Pagination{
				TotalCount:    2,
				FilteredCount: 2,
			},
		}

		authManager.On(reflection.GetMethodName(authManager.GetActiveSessionsForUser), mock.Anything, sessionData.GetUserID(), mock.AnythingOfType("*filtering.QueryFilter")).Return(fakeSessions, nil)

		request := &authsvc.ListActiveSessionsRequest{}

		response, err := service.ListActiveSessions(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.Len(t, response.Sessions, 2)

		// First session should be marked as current.
		assert.Equal(t, currentSessionID, response.Sessions[0].Id)
		assert.True(t, response.Sessions[0].IsCurrent)

		// Second session should NOT be marked as current.
		assert.Equal(t, otherSessionID, response.Sessions[1].Id)
		assert.False(t, response.Sessions[1].IsCurrent)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context()

		request := &authsvc.ListActiveSessionsRequest{}

		response, err := service.ListActiveSessions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error from auth manager", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx, sessionData := buildContextWithSessionDataAndSessionID(t)

		authManager.On(reflection.GetMethodName(authManager.GetActiveSessionsForUser), mock.Anything, sessionData.GetUserID(), mock.AnythingOfType("*filtering.QueryFilter")).Return((*filtering.QueryFilteredResult[auth.UserSession])(nil), errors.New("database error"))

		request := &authsvc.ListActiveSessionsRequest{}

		response, err := service.ListActiveSessions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RevokeSession(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx, sessionData := buildContextWithSessionDataAndSessionID(t)

		targetSessionID := identityfakes.BuildFakeID()

		authManager.On(reflection.GetMethodName(authManager.RevokeSession), mock.Anything, targetSessionID, sessionData.GetUserID()).Return(nil)

		request := &authsvc.RevokeSessionRequest{
			SessionId: targetSessionID,
		}

		response, err := service.RevokeSession(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error when session_id is empty", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		request := &authsvc.RevokeSessionRequest{
			SessionId: "",
		}

		response, err := service.RevokeSession(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, grpcErr.Code())
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context()

		request := &authsvc.RevokeSessionRequest{
			SessionId: identityfakes.BuildFakeID(),
		}

		response, err := service.RevokeSession(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})
}

func TestServiceImpl_RevokeAllOtherSessions(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx, sessionData := buildContextWithSessionDataAndSessionID(t)

		authManager.On(reflection.GetMethodName(authManager.RevokeAllSessionsForUserExcept), mock.Anything, sessionData.GetUserID(), sessionData.GetSessionID()).Return(nil)

		request := &authsvc.RevokeAllOtherSessionsRequest{}

		response, err := service.RevokeAllOtherSessions(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context()

		request := &authsvc.RevokeAllOtherSessionsRequest{}

		response, err := service.RevokeAllOtherSessions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})
}

func TestServiceImpl_RevokeCurrentSession(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx, sessionData := buildContextWithSessionDataAndSessionID(t)

		authManager.On(reflection.GetMethodName(authManager.RevokeSession), mock.Anything, sessionData.GetSessionID(), sessionData.GetUserID()).Return(nil)

		request := &authsvc.RevokeCurrentSessionRequest{}

		response, err := service.RevokeCurrentSession(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error when session ID not available", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		request := &authsvc.RevokeCurrentSessionRequest{}

		response, err := service.RevokeCurrentSession(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, grpcErr.Code())
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _, _ := buildTestService(t)
		ctx := t.Context()

		request := &authsvc.RevokeCurrentSessionRequest{}

		response, err := service.RevokeCurrentSession(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error revoking session", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _, _ := buildTestService(t)
		ctx, sessionData := buildContextWithSessionDataAndSessionID(t)

		authManager.On(reflection.GetMethodName(authManager.RevokeSession), mock.Anything, sessionData.GetSessionID(), sessionData.GetUserID()).Return(errors.New("oh no"))

		request := &authsvc.RevokeCurrentSessionRequest{}

		response, err := service.RevokeCurrentSession(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}
