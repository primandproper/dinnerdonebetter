/*
Package identitystore wires platform-go's identity service into this application's
recording spine and its users search index.

The recording is platform's. identity.RecordingHooks writes the audit entries and
publishes the events every identity operation owes, on the operation's own
database.Tx, so they commit with the rows they describe or not at all. Nothing here
decides what an operation records; see platform-go's identity/recording.go.

What is this application's is the users search index, and the one reason this
package still has a type of its own. The index is fed by platform's searchsync side
effect, which derives index events from the messages the outbox writer is handed
by asserting searchsync.Change on each payload — and no payload platform's own hooks
emit implements it (platform-go #1136). Until that lands, Hooks wraps platform's and
runs the index rules itself for the four user writes the index cares about, through
the writer's EnqueueDerived, so a registration or a profile change still reaches the
index. When #1136 lands the index rules match platform's events on their own, this
wrapper deletes down to platform's constructor, and nothing else here changes.
*/
package identitystore

import (
	"context"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// ErrNilEmitter indicates a nil events.Emitter handed to ProvideHooks.
var ErrNilEmitter = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil identity events emitter")

// Hooks is platform's identity.RecordingHooks with the users search index fed beside it.
//
// Every method is platform's, recorded and published as platform records it. The four this
// type overrides — a registration through either door, a profile change, and an archival —
// are the writes that change what the users index says about somebody, and each one runs
// platform's recording first and then derives the index event the recording cannot yet
// derive for itself (platform-go #1136). The override runs after the recording rather than
// before so that a refused recording leaves no index event behind it.
type Hooks struct {
	*platformidentity.RecordingHooks

	emitter *events.Emitter
}

var _ platformidentity.Hooks = (*Hooks)(nil)

// ProvideHooks builds platform's recording hooks over the spine's Recorder, with the users
// index bridge around them.
func ProvideHooks(emitter *events.Emitter) (*Hooks, error) {
	if emitter == nil {
		return nil, ErrNilEmitter
	}

	recording, err := platformidentity.NewRecordingHooks(emitter.Recorder())
	if err != nil {
		return nil, platformerrors.Wrap(err, "building identity recording hooks")
	}

	return &Hooks{RecordingHooks: recording, emitter: emitter}, nil
}

// AfterRegister records the registration as platform does, then indexes the new user.
func (h *Hooks) AfterRegister(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	registration *platformidentity.Registration,
) error {
	if err := h.RecordingHooks.AfterRegister(ctx, tx, scope, registration); err != nil {
		return err
	}

	return h.index(ctx, tx, platformidentity.EventUserRegistered, registration.User.ID)
}

// AfterRegisterWithInvitation records the registration as platform does, then indexes the new
// user.
func (h *Hooks) AfterRegisterWithInvitation(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	registration *platformidentity.InvitedRegistration,
) error {
	if err := h.RecordingHooks.AfterRegisterWithInvitation(ctx, tx, scope, registration); err != nil {
		return err
	}

	return h.index(ctx, tx, platformidentity.EventUserRegistered, registration.User.ID)
}

// AfterUpdateProfile records the change as platform does, then reindexes the user. Every
// handle and name a search can match on is a profile field, and both handle doors on sign-in
// write through this operation.
func (h *Hooks) AfterUpdateProfile(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	user *platformidentity.User,
	changed []string,
) error {
	if err := h.RecordingHooks.AfterUpdateProfile(ctx, tx, scope, user, changed); err != nil {
		return err
	}

	return h.index(ctx, tx, platformidentity.EventUserProfileUpdated, user.ID)
}

// AfterArchiveUser records the archival as platform does, then removes the user's document.
func (h *Hooks) AfterArchiveUser(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	user *platformidentity.User,
	endedMemberships []*platformidentity.Membership,
) error {
	if err := h.RecordingHooks.AfterArchiveUser(ctx, tx, scope, user, endedMemberships); err != nil {
		return err
	}

	return h.index(ctx, tx, platformidentity.EventUserArchived, user.ID)
}

// index derives the index event platform's eventType implies for userID, under the rule
// internal/indexevents tables for that event, without announcing anything a second time: the
// announcement was platform's, a moment ago, on the same transaction.
func (h *Hooks) index(ctx context.Context, tx database.Tx, eventType webhooks.EventType, userID string) error {
	if err := h.emitter.EmitIndex(ctx, tx, eventType.String(), map[string]any{identitykeys.UserIDKey: userID}); err != nil {
		return platformerrors.Wrapf(err, "indexing user for %s", eventType)
	}

	return nil
}
