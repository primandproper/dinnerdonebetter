package auth

import (
	"github.com/primandproper/platform-go/v13/sessions"
	"github.com/primandproper/platform-go/v13/tenancy"
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
	// SessionPayload is what a session record carries beyond what sessions/database
	// stores for every session — the holder, the device metadata, and the two
	// timestamps expiry is measured from.
	//
	// Both fields are the JTI of a token this session was last issued alongside,
	// and they are here so that only the newest pair works. A session is not a
	// bearer credential in this application: the credential is the token, and the
	// session identifier merely rides along in its `sid` claim. Without the JTIs
	// an access token from before a refresh would still name a live session and
	// still be accepted, and rotating a refresh token would rotate nothing.
	SessionPayload struct {
		_ struct{} `json:"-"`

		// SessionTokenID is the JTI of the access token this session was last
		// issued alongside.
		SessionTokenID string `json:"sessionTokenID"`
		// RefreshTokenID is the JTI of the refresh token this session was last
		// issued alongside.
		RefreshTokenID string `json:"refreshTokenID"`
	}

	// UserSession is one live session as the store hands it back.
	//
	// It is an alias rather than a type of this repository's own because there is
	// nothing here to add: the store already carries the device metadata a security
	// page renders, the two anchors expiry is measured from, and the IsCurrent flag
	// that says which of the listed sessions is the reader's. A struct beside it
	// would be a second account of the same row.
	UserSession = sessions.Session[SessionPayload]

	// SessionStore is the session store this application holds.
	SessionStore = sessions.Store[SessionPayload]
)

// SessionHolder names whose sessions a call is about.
//
// The scope is global rather than the active account, because a session belongs to
// a person and not to one of the accounts they can act in: the active account is a
// claim on the token, a user switches it without signing in again, and a sign-out
// that only reached the sessions opened under one account would leave the others
// live. The principal is therefore the whole key, and the scope is there because
// sessions.Holder refuses to be half a key — see its documentation.
func SessionHolder(userID string) sessions.Holder {
	return sessions.Holder{Scope: tenancy.Global(), Principal: userID}
}
