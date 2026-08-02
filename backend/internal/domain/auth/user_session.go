package auth

import (
	"context"
	"time"

	"github.com/primandproper/platform-go/v9/filtering"
)

const (
	// UserSessionCreatedEventType indicates a user session was created.
	UserSessionCreatedEventType = "user_session_created"
	// UserSessionRevokedEventType indicates a user session was revoked.
	UserSessionRevokedEventType = "user_session_revoked"

	// LoginMethodPassword indicates the session was created via password login.
	LoginMethodPassword = "password"
	// LoginMethodPasskey indicates the session was created via passkey login.
	LoginMethodPasskey = "passkey"
)

type (
	// UserSession represents an active login session for a user.
	UserSession struct {
		_ struct{} `json:"-"`

		CreatedAt      time.Time  `json:"createdAt"`
		LastActiveAt   time.Time  `json:"lastActiveAt"`
		ExpiresAt      time.Time  `json:"expiresAt"`
		RevokedAt      *time.Time `json:"revokedAt"`
		ID             string     `json:"id"`
		BelongsToUser  string     `json:"belongsToUser"`
		SessionTokenID string     `json:"-"`
		RefreshTokenID string     `json:"-"`
		ClientIP       string     `json:"clientIP"`
		UserAgent      string     `json:"userAgent"`
		DeviceName     string     `json:"deviceName"`
		LoginMethod    string     `json:"loginMethod"`
		IsCurrent      bool       `json:"isCurrent"`
	}

	// UserSessionDatabaseCreationInput represents the input for creating a user session in the database.
	UserSessionDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ExpiresAt      time.Time `json:"-"`
		ID             string    `json:"-"`
		BelongsToUser  string    `json:"-"`
		SessionTokenID string    `json:"-"`
		RefreshTokenID string    `json:"-"`
		ClientIP       string    `json:"-"`
		UserAgent      string    `json:"-"`
		DeviceName     string    `json:"-"`
		LoginMethod    string    `json:"-"`
	}

	// UserSessionDataManager describes a structure capable of storing user sessions.
	UserSessionDataManager interface {
		CreateUserSession(ctx context.Context, input *UserSessionDatabaseCreationInput) (*UserSession, error)
		GetUserSessionBySessionTokenID(ctx context.Context, sessionTokenID string) (*UserSession, error)
		GetUserSessionByRefreshTokenID(ctx context.Context, refreshTokenID string) (*UserSession, error)
		GetActiveSessionsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[UserSession], error)
		RevokeUserSession(ctx context.Context, sessionID, userID string) error
		RevokeAllSessionsForUser(ctx context.Context, userID string) error
		RevokeAllSessionsForUserExcept(ctx context.Context, userID, sessionID string) error
		UpdateSessionTokenIDs(ctx context.Context, sessionID, newSessionTokenID, newRefreshTokenID string, newExpiresAt time.Time) error
		TouchSessionLastActive(ctx context.Context, sessionTokenID string) error
	}
)
