package authentication

import (
	platformauth "github.com/primandproper/platform-go/v5/authentication"
	"github.com/primandproper/platform-go/v5/authentication/argon2"
	"github.com/primandproper/platform-go/v5/authentication/totp"
	platformerrors "github.com/primandproper/platform-go/v5/errors"
)

// Re-exports of types that now live in github.com/primandproper/platform-go/v5/authentication.
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

	// NewArgon2Authenticator returns an argon2 powered Authenticator.
	NewArgon2Authenticator = argon2.NewArgon2Authenticator
)
