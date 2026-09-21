package webauthn

import (
	"bytes"
	"strings"

	"github.com/primandproper/platform-go/v14/authentication/passkeys"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformwebauthn "github.com/primandproper/primitives-go/v2/authentication/webauthn"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnUser adapts a directory user and their credentials to the platformwebauthn.User interface.
type WebAuthnUser struct {
	User        *platformidentity.User
	Credentials []*passkeys.Credential
	UserID      []byte // WebAuthn user handle - use user ID bytes
	// Optional: when set, credentials matching AssertionCredID use AssertionFlags for BackupEligible/BackupState.
	// Used to satisfy go-webauthn's consistency check during login (avoids "Backup Eligible flag inconsistency").
	AssertionCredID []byte
	AssertionFlags  protocol.AuthenticatorFlags
}

// Ensure WebAuthnUser implements platformwebauthn.User.
var _ platformwebauthn.User = (*WebAuthnUser)(nil)

// WebAuthnID returns the user handle (opaque byte sequence, max 64 bytes).
func (u *WebAuthnUser) WebAuthnID() []byte {
	return u.UserID
}

// WebAuthnName returns the username for display during registration.
func (u *WebAuthnUser) WebAuthnName() string {
	return u.User.Username
}

// WebAuthnDisplayName returns the display name for registration.
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.User.FirstName != "" || u.User.LastName != "" {
		return strings.TrimSpace(u.User.FirstName + " " + u.User.LastName)
	}
	return u.User.Username
}

// WebAuthnCredentials returns credentials in platformwebauthn.Credential format.
func (u *WebAuthnUser) WebAuthnCredentials() []platformwebauthn.Credential {
	creds := make([]platformwebauthn.Credential, 0, len(u.Credentials))
	for _, c := range u.Credentials {
		creds = append(creds, domainCredentialToWebAuthnWithAssertionFlags(c, u.AssertionCredID, u.AssertionFlags))
	}
	return creds
}

// domainCredentialToWebAuthnWithAssertionFlags converts a domain credential to platformwebauthn.Credential.
// When assertionCredID and flags are provided and the credential matches, uses the assertion's
// BackupEligible/BackupState to satisfy go-webauthn's consistency check (avoids "Backup Eligible
// flag inconsistency" for passkeys that report different flags than our stored default, e.g. iCloud-synced).
func domainCredentialToWebAuthnWithAssertionFlags(c *passkeys.Credential, assertionCredID []byte, flags protocol.AuthenticatorFlags) platformwebauthn.Credential {
	transports := parseTransports(c.Transports)
	backupEligible, backupState := false, false
	if len(assertionCredID) > 0 && bytes.Equal(c.CredentialID, assertionCredID) {
		backupEligible = flags.HasBackupEligible()
		backupState = flags.HasBackupState()
	}
	return platformwebauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: "none",
		Transport:       transports,
		Flags: gowebauthn.CredentialFlags{
			UserPresent:    true,
			UserVerified:   true,
			BackupEligible: backupEligible,
			BackupState:    backupState,
		},
		Authenticator: gowebauthn.Authenticator{
			SignCount:    c.SignCount,
			CloneWarning: false,
		},
	}
}

// parseTransports renames a credential's transports into the protocol's type.
//
// It used to decode JSON, because the column held a JSON array in a text field. platform
// stores them as a list, so there is nothing to decode and nothing to fail at decoding —
// which is what the silent nil on a malformed value used to hide.
func parseTransports(transports []string) []protocol.AuthenticatorTransport {
	result := make([]protocol.AuthenticatorTransport, 0, len(transports))
	for _, t := range transports {
		result = append(result, protocol.AuthenticatorTransport(t))
	}

	return result
}
