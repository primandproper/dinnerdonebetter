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
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformtotp "github.com/primandproper/primitives-go/v2/authentication/totp"
	"github.com/primandproper/primitives-go/v2/database"
	perrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	"github.com/primandproper/primitives-go/v2/observability"
	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/qrcodes"
	"github.com/primandproper/primitives-go/v2/random"
	"github.com/primandproper/primitives-go/v2/tenancy"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

const (
	o11yName       = "auth_manager"
	totpSecretSize = 64

	// emailVerificationTokenSize is how many bytes of entropy a verification link carries.
	// It is a bearer credential that proves an address, so it is sized like the TOTP secret
	// beside it rather than like an identifier.
	emailVerificationTokenSize = 64

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
	// db supplies the transaction the password reset store's writes now take.
	// Each of the three is a single write, so each gets one transaction — and
	// the audit entry the repository wraps around it commits with it.
	db                    database.Client
	passwordResetTokens   passwordreset.Store
	sessionStore          auth.SessionStore
	directory             *platformidentity.Service
	users                 platformidentity.Store
	signIn                *signin.Service
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
	db database.Client,
	passwordResetTokens passwordreset.Store,
	sessionStore auth.SessionStore,
	directory *platformidentity.Service,
	users platformidentity.Store,
	signInService *signin.Service,
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
		db:                    db,
		passwordResetTokens:   passwordResetTokens,
		sessionStore:          sessionStore,
		directory:             directory,
		users:                 users,
		signIn:                signInService,
		authenticator:         authenticator,
		totpVerifier:          totpVerifier,
		secretGenerator:       secretGenerator,
		qrCodeBuilder:         qrCodeBuilder,
		dataChangesPublisher:  dataChangesPublisher,
		minimumPasswordLength: 0,
	}, nil
}

func (l *AuthManager) Self(ctx context.Context) (*platformidentity.User, error) {
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
	user, err := l.users.GetUser(ctx, l.db.Reader(), ddbidentity.Scope(), requester)
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
	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.UserID)

	// The read, the already-verified check, the code comparison and the write were four
	// steps here and are one call now. platform refuses a replayed verification with a
	// sentinel of its own rather than with a bare errors.New, which is what this returned
	// — so "already verified" is something a caller can match on instead of a string.
	if err := l.signIn.VerifyTOTPSecret(ctx, ddbidentity.Scope(), input.UserID, input.TOTPToken); err != nil {
		return observability.PrepareError(err, span, "verifying two factor secret")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.TwoFactorSecretVerifiedServiceEventType,
		UserID:    input.UserID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	logger.Info("two factor secret verified")

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
	tracing.AttachToSpan(span, platformkeys.RequesterIDKey, sessionContextData.GetUserID())

	// The password check, the code check, the secret generation and the write are one call.
	//
	// One rule loosened with it, deliberately. This refused a user whose current secret was
	// not yet proven — "two factor secret not yet verified" — which left somebody who
	// registered, never finished enrollment and lost the secret with no way to get another;
	// the only credential they could still prove was the password, and this is the call a
	// password proves. platform asks for the password always and for a code only from
	// somebody who holds a proven secret, which is the same protection without the corner.
	enrollment, err := l.signIn.RefreshTOTPSecret(ctx, ddbidentity.Scope(), sessionContextData.GetUserID(),
		&signin.SecretRefresh{
			CurrentPassword: input.CurrentPassword,
			TOTPCode:        input.TOTPToken,
		})
	if err != nil {
		return nil, observability.PrepareError(err, span, "refreshing two factor secret")
	}

	// The username, for the QR code's label. platform hands back an otpauth URI carrying
	// the same thing, and rendering that instead would mean a builder that takes a URI —
	// worth doing when something else wants one.
	user, err := l.users.GetUser(ctx, l.db.Reader(), ddbidentity.Scope(), sessionContextData.GetUserID())
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "reading the user the secret belongs to")
	}

	tracing.AttachToSpan(span, identitykeys.UsernameKey, user.Username)

	dcm := &audit.DataChangeMessage{
		EventType: auth.TwoFactorSecretChangedServiceEventType,
		UserID:    user.ID,
	}

	l.dataChangesPublisher.PublishAsync(ctx, dcm)

	qrCode, err := l.qrCodeBuilder.BuildQRCode(ctx, user.Username, enrollment.Secret)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building QR code")
	}

	return &auth.TOTPSecretRefreshResponse{
		TwoFactorSecret: enrollment.Secret,
		TwoFactorQRCode: qrCode,
	}, nil
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

	// The password policy stays here, and platform says so: PasswordUpdate's doc notes
	// that whether a password is long enough or unusual enough is the consumer's rule,
	// applied before the call, because a policy inside the package is one every consumer
	// then has to work around. Ours is the entropy floor below.
	if err = passwordvalidator.Validate(input.NewPassword, minimumPasswordEntropy); err != nil {
		return observability.PrepareError(err, span, "invalid password provided")
	}

	// What moves is everything after it: the credential check, the hash and the write.
	// The current password is required whatever the session says, which is platform's
	// rule and was this application's too — a session is not proof enough to change the
	// credential the session was obtained with.
	if err = l.signIn.UpdatePassword(ctx, ddbidentity.Scope(), sessionContextData.GetUserID(),
		&signin.PasswordUpdate{
			CurrentPassword: input.CurrentPassword,
			NewPassword:     strings.TrimSpace(input.NewPassword),
			TOTPCode:        input.TOTPToken,
		}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating password")
	}

	dcm := &audit.DataChangeMessage{
		EventType: auth.PasswordChangedEventType,
		UserID:    sessionContextData.GetUserID(),
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
	if _, err = l.directory.UpdateProfile(ctx, ddbidentity.Scope(), user.ID, &platformidentity.ProfileUpdate{
		EmailAddress: &input.NewEmailAddress,
	}); err != nil {
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
	if _, err = l.directory.UpdateProfile(ctx, ddbidentity.Scope(), user.ID, &platformidentity.ProfileUpdate{
		Username: &input.NewUsername,
	}); err != nil {
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

	u, err := l.users.GetUserByEmailAddress(ctx, l.db.Reader(), ddbidentity.Scope(), input.EmailAddress)
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

	u, err := l.users.GetUserByEmailAddress(ctx, l.db.Reader(), ddbidentity.Scope(), input.EmailAddress)
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
	issuance, err := inTransaction(ctx, l.db, func(tx database.Tx) (*passwordreset.Issuance, error) {
		return l.passwordResetTokens.Issue(ctx, tx, tenancy.Global(), u.ID, passwordResetTokenLifetime)
	})
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
	t, err := inTransaction(ctx, l.db, func(tx database.Tx) (*passwordreset.Token, error) {
		return l.passwordResetTokens.Consume(ctx, tx, tenancy.Global(), input.Token)
	})
	if err != nil {
		return observability.PrepareError(err, span, "redeeming password reset token")
	}
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, t.ID)

	u, err := l.users.GetUser(ctx, l.db.Reader(), ddbidentity.Scope(), t.UserID)
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
	if _, err = l.directory.UpdateUserPassword(ctx, ddbidentity.Scope(), u.ID, newPasswordHash); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating user")
		if errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareError(err, span, "user not found")
		}

		return observability.PrepareError(err, span, "updating user")
	}

	// Every other link this user was holding stops working. Somebody who asked for a reset
	// twice and completed the second one should not be left with a first link that still
	// resets the password they just chose.
	if _, err = inTransaction(ctx, l.db, func(tx database.Tx) (int64, error) {
		return l.passwordResetTokens.RevokeForUser(ctx, tx, tenancy.Global(), u.ID)
	}); err != nil {
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

	// Minted here rather than read back, because there is nothing to read back: the
	// column holds a digest, and no read fills the secret in. That is a change in
	// behaviour and the right one — asking for the mail again issues a fresh link and
	// retires the one that went missing, where the read this replaced re-sent whatever
	// token was already outstanding forever.
	verificationToken, err := l.secretGenerator.GenerateBase32EncodedString(ctx, emailVerificationTokenSize)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "generating email verification token")
	}

	if _, err = l.directory.SetUserEmailAddressVerificationToken(ctx, ddbidentity.Scope(),
		sessionContextData.GetUserID(), verificationToken); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "storing email verification token")
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

	// Read first, for the user id the event below names — VerifyEmailAddress answers
	// with an error and nothing else, and the link is spent by the time it returns.
	user, err := l.users.GetUserByEmailVerificationToken(ctx, l.db.Reader(), ddbidentity.Scope(), input.Token)
	if err != nil {
		// Deliberately not told apart from a token that simply does not match. This used
		// to answer "user not found" for an unknown token and something else for a write
		// that matched no row, which made the endpoint a way to ask whether a given
		// verification token was live. signin's rule is that expired, already spent,
		// never issued and simply wrong are one answer, because the caller's remedy is
		// the same in every case and telling them apart tells whoever is guessing which
		// guesses are getting warm.
		return observability.PrepareError(signin.ErrInvalidVerificationToken, span, "verifying email address")
	}

	if err = l.signIn.VerifyEmailAddress(ctx, ddbidentity.Scope(), input.Token); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "verifying email address")
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

	// Read first, for the user id the event below names — VerifyEmailAddress answers
	// with an error and nothing else, and the link is spent by the time it returns.
	user, err := l.users.GetUserByEmailVerificationToken(ctx, l.db.Reader(), ddbidentity.Scope(), token)
	if err != nil {
		// Deliberately not told apart from a token that simply does not match. This used
		// to answer "user not found" for an unknown token and something else for a write
		// that matched no row, which made the endpoint a way to ask whether a given
		// verification token was live. signin's rule is that expired, already spent,
		// never issued and simply wrong are one answer, because the caller's remedy is
		// the same in every case and telling them apart tells whoever is guessing which
		// guesses are getting warm.
		return observability.PrepareError(signin.ErrInvalidVerificationToken, span, "verifying email address")
	}

	if err = l.signIn.VerifyEmailAddress(ctx, ddbidentity.Scope(), token); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "verifying email address")
	}

	l.dataChangesPublisher.PublishAsync(ctx, &audit.DataChangeMessage{
		EventType: auth.UserEmailAddressVerifiedEventType,
		UserID:    user.ID,
	})

	return nil
}

// validateCredentialsForUpdateRequest takes a user's credentials and determines if they match what is on record.
func (l *AuthManager) validateCredentialsForUpdateRequest(ctx context.Context, userID, password, totpToken string) (*platformidentity.User, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	logger := l.logger.WithValue(identitykeys.UserIDKey, userID)

	// fetch user data.
	user, err := l.users.GetUser(ctx, l.db.Reader(), ddbidentity.Scope(), userID)
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

// inTransaction runs one store write on a transaction of its own.
//
// As of platform-go v14 a store holds no database handle: a write takes the
// caller's database.Tx. Every write in this service is a single store call, so
// each gets one transaction — which is exactly what the store opened for itself
// before the caller was required to supply it. A handler that ever writes twice
// should take one transaction across both rather than call this twice.
func inTransaction[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
