package sessions

import (
	platformerrors "github.com/primandproper/platform-go/v9/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
