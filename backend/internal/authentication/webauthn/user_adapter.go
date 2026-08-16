package webauthn

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformwebauthn "github.com/primandproper/platform-go/v10/authentication/webauthn"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnUser adapts identity.User and credentials to the platformwebauthn.User interface.
type WebAuthnUser struct {
	User        *identity.User
	Credentials []*identity.WebAuthnCredential
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
func domainCredentialToWebAuthnWithAssertionFlags(c *identity.WebAuthnCredential, assertionCredID []byte, flags protocol.AuthenticatorFlags) platformwebauthn.Credential {
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

func parseTransports(s string) []protocol.AuthenticatorTransport {
	if s == "" {
		return nil
	}
	var transports []string
	if err := json.Unmarshal([]byte(s), &transports); err != nil {
		return nil
	}
	result := make([]protocol.AuthenticatorTransport, 0, len(transports))
	for _, t := range transports {
		result = append(result, protocol.AuthenticatorTransport(t))
	}
	return result
}
