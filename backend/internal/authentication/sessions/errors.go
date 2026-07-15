package sessions

import (
	platformerrors "github.com/primandproper/platform-go/v4/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
