package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/localdev"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auth"

	"github.com/primandproper/platform-go/v4/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v4/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v4/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuth_LoginForToken_DesiredAccount(T *testing.T) {
	T.Parallel()

	T.Run("login for non-default account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create inviter user + client; inviter gets account A (default from registration).
		_, testClient := createUserAndClientForTest(t)
		accountRes, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		inviterAccountID := accountRes.Result.Id

		// Create invitee user + client; invitee gets account B (their default from registration).
		inviteeEmailAddress := fmt.Sprintf("invitee%d@testing.com", time.Now().UnixMicro())
		input := &identity.UserRegistrationInput{
			Birthday:              new(time.Now()),
			EmailAddress:          inviteeEmailAddress,
			FirstName:             fmt.Sprintf("invitee_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AccountName:           fmt.Sprintf("invitee_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			LastName:              fmt.Sprintf("invitee_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Password:              fmt.Sprintf("invitee_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			Username:              fmt.Sprintf("invitee_%d", hashStringToNumber(t.Name()+time.Now().Format(time.RFC3339Nano))),
			AcceptedPrivacyPolicy: true,
			AcceptedTOS:           true,
		}
		invitee, inviteeClient := createUserAndClientForTestWithRegistrationInput(t, input)

		// Inviter creates invitation for invitee.
		invitation, err := testClient.CreateAccountInvitation(ctx, &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				Note:    t.Name(),
				ToName:  t.Name(),
				ToEmail: inviteeEmailAddress,
			},
		})
		require.NoError(t, err)

		// Invitee accepts invitation. Invitee now has accounts A and B; B remains default.
		_, err = inviteeClient.AcceptAccountInvitation(ctx, &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: invitation.Created.Id,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: invitation.Created.Token,
				Note:  t.Name(),
			},
		})
		require.NoError(t, err)

		// Login with DesiredAccountId set to inviter's account (non-default for invitee).
		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:         invitee.Username,
				Password:         invitee.HashedPassword,
				TotpToken:        generateTOTPCodeForUserForTest(t, invitee),
				DesiredAccountId: inviterAccountID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, tokenRes)
		require.NotEmpty(t, tokenRes.Result.AccessToken)
		assert.Equal(t, inviterAccountID, tokenRes.Result.AccountId, "token response should include the desired (non-default) account")

		// Use JWT directly as Bearer so session context uses the token's account_id claim.
		jwtClient, err := buildAuthedGRPCClientWithBearerToken(tokenRes.Result.AccessToken)
		require.NoError(t, err)
		authStatus, err := jwtClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})

		require.NoError(t, err)
		require.NotNil(t, authStatus)
		assert.Equal(t, inviterAccountID, authStatus.ActiveAccount, "session context should use account from token")
	})

	T.Run("login with invalid desired account returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		// Use a made-up account ID the user is not a member of.
		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:         user.Username,
				Password:         user.HashedPassword,
				TotpToken:        generateTOTPCodeForUserForTest(t, user),
				DesiredAccountId: nonexistentID,
			},
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})
}

func TestAuth_LoginForToken(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		user := createServiceUserForTest(t, true, fakes.BuildFakeUserRegistrationInput())
		actual := fetchLoginTokenForUserForTest(t, user)

		assert.NotEmpty(t, actual)
	})

	T.Run("2FA is not required for non-admin users who haven't verified their secrets", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user := createServiceUserForTest(t, false, fakes.BuildFakeUserRegistrationInput())

		loginInput := &authsvc.UserLoginInput{
			Username: user.Username,
			Password: user.HashedPassword,
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: loginInput,
		})
		assert.NoError(t, err)
		assert.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})

	T.Run("with bogus input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  " ",
			Password:  "1",
			TotpToken: "otp scode",
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})
}

func TestAuth_AdminLoginForToken(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  adminUserPassword,
			TotpToken: generateTOTPCodeForUserForTest(t, premadeAdminUser),
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.NoError(t, err)
		assert.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})

	T.Run("non-admin users cannot login via this route", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user := createServiceUserForTest(t, true, fakes.BuildFakeUserRegistrationInput())

		loginInput := &authsvc.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TotpToken: generateTOTPCodeForUserForTest(t, user),
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})

	T.Run("with incorrect password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  adminUserPassword + "blah",
			TotpToken: generateTOTPCodeForUserForTest(t, premadeAdminUser),
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})

	T.Run("with incorrect TOTP Code", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  adminUserPassword,
			TotpToken: "000000",
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})

	T.Run("with bogus input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  " ",
			Password:  "1",
			TotpToken: "otp scode",
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})
}

func TestAuth_EvaluateFeatureFlags(T *testing.T) {
	T.Parallel()

	// Integration tests use NoopFeatureFlagManager (no actual FF backend), so all values are zero.
	// Purpose: validate client methods exist and server responds correctly.
	T.Run("EvaluateBooleanFeatureFlag returns false with noop provider", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.EvaluateBooleanFeatureFlag(ctx, &authsvc.EvaluateBooleanFeatureFlagRequest{
			FeatureFlag: "some-flag",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.False(t, res.Enabled)
	})

	T.Run("EvaluateStringFeatureFlag returns empty with noop provider", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.EvaluateStringFeatureFlag(ctx, &authsvc.EvaluateStringFeatureFlagRequest{
			FeatureFlag: "some-flag",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Empty(t, res.Value)
	})

	T.Run("EvaluateInt64FeatureFlag returns zero with noop provider", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.EvaluateInt64FeatureFlag(ctx, &authsvc.EvaluateInt64FeatureFlagRequest{
			FeatureFlag: "some-flag",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, int64(0), res.Value)
	})

	T.Run("feature flag methods require authentication", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		_, err := unauthedClient.EvaluateBooleanFeatureFlag(ctx, &authsvc.EvaluateBooleanFeatureFlagRequest{
			FeatureFlag: "some-flag",
		})
		assert.Error(t, err)
	})

	T.Run("empty feature_flag returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.EvaluateBooleanFeatureFlag(ctx, &authsvc.EvaluateBooleanFeatureFlagRequest{
			FeatureFlag: "",
		})
		assert.Error(t, err)
	})
}

func TestAuth_GetAuthStatus(T *testing.T) {
	T.Parallel()

	T.Run("for unauthenticated user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("for logged in user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, user.ID, res.UserId)
	})
}

func TestAuth_GetSelf(T *testing.T) {
	T.Parallel()

	T.Run("for authenticated user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.GetSelf(ctx, &authsvc.GetSelfRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Result)
		assert.Equal(t, user.ID, res.Result.Id)
		assert.Equal(t, user.Username, res.Result.Username)
	})

	T.Run("for unauthenticated user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.GetSelf(ctx, &authsvc.GetSelfRequest{})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestAuth_ChangingPassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.UpdatePassword(ctx, &authsvc.UpdatePasswordRequest{
			NewPassword:     user.HashedPassword + "blah",
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  user.Username,
				Password:  user.HashedPassword + "blah",
				TotpToken: generateTOTPCodeForUserForTest(t, user),
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "updated", ResourceType: "users", RelevantID: user.ID},
		})
	})

	T.Run("with inadequate new password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.UpdatePassword(ctx, &authsvc.UpdatePasswordRequest{
			NewPassword:     "b",
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		assert.Error(t, err)
	})
}

func TestAuth_ChangingTOTPSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.RefreshTOTPSecret(ctx, &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)
		res.Result.TwoFactorSecret = user.TwoFactorSecret

		_, err = testClient.VerifyTOTPSecret(ctx, &authsvc.VerifyTOTPSecretRequest{
			TotpToken: generateTOTPCodeForUserForTest(t, user),
			UserId:    user.ID,
		})
		assert.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "updated", ResourceType: "users", RelevantID: user.ID},
		})
	})

	T.Run("fails with invalid token", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.RefreshTOTPSecret(ctx, &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: user.HashedPassword,
			TotpToken:       "000000",
		})
		assert.Error(t, err)
	})

	T.Run("fails with invalid password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.RefreshTOTPSecret(ctx, &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: user.HashedPassword + "blah",
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		assert.Error(t, err)
	})
}

func TestAuth_RequestingPasswordReset(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.RequestPasswordResetToken(ctx, &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: user.EmailAddress,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		// boo, hiss, we're talking directly to the database in an _integration test?_ for shame, for shame.
		var token string
		queryErr := databaseClient.ReadDB().QueryRow(`SELECT token FROM password_reset_tokens WHERE belongs_to_user = $1`, user.ID).Scan(&token)
		require.NoError(t, queryErr)

		auditLogRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), databaseClient)
		authRepo := authrepo.ProvideAuthRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogRepo, databaseClient)

		resetToken, err := authRepo.GetPasswordResetTokenByToken(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, resetToken)

		_, err = testClient.RedeemPasswordResetToken(ctx, &authsvc.RedeemPasswordResetTokenRequest{
			Token:       resetToken.Token,
			NewPassword: user.HashedPassword + "blah",
		})
		require.NoError(t, err)

		tokenRes, err := testClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  user.Username,
				Password:  user.HashedPassword + "blah",
				TotpToken: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)

		// verify that we can't do it twice
		_, err = testClient.RedeemPasswordResetToken(ctx, &authsvc.RedeemPasswordResetTokenRequest{
			Token:       resetToken.Token,
			NewPassword: user.HashedPassword + "blah",
		})
		require.Error(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "password_reset_tokens"},
			{EventType: "updated", ResourceType: "password_reset_tokens"},
		})
	})

	T.Run("unauthenticated flow", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.RequestPasswordResetToken(ctx, &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: user.EmailAddress,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		var token string
		queryErr := databaseClient.ReadDB().QueryRow(`SELECT token FROM password_reset_tokens WHERE belongs_to_user = $1 ORDER BY created_at DESC LIMIT 1`, user.ID).Scan(&token)
		require.NoError(t, queryErr)

		_, err = unauthedClient.RedeemPasswordResetToken(ctx, &authsvc.RedeemPasswordResetTokenRequest{
			Token:       token,
			NewPassword: "newpassword123!",
		})
		require.NoError(t, err)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  user.Username,
				Password:  "newpassword123!",
				TotpToken: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})
}

func TestAuth_InvalidateToken(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		loginInput := &authsvc.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TotpToken: generateTOTPCodeForUserForTest(t, user),
		}

		oauth2Token, err := localdev.FetchOAuth2TokenForUser(
			ctx,
			httpTestServerAddress,
			fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port),
			createdClientID,
			createdClientSecret,
			loginInput,
		)
		require.NoError(t, err)
		require.NotNil(t, oauth2Token)
		require.NotEmpty(t, oauth2Token.AccessToken)

		// Verify token works before revocation
		clientWithToken, err := buildAuthedGRPCClientWithBearerToken(oauth2Token.AccessToken)
		require.NoError(t, err)
		_, err = clientWithToken.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)

		// Revoke the access token
		form := url.Values{}
		form.Set("token", oauth2Token.AccessToken)
		form.Set("token_type_hint", "access_token")
		form.Set("client_id", createdClientID)
		form.Set("client_secret", createdClientSecret)

		revokeReq, err := http.NewRequestWithContext(ctx, http.MethodPost, httpTestServerAddress+"/oauth2/revoke", strings.NewReader(form.Encode()))
		require.NoError(t, err)
		revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		httpClient := &http.Client{}
		revokeRes, err := httpClient.Do(revokeReq)
		require.NoError(t, err)
		defer revokeRes.Body.Close()

		assert.Equal(t, http.StatusOK, revokeRes.StatusCode, "revoke endpoint should return 200")

		// Token should be invalid after revocation
		_, err = clientWithToken.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err, "API calls with revoked token should fail")
	})
}

func TestAuth_Passkey(T *testing.T) {
	T.Parallel()

	T.Run("BeginPasskeyAuthentication discoverable returns options and challenge", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
			Username: "",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Challenge)
		require.NotEmpty(t, res.PublicKeyCredentialRequestOptions)
	})

	T.Run("BeginPasskeyAuthentication with existing user and no passkeys returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		// User exists but has no passkeys; go-webauthn requires at least one credential for username-based login
		res, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
			Username: user.Username,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("BeginPasskeyAuthentication with nonexistent user returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
			Username: "nonexistent_user_12345",
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("BeginPasskeyRegistration unauthenticated returns Unauthenticated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("BeginPasskeyRegistration authenticated returns options and challenge", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Challenge)
		require.NotEmpty(t, res.PublicKeyCredentialCreationOptions)
	})

	T.Run("FinishPasskeyAuthentication empty assertion returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		// Get a valid challenge first
		beginRes, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{Username: ""})
		require.NoError(t, err)
		require.NotEmpty(t, beginRes.Challenge)

		res, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Challenge:         beginRes.Challenge,
			AssertionResponse: nil,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("FinishPasskeyAuthentication empty challenge returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Challenge:         "",
			AssertionResponse: []byte("some-bytes"),
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("FinishPasskeyAuthentication invalid assertion returns Unauthenticated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		beginRes, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{Username: ""})
		require.NoError(t, err)
		require.NotEmpty(t, beginRes.Challenge)

		// Malformed assertion - valid JSON structure but not a real WebAuthn assertion
		invalidAssertion := []byte(`{"id":"x","rawId":"eA==","type":"public-key","response":{"clientDataJSON":"eA==","authenticatorData":"eA==","signature":"eA=="}}`)

		res, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
			Challenge:         beginRes.Challenge,
			AssertionResponse: invalidAssertion,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		// Invalid/malformed assertion fails validation (signature check, etc.)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("FinishPasskeyRegistration unauthenticated returns Unauthenticated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			Challenge:           "some-challenge",
			AttestationResponse: []byte("some-bytes"),
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("FinishPasskeyRegistration empty attestation returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		beginRes, err := testClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, beginRes.Challenge)

		res, err := testClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			Challenge:           beginRes.Challenge,
			AttestationResponse: nil,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("FinishPasskeyRegistration empty challenge returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			Challenge:           "",
			AttestationResponse: []byte("some-bytes"),
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("FinishPasskeyRegistration invalid attestation returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		beginRes, err := testClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, beginRes.Challenge)

		// Malformed attestation - not a valid WebAuthn attestation
		invalidAttestation := []byte(`{"id":"x","rawId":"eA==","type":"public-key","response":{"clientDataJSON":"eA==","attestationObject":"eA=="}}`)

		res, err := testClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
			Challenge:           beginRes.Challenge,
			AttestationResponse: invalidAttestation,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("ListPasskeys unauthenticated returns Unauthenticated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.ListPasskeys(ctx, &authsvc.ListPasskeysRequest{})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("ListPasskeys for user with no passkeys returns empty results", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.ListPasskeys(ctx, &authsvc.ListPasskeysRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Empty(t, res.Results)
	})

	T.Run("ListPasskeys for user with passkeys returns credentials", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		credID := insertWebAuthnCredentialForTest(t, user.ID, "My MacBook")

		res, err := testClient.ListPasskeys(ctx, &authsvc.ListPasskeysRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Results, 1)
		assert.Equal(t, credID, res.Results[0].Id)
		assert.Equal(t, "My MacBook", res.Results[0].FriendlyName)
		assert.NotNil(t, res.Results[0].CreatedAt)
	})

	T.Run("ArchivePasskey unauthenticated returns Unauthenticated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.ArchivePasskey(ctx, &authsvc.ArchivePasskeyRequest{
			CredentialId: "some-cred-id",
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("ArchivePasskey empty credential_id returns InvalidArgument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.ArchivePasskey(ctx, &authsvc.ArchivePasskeyRequest{
			CredentialId: "",
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("ArchivePasskey for user's own credential succeeds", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		credID := insertWebAuthnCredentialForTest(t, user.ID, "Test Passkey")

		_, err := testClient.ArchivePasskey(ctx, &authsvc.ArchivePasskeyRequest{
			CredentialId: credID,
		})
		require.NoError(t, err)

		// Verify credential no longer appears in list
		res, err := testClient.ListPasskeys(ctx, &authsvc.ListPasskeysRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Empty(t, res.Results)
	})
}

// insertWebAuthnCredentialForTest inserts a test WebAuthn credential for the given user.
// Returns the credential's internal ID (used by ListPasskeys/ArchivePasskey).
func insertWebAuthnCredentialForTest(t *testing.T, userID, friendlyName string) string {
	t.Helper()

	credID := identifiers.New()
	credentialIDBytes := fmt.Appendf(nil, "test-cred-%s-%d", credID, time.Now().UnixNano())
	publicKeyBytes := []byte("test-public-key-data")

	_, err := databaseClient.WriteDB().ExecContext(
		t.Context(),
		`INSERT INTO webauthn_credentials (id, belongs_to_user, credential_id, public_key, sign_count, transports, friendly_name)
		 VALUES ($1, $2, $3, $4, 0, '', $5)`,
		credID, userID, credentialIDBytes, publicKeyBytes, friendlyName,
	)
	require.NoError(t, err)

	return credID
}
