/*
Package oauth2clientsstore is what a registration of an OAuth2 client means to the
rest of this application. The registry itself is platform-go's: the table, the
credentials, the lifecycle and the transaction all live there, and this package
neither reimplements nor wraps them.

It replaces internal/repositories/postgres/oauth, which owned the oauth2_clients
table. The table is platform's now — see migrate.go's
adopt_oauth2_registered_clients — and what is left here is the part that was
never platform's: the audit entry and the data change event a registration owes.

Both are written from platform's oauth2clients.Hooks, which the registry Service
calls on the transaction it opened once each write has landed. A hook's error
rolls the operation back, so an entry refused takes the credential with it, and
a registration whose event failed to enqueue is not a client that exists and
that nothing downstream was told about.

The hooks hang off the Service rather than the Store because the Service is the
one that owns the transaction, and every write in this application goes through
it. The Store is registered undecorated, for its reads.
*/
package oauth2clientsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformoauth2clients "github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// resourceTypeOAuth2Clients is what an audit entry about a registration names.
//
// It keeps the old table's name rather than taking the new one's, for the reason
// webhooksstore keeps webhooks: an audit log is read across the change, and an
// investigation asking what happened to a client should find the entries written
// before the store moved as well as the ones after.
const resourceTypeOAuth2Clients = "oauth2_clients"

// hooks records every registration, revision and withdrawal.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
//
// Store.DeleteClientsForOwner has no hook and is not recorded: it is the erasure,
// and the erasure writes its own entry through dataprivacy. A second one naming
// the registrations it destroyed would be personal data surviving the erasure that
// removed it — the same rule notificationsstore applies to its two delete paths.
//
// Each method records off the row platform hands it rather than off the
// operation's arguments: an update is an id and an UpdateInput, and an archive an
// id alone, so the row is the only place the name and the client_id exist.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformoauth2clients.Hooks = (*hooks)(nil)

// AfterCreateClient records a registration.
func (h *hooks) AfterCreateClient(ctx context.Context, tx database.Tx, _ tenancy.Scope, client *platformoauth2clients.Client) error {
	return h.record(ctx, tx, client, platformaudit.EventCreated, ddboauth.OAuth2ClientCreatedServiceEventType, nil)
}

// AfterUpdateClient records a revision, and which of its fields moved.
//
// The secret digest is never among them: it is tagged json:"-", which Diff
// skips, and a revision cannot rotate it anyway.
func (h *hooks) AfterUpdateClient(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformoauth2clients.Client) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated oauth2 client")
	}

	return h.record(ctx, tx, after, platformaudit.EventUpdated, ddboauth.OAuth2ClientUpdatedServiceEventType, changes)
}

// AfterArchiveClient records a withdrawal. The row is the one the archive left,
// which still names what was withdrawn: every read of this registry but the
// client_id resolution filters archived rows out, so a read afterwards would find
// nothing.
func (h *hooks) AfterArchiveClient(ctx context.Context, tx database.Tx, _ tenancy.Scope, client *platformoauth2clients.Client) error {
	return h.record(ctx, tx, client, platformaudit.EventArchived, ddboauth.OAuth2ClientArchivedServiceEventType, nil)
}

// record writes the audit entry and enqueues the data change event on the
// operation's transaction, so they commit with the row they describe or not at
// all. changes is the field-level diff a revision carries, and nil otherwise.
//
// The entry belongs to the registration's owner where it has one and to nobody where it
// does not. An administered client is minted by an operator to speak for the service on
// behalf of whoever signs in, so there is no person it is about; filing it under the
// operator who happened to run the call would make the log answer "whose client is this"
// with the wrong name.
//
// The client_id is on the metadata and the secret digest is on neither. The client_id is
// the protocol identifier and travels in every /authorize URL, so recording it is what
// makes an entry joinable to a token; a digest is a credential and belongs in no log.
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	client *platformoauth2clients.Client,
	auditEventType platformaudit.EventType,
	changeEventType string,
	changes map[string]platformaudit.Change,
) error {
	if client == nil {
		// A write that answered with no row and no error. platform's SQL store cannot,
		// but Store is a seam, and an entry naming nothing is worse than none.
		return nil
	}

	logger := h.logger.WithValue(oauthkeys.OAuth2ClientIDKey, client.ID)

	entry := audit.NewEntry(client.BelongsToUser, "", resourceTypeOAuth2Clients, client.ID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, "", map[string]any{
		oauthkeys.OAuth2ClientIDKey:       client.ID,
		oauthkeys.OAuth2ClientClientIDKey: client.ClientID,
	})
}
