package managers

import (
	"context"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
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

// errNoSecondFactorMinted is a registration that came back without the second factor
// authentication.RegistrationPolicy asks for — a sign-in service built without that policy.
// It is a wiring failure, and refusing to answer is better than answering a registrant with
// an empty secret they would be asked for at their first sign-in.
var errNoSecondFactorMinted = perrors.New("registration minted no second factor")

// RegisterUser signs somebody up: a user, the credentials they will sign in with, and
// either the account they now own or the one the invitation they answered admits them to.
//
// It is here rather than on the identity surface because that surface is platform's, and
// platform's Register is made on a registrar's behalf. What this door adds over signin's is
// the wire it answers on — the agreements as two booleans, the QR code rendered from the
// second factor — and the input's own validation. Who may sign up, and as what, is
// authentication.RegistrationPolicy's, which signin applies here and on its own Register.
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

	if err := passwordvalidator.Validate(input.Password, authentication.MinimumPasswordEntropy); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "weak password provided for user creation")
	}

	// What this application adds to a registrant — good standing, the service role, the
	// second factor, the account's name and its owner's role, and the agreements it insists
	// on — is authentication.RegistrationPolicy's, which signin applies to every
	// registration it writes. It is not repeated here, because platform's SignInService
	// registers through the same policy and two statements of it would be two registrations
	// that could come to disagree.
	//
	// One call for both registrations. The hash, the verification token, the second factor
	// and the directory write happen inside it, on one transaction: naming an invitation
	// runs identity's invitation registration and mints no account, naming none runs the
	// ordinary one. The agreements are stamped on that transaction too, rather than by a
	// second write afterwards that could fail and leave a registrant with no record of
	// having accepted anything.
	//
	// The password arrives as a Credential rather than a hash because a caller who could
	// supply a hash is a caller who could choose somebody's secret, which is why the field
	// on the user is ignored. NoPassword is the other member of that set and this
	// application has no passwordless arrival, so naming it is not a branch here.
	registered, err := l.signIn.Register(ctx, ddbidentity.Scope(), &signin.Registration{
		User: &platformidentity.User{
			ID:           identifiers.New(),
			Username:     input.Username,
			EmailAddress: input.EmailAddress,
			FirstName:    input.FirstName,
			LastName:     input.LastName,
		},
		Account:         &platformidentity.Account{Name: input.AccountName},
		Credential:      signin.Password(input.Password),
		InvitationID:    input.InvitationID,
		InvitationToken: input.InvitationToken,
		Agreements:      []platformidentity.Agreement{platformidentity.TermsOfService, platformidentity.PrivacyPolicy},
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "registering user")
	}

	if registered.TOTPEnrollment == nil {
		return nil, observability.PrepareAndLogError(errNoSecondFactorMinted, logger, span, "registering user")
	}

	user := registered.User
	accountID := registered.Membership.BelongsToAccount
	twoFactorSecret := registered.TOTPEnrollment.Secret

	tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)

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
