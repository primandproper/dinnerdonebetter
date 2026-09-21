package sessions

import (
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

var (
	ErrAuthenticationNotFound = platformerrors.New("authentication not found")
)
