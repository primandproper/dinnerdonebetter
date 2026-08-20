package managers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	platformtotp "github.com/primandproper/platform-go/v12/authentication/totp"
	perrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/identifiers"
	"github.com/primandproper/platform-go/v12/messagequeue"
	"github.com/primandproper/platform-go/v12/observability"
	platformkeys "github.com/primandproper/platform-go/v12/observability/keys"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	"github.com/primandproper/platform-go/v12/qrcodes"
	"github.com/primandproper/platform-go/v12/random"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	o11yName               = "auth_manager"
	passwordResetTokenSize = 32
	totpSecretSize         = 64
	minimumPasswordEntropy = 60
)

// sessionContextDataForTracing adapts *sessions.ContextData to the tracing package's
// sessionContextData interface (which uses a minimal servicePermissionChecker to avoid
// platform depending on authorization).
type sessionContextDataForTracing struct {
	*sessions.ContextData
}

func (s *sessionContextDataForTracing) GetServicePermissions() tracing.ServicePermissionChecker {
	return servicePermissionCheckerAdapter{inner: s.ContextData.GetServicePermissions()}
}

// servicePermissionCheckerAdapter adapts authorization.ServiceRolePermissionChecker to tracing.ServicePermissionChecker.
type servicePermissionCheckerAdapter struct {
	inner authorization.ServiceRolePermissionChecker
}

func (a servicePermissionCheckerAdapter) IsServiceAdmin() bool {
	if a.inner == nil {
		return false
	}
	return a.inner.IsServiceAdmin()
}

type AuthManager struct {
	passwordResetTokenDataManager auth.PasswordResetTokenDataManager
	sessionDataManager            auth.UserSessionDataManager
	userDataManager               identity.UserDataManager
	tracer                        tracing.Tracer
	authenticator                 authentication.Authenticator
	totpVerifier                  platformtotp.Verifier
	logger                        logging.Logger
	dataChangesPublisher          messagequeue.Publisher
	secretGenerator               random.Generator
	qrCodeBuilder                 qrcodes.Builder
	minimumPasswordLength         uint8
}

func ProvideAuthManager(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	passwordResetTokenDataManager auth.PasswordResetTokenDataManager,
	sessionDataManager auth.UserSessionDataManager,
	userDataManager identity.UserDataManager,
	authenticator authentication.Authenticator,
	totpVerifier platformtotp.Verifier,
	publisherProvider messagequeue.PublisherProvider,
	secretGenerator random.Generator,
	qrCodeBuilder qrcodes.Builder,
	queueConfig *queuescfg.Config,
) (AuthManagerInterface, error) {
	if queueConfig == nil {
		return nil, perrors.ErrNilInputParameter
	}

	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, queueConfig.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide data changes publisher: %w", err)
	}

	return &AuthManager{
		logger:                        logging.NewNamedLogger(logger, o11yName),
		tracer:                        tracing.NewNamedTracer(tracerProvider, o11yName),
		passwordResetTokenDataManager: passwordResetTokenDataManager,
		sessionDataManager:            sessionDataManager,
		userDataManager:               userDataManager,
		authenticator:                 authenticator,
		totpVerifier:                  totpVerifier,
		secretGenerator:               secretGenerator,
		qrCodeBuilder:                 qrCodeBuilder,
		dataChangesPublisher:          dataChangesPublisher,
		minimumPasswordLength:         0,
	}, nil
}

func (l *AuthManager) Self(ctx context.Context) (*identity.User, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching session context data")
	}
	tracing.AttachSessionContextDataToSpan(span, &sessionContextDataForTracing{sessionContextData})
	logger := sessionContextData.AttachToLogger(l.logger)

	// figure out who this is all for.
	requester := sessionContextData.GetUserID()
	tracing.AttachToSpan(span, platformkeys.RequesterIDKey, requester)

	// fetch user data.
	user, err := l.userDataManager.GetUser(ctx, requester)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("no such user")
		return nil, observability.PrepareError(err, span, "no such user")
	} else if err != nil {
		return nil, observability.PrepareError(err, span, "fetching user")
	}

	return user, nil
}

func (l *AuthManager) CheckUserPermissions(ctx context.Context, input *auth.UserPermissionsRequestInput) (*auth.UserPermissionsResponse, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, observability.PrepareError(perrors.ErrNilInputParameter, span, "nil input provided")
	}

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching session context data")
	}

	body := &auth.UserPermissionsResponse{
		Permissions: make(map[string]bool),
	}

	// A service-admin session may have no membership in the active account, so the account
	// permission checker can be absent from the map; comma-ok the lookup and guard the nil
	// interface value before calling a method on it.
	accountChecker, hasAccountChecker := sessionContextData.AccountPermissions[sessionContextData.GetActiveAccountID()]

	for _, perm := range input.Permissions {
		p := authorization.Permission(perm)
		hasAccountPerm := hasAccountChecker && accountChecker != nil && accountChecker.HasPermission(p)
		hasServicePerm := sessionContextData.GetServicePermissions().HasPermission(p)
		body.Permissions[perm] = hasAccountPerm || hasServicePerm
	}

	return body, nil
}

func (l *AuthManager) TOTPSecretVerification(ctx context.Context, input *auth.TOTPSecretVerificationInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	logger = logger.WithValue(identitykeys.UserIDKey, input.UserID)
	logger.Info("validated input, getting user")

	user, err := l.userDataManager.GetUserWithUnverifiedTwoFactorSecret(ctx, input.UserID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching user to verify two factor secret")
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)
	tracing.AttachToSpan(span, identitykeys.UsernameKey, user.Username)
	logger = logger.WithValue(identitykeys.UsernameKey, user.Username)

	if user.TwoFactorSecretVerifiedAt != nil {
		// I suppose if this happens too many times, we might want to keep track of that?
		return errors.New("two factor secret already verified")
	}

	// Verify through the injected verifier (rather than calling totp.Validate directly) so the
	// configured verifier is honored, and pass the non-nil verification error: PrepareError returns
	// nil on a nil error, which would otherwise report success on an invalid code.
	if verifyErr := l.totpVerifier.Verify(ctx, user.TwoFactorSecret, input.TOTPToken); verifyErr != nil {
		return observability.PrepareError(verifyErr, span, "TOTP code was invalid")
	}

	if err = l.userDataManager.MarkUserTwoFactorSecretAsVerified(ctx, user.ID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "verifying user two factor secret")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.TwoFactorSecretVerifiedServiceEventType,
		UserID:    user.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) NewTOTPSecret(ctx context.Context, input *auth.TOTPSecretRefreshInput) (*auth.TOTPSecretRefreshResponse, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "provided input was invalid")
	}

	tracing.AttachSessionContextDataToSpan(span, &sessionContextDataForTracing{sessionContextData})
	logger = sessionContextData.AttachToLogger(logger)

	// fetch user
	user, err := l.userDataManager.GetUser(ctx, sessionContextData.GetUserID())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareError(err, span, "user does not exist")
		}
		return nil, observability.PrepareError(err, span, "retrieving user from database")
	}

	if user.TwoFactorSecretVerifiedAt != nil {
		matches, validationErr := l.authenticator.PasswordMatches(ctx, user.HashedPassword, input.CurrentPassword)
		if validationErr != nil {
			return nil, observability.PrepareError(validationErr, span, "validating credentials")
		}

		if !matches {
			// Use an explicit error instead of the nil validationErr
			return nil, observability.PrepareError(errors.New("password mismatch"), span, "invalid credentials")
		}

		if verifyErr := l.totpVerifier.Verify(ctx, user.TwoFactorSecret, input.TOTPToken); verifyErr != nil {
			return nil, observability.PrepareError(verifyErr, span, "invalid credentials")
		}
	} else {
		return nil, observability.PrepareError(errors.New("unverified secret"), span, "two factor secret not yet verified")
	}

	// document who this is for.
	tracing.AttachToSpan(span, platformkeys.RequesterIDKey, sessionContextData.GetUserID())
	tracing.AttachToSpan(span, identitykeys.UsernameKey, user.Username)
	logger = logger.WithValue(identitykeys.UserIDKey, user.ID)

	// set the two factor secret.
	tfs, err := l.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating 2FA secret")
	}

	// update the user in the database.
	if err = l.userDataManager.MarkUserTwoFactorSecretAsUnverified(ctx, user.ID, tfs); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating 2FA secret")
	}

	user.TwoFactorSecret = tfs
	user.TwoFactorSecretVerifiedAt = nil

	dcm := &audit.DataChangeMessage{
		EventType: auth.TwoFactorSecretChangedServiceEventType,
		UserID:    user.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	qrCode, err := l.qrCodeBuilder.BuildQRCode(ctx, user.Username, user.TwoFactorSecret)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building QR code")
	}

	result := &auth.TOTPSecretRefreshResponse{
		TwoFactorSecret: user.TwoFactorSecret,
		TwoFactorQRCode: qrCode,
	}

	return result, nil
}

func (l *AuthManager) UpdatePassword(ctx context.Context, input *auth.PasswordUpdateInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return observability.PrepareError(err, span, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	if err = input.ValidateWithContext(ctx, l.minimumPasswordLength); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	// determine relevant user ID.
	tracing.AttachToSpan(span, platformkeys.RequesterIDKey, sessionContextData.GetUserID())
	logger = sessionContextData.AttachToLogger(logger)

	user, err := l.validateCredentialsForUpdateRequest(
		ctx,
		sessionContextData.GetUserID(),
		input.CurrentPassword,
		input.TOTPToken,
	)
	if err != nil {
		return observability.PrepareError(err, span, "validating credentials")
	}
	tracing.AttachToSpan(span, identitykeys.UsernameKey, user.Username)

	// ensure the password isn't garbage-tier
	if err = passwordvalidator.Validate(input.NewPassword, minimumPasswordEntropy); err != nil {
		return observability.PrepareError(err, span, "invalid password provided")
	}

	// hash the new password.
	newPasswordHash, err := l.authenticator.HashPassword(ctx, strings.TrimSpace(input.NewPassword))
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "hashing password")
	}

	// update the user.
	if err = l.userDataManager.UpdateUserPassword(ctx, user.ID, newPasswordHash); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.PasswordChangedEventType,
		UserID:    user.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) UpdateUserEmailAddress(ctx context.Context, input *auth.UserEmailAddressUpdateInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return observability.PrepareError(err, span, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	if err = input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}
	tracing.AttachToSpan(span, identitykeys.UserEmailAddressKey, input.NewEmailAddress)

	// determine relevant user ID.
	tracing.AttachToSpan(span, platformkeys.RequesterIDKey, sessionContextData.GetUserID())
	logger = sessionContextData.AttachToLogger(logger)

	user, err := l.validateCredentialsForUpdateRequest(
		ctx,
		sessionContextData.GetUserID(),
		input.CurrentPassword,
		input.TOTPToken,
	)
	if err != nil {
		return observability.PrepareError(err, span, "validating credentials")
	}

	// update the user.
	if err = l.userDataManager.UpdateUserEmailAddress(ctx, user.ID, input.NewEmailAddress); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.EmailAddressChangedEventType,
		UserID:    user.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) UpdateUserUsername(ctx context.Context, input *auth.UsernameUpdateInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return observability.PrepareError(err, span, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	if err = input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}
	tracing.AttachToSpan(span, identitykeys.UsernameKey, input.NewUsername)

	// determine relevant user ID.
	tracing.AttachToSpan(span, platformkeys.RequesterIDKey, sessionContextData.GetUserID())
	logger = sessionContextData.AttachToLogger(logger)

	user, err := l.validateCredentialsForUpdateRequest(
		ctx,
		sessionContextData.GetUserID(),
		input.CurrentPassword,
		input.TOTPToken,
	)
	if err != nil {
		return observability.PrepareError(err, span, "validating credentials")
	}

	// update the user.
	if err = l.userDataManager.UpdateUserUsername(ctx, user.ID, input.NewUsername); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.UsernameChangedEventType,
		UserID:    user.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) RequestUsernameReminder(ctx context.Context, input *auth.UsernameReminderRequestInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	// The session is optional: a user who forgot their username can't be authenticated, so this
	// mirrors the password-reset flow and only decorates the logger when a session happens to exist.
	// This flow is reachable without authentication, so a missing session is expected rather than
	// an error; attribute the log line only when one is present.
	if sessionContextData := sessions.FromContext(ctx); sessionContextData != nil {
		logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	u, err := l.userDataManager.GetUserByEmail(ctx, input.EmailAddress)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// Do not leak user existence; return success without sending a reminder.
		return nil
	} else if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching user")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.UsernameReminderRequestedEventType,
		UserID:    u.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) CreatePasswordResetToken(ctx context.Context, input *auth.PasswordResetTokenCreationRequestInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)
	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	u, err := l.userDataManager.GetUserByEmail(ctx, input.EmailAddress)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// Do not leak user existence; return success without sending email.
		return nil
	}
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching user")
	}

	token, err := l.secretGenerator.GenerateBase32EncodedString(ctx, passwordResetTokenSize)
	if err != nil {
		return observability.PrepareError(err, span, "generating secret")
	}

	dbInput := &auth.PasswordResetTokenDatabaseCreationInput{
		ID:            identifiers.New(),
		Token:         token,
		BelongsToUser: u.ID,
		ExpiresAt:     time.Now().Add(30 * time.Minute),
	}

	t, err := l.passwordResetTokenDataManager.CreatePasswordResetToken(ctx, dbInput)
	if err != nil {
		return observability.PrepareError(err, span, "creating password reset token")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.PasswordResetTokenCreatedEventType,
		UserID:    u.ID,
		Context: map[string]any{
			authkeys.PasswordResetTokenIDKey: t.ID,
		},
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) PasswordResetTokenRedemption(ctx context.Context, input *auth.PasswordResetTokenRedemptionRequestInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)
	// This flow is reachable without authentication, so a missing session is expected rather than
	// an error; attribute the log line only when one is present.
	if sessionContextData := sessions.FromContext(ctx); sessionContextData != nil {
		logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	t, err := l.passwordResetTokenDataManager.GetPasswordResetTokenByToken(ctx, input.Token)
	if errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "password reset token not found")
	} else if err != nil {
		return observability.PrepareError(err, span, "fetching password reset token")
	}

	u, err := l.userDataManager.GetUser(ctx, t.BelongsToUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}
		return observability.PrepareError(err, span, "fetching user")
	}

	// ensure the password isn't garbage-tier
	if err = passwordvalidator.Validate(strings.TrimSpace(input.NewPassword), minimumPasswordEntropy); err != nil {
		return observability.PrepareError(err, span, "provided password was invalid")
	}

	// hash the new password.
	newPasswordHash, err := l.authenticator.HashPassword(ctx, strings.TrimSpace(input.NewPassword))
	if err != nil {
		return observability.PrepareError(err, span, "hashing password")
	}

	// update the user.
	if err = l.userDataManager.UpdateUserPassword(ctx, u.ID, newPasswordHash); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating user")
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}

		return observability.PrepareError(err, span, "updating user")
	}

	if redemptionErr := l.passwordResetTokenDataManager.RedeemPasswordResetToken(ctx, t.ID); redemptionErr != nil {
		observability.AcknowledgeError(redemptionErr, logger, span, "redeeming password reset token")
		if errors.Is(redemptionErr, sql.ErrNoRows) {
			return observability.PrepareError(redemptionErr, span, "redeeming password reset token not found")
		}

		return observability.PrepareError(redemptionErr, span, "redeeming password reset token")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.PasswordResetTokenRedeemedEventType,
		UserID:    t.BelongsToUser,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

func (l *AuthManager) RequestEmailVerificationEmail(ctx context.Context) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return observability.PrepareError(err, span, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	verificationToken, err := l.userDataManager.GetEmailAddressVerificationTokenForUser(ctx, sessionContextData.GetUserID())
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "email verification token not found")
	} else if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching email verification token")
	}

	l.dataChangesPublisher.PublishAsync(ctx, &audit.DataChangeMessage{
		EventType: auth.UserEmailAddressVerificationEmailRequestedEventType,
		UserID:    sessionContextData.GetUserID(),
		Context: map[string]any{
			identitykeys.UserEmailVerificationTokenKey: verificationToken,
		},
	})

	return nil
}

func (l *AuthManager) VerifyUserEmailAddress(ctx context.Context, input *auth.EmailAddressVerificationRequestInput) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return observability.PrepareError(err, span, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	if err = input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	user, err := l.userDataManager.GetUserByEmailAddressVerificationToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}
		return observability.PrepareAndLogError(err, logger, span, "fetching user")
	}

	if err = l.userDataManager.MarkUserEmailAddressAsVerified(ctx, user.ID, input.Token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}
		return observability.PrepareAndLogError(err, logger, span, "marking user email as verified")
	}

	l.dataChangesPublisher.PublishAsync(ctx, &audit.DataChangeMessage{
		EventType: auth.UserEmailAddressVerifiedEventType,
		UserID:    user.ID,
	})

	return nil
}

// VerifyUserEmailAddressByToken verifies a user's email address using only the verification token.
// It does not require session context and is used for unauthenticated verification (e.g., from email links).
func (l *AuthManager) VerifyUserEmailAddressByToken(ctx context.Context, token string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithSpan(span)

	input := &auth.EmailAddressVerificationRequestInput{Token: token}
	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "provided input was invalid")
	}

	user, err := l.userDataManager.GetUserByEmailAddressVerificationToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}
		return observability.PrepareAndLogError(err, logger, span, "fetching user")
	}

	if err = l.userDataManager.MarkUserEmailAddressAsVerified(ctx, user.ID, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}
		return observability.PrepareAndLogError(err, logger, span, "marking user email as verified")
	}

	l.dataChangesPublisher.PublishAsync(ctx, &audit.DataChangeMessage{
		EventType: auth.UserEmailAddressVerifiedEventType,
		UserID:    user.ID,
	})

	return nil
}

// validateCredentialsForUpdateRequest takes a user's credentials and determines if they match what is on record.
func (l *AuthManager) validateCredentialsForUpdateRequest(ctx context.Context, userID, password, totpToken string) (*identity.User, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithValue(identitykeys.UserIDKey, userID)

	// fetch user data.
	user, err := l.userDataManager.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		logger.Error("error encountered fetching user", err)
		return nil, observability.PrepareError(err, span, "fetching user")
	}

	if user.TwoFactorSecretVerifiedAt != nil && totpToken == "" {
		// Pass an explicit error: PrepareError returns nil on a nil error, which would make callers
		// treat a missing TOTP code as success and then dereference the nil *identity.User below.
		return nil, observability.PrepareError(ErrTOTPTokenRequired, span, "two factor secret not provided")
	}

	tfs := user.TwoFactorSecret
	if user.TwoFactorSecretVerifiedAt == nil {
		tfs = ""
		totpToken = ""
	}

	// validate password.
	matches, err := l.authenticator.PasswordMatches(ctx, user.HashedPassword, password)
	if err != nil {
		return nil, observability.PrepareError(err, span, "error validating credentials")
	} else if !matches {
		// PasswordMatches returns (false, nil) on mismatch; pass an explicit error so callers don't
		// treat the mismatch as success and dereference the nil *identity.User.
		return nil, observability.PrepareError(ErrInvalidCredentials, span, "credentials are not valid")
	}

	// verify TOTP code (if applicable). If TOTP is not enabled on the user, tfs is
	// empty and totpToken is empty, so we skip TOTP verification entirely.
	if tfs != "" {
		if verifyErr := l.totpVerifier.Verify(ctx, tfs, totpToken); verifyErr != nil {
			return nil, observability.PrepareError(verifyErr, span, "credentials are not valid")
		}
	}

	return user, nil
}

// GetActiveSessionsForUser returns all active sessions for a user.
func (l *AuthManager) GetActiveSessionsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	return l.sessionDataManager.GetActiveSessionsForUser(ctx, userID, filter)
}

// RevokeSession revokes a specific user session.
func (l *AuthManager) RevokeSession(ctx context.Context, sessionID, userID string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	return l.sessionDataManager.RevokeUserSession(ctx, sessionID, userID)
}

// RevokeAllSessionsForUserExcept revokes all sessions for a user except the specified one.
func (l *AuthManager) RevokeAllSessionsForUserExcept(ctx context.Context, userID, currentSessionID string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	return l.sessionDataManager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID)
}

// RevokeAllSessionsForUser revokes all sessions for a user.
func (l *AuthManager) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	return l.sessionDataManager.RevokeAllSessionsForUser(ctx, userID)
}
