package authentication

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/authentication/passkeys"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// passkeyHooks is what this application records when platform's passkeys service adds or
// removes a passkey: an audit entry filed under the passkey's owner, and the event beside it,
// on the write's own transaction.
//
// A passkey is a credential like a password or a second factor, and its writes are recorded
// for signInHooks' reason: "who added a key to my account" is the entry an investigation reads
// first. The entry names the passkey by its ID alone. Its friendly name is whatever the owner
// typed, and its public key and credential ID say nothing an investigation needs.
//
// It implements passkeys.Hooks outright rather than embedding NoopHooks, as the repository
// hooks do, so a write platform adds later fails to compile here until somebody decides what
// it records.
type passkeyHooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ passkeys.Hooks = (*passkeyHooks)(nil)

// NewPasskeyHooks builds the hooks platform's passkeys service runs for this application.
func NewPasskeyHooks(logger logging.Logger, recorder *recording.Recorder) passkeys.Hooks {
	return &passkeyHooks{
		logger:   logging.NewNamedLogger(logger, "passkey_hooks"),
		recorder: recorder,
	}
}

// AfterRegisterPasskey records a passkey being added to somebody's account.
func (h *passkeyHooks) AfterRegisterPasskey(ctx context.Context, tx database.Tx, _ tenancy.Scope, credential *passkeys.Credential) error {
	return h.record(ctx, tx, credential, platformaudit.EventCreated, ddbidentity.PasskeyRegisteredServiceEventType, "recording a passkey registration")
}

// AfterArchivePasskey records a passkey being removed from somebody's account.
func (h *passkeyHooks) AfterArchivePasskey(ctx context.Context, tx database.Tx, _ tenancy.Scope, credential *passkeys.Credential) error {
	return h.record(ctx, tx, credential, platformaudit.EventArchived, ddbidentity.PasskeyArchivedServiceEventType, "recording a passkey removal")
}

// AfterFailedPasskeyLogin records nothing, as signInHooks records nothing for a refused
// password: a failed login is not yet something this application keeps. A lockout or an alert
// belongs on both doors at once, and this is where the passkey half of it would go.
func (*passkeyHooks) AfterFailedPasskeyLogin(context.Context, database.Tx, tenancy.Scope, *passkeys.FailedLogin) error {
	return nil
}

// record writes the audit entry and the event for one passkey write.
func (h *passkeyHooks) record(
	ctx context.Context,
	tx database.Tx,
	credential *passkeys.Credential,
	auditEventType platformaudit.EventType,
	changeEventType string,
	describing string,
) error {
	if credential == nil || credential.BelongsToUser == "" {
		return platformerrors.Newf("%s: hook called with no passkey owner", describing)
	}

	logger := h.logger.WithValue(identitykeys.UserIDKey, credential.BelongsToUser).WithValue(identitykeys.PasskeyIDKey, credential.ID)

	entry := audit.NewEntry(credential.BelongsToUser, "", passkeysResourceType, credential.ID, auditEventType)

	if err := h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, "",
		map[string]any{
			identitykeys.UserIDKey:    credential.BelongsToUser,
			identitykeys.PasskeyIDKey: credential.ID,
		}, events.WithUserID(credential.BelongsToUser)); err != nil {
		return platformerrors.Wrap(err, describing)
	}

	return nil
}

// passkeysResourceType is what an audit entry about a passkey names.
const passkeysResourceType = "passkeys"
