package sessions

import (
	platformerrors "github.com/primandproper/platform-go/v6/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
