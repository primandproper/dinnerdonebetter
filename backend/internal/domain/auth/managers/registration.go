package managers

import (
	"context"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v14/authentication/signin"
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

	twoFactorSecret, err := l.secretGenerator.GenerateBase32EncodedString(ctx, totpSecretSize)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating two factor secret")
	}

	registrant := &platformidentity.User{
		ID: identifiers.New(),

		// Good standing from the moment they register, which disagrees with platform's
		// default and is this application's call to make.
		//
		// platform defaults a new user to unverified and refuses sign-in to anything but
		// good, on the reading that "admitting them by default is the mistake worth making
		// unspellable". That is the right default for a directory. It is not this
		// application's policy: nothing here has ever gated use on a verified email — the
		// verification mail exists, and a user who ignores it keeps cooking — so adopting
		// platform's reading would be a product change arriving as a side effect of a
		// library upgrade, and would lock out everybody who registered and closed the tab.
		//
		// The status is the consumer's to set, which is what makes this a statement rather
		// than a workaround: platform supplies a default for a caller who names none.
		// Gating on verification later means writing unverified here and promoting on
		// MarkUserEmailAddressVerified, and nothing else.
		AccountStatus: platformidentity.StatusGood,
		Username:      input.Username,
		EmailAddress:  input.EmailAddress,
		FirstName:     input.FirstName,
		LastName:      input.LastName,

		// The second factor is this application's to mint and is not a credential signin
		// knows about: it overwrites HashedPassword and EmailAddressVerificationToken on
		// the user it is handed, and leaves everything else — this secret, the standing
		// above, the service role below — exactly as given.
		TwoFactorSecret: twoFactorSecret,
		ServiceRoles:    []string{authorization.ServiceUserRoleName},
	}

	// One call for both registrations. The hash, the verification token and the directory
	// write happen inside it, on one transaction: naming an invitation runs identity's
	// invitation registration and mints no account, naming none runs the ordinary one.
	//
	// The password arrives as a Credential rather than a hash because a caller who could
	// supply a hash is a caller who could choose somebody's secret, which is why the field
	// on the user is ignored. NoPassword is the other member of that set and this
	// application has no passwordless arrival, so naming it is not a branch here.
	accountName := strings.TrimSpace(input.AccountName)
	if accountName == "" {
		accountName = input.Username + "'s account"
	}

	registered, err := l.signIn.Register(ctx, ddbidentity.Scope(), &signin.Registration{
		User:            registrant,
		Account:         &platformidentity.Account{Name: accountName},
		Credential:      signin.Password(input.Password),
		InvitationID:    input.InvitationID,
		InvitationToken: input.InvitationToken,
		OwnerRoles:      []string{authorization.AccountAdminRoleName},
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "registering user")
	}

	user := registered.User
	accountID := registered.Membership.BelongsToAccount

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
