package sessions

import (
	platformerrors "github.com/primandproper/platform-go/v10/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
