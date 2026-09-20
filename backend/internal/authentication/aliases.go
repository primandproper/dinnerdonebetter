package authentication

import (
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformauth "github.com/primandproper/primitives-go/v2/authentication"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// Re-exports of types that now live in github.com/primandproper/primitives-go/v2/authentication.
// These aliases let existing consumers keep using the `authentication` package while the
// interface and argon2 provider definitions live in the shared platform module.
type (
	// Authenticator hashes passwords and verifies them against a stored hash.
	Authenticator = platformauth.Authenticator
	// Hasher hashes passwords.
	Hasher = platformauth.Hasher
)

var (
	// ErrInvalidTOTPToken indicates that a provided two-factor code is invalid.
	// Alias for totp.ErrInvalidCode, retained so existing callers / error mappers keep working.
	ErrInvalidTOTPToken = totp.ErrInvalidCode
	// ErrTOTPRequired indicates that the user has TOTP enabled but did not provide a code.
	// Alias for totp.ErrCodeRequired, retained for the same reason as ErrInvalidTOTPToken.
	ErrTOTPRequired = totp.ErrCodeRequired
	// ErrPasswordDoesNotMatch is returned by login flows when a password does not match the stored hash.
	// Platform no longer exports a dedicated sentinel (platformauth.Authenticator.PasswordMatches returns
	// (false, nil) on mismatch); we keep a local sentinel so the HTTP/gRPC error mappers can continue to
	// convert password mismatches into 401 responses.
	ErrPasswordDoesNotMatch = platformerrors.New("password does not match")
	// ErrUserBanned is what a login gets when the user's account status does not admit
	// signing in. It is platform's sentinel rather than one of this package's: the check
	// is no longer made here — Store.GetPrincipal refuses before it reads a membership —
	// so a local error would be one nothing ever returns, and errors.Is against it would
	// quietly stop matching.
	ErrUserBanned = platformidentity.ErrSignInNotAdmitted

	// ErrUserNotAdmin is what the administrator-only login door gets when the user is
	// perfectly real and holds no administrator role.
	//
	// It is distinct from "no such user" on purpose. The read it replaced filtered
	// non-administrators out inside the query, so the two came back as one answer and
	// an operator reading a log could not tell a typo from a missing grant.
	ErrUserNotAdmin = platformerrors.New("user is not a service administrator")
	// ErrSessionSuperseded is returned when a refresh token names a live session that has
	// since been issued a newer pair of tokens. The session is fine; this particular
	// refresh token has been spent, and replaying it must not mint a third pair.
	ErrSessionSuperseded = platformerrors.New("session has been issued newer tokens")

	// NewArgon2Authenticator returns an argon2 powered Authenticator.
	NewArgon2Authenticator = argon2.NewArgon2Authenticator
)
