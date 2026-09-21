package auth

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type (
	// UserRegistrationInput is what somebody signing up supplies.
	//
	// It lives on the auth surface rather than beside the directory's types, and that is
	// where registration moved to rather than an accident of packaging. platform's
	// identity service registers somebody on a registrar's behalf and holds no policy
	// about who may sign up; the password, the agreements, and the invitation a
	// registrant may be answering are all this application's, and they belong with the
	// rest of what a caller with no session can reach.
	UserRegistrationInput struct {
		_ struct{} `json:"-"`

		Password              string `json:"password"`
		EmailAddress          string `json:"emailAddress"`
		InvitationToken       string `json:"invitationToken,omitempty"`
		InvitationID          string `json:"invitationID,omitempty"`
		Username              string `json:"username"`
		FirstName             string `json:"firstName"`
		LastName              string `json:"lastName"`
		AccountName           string `json:"accountName"`
		AcceptedTOS           bool   `json:"acceptedTOS"`
		AcceptedPrivacyPolicy bool   `json:"acceptedPrivacyPolicy"`
	}

	// UserCreationResponse is what a registration answers with.
	//
	// It carries the two-factor secret, which is the one time it is ever handed back: the
	// directory stores it and no read returns it, so a registrant who closes this response
	// without scanning the QR code has to refresh the secret rather than ask for it again.
	UserCreationResponse struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time `json:"createdAt"`
		Username         string    `json:"username"`
		EmailAddress     string    `json:"emailAddress"`
		TwoFactorQRCode  string    `json:"qrCode"`
		CreatedUserID    string    `json:"createdUserID"`
		CreatedAccountID string    `json:"createdAccountID"`
		AccountStatus    string    `json:"accountStatus"`
		TwoFactorSecret  string    `json:"twoFactorSecret"`
		FirstName        string    `json:"firstName"`
		LastName         string    `json:"lastName"`
	}
)

var _ validation.ValidatableWithContext = (*UserRegistrationInput)(nil)

// ValidateWithContext checks a registration before anything is written.
//
// The agreements are required rather than merely recorded. A registration that did not
// accept them is one this application may not act on, and refusing it here is the only
// place that reading is enforced — the directory stores when somebody last agreed and has
// no opinion about whether they had to.
func (i *UserRegistrationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Username, validation.Required),
		validation.Field(&i.EmailAddress, validation.Required, is.EmailFormat),
		validation.Field(&i.Password, validation.Required),
		validation.Field(&i.AcceptedTOS, validation.Required),
		validation.Field(&i.AcceptedPrivacyPolicy, validation.Required),
	)
}
