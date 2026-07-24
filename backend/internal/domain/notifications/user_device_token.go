package notifications

import (
	"context"
	"encoding/gob"
	"regexp"
	"time"

	"github.com/primandproper/platform-go/v6/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// iOS APNs device tokens are 32 bytes, represented as 64 hex characters.
var iosDeviceTokenHexPattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

const (
	// UserDeviceTokenPlatformIOS represents the iOS platform.
	UserDeviceTokenPlatformIOS = "ios"
	// UserDeviceTokenPlatformAndroid represents the Android platform.
	UserDeviceTokenPlatformAndroid = "android"

	// UserDeviceTokenCreatedServiceEventType indicates a user device token was created.
	UserDeviceTokenCreatedServiceEventType = "user_device_token_created"
	// UserDeviceTokenUpdatedServiceEventType indicates a user device token was updated.
	UserDeviceTokenUpdatedServiceEventType = "user_device_token_updated"
	// UserDeviceTokenArchivedServiceEventType indicates a user device token was archived.
	UserDeviceTokenArchivedServiceEventType = "user_device_token_archived"
)

func init() {
	gob.Register(new(UserDeviceToken))
	gob.Register(new(UserDeviceTokenDatabaseCreationInput))
	gob.Register(new(UserDeviceTokenUpdateRequestInput))
}

type (
	// UserDeviceToken represents a push notification device token.
	UserDeviceToken struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time  `json:"createdAt"`
		LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time `json:"archivedAt"`
		ID            string     `json:"id"`
		DeviceToken   string     `json:"deviceToken"`
		Platform      string     `json:"platform"`
		BelongsToUser string     `json:"belongsToUser"`
	}

	// UserDeviceTokenDatabaseCreationInput represents what is needed to create a user device token.
	UserDeviceTokenDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID            string `json:"-"`
		DeviceToken   string `json:"-"`
		Platform      string `json:"-"`
		BelongsToUser string `json:"-"`
	}

	// UserDeviceTokenUpdateRequestInput represents what a user could set as input for updating a device token.
	UserDeviceTokenUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Platform *string `json:"platform,omitempty"`
	}

	// UserDeviceTokenDataManager describes a structure capable of storing user device tokens permanently.
	UserDeviceTokenDataManager interface {
		UserDeviceTokenExists(ctx context.Context, userID, tokenID string) (bool, error)
		GetUserDeviceToken(ctx context.Context, userID, tokenID string) (*UserDeviceToken, error)
		GetUserDeviceTokens(ctx context.Context, userID string, filter *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[UserDeviceToken], error)
		CreateUserDeviceToken(ctx context.Context, input *UserDeviceTokenDatabaseCreationInput) (*UserDeviceToken, error)
		UpdateUserDeviceToken(ctx context.Context, updated *UserDeviceToken) error
		ArchiveUserDeviceToken(ctx context.Context, userID, tokenID string) error
	}
)

// Update merges a UserDeviceTokenUpdateRequestInput with a user device token.
func (x *UserDeviceToken) Update(input *UserDeviceTokenUpdateRequestInput) {
	if input.Platform != nil && *input.Platform != x.Platform {
		x.Platform = *input.Platform
	}
}

var _ validation.ValidatableWithContext = (*UserDeviceTokenDatabaseCreationInput)(nil)

// ValidateWithContext validates a UserDeviceTokenDatabaseCreationInput.
func (x *UserDeviceTokenDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.DeviceToken, validation.Required, validation.By(validateDeviceTokenFormat(x.Platform))),
		validation.Field(&x.Platform, validation.Required, validation.In(UserDeviceTokenPlatformIOS, UserDeviceTokenPlatformAndroid)),
		validation.Field(&x.BelongsToUser, validation.Required),
	)
}

func validateDeviceTokenFormat(platform string) validation.RuleFunc {
	return func(value any) error {
		if s, ok := value.(string); ok {
			switch platform {
			case UserDeviceTokenPlatformIOS:
				if !iosDeviceTokenHexPattern.MatchString(s) {
					return validation.NewError("validation_device_token_format", "iOS device token must be 64 hexadecimal characters")
				}
			case UserDeviceTokenPlatformAndroid:
				// FCM tokens are typically 152+ chars; minimal sanity check
				if len(s) < 50 {
					return validation.NewError("validation_device_token_format", "Android device token appears too short")
				}
			}
		}
		return nil
	}
}

var _ validation.ValidatableWithContext = (*UserDeviceTokenUpdateRequestInput)(nil)

// ValidateWithContext validates a UserDeviceTokenUpdateRequestInput.
func (x *UserDeviceTokenUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Platform, validation.When(x.Platform != nil, validation.In(UserDeviceTokenPlatformIOS, UserDeviceTokenPlatformAndroid))),
	)
}
