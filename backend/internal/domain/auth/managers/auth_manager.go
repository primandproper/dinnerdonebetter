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

	"github.com/primandproper/platform-go/v13/authentication/passwordreset"
	platformtotp "github.com/primandproper/platform-go/v13/authentication/totp"
	perrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/messagequeue"
	"github.com/primandproper/platform-go/v13/observability"
	platformkeys "github.com/primandproper/platform-go/v13/observability/keys"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/qrcodes"
	"github.com/primandproper/platform-go/v13/random"
	"github.com/primandproper/platform-go/v13/tenancy"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	o11yName               = "auth_manager"
	totpSecretSize         = 64
	minimumPasswordEntropy = 60

	// passwordResetTokenLifetime is how long a reset link is good for.
	//
	// It is passed per issuance because the store takes it that way: a TTL is a policy
	// decision, and a configured default would be the value nobody chose. Thirty minutes is
	// long enough to walk to a phone and short enough that a link left in an inbox is not a
	// standing key to the account.
	passwordResetTokenLifetime = 30 * time.Minute
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
	passwordResetTokens   passwordreset.Store
	sessionStore          auth.SessionStore
	userDataManager       identity.UserDataManager
	tracer                tracing.Tracer
	authenticator         authentication.Authenticator
	totpVerifier          platformtotp.Verifier
	logger                logging.Logger
	dataChangesPublisher  messagequeue.Publisher
	secretGenerator       random.Generator
	qrCodeBuilder         qrcodes.Builder
	minimumPasswordLength uint8
}

func ProvideAuthManager(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	passwordResetTokens passwordreset.Store,
	sessionStore auth.SessionStore,
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
		logger:                logging.NewNamedLogger(logger, o11yName),
		tracer:                tracing.NewNamedTracer(tracerProvider, o11yName),
		passwordResetTokens:   passwordResetTokens,
		sessionStore:          sessionStore,
		userDataManager:       userDataManager,
		authenticator:         authenticator,
		totpVerifier:          totpVerifier,
		secretGenerator:       secretGenerator,
		qrCodeBuilder:         qrCodeBuilder,
		dataChangesPublisher:  dataChangesPublisher,
		minimumPasswordLength: 0,
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

	// Global scope: a reset identifies a person, not a person within an account. The user
	// asking for one is not signed in and has no active account to name, and a link that
	// only worked in the account they happened to have selected last would be a link that
	// stops working when they switch.
	issuance, err := l.passwordResetTokens.Issue(ctx, tenancy.Global(), u.ID, passwordResetTokenLifetime)
	if err != nil {
		return observability.PrepareError(err, span, "creating password reset token")
	}

	// The secret rides on the message rather than being fetched back out of the store,
	// because it cannot be fetched back: the row holds a digest, and this issuance is the
	// only place the secret will ever exist. The email handler is its one consumer, and it
	// puts it in a link and nowhere else. The email verification token travels the same
	// way for the same reason.
	dcm := &audit.DataChangeMessage{
		EventType: auth.PasswordResetTokenCreatedEventType,
		UserID:    u.ID,
		Context: map[string]any{
			authkeys.PasswordResetTokenIDKey:     issuance.Token.ID,
			authkeys.PasswordResetTokenSecretKey: issuance.Secret,
		},
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	return nil
}

// PasswordResetTokenRedemption spends a reset link and writes the password it was issued for.
//
// The order is the one that fails safe. The password is vetted first, because rejecting a
// weak password should not cost the user their link; then the token is consumed, which is
// the store's atomic decision about which of two racing requests owns it; then the password
// is written. Consuming after the write would leave a live reset link for an account whose
// password has just changed, which is the worse of the two failures — the other costs an
// email.
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

	// ensure the password isn't garbage-tier
	newPassword := strings.TrimSpace(input.NewPassword)
	if err := passwordvalidator.Validate(newPassword, minimumPasswordEntropy); err != nil {
		return observability.PrepareError(err, span, "provided password was invalid")
	}

	// Single use is decided here, by the store, in one statement. Two requests answering the
	// same link at the same instant both find the row live; exactly one of them gets a token
	// back and the other is told it has already been redeemed.
	t, err := l.passwordResetTokens.Consume(ctx, tenancy.Global(), input.Token)
	if err != nil {
		return observability.PrepareError(err, span, "redeeming password reset token")
	}
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, t.ID)

	u, err := l.userDataManager.GetUser(ctx, t.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}
		return observability.PrepareError(err, span, "fetching user")
	}

	// hash the new password.
	newPasswordHash, err := l.authenticator.HashPassword(ctx, newPassword)
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

	// Every other link this user was holding stops working. Somebody who asked for a reset
	// twice and completed the second one should not be left with a first link that still
	// resets the password they just chose.
	if _, err = l.passwordResetTokens.RevokeForUser(ctx, tenancy.Global(), u.ID); err != nil {
		// The reset itself succeeded, so this is reported rather than returned: failing the
		// request here would tell the user their password did not change when it did.
		observability.AcknowledgeError(err, logger, span, "revoking outstanding password reset tokens")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.PasswordResetTokenRedeemedEventType,
		UserID:    t.UserID,
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

// GetActiveSessionsForUser returns the live sessions a user holds, newest first.
//
// currentSessionID is the session the caller is asking with, and decides which of the
// returned sessions is flagged as theirs. It is not a filter — a security page that hid the
// session you are reading it from would be listing the wrong set — and the empty string is
// what the admin read passes, since none of the listed sessions is the administrator's.
//
// There is no page here. A person's live sessions are the devices they are signed in on,
// which is a handful, and the store returns the set rather than a window onto it.
func (l *AuthManager) GetActiveSessionsForUser(ctx context.Context, userID, currentSessionID string) ([]*auth.UserSession, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	return l.sessionStore.List(ctx, auth.SessionHolder(userID), currentSessionID)
}

// RevokeSession revokes a specific user session.
//
// The user is part of the question rather than checked beforehand: the store decides "this
// session, and it is theirs" where the row goes, and answers a session that is not the
// named user's as absent rather than as forbidden — so the answer does not confirm that
// somebody else's identifier names anything.
func (l *AuthManager) RevokeSession(ctx context.Context, sessionID, userID string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	return l.sessionStore.Revoke(ctx, auth.SessionHolder(userID), sessionID)
}

// RevokeAllSessionsForUserExcept revokes all sessions for a user except the specified one.
func (l *AuthManager) RevokeAllSessionsForUserExcept(ctx context.Context, userID, currentSessionID string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	_, err := l.sessionStore.RevokeAllExcept(ctx, auth.SessionHolder(userID), currentSessionID)

	return err
}

// RevokeAllSessionsForUser revokes all sessions for a user.
func (l *AuthManager) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	_, err := l.sessionStore.RevokeAll(ctx, auth.SessionHolder(userID))

	return err
}
