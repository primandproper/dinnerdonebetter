package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/primandproper/platform-go/v5/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// AccountCreatedServiceEventType indicates an account was created.
	AccountCreatedServiceEventType = "account_created"
	// AccountUpdatedServiceEventType indicates an account was updated.
	AccountUpdatedServiceEventType = "account_updated"
	// AccountArchivedServiceEventType indicates an account was archived.
	AccountArchivedServiceEventType = "account_archived"
	// AccountMemberRemovedServiceEventType indicates an account member was removed.
	AccountMemberRemovedServiceEventType = "account_member_removed"
	// AccountMembershipPermissionsUpdatedServiceEventType indicates an account member's permissions were modified.
	AccountMembershipPermissionsUpdatedServiceEventType = "account_membership_permissions_updated"
	// AccountSetAsDefaultServiceEventType indicates an account was selected as a user's default.
	AccountSetAsDefaultServiceEventType = "account_set_as_default"
	// AccountOwnershipTransferredServiceEventType indicates an account was transferred to another owner.
	AccountOwnershipTransferredServiceEventType = "account_ownership_transferred"

	// UnpaidAccountBillingStatus indicates an account is not paid.
	UnpaidAccountBillingStatus = "unpaid"
	// PaidAccountBillingStatus indicates an account is paid.
	PaidAccountBillingStatus = "paid"
	// TrialAccountBillingStatus indicates an account is on a trial.
	TrialAccountBillingStatus = "trial"
	// SuspendedAccountBillingStatus indicates an account is suspended.
	SuspendedAccountBillingStatus = "suspended"
)

type (
	// Account represents an account.
	Account struct {
		_ struct{} `json:"-"`

		CreatedAt                  time.Time                        `json:"createdAt"`
		SubscriptionPlanID         *string                          `json:"subscriptionPlanID"`
		LastUpdatedAt              *time.Time                       `json:"lastUpdatedAt"`
		ArchivedAt                 *time.Time                       `json:"archivedAt"`
		Longitude                  *float64                         `json:"longitude"`
		Latitude                   *float64                         `json:"latitude"`
		State                      string                           `json:"state"`
		ContactPhone               string                           `json:"contactPhone"`
		City                       string                           `json:"city"`
		AddressLine1               string                           `json:"addressLine1"`
		ZipCode                    string                           `json:"zipCode"`
		Country                    string                           `json:"country"`
		BillingStatus              string                           `json:"billingStatus"`
		AddressLine2               string                           `json:"addressLine2"`
		PaymentProcessorCustomerID string                           `json:"paymentProcessorCustomer"`
		BelongsToUser              string                           `json:"belongsToUser"`
		ID                         string                           `json:"id"`
		Name                       string                           `json:"name"`
		WebhookEncryptionKey       string                           `json:"-"`
		Members                    []*AccountUserMembershipWithUser `json:"members"`
	}

	// AccountCreationRequestInput represents what a User could set as input for creating accounts.
	AccountCreationRequestInput struct {
		_ struct{} `json:"-"`

		Latitude      *float64 `json:"latitude"`
		Longitude     *float64 `json:"longitude"`
		Name          string   `json:"name"`
		ContactPhone  string   `json:"contactPhone"`
		AddressLine1  string   `json:"addressLine1"`
		AddressLine2  string   `json:"addressLine2"`
		City          string   `json:"city"`
		State         string   `json:"state"`
		ZipCode       string   `json:"zipCode"`
		Country       string   `json:"country"`
		BelongsToUser string   `json:"belongsToUser"`
	}

	// AccountDatabaseCreationInput represents what a User could set as input for creating accounts.
	AccountDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                   string   `json:"-"`
		Name                 string   `json:"-"`
		AddressLine1         string   `json:"-"`
		AddressLine2         string   `json:"-"`
		City                 string   `json:"-"`
		State                string   `json:"-"`
		ZipCode              string   `json:"-"`
		Country              string   `json:"-"`
		Latitude             *float64 `json:"-"`
		Longitude            *float64 `json:"-"`
		ContactPhone         string   `json:"-"`
		BelongsToUser        string   `json:"-"`
		WebhookEncryptionKey string   `json:"-"`
	}

	// AccountUpdateRequestInput represents what a User could set as input for updating accounts.
	AccountUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name          *string  `json:"name,omitempty"`
		ContactPhone  *string  `json:"contactPhone,omitempty"`
		AddressLine1  *string  `json:"addressLine1"`
		AddressLine2  *string  `json:"addressLine2"`
		City          *string  `json:"city"`
		State         *string  `json:"state"`
		ZipCode       *string  `json:"zipCode"`
		Country       *string  `json:"country"`
		Latitude      *float64 `json:"latitude"`
		Longitude     *float64 `json:"longitude"`
		BelongsToUser string   `json:"-"`
	}

	// AccountDataManager describes a structure capable of storing accounts permanently.
	AccountDataManager interface {
		GetAccount(ctx context.Context, accountID string) (*Account, error)
		GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Account], error)
		CreateAccount(ctx context.Context, input *AccountDatabaseCreationInput) (*Account, error)
		UpdateAccount(ctx context.Context, updated *Account) error
		UpdateAccountBillingFields(ctx context.Context, accountID string, billingStatus, subscriptionPlanID, paymentProcessorCustomerID *string, lastPaymentProviderSyncOccurredAt *time.Time) error
		ArchiveAccount(ctx context.Context, accountID string, userID string) error
	}
)

// Update merges a AccountUpdateRequestInput with an account.
func (x *Account) Update(input *AccountUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.ContactPhone != nil && *input.ContactPhone != x.ContactPhone {
		x.ContactPhone = *input.ContactPhone
	}

	if input.AddressLine1 != nil && *input.AddressLine1 != x.AddressLine1 {
		x.AddressLine1 = *input.AddressLine1
	}

	if input.AddressLine2 != nil && *input.AddressLine2 != x.AddressLine2 {
		x.AddressLine2 = *input.AddressLine2
	}

	if input.City != nil && *input.City != x.City {
		x.City = *input.City
	}

	if input.State != nil && *input.State != x.State {
		x.State = *input.State
	}

	if input.ZipCode != nil && *input.ZipCode != x.ZipCode {
		x.ZipCode = *input.ZipCode
	}

	if input.Country != nil && *input.Country != x.Country {
		x.Country = *input.Country
	}

	if input.Latitude != nil && input.Latitude != x.Latitude {
		x.Latitude = input.Latitude
	}

	if input.Longitude != nil && input.Longitude != x.Longitude {
		x.Longitude = input.Longitude
	}
}

var _ validation.ValidatableWithContext = (*AccountCreationRequestInput)(nil)

// ValidateWithContext validates a AccountCreationRequestInput.
func (x *AccountCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Latitude, validation.NilOrNotEmpty),
		validation.Field(&x.Longitude, validation.NilOrNotEmpty),
	)
}

var _ validation.ValidatableWithContext = (*AccountUpdateRequestInput)(nil)

// ValidateWithContext validates a AccountUpdateRequestInput.
func (x *AccountUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Latitude, validation.NilOrNotEmpty),
		validation.Field(&x.Longitude, validation.NilOrNotEmpty),
	)
}

// AccountCreationInputForNewUser creates a new AccountInputCreation struct for a given user.
func AccountCreationInputForNewUser(u *User) *AccountDatabaseCreationInput {
	return &AccountDatabaseCreationInput{
		Name:          fmt.Sprintf("%s's cool account", u.Username),
		BelongsToUser: u.ID,
	}
}
