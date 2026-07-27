package managers

import (
	platformerrors "github.com/primandproper/platform-go/v7/errors"
)

var (
	// ErrInvalidCredentials indicates a credential validation failure (e.g. a password mismatch).
	// It must be an explicit non-nil error because observability.PrepareError returns nil when handed
	// a nil error, which would otherwise turn a failed credential check into a silent success.
	ErrInvalidCredentials = platformerrors.New("credentials are not valid")

	// ErrTOTPTokenRequired indicates the user has a verified two-factor secret but supplied no code.
	ErrTOTPTokenRequired = platformerrors.New("two factor code required")
)
