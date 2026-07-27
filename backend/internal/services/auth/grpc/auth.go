package grpc

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/grpc/converters"
	_ "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/errors"
	identityconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	errorsgrpc "github.com/primandproper/platform-go/v7/errors/grpc"
	"github.com/primandproper/platform-go/v7/featureflags"
	"github.com/primandproper/platform-go/v7/observability"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func (s *serviceImpl) EvaluateBooleanFeatureFlag(ctx context.Context, req *authsvc.EvaluateBooleanFeatureFlagRequest) (*authsvc.EvaluateBooleanFeatureFlagResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	featureFlag := strings.TrimSpace(req.GetFeatureFlag())
	if featureFlag == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("feature_flag is required"), logger, span, codes.InvalidArgument, "feature_flag is required")
	}

	enabled, err := s.featureFlagManager.CanUseFeature(ctx, featureFlag, featureflags.EvaluationContext{TargetingKey: sessionContextData.GetUserID()})
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to evaluate feature flag")
	}

	return &authsvc.EvaluateBooleanFeatureFlagResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Enabled: enabled,
	}, nil
}

func (s *serviceImpl) EvaluateInt64FeatureFlag(ctx context.Context, req *authsvc.EvaluateInt64FeatureFlagRequest) (*authsvc.EvaluateInt64FeatureFlagResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	featureFlag := strings.TrimSpace(req.GetFeatureFlag())
	if featureFlag == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("feature_flag is required"), logger, span, codes.InvalidArgument, "feature_flag is required")
	}

	value, err := s.featureFlagManager.GetInt64Value(ctx, featureFlag, 0, featureflags.EvaluationContext{TargetingKey: sessionContextData.GetUserID()})
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to evaluate feature flag")
	}

	return &authsvc.EvaluateInt64FeatureFlagResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Value: value,
	}, nil
}

func (s *serviceImpl) EvaluateStringFeatureFlag(ctx context.Context, req *authsvc.EvaluateStringFeatureFlagRequest) (*authsvc.EvaluateStringFeatureFlagResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	featureFlag := strings.TrimSpace(req.GetFeatureFlag())
	if featureFlag == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("feature_flag is required"), logger, span, codes.InvalidArgument, "feature_flag is required")
	}

	value, err := s.featureFlagManager.GetStringValue(ctx, featureFlag, "", featureflags.EvaluationContext{TargetingKey: sessionContextData.GetUserID()})
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to evaluate feature flag")
	}

	return &authsvc.EvaluateStringFeatureFlagResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Value: value,
	}, nil
}

func (s *serviceImpl) GetAuthStatus(ctx context.Context, _ *authsvc.GetAuthStatusRequest) (*authsvc.GetAuthStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	requiresPasswordChange, pcErr := s.identityDataManager.UserRequiresPasswordChange(ctx, sessionContextData.GetUserID())
	if pcErr != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(pcErr, logger, span, codes.Internal, "checking password change requirement")
	}

	x := &authsvc.GetAuthStatusResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserId:                   sessionContextData.GetUserID(),
		AccountStatus:            sessionContextData.Requester.AccountStatus,
		AccountStatusExplanation: sessionContextData.Requester.AccountStatusExplanation,
		ActiveAccount:            sessionContextData.GetActiveAccountID(),
		RequiresPasswordChange:   requiresPasswordChange,
	}

	return x, nil
}

func (s *serviceImpl) ExchangeToken(ctx context.Context, request *authsvc.ExchangeTokenRequest) (*authsvc.ExchangeTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	_, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	newToken, err := s.authenticationManager.ExchangeTokenForUser(ctx, request.RefreshToken, request.DesiredAccountId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to exchange token")
	}

	x := &authsvc.ExchangeTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserId:       newToken.UserID,
		AccountId:    newToken.AccountID,
		RefreshToken: newToken.RefreshToken,
		AccessToken:  newToken.AccessToken,
		ExpiresUtc:   grpcconverters.ConvertTimeToPBTimestamp(newToken.ExpiresUTC),
	}

	return x, nil
}

func (s *serviceImpl) LoginForToken(ctx context.Context, request *authsvc.LoginForTokenRequest) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	return s.loginForToken(ctx, false, converters.ConvertGRPCUserLoginInputToUserLoginInput(request.Input))
}

func (s *serviceImpl) AdminLoginForToken(ctx context.Context, request *authsvc.AdminLoginForTokenRequest) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	return s.loginForToken(ctx, true, converters.ConvertGRPCUserLoginInputToUserLoginInput(request.Input))
}

func (s *serviceImpl) loginForToken(ctx context.Context, admin bool, input *auth.UserLoginInput) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	tokenResponse, err := s.authenticationManager.ProcessLogin(ctx, admin, input, extractLoginMetadata(ctx))
	if err != nil {
		// Log the specific failure reason server-side, but return a single generic status to the
		// unauthenticated caller so distinct outcomes (unknown user, wrong password, banned account)
		// can't be used to enumerate usernames or account state.
		observability.AcknowledgeError(err, logger, span, "processing login request")
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	x := &authsvc.LoginForTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertTokenResponseToGRPCTokenResponse(tokenResponse),
	}

	return x, nil
}

func (s *serviceImpl) CheckPermissions(ctx context.Context, request *authsvc.UserPermissionsRequestInput) (*authsvc.UserPermissionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCCheckPermissionsRequestToUserPermissionsRequestInput(request)
	result, err := s.authManager.CheckUserPermissions(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "checking user permissions")
	}

	x := &authsvc.UserPermissionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Permissions: result.Permissions,
	}

	return x, nil
}

func (s *serviceImpl) GetActiveAccount(ctx context.Context, request *authsvc.GetActiveAccountRequest) (*authsvc.GetActiveAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	account, err := s.identityDataManager.GetAccount(ctx, sessionContextData.GetActiveAccountID())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "failed to get active account")
		}
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get active account")
	}

	x := &authsvc.GetActiveAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: identityconverters.ConvertAccountToGRPCAccount(account),
	}

	return x, nil
}

func (s *serviceImpl) GetSelf(ctx context.Context, request *authsvc.GetSelfRequest) (*authsvc.GetSelfResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	user, err := s.identityDataManager.GetUser(ctx, sessionContextData.GetUserID())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "failed to get user")
		}
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get user")
	}

	x := &authsvc.GetSelfResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: identityconverters.ConvertUserToGRPCUser(user),
	}

	return x, nil
}

func (s *serviceImpl) RedeemPasswordResetToken(ctx context.Context, request *authsvc.RedeemPasswordResetTokenRequest) (*authsvc.RedeemPasswordResetTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	input := converters.ConvertGRPCRedeemPasswordResetTokenRequestToPasswordResetTokenRedemptionRequestInput(request)
	if err := s.authManager.PasswordResetTokenRedemption(ctx, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "redeeming password reset token")
	}

	x := &authsvc.RedeemPasswordResetTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) RefreshTOTPSecret(ctx context.Context, request *authsvc.RefreshTOTPSecretRequest) (*authsvc.RefreshTOTPSecretResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCRefreshTOTPSecretRequestToTOTPSecretRefreshInput(request)
	totpSecretResponse, err := s.authManager.NewTOTPSecret(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "refreshing totp secret")
	}

	x := &authsvc.RefreshTOTPSecretResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertTOTPSecretRefreshResponseToGRPCTOTPSecretRefreshResponse(totpSecretResponse),
	}

	return x, nil
}

func (s *serviceImpl) RequestEmailVerificationEmail(ctx context.Context, request *authsvc.RequestEmailVerificationEmailRequest) (*authsvc.RequestEmailVerificationEmailResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	if err = s.authManager.RequestEmailVerificationEmail(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to request email verification email")
	}

	x := &authsvc.RequestEmailVerificationEmailResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) RequestPasswordResetToken(ctx context.Context, request *authsvc.RequestPasswordResetTokenRequest) (*authsvc.RequestPasswordResetTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	input := converters.ConvertGRPCRequestPasswordResetTokenRequestToPasswordResetTokenCreationRequestInput(request)
	if err := s.authManager.CreatePasswordResetToken(ctx, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create password reset token")
	}

	x := &authsvc.RequestPasswordResetTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) RequestUsernameReminder(ctx context.Context, request *authsvc.RequestUsernameReminderRequest) (*authsvc.RequestUsernameReminderResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	// A user who forgot their username can't authenticate, so this endpoint is unauthenticated
	// (mirroring the password-reset flow); only decorate the logger if a session happens to exist.
	if sessionContextData, scErr := s.fetchSessionContext(ctx); scErr == nil {
		logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())
	}

	input := converters.ConvertGRPCRequestUsernameReminderRequestToUsernameReminderRequestInput(request)
	if err := s.authManager.RequestUsernameReminder(ctx, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to issue username reminder")
	}

	x := &authsvc.RequestUsernameReminderResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) VerifyEmailAddress(ctx context.Context, request *authsvc.VerifyEmailAddressRequest) (*authsvc.VerifyEmailAddressResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if err := s.authManager.VerifyUserEmailAddressByToken(ctx, request.GetToken()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to verify user's email address")
	}

	x := &authsvc.VerifyEmailAddressResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Verified: true,
	}

	return x, nil
}

func (s *serviceImpl) VerifyTOTPSecret(ctx context.Context, request *authsvc.VerifyTOTPSecretRequest) (*authsvc.VerifyTOTPSecretResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, request.UserId)

	input := converters.ConvertGRPCVerifyTOTPSecretRequestToTOTPSecretVerificationInput(request)
	if err := s.authManager.TOTPSecretVerification(ctx, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to verify TOTP secret")
	}

	x := &authsvc.VerifyTOTPSecretResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Accepted: true,
	}

	return x, nil
}

func (s *serviceImpl) UpdatePassword(ctx context.Context, request *authsvc.UpdatePasswordRequest) (*authsvc.UpdatePasswordResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCUpdatePasswordRequestToPasswordUpdateInput(request)
	if err = s.authManager.UpdatePassword(ctx, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update password")
	}

	x := &authsvc.UpdatePasswordResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) BeginPasskeyRegistration(ctx context.Context, _ *authsvc.BeginPasskeyRegistrationRequest) (*authsvc.BeginPasskeyRegistrationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	creation, session, err := s.passkeyService.BeginRegistrationOptions(ctx, sessionContextData.GetUserID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to begin passkey registration")
	}

	optionsJSON := s.jsonEncoder.MustEncodeJSON(ctx, creation)

	return &authsvc.BeginPasskeyRegistrationResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		PublicKeyCredentialCreationOptions: optionsJSON,
		Challenge:                          session.Challenge,
	}, nil
}

func (s *serviceImpl) FinishPasskeyRegistration(ctx context.Context, request *authsvc.FinishPasskeyRegistrationRequest) (*authsvc.FinishPasskeyRegistrationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if len(request.AttestationResponse) == 0 {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("attestation_response is required"), logger, span, codes.InvalidArgument, "attestation_response is required")
	}
	if request.Challenge == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("challenge is required"), logger, span, codes.InvalidArgument, "challenge is required")
	}

	if err = s.passkeyService.FinishRegistrationFromBytes(ctx, sessionContextData.GetUserID(), request.AttestationResponse, request.Challenge); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to finish passkey registration")
	}

	return &authsvc.FinishPasskeyRegistrationResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) BeginPasskeyAuthentication(ctx context.Context, request *authsvc.BeginPasskeyAuthenticationRequest) (*authsvc.BeginPasskeyAuthenticationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	username := strings.TrimSpace(request.GetUsername())

	assertion, session, err := s.passkeyService.BeginAuthenticationOptions(ctx, username)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to begin passkey authentication")
	}

	optionsJSON := s.jsonEncoder.MustEncodeJSON(ctx, assertion)

	return &authsvc.BeginPasskeyAuthenticationResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		PublicKeyCredentialRequestOptions: optionsJSON,
		Challenge:                         session.Challenge,
	}, nil
}

func (s *serviceImpl) FinishPasskeyAuthentication(ctx context.Context, request *authsvc.FinishPasskeyAuthenticationRequest) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if len(request.AssertionResponse) == 0 {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("assertion_response is required"), logger, span, codes.InvalidArgument, "assertion_response is required")
	}
	if request.Challenge == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("challenge is required"), logger, span, codes.InvalidArgument, "challenge is required")
	}

	username := strings.TrimSpace(request.GetUsername())

	result, err := s.passkeyService.FinishAuthenticationFromBytes(ctx, username, request.AssertionResponse, request.Challenge)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to finish passkey authentication")
	}

	tokenResponse, err := s.authenticationManager.ProcessPasskeyLogin(ctx, result.UserID, "", extractLoginMetadata(ctx))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to process passkey login")
	}

	return &authsvc.LoginForTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertTokenResponseToGRPCTokenResponse(tokenResponse),
	}, nil
}

func (s *serviceImpl) ListPasskeys(ctx context.Context, _ *authsvc.ListPasskeysRequest) (*authsvc.ListPasskeysResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	creds, err := s.passkeyService.GetCredentialsForUser(ctx, sessionContextData.GetUserID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to list passkeys")
	}

	results := make([]*authsvc.PasskeyCredential, 0, len(creds))
	for _, c := range creds {
		friendlyName := c.FriendlyName
		if friendlyName == "" {
			friendlyName = "Passkey"
		}
		pc := &authsvc.PasskeyCredential{
			Id:           c.ID,
			FriendlyName: friendlyName,
			CreatedAt:    grpcconverters.ConvertTimeToPBTimestamp(c.CreatedAt),
		}
		if c.LastUsedAt != nil {
			pc.LastUsedAt = grpcconverters.ConvertTimeToPBTimestamp(*c.LastUsedAt)
		}
		results = append(results, pc)
	}

	return &authsvc.ListPasskeysResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Results: results,
	}, nil
}

// extractLoginMetadata extracts client IP and User-Agent from gRPC request metadata.
func extractLoginMetadata(ctx context.Context) *authentication.LoginMetadata {
	meta := &authentication.LoginMetadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("user-agent"); len(vals) > 0 {
			meta.UserAgent = vals[0]
		}
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 {
			meta.ClientIP = vals[0]
		}
	}

	if meta.ClientIP == "" {
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			meta.ClientIP = p.Addr.String()
		}
	}

	return meta
}

func (s *serviceImpl) ArchivePasskey(ctx context.Context, request *authsvc.ArchivePasskeyRequest) (*authsvc.ArchivePasskeyResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	credentialID := strings.TrimSpace(request.GetCredentialId())
	if credentialID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("credential_id is required"), logger, span, codes.InvalidArgument, "credential_id is required")
	}

	if err = s.passkeyService.ArchiveCredentialForUser(ctx, credentialID, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive passkey")
	}

	return &authsvc.ArchivePasskeyResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) ListActiveSessions(ctx context.Context, request *authsvc.ListActiveSessionsRequest) (*authsvc.ListActiveSessionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	sessionsResult, err := s.authManager.GetActiveSessionsForUser(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to list active sessions")
	}

	results := make([]*authsvc.UserSession, 0, len(sessionsResult.Data))
	for _, sess := range sessionsResult.Data {
		results = append(results, &authsvc.UserSession{
			Id:           sess.ID,
			ClientIp:     sess.ClientIP,
			UserAgent:    sess.UserAgent,
			DeviceName:   sess.DeviceName,
			LoginMethod:  sess.LoginMethod,
			CreatedAt:    grpcconverters.ConvertTimeToPBTimestamp(sess.CreatedAt),
			LastActiveAt: grpcconverters.ConvertTimeToPBTimestamp(sess.LastActiveAt),
			ExpiresAt:    grpcconverters.ConvertTimeToPBTimestamp(sess.ExpiresAt),
			IsCurrent:    sess.ID == sessionContextData.GetSessionID(),
		})
	}

	return &authsvc.ListActiveSessionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(sessionsResult.Pagination, filter),
		Sessions:   results,
	}, nil
}

func (s *serviceImpl) RevokeSession(ctx context.Context, request *authsvc.RevokeSessionRequest) (*authsvc.RevokeSessionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	sessionID := strings.TrimSpace(request.GetSessionId())
	if sessionID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("session_id is required"), logger, span, codes.InvalidArgument, "session_id is required")
	}

	if err = s.authManager.RevokeSession(ctx, sessionID, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to revoke session")
	}

	return &authsvc.RevokeSessionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) RevokeAllOtherSessions(ctx context.Context, _ *authsvc.RevokeAllOtherSessionsRequest) (*authsvc.RevokeAllOtherSessionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if err = s.authManager.RevokeAllSessionsForUserExcept(ctx, sessionContextData.GetUserID(), sessionContextData.GetSessionID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to revoke all other sessions")
	}

	return &authsvc.RevokeAllOtherSessionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) RevokeCurrentSession(ctx context.Context, _ *authsvc.RevokeCurrentSessionRequest) (*authsvc.RevokeCurrentSessionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	sessionID := sessionContextData.GetSessionID()
	if sessionID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("session ID not available"), logger, span, codes.FailedPrecondition, "session ID not available — use JWT auth")
	}

	if err = s.authManager.RevokeSession(ctx, sessionID, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to revoke current session")
	}

	return &authsvc.RevokeCurrentSessionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) AdminListSessionsForUser(ctx context.Context, request *authsvc.AdminListSessionsForUserRequest) (*authsvc.ListActiveSessionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if !sessionContextData.GetServicePermissions().CanManageUserSessions() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("insufficient permissions"), logger, span, codes.PermissionDenied, "insufficient permissions")
	}

	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("user_id is required"), logger, span, codes.InvalidArgument, "user_id is required")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	sessionsResult, err := s.authManager.GetActiveSessionsForUser(ctx, userID, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to list active sessions for user")
	}

	results := make([]*authsvc.UserSession, 0, len(sessionsResult.Data))
	for _, sess := range sessionsResult.Data {
		results = append(results, &authsvc.UserSession{
			Id:           sess.ID,
			ClientIp:     sess.ClientIP,
			UserAgent:    sess.UserAgent,
			DeviceName:   sess.DeviceName,
			LoginMethod:  sess.LoginMethod,
			CreatedAt:    grpcconverters.ConvertTimeToPBTimestamp(sess.CreatedAt),
			LastActiveAt: grpcconverters.ConvertTimeToPBTimestamp(sess.LastActiveAt),
			ExpiresAt:    grpcconverters.ConvertTimeToPBTimestamp(sess.ExpiresAt),
		})
	}

	return &authsvc.ListActiveSessionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(sessionsResult.Pagination, filter),
		Sessions:   results,
	}, nil
}

func (s *serviceImpl) AdminRevokeUserSession(ctx context.Context, request *authsvc.AdminRevokeUserSessionRequest) (*authsvc.RevokeSessionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if !sessionContextData.GetServicePermissions().CanManageUserSessions() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("insufficient permissions"), logger, span, codes.PermissionDenied, "insufficient permissions")
	}

	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("user_id is required"), logger, span, codes.InvalidArgument, "user_id is required")
	}

	sessionID := strings.TrimSpace(request.GetSessionId())
	if sessionID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("session_id is required"), logger, span, codes.InvalidArgument, "session_id is required")
	}

	if err = s.authManager.RevokeSession(ctx, sessionID, userID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to revoke user session")
	}

	return &authsvc.RevokeSessionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) AdminRevokeAllUserSessions(ctx context.Context, request *authsvc.AdminRevokeAllUserSessionsRequest) (*authsvc.RevokeAllOtherSessionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if !sessionContextData.GetServicePermissions().CanManageUserSessions() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("insufficient permissions"), logger, span, codes.PermissionDenied, "insufficient permissions")
	}

	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("user_id is required"), logger, span, codes.InvalidArgument, "user_id is required")
	}

	if err = s.authManager.RevokeAllSessionsForUser(ctx, userID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to revoke all sessions for user")
	}

	return &authsvc.RevokeAllOtherSessionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}
