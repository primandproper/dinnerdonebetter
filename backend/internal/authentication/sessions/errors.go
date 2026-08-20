package sessions

import (
	platformerrors "github.com/primandproper/platform-go/v12/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
