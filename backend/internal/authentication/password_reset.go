package authentication

import (
	"context"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// PasswordResetDirectory is the directory platform's password reset flow writes a new password
// through: the identity store, plus the event that tells the person their password was reset.
//
// passwordreset.Service writes the password on the transaction that spends the link and
// revokes the others, and identity's service would open a transaction of its own, so its
// AfterUpdateUserPassword hook does not fire on this path. The event is written here instead,
// on that same transaction — the reset, the revocation and the mail announcing them commit
// together or not at all.
type PasswordResetDirectory struct {
	platformidentity.Store

	emitter *events.Emitter
	logger  logging.Logger
}

var _ passwordreset.Directory = (*PasswordResetDirectory)(nil)

// NewPasswordResetDirectory builds the directory platform's password reset flow writes through.
func NewPasswordResetDirectory(logger logging.Logger, store platformidentity.Store, emitter *events.Emitter) *PasswordResetDirectory {
	return &PasswordResetDirectory{
		Store:   store,
		emitter: emitter,
		logger:  logging.NewNamedLogger(logger, "password_reset_directory"),
	}
}

// UpdateUserPassword writes the new password and queues the mail saying it changed.
func (d *PasswordResetDirectory) UpdateUserPassword(ctx context.Context, tx database.Tx, scope tenancy.Scope, userID, hashedPassword string) error {
	if err := d.Store.UpdateUserPassword(ctx, tx, scope, userID, hashedPassword); err != nil {
		return err
	}

	logger := d.logger.WithValue(identitykeys.UserIDKey, userID)

	if err := d.emitter.Emit(ctx, tx, logger, ddbidentity.PasswordResetTokenRedeemedEventType, "", nil, events.WithUserID(userID)); err != nil {
		return platformerrors.Wrap(err, "recording a password reset")
	}

	return nil
}
