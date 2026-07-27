package identity

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v7/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pquerna/otp/totp"
)

const (
	// UserSignedUpServiceEventType indicates a user signed up.
	UserSignedUpServiceEventType = "user_signed_up"
	// UserArchivedServiceEventType indicates a user archived their account.
	UserArchivedServiceEventType = "user_archived"
	// UserStatusChangedServiceEventType indicates a user had their user status changed.
	UserStatusChangedServiceEventType = "user_status_changed"
	// UserPasswordChangeRequiredServiceEventType indicates a user was marked as requiring a password change.
	UserPasswordChangeRequiredServiceEventType = "user_password_change_required"

	// GoodStandingUserAccountStatus indicates a User's account is in good standing.
	GoodStandingUserAccountStatus userAccountStatus = "good"
	// UnverifiedAccountStatus indicates a User's account requires two factor secret verification.
	UnverifiedAccountStatus userAccountStatus = "unverified"
	// BannedUserAccountStatus indicates a User's account is banned.
	BannedUserAccountStatus userAccountStatus = "banned"
	// TerminatedUserAccountStatus indicates a User's account is banned.
	TerminatedUserAccountStatus userAccountStatus = "terminated"

	validTOTPTokenLength = 6
)

var (
	totpTokenLengthRule = validation.Length(validTOTPTokenLength, validTOTPTokenLength)
)

type (
	userAccountStatus string

	// User represents a User.
	User struct {
		_                          struct{}                     `json:"-"`
		CreatedAt                  time.Time                    `json:"createdAt"`
		EmailAddressVerifiedAt     *time.Time                   `json:"emailAddressVerifiedAt"`
		PasswordLastChangedAt      *time.Time                   `json:"passwordLastChangedAt"`
		LastUpdatedAt              *time.Time                   `json:"lastUpdatedAt"`
		LastAcceptedTermsOfService *time.Time                   `json:"lastAcceptedTOS"`
		LastAcceptedPrivacyPolicy  *time.Time                   `json:"lastAcceptedPrivacyPolicy"`
		TwoFactorSecretVerifiedAt  *time.Time                   `json:"twoFactorSecretVerifiedAt"`
		Avatar                     *uploadedmedia.UploadedMedia `json:"avatar"`
		Birthday                   *time.Time                   `json:"birthday"`
		ArchivedAt                 *time.Time                   `json:"archivedAt"`
		AccountStatusExplanation   string                       `json:"accountStatusExplanation"`
		HashedPassword             string                       `json:"-"`
		ID                         string                       `json:"id"`
		AccountStatus              string                       `json:"accountStatus"`
		Username                   string                       `json:"username"`
		FirstName                  string                       `json:"firstName"`
		LastName                   string                       `json:"lastName"`
		EmailAddress               string                       `json:"emailAddress"`
		TwoFactorSecret            string                       `json:"-"`
		RequiresPasswordChange     bool                         `json:"requiresPasswordChange"`
	}

	// UserRegistrationInput represents the input required from users to register an account.
	UserRegistrationInput struct {
		_ struct{} `json:"-"`

		Birthday              *time.Time `json:"birthday,omitempty"`
		Password              string     `json:"password"`
		EmailAddress          string     `json:"emailAddress"`
		InvitationToken       string     `json:"invitationToken,omitempty"`
		InvitationID          string     `json:"invitationID,omitempty"`
		Username              string     `json:"username"`
		FirstName             string     `json:"firstName"`
		LastName              string     `json:"lastName"`
		AccountName           string     `json:"accountName"`
		AcceptedTOS           bool       `json:"acceptedTOS"`
		AcceptedPrivacyPolicy bool       `json:"acceptedPrivacyPolicy"`
	}

	// UserDatabaseCreationInput is used by the User creation route to communicate with the data store.
	UserDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		Birthday              *time.Time `json:"-"`
		ID                    string     `json:"-"`
		HashedPassword        string     `json:"-"`
		TwoFactorSecret       string     `json:"-"`
		InvitationToken       string     `json:"-"`
		DestinationAccountID  string     `json:"-"`
		Username              string     `json:"-"`
		EmailAddress          string     `json:"-"`
		AccountName           string     `json:"-"`
		FirstName             string     `json:"-"`
		LastName              string     `json:"-"`
		AcceptedTOS           bool       `json:"-"`
		AcceptedPrivacyPolicy bool       `json:"-"`
	}

	// UserCreationResponse is a response structure for Users that doesn't contain passwords fields, but does contain the two factor secret.
	UserCreationResponse struct {
		_ struct{} `json:"-"`

		CreatedAt       time.Time  `json:"createdAt"`
		Birthday        *time.Time `json:"birthday"`
		Username        string     `json:"username"`
		EmailAddress    string     `json:"emailAddress"`
		TwoFactorQRCode string     `json:"qrCode"`
		CreatedUserID   string     `json:"createdUserID"`
		AccountStatus   string     `json:"accountStatus"`
		TwoFactorSecret string     `json:"twoFactorSecret"`
		FirstName       string     `json:"firstName"`
		LastName        string     `json:"lastName"`
	}

	// UserDetailsUpdateRequestInput represents input a User would provide when updating their information.
	UserDetailsUpdateRequestInput struct {
		_ struct{} `json:"-"`

		FirstName       string    `json:"firstName"`
		LastName        string    `json:"lastName"`
		Birthday        time.Time `json:"birthday"`
		CurrentPassword string    `json:"currentPassword"`
		TOTPToken       string    `json:"totpToken"`
	}

	// UserDetailsDatabaseUpdateInput represents input a User would provide when updating their information.
	UserDetailsDatabaseUpdateInput struct {
		_ struct{} `json:"-"`

		Birthday  time.Time `json:"-"`
		FirstName string    `json:"-"`
		LastName  string    `json:"-"`
	}

	// AvatarUpdateInput represents input a User would provide when updating their passwords.
	AvatarUpdateInput struct {
		_ struct{} `json:"-"`

		Base64EncodedData string `json:"base64EncodedData"`
	}

	// UserAccountStatusUpdateInput represents what an admin User could provide as input for changing statuses.
	UserAccountStatusUpdateInput struct {
		_ struct{} `json:"-"`

		NewStatus    string `json:"newStatus"`
		Reason       string `json:"reason"`
		TargetUserID string `json:"targetUserID"`
	}

	// UserDataManager describes a structure which can manage users in persistent storage.
	UserDataManager interface {
		UpdateUserAccountStatus(ctx context.Context, userID string, input *UserAccountStatusUpdateInput) error
		SetUserRequiresPasswordChange(ctx context.Context, userID string, requiresChange bool) error
		UserRequiresPasswordChange(ctx context.Context, userID string) (bool, error)
		GetUser(ctx context.Context, userID string) (*User, error)
		GetUserByUsername(ctx context.Context, username string) (*User, error)
		GetAdminUserByUsername(ctx context.Context, username string) (*User, error)
		GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[User], error)
		GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[User], error)
		GetUsersWithIDs(ctx context.Context, ids []string) ([]*User, error)
		GetUserByEmail(ctx context.Context, email string) (*User, error)
		SearchForUsersByUsername(ctx context.Context, usernameQuery string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[User], error)
		CreateUser(ctx context.Context, input *UserDatabaseCreationInput) (*User, error)
		SetUserAvatar(ctx context.Context, userID, uploadedMediaID string) error
		UpdateUserUsername(ctx context.Context, userID, newUsername string) error
		UpdateUserEmailAddress(ctx context.Context, userID, newEmailAddress string) error
		UpdateUserDetails(ctx context.Context, userID string, input *UserDetailsDatabaseUpdateInput) error
		UpdateUserPassword(ctx context.Context, userID, newHash string) error
		ArchiveUser(ctx context.Context, userID string) error
		DeleteUser(ctx context.Context, userID string) error
		GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID string) (*User, error)
		MarkUserTwoFactorSecretAsVerified(ctx context.Context, userID string) error
		MarkUserTwoFactorSecretAsUnverified(ctx context.Context, userID, newSecret string) error
		GetEmailAddressVerificationTokenForUser(ctx context.Context, userID string) (string, error)
		GetUserByEmailAddressVerificationToken(ctx context.Context, token string) (*User, error)
		MarkUserEmailAddressAsVerified(ctx context.Context, userID, token string) error
		MarkUserEmailAddressAsUnverified(ctx context.Context, userID string) error
		GetUserIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error)
		MarkUserAsIndexed(ctx context.Context, userID string) error
	}
)

func (x userAccountStatus) String() string {
	return string(x)
}

// Update accepts a User as input and merges those values if they're set.
func (u *User) Update(input *User) {
	if input.Username != "" && input.Username != u.Username {
		u.Username = input.Username
	}

	if input.FirstName != "" && input.FirstName != u.FirstName {
		u.FirstName = input.FirstName
	}

	if input.LastName != "" && input.LastName != u.LastName {
		u.LastName = input.LastName
	}

	if input.HashedPassword != "" && input.HashedPassword != u.HashedPassword {
		u.HashedPassword = input.HashedPassword
	}

	if input.TwoFactorSecret != "" && input.TwoFactorSecret != u.TwoFactorSecret {
		u.TwoFactorSecret = input.TwoFactorSecret
	}

	if input.Birthday != nil && input.Birthday != u.Birthday {
		u.Birthday = input.Birthday
	}
}

// IsBanned is a handy helper function.
func (u *User) IsBanned() bool {
	return u.AccountStatus == string(BannedUserAccountStatus)
}

// ValidateWithContext ensures our provided UserRegistrationInput meets expectations.
func (i *UserRegistrationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.EmailAddress, validation.Required, is.EmailFormat),
		validation.Field(&i.Username, validation.Required),
		validation.Field(&i.Password, validation.Required, validation.Length(8, math.MaxInt8)),
	)
}

var _ validation.ValidatableWithContext = (*UserDetailsUpdateRequestInput)(nil)

// ValidateWithContext ensures our provided UserDetailsUpdateRequestInput meets expectations.
func (i *UserDetailsUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.FirstName, validation.Required),
		validation.Field(&i.CurrentPassword, validation.Required),
		validation.Field(&i.TOTPToken, validation.When(i.TOTPToken != "", totpTokenLengthRule)),
	)
}

var _ validation.ValidatableWithContext = (*AvatarUpdateInput)(nil)

// ValidateWithContext ensures our provided AvatarUpdateInput meets expectations.
func (i *AvatarUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Base64EncodedData, validation.Required),
	)
}

func (u *User) GenerateTOTPCode() (string, error) {
	code, err := totp.GenerateCode(strings.ToUpper(u.TwoFactorSecret), time.Now().UTC())
	if err != nil {
		return "", fmt.Errorf("generating totp code: %w", err)
	}

	return code, nil
}

// begin obligatory getter implementations

// GetID fetches the User's ID value.
func (u *User) GetID() string {
	return u.ID
}

// GetEmail fetches the User's EmailAddress value.
func (u *User) GetEmail() string {
	return u.EmailAddress
}

// GetUsername fetches the User's Username value.
func (u *User) GetUsername() string {
	return u.Username
}

// GetFirstName fetches the User's FirstName value.
func (u *User) GetFirstName() string {
	return u.FirstName
}

// GetLastName fetches the User's LastName value.
func (u *User) GetLastName() string {
	return u.LastName
}

// FullName tries to construct the user's full name.
func (u *User) FullName() string {
	out := ""
	if u.FirstName != "" {
		out = u.FirstName
	}

	if u.LastName != "" {
		out += fmt.Sprintf(" %s", u.LastName)
	}

	return out
}
