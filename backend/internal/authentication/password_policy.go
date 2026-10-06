package authentication

import (
	"context"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	grpcerrors "github.com/primandproper/primitives-go/v2/errors/grpc"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

// MinimumPasswordEntropy is the floor, in bits, every password this application writes has to
// clear.
//
// One number for every door that writes a password — registration, change and reset — through
// PasswordPolicy below. Two copies of it would be two doors that could come to disagree about
// what a weak password is, and a caller would find the weaker one.
const MinimumPasswordEntropy = 60

// PasswordPolicy is this application's password rule, in the shape signin.WithPasswordPolicy
// takes.
//
// It is what lets platform's servers write a password at all: without it, a door mounted from
// platform would accept any non-empty password.
func PasswordPolicy(_ context.Context, password string) error {
	return passwordvalidator.Validate(password, MinimumPasswordEntropy)
}

// ErrPasswordUnchanged is a password change to the password the account already holds. Its words
// are the ones the person changing it reads, so it is registered as client-safe below; nobody
// reaches it without having just proven the current password.
var ErrPasswordUnchanged = platformerrors.New("the new password is the current password")

func init() {
	grpcerrors.RegisterClientSafeSentinels(ErrPasswordUnchanged)
}

// AccountPasswordPolicy is this application's rule for a password written to an account that
// already exists, in the shape signin.WithAccountPasswordPolicy takes: it may not be the password
// the account holds now.
//
// It is asked only after the caller has proven the current password, which is why it may ask
// whether the two match at all: before that, the answer would be an oracle on the current
// password for whoever held the session. The comparison is the sign-in service's own, through
// MatchesCurrent, so the hash never crosses this seam.
func AccountPasswordPolicy(ctx context.Context, change *signin.PasswordChange) error {
	if change == nil || change.MatchesCurrent == nil {
		return nil
	}

	same, err := change.MatchesCurrent(ctx, change.NewPassword)
	if err != nil {
		return platformerrors.Wrap(err, "comparing a new password with the current one")
	}

	if same {
		return ErrPasswordUnchanged
	}

	return nil
}
