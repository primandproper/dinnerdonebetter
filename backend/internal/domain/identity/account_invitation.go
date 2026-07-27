package identity

import (
	"context"
	"time"

	"github.com/primandproper/platform-go/v7/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// AccountInvitationCreatedServiceEventType indicates an account invitation was created.
	AccountInvitationCreatedServiceEventType = "account_invitation_created"
	// AccountInvitationCanceledServiceEventType indicates an account invitation was created.
	AccountInvitationCanceledServiceEventType = "account_invitation_canceled"
	// AccountInvitationAcceptedServiceEventType indicates an account invitation was created.
	AccountInvitationAcceptedServiceEventType = "account_invitation_accepted"
	// AccountInvitationRejectedServiceEventType indicates an account invitation was created.
	AccountInvitationRejectedServiceEventType = "account_invitation_rejected"

	// PendingAccountInvitationStatus indicates an account invitation is pending.
	PendingAccountInvitationStatus AccountInvitationStatus = "pending"
	// CancelledAccountInvitationStatus indicates an account invitation was accepted.
	CancelledAccountInvitationStatus AccountInvitationStatus = "cancelled"
	// AcceptedAccountInvitationStatus indicates an account invitation was accepted.
	AcceptedAccountInvitationStatus AccountInvitationStatus = "accepted"
	// RejectedAccountInvitationStatus indicates an account invitation was rejected.
	RejectedAccountInvitationStatus AccountInvitationStatus = "rejected"

	// MobileNotificationRequestTypeHouseholdInvitationAccepted indicates a household invitation accepted notification.
	MobileNotificationRequestTypeHouseholdInvitationAccepted = "household_invitation_accepted"
	// ExcludedUserIDContextKey is the key used in MobileNotificationRequest.Context for the user to exclude.
	ExcludedUserIDContextKey = "excludedUserID"
)

type (
	// AccountInvitationStatus is the type to use/compare against when checking invitation status.
	AccountInvitationStatus string

	// AccountInvitationCreationRequestInput represents what a User could set as input for creating account invitations.
	AccountInvitationCreationRequestInput struct {
		_ struct{} `json:"-"`

		ExpiresAt *time.Time `json:"expiresAt"`
		Note      string     `json:"note"`
		ToEmail   string     `json:"toEmail"`
		ToName    string     `json:"toName"`
	}

	// AccountInvitationDatabaseCreationInput represents what a User could set as input for creating account invitations.
	AccountInvitationDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                   string    `json:"-"`
		FromUser             string    `json:"-"`
		ToUser               *string   `json:"-"`
		Note                 string    `json:"-"`
		ToEmail              string    `json:"-"`
		Token                string    `json:"-"`
		ToName               string    `json:"-"`
		ExpiresAt            time.Time `json:"-"`
		DestinationAccountID string    `json:"-"`
	}

	// AccountInvitation represents an account invitation.
	AccountInvitation struct {
		_                  struct{}   `json:"-"`
		ExpiresAt          time.Time  `json:"expiresAt"`
		CreatedAt          time.Time  `json:"createdAt"`
		LastUpdatedAt      *time.Time `json:"lastUpdatedAt"`
		ArchivedAt         *time.Time `json:"archivedAt"`
		ToUser             *string    `json:"toUser"`
		FromUser           User       `json:"fromUser"`
		ToEmail            string     `json:"toEmail"`
		Token              string     `json:"token"`
		ID                 string     `json:"id"`
		Note               string     `json:"note"`
		ToName             string     `json:"toName"`
		StatusNote         string     `json:"statusNote"`
		Status             string     `json:"status"`
		DestinationAccount Account    `json:"destinationAccount"`
	}

	// AccountInvitationUpdateRequestInput is used by users to update the status of a given account invitation.
	AccountInvitationUpdateRequestInput struct {
		Token string `json:"token"`
		Note  string `json:"note"`
	}

	// AccountInvitationDataManager describes a structure capable of storing account invitations permanently.
	AccountInvitationDataManager interface {
		AccountInvitationExists(ctx context.Context, accountInvitationID string) (bool, error)
		GetAccountInvitationByAccountAndID(ctx context.Context, accountID, accountInvitationID string) (*AccountInvitation, error)
		GetAccountInvitationByTokenAndID(ctx context.Context, token, invitationID string) (*AccountInvitation, error)
		GetAccountInvitationByEmailAndToken(ctx context.Context, emailAddress, token string) (*AccountInvitation, error)
		GetPendingAccountInvitationsFromUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AccountInvitation], error)
		GetPendingAccountInvitationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AccountInvitation], error)
		CreateAccountInvitation(ctx context.Context, input *AccountInvitationDatabaseCreationInput) (*AccountInvitation, error)
		CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error
		AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID, token, note string) error
		RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error
	}
)

var _ validation.ValidatableWithContext = (*AccountInvitationCreationRequestInput)(nil)

// ValidateWithContext validates a AccountCreationRequestInput.
func (x *AccountInvitationCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.ToName, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*AccountInvitationUpdateRequestInput)(nil)

// ValidateWithContext validates a AccountInvitationUpdateRequestInput.
func (x *AccountInvitationUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Token, validation.Required),
	)
}
