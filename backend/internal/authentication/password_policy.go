package authentication

import (
	"context"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

// MinimumPasswordEntropy is the floor, in bits, every password this application writes has to
// clear.
//
// One number for every door that writes a password: AuthService's registration, change and
// reset, and platform's SignInService through PasswordPolicy below. Two copies of it would be
// two doors that could come to disagree about what a weak password is, and a caller would
// find the weaker one.
const MinimumPasswordEntropy = 60

// PasswordPolicy is this application's password rule, in the shape signin.WithPasswordPolicy
// takes.
//
// It is what lets platform's server write a password at all: without it, a door mounted from
// platform would accept any non-empty password that AuthService refuses.
func PasswordPolicy(_ context.Context, password string) error {
	return passwordvalidator.Validate(password, MinimumPasswordEntropy)
}
