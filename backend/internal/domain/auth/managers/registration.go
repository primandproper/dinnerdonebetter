package managers

import (
	"context"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	perrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

// RegisterUser signs somebody up: a user, the credentials they will sign in with, and
// either the account they now own or the one the invitation they answered admits them to.
//
// It is here rather than on the identity surface because that surface is platform's, and
// platform's Register is made on a registrar's behalf — it holds no policy about who may
// sign up, which is the whole of what this method adds. The password rules, the agreements
// this application requires, the TOTP secret and the QR code rendered from it are all the
// consumer's, and none of them is something a directory should have an opinion about.
func (l *AuthManager) RegisterUser(ctx context.Context, input *auth.UserRegistrationInput) (*auth.UserCreationResponse, error) {
	ctx, span := l.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, observability.PrepareError(perrors.ErrNilInputParameter, span, "nil registration input")
	}

	// Trimmed before validation rather than after, so the address the email rule parses is
	// the address the column receives. The handle is not folded here — the store folds it
	// on write and on every lookup, and a second fold in front of it is a second place for
	// the two to disagree.
	input.Username = strings.TrimSpace(input.Username)
	input.EmailAddress = strings.TrimSpace(strings.ToLower(input.EmailAddress))
	input.Password = strings.TrimSpace(input.Password)

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UsernameKey:            input.Username,
		identitykeys.UserEmailAddressKey:    input.EmailAddress,
		identitykeys.AccountInvitationIDKey: input.InvitationID,
	}, span, l.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(platformkeys.ValidationErrorKey, err).Debug("provided registration input was invalid")

		return nil, observability.PrepareError(err, span, "invalid registration input provided")
	}

	if err := passwordvalidator.Validate(input.Password, minimumPasswordEntropy); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "weak password provided for user creation")
	}

	hashedPassword, err := l.authenticator.HashPassword(ctx, input.Password)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "hashing user creation password")
	}

	twoFactorSecret, err := l.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating two factor secret")
	}

	// The verification token rides in on the user rather than being set afterwards,
	// because the store holds only its digest and the mail that carries the secret is
	// queued from the registration's own hook. A token set in a second call would be a
	// second transaction, and a link that committed without the user it proves.
	verificationToken, err := l.secretGenerator.GenerateBase32EncodedString(ctx, emailVerificationTokenSize)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating email verification token")
	}

	registrant := &platformidentity.User{
		ID:                            identifiers.New(),
		Username:                      input.Username,
		EmailAddress:                  input.EmailAddress,
		FirstName:                     input.FirstName,
		LastName:                      input.LastName,
		HashedPassword:                hashedPassword,
		TwoFactorSecret:               twoFactorSecret,
		EmailAddressVerificationToken: verificationToken,
		ServiceRoles:                  []string{authorization.ServiceUserRoleName},
	}

	user, accountID, err := l.register(ctx, input, registrant)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "registering user")
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)

	// Recorded after the registration rather than as part of it. The input requires both,
	// so a user in the directory without them is not a state this path can produce; what a
	// failure here costs is the stamp, not the registration, which is why it is reported
	// rather than returned.
	if _, agreementErr := l.directory.RecordAgreement(ctx, ddbidentity.Scope(), user.ID,
		platformidentity.TermsOfService, platformidentity.PrivacyPolicy); agreementErr != nil {
		observability.AcknowledgeError(agreementErr, logger, span, "recording registrant's agreements")
	}

	// The QR code is a convenience rendering of the secret, which is returned beside it;
	// don't fail a signup over it.
	twoFactorQRCode, qrCodeErr := l.qrCodeBuilder.BuildQRCode(ctx, user.Username, twoFactorSecret)
	if qrCodeErr != nil {
		observability.AcknowledgeError(qrCodeErr, logger, span, "building two factor QR code")
	}

	// No signup event is published here, and its absence is deliberate. The registration's
	// hook writes user_signed_up into the outbox on the transaction that made the user, so
	// the event and the rows it announces commit together — see
	// internal/repositories/postgres/identitystore/hooks.go. A second publish from out here
	// would tell every subscriber twice.

	return &auth.UserCreationResponse{
		CreatedAt:        user.CreatedAt,
		Username:         user.Username,
		EmailAddress:     user.EmailAddress,
		TwoFactorQRCode:  twoFactorQRCode,
		CreatedUserID:    user.ID,
		CreatedAccountID: accountID,
		AccountStatus:    string(user.AccountStatus),
		TwoFactorSecret:  twoFactorSecret,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
	}, nil
}

// register makes the registration the input describes, and answers with the registrant and
// the account they landed in.
//
// Two operations rather than one with a flag, because platform ships two and the difference
// is not a detail: a registration by invitation mints no account — the registrant is
// joining one that already exists — and an invitation that no longer admits them takes the
// whole registration down with it rather than leaving a user committed against a dead link.
func (l *AuthManager) register(
	ctx context.Context,
	input *auth.UserRegistrationInput,
	registrant *platformidentity.User,
) (*platformidentity.User, string, error) {
	if input.InvitationID != "" && input.InvitationToken != "" {
		registration, err := l.directory.RegisterWithInvitation(ctx, ddbidentity.Scope(), registrant,
			input.InvitationID, input.InvitationToken, "")
		if err != nil {
			return nil, "", err
		}

		return registration.User, registration.Membership.BelongsToAccount, nil
	}

	accountName := strings.TrimSpace(input.AccountName)
	if accountName == "" {
		accountName = input.Username + "'s account"
	}

	registration, err := l.directory.Register(ctx, ddbidentity.Scope(), registrant, &platformidentity.Account{
		Name: accountName,
	}, []string{authorization.AccountAdminRoleName})
	if err != nil {
		return nil, "", err
	}

	return registration.User, registration.Account.ID, nil
}
