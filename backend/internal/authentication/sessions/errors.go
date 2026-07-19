package sessions

import (
	platformerrors "github.com/primandproper/platform-go/v5/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
