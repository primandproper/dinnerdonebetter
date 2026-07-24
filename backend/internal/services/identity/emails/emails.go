package emails

import (
	"errors"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/branding"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v6/email"

	"github.com/matcornic/hermes/v2"
)

var (
	ErrUnverifiedEmailRecipient = errors.New("missing email address verification for user")
)

// BuildInviteMemberEmail builds an email notifying a user that they've been invited to join an account.
func BuildInviteMemberEmail(recipient *identity.User, accountInvitation *identity.AccountInvitation, baseURL string) (*email.OutboundEmailMessage, error) {
	e := hermes.Email{
		Body: hermes.Body{
			Name: accountInvitation.ToEmail,
			Intros: []string{
				fmt.Sprintf("You've been invited to join an account on %s!", branding.CompanyName),
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Text: "Join the fun",
						Link: fmt.Sprintf("%s/accept_invitation?i=%s&t=%s", baseURL, accountInvitation.ID, accountInvitation.Token),
					},
				},
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		UserID:      recipient.ID,
		ToAddress:   accountInvitation.ToEmail,
		ToName:      recipient.FullName(),
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     "You've been invited!",
		HTMLContent: htmlContent,
	}

	return msg, nil
}

// BuildGeneratedPasswordResetTokenEmail builds an email notifying a user that they've been invited to join an account.
func BuildGeneratedPasswordResetTokenEmail(recipient *identity.User, passwordResetToken *auth.PasswordResetToken, baseURL string) (*email.OutboundEmailMessage, error) {
	if recipient.EmailAddressVerifiedAt == nil {
		return nil, ErrUnverifiedEmailRecipient
	}

	e := hermes.Email{
		Body: hermes.Body{
			Name: recipient.Username,
			Intros: []string{
				"You have received this email because a password reset was requested.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Text: "Reset your password",
						Link: fmt.Sprintf("%s/reset_password?t=%s", baseURL, passwordResetToken.Token),
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		UserID:      recipient.ID,
		ToAddress:   recipient.EmailAddress,
		ToName:      recipient.FullName(),
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("A password reset link was requested for your %s account", branding.CompanyName),
		HTMLContent: htmlContent,
	}

	return msg, nil
}

// BuildUsernameReminderEmail builds an email notifying a user that they've been invited to join an account.
func BuildUsernameReminderEmail(recipient *identity.User, baseURL string) (*email.OutboundEmailMessage, error) {
	if recipient.EmailAddressVerifiedAt == nil {
		return nil, ErrUnverifiedEmailRecipient
	}

	e := hermes.Email{
		Body: hermes.Body{
			Name: recipient.Username,
			Intros: []string{
				fmt.Sprintf("A username reminder for your %s account was requested. Your username is <b>%s</b>.", branding.CompanyName, recipient.Username),
			},
			Outros: []string{
				"If you did not request a username reminder, no further action is required on your part.",
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		UserID:      recipient.ID,
		ToName:      recipient.FullName(),
		ToAddress:   recipient.EmailAddress,
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("A password reset link was requested for your %s account", branding.CompanyName),
		HTMLContent: htmlContent,
	}

	return msg, nil
}

// BuildPasswordResetTokenRedeemedEmail builds an email notifying a user that they've been invited to join an account.
func BuildPasswordResetTokenRedeemedEmail(recipient *identity.User, baseURL string) (*email.OutboundEmailMessage, error) {
	if recipient.EmailAddressVerifiedAt == nil {
		return nil, ErrUnverifiedEmailRecipient
	}

	e := hermes.Email{
		Body: hermes.Body{
			Name: recipient.Username,
			Intros: []string{
				"This is to inform you that your password has been changed upon successful redemption of a reset token.",
			},
			Outros: []string{
				"If you did not request a password reset, please contact support.",
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		UserID:      recipient.ID,
		ToAddress:   recipient.EmailAddress,
		ToName:      recipient.FullName(),
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("Your %s account password has been changed.", branding.CompanyName),
		HTMLContent: htmlContent,
	}

	return msg, nil
}

// BuildPasswordChangedEmail builds an email notifying a user that they've been invited to join an account.
func BuildPasswordChangedEmail(recipient *identity.User, baseURL string) (*email.OutboundEmailMessage, error) {
	if recipient.EmailAddressVerifiedAt == nil {
		return nil, ErrUnverifiedEmailRecipient
	}

	e := hermes.Email{
		Body: hermes.Body{
			Name: recipient.Username,
			Intros: []string{
				"This is to inform you that your password has been changed.",
			},
			Outros: []string{
				"If you did not perform this action, please contact support.",
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		UserID:      recipient.ID,
		ToAddress:   recipient.EmailAddress,
		ToName:      recipient.FullName(),
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("Your %s account password has been changed.", branding.CompanyName),
		HTMLContent: htmlContent,
	}

	return msg, nil
}

var errEmailVerificationTokenRequired = errors.New("email verification token required")

// BuildVerifyEmailAddressEmail builds an email notifying a user that they've been invited to join an account.
func BuildVerifyEmailAddressEmail(recipient *identity.User, emailVerificationToken, baseURL string) (*email.OutboundEmailMessage, error) {
	if emailVerificationToken == "" {
		return nil, errEmailVerificationTokenRequired
	}

	if recipient.EmailAddressVerifiedAt != nil {
		return nil, fmt.Errorf("user %s has already verified their email address", recipient.ID)
	}

	e := hermes.Email{
		Body: hermes.Body{
			Name: recipient.Username,
			Intros: []string{
				fmt.Sprintf("You recently signed up for an account on %s. Please verify your email address.", branding.CompanyName),
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to verify your email address:",
					Button: hermes.Button{
						Text: "Verify my email",
						Link: fmt.Sprintf("%s/verify_email_address?t=%s", baseURL, emailVerificationToken),
					},
				},
			},
			Outros: []string{
				"If you did not sign up for an account, please contact support.",
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	msg := &email.OutboundEmailMessage{
		UserID:      recipient.ID,
		ToAddress:   recipient.EmailAddress,
		ToName:      recipient.FullName(),
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("Verify your email with %s", branding.CompanyName),
		HTMLContent: htmlContent,
	}

	return msg, nil
}
