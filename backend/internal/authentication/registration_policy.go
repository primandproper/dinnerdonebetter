package authentication

import (
	"context"
	"slices"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	grpcerrors "github.com/primandproper/primitives-go/v2/errors/grpc"
)

// ErrAgreementsRequired is a registration that did not accept both the terms of service and the
// privacy policy. Its words are the ones the registrant reads, so it is registered as
// client-safe below.
var ErrAgreementsRequired = platformerrors.New("the terms of service and the privacy policy must both be accepted to register")

// ErrPasswordRequired is a registration that chose no password. This application has no
// passwordless arrival yet: every door it serves but a passkey proves a password, and a
// registrant cannot hold a passkey before they can sign in to enroll one.
var ErrPasswordRequired = platformerrors.New("a password is required to register")

func init() {
	// This repository's own sentinels register from an init — see the note beside
	// platformerrormappers.Register in the API server's injector.
	grpcerrors.RegisterClientSafeSentinels(ErrAgreementsRequired, ErrPasswordRequired)
}

// RegistrationPolicy is this application's registration, in the shape signin.WithRegistrationPolicy
// takes: what it adds to somebody signing up, and what it refuses.
//
// It is what stands in front of SignInService.Register, which is open to anybody: signin runs it
// on every registration, before anything is hashed, minted or written. Without it the door would
// write a user this application does not recognize — unverified, holding no service role, with no
// second factor and no agreements. What it says:
//
//   - A registrant starts in good standing. platform's default is unverified, which is the
//     right default for a directory and not this application's policy: nothing here gates use
//     on a verified address, and a user who ignores the verification mail keeps cooking.
//   - They hold the service_user role, which is what makes them a user of this service.
//   - They are issued a second factor with the registration. It is unproven until they prove
//     it, and signin asks for it from then on.
//   - The account they own is named for them when they did not name it, and they hold
//     account_admin in it — the service's default owner role too, named here so the policy
//     says the whole of what a registrant is.
//   - They have accepted the terms of service and the privacy policy. A registration that has
//     not is refused before anything is written, with ErrAgreementsRequired.
//   - They chose a password. platform supports registering with none, which this application
//     does not yet: such a registrant could only get in by attaching a password from the
//     verification link, so they are refused with ErrPasswordRequired instead.
//
// The password is not here. It is PasswordPolicy's, which signin applies to every password it
// writes, this one included.
func RegistrationPolicy(_ context.Context, registration *signin.Registration) error {
	// The credential is a sealed interface; its passwordless arm is a value, compared as one. A
	// registration naming no credential at all is platform's to refuse, in its own words.
	if registration.Credential == signin.NoPassword() {
		return ErrPasswordRequired
	}

	if !slices.Contains(registration.Agreements, platformidentity.TermsOfService) ||
		!slices.Contains(registration.Agreements, platformidentity.PrivacyPolicy) {
		return ErrAgreementsRequired
	}

	if registration.User != nil {
		registration.User.AccountStatus = platformidentity.StatusGood
		registration.User.ServiceRoles = []string{authorization.ServiceUserRoleName}
	}

	registration.EnrollTOTP = true
	registration.OwnerRoles = []string{authorization.AccountAdminRoleName}

	if registration.Account == nil {
		registration.Account = &platformidentity.Account{}
	}

	registration.Account.Name = strings.TrimSpace(registration.Account.Name)
	if registration.Account.Name == "" && registration.User != nil {
		registration.Account.Name = registration.User.Username + "'s account"
	}

	return nil
}
