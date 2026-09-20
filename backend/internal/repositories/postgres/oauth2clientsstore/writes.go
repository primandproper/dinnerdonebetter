package oauth2clientsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/keys"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// The three writes, and the one that is not recorded.
//
// DeleteClientsForOwner is unrecorded because it is the erasure, and the erasure writes its
// own entry through dataprivacy. A second one naming the registrations it destroyed would
// be personal data surviving the erasure that removed it — the same rule notificationsstore
// applies to its two delete paths.
//
// Each of the three records off the row the write returned rather than off its argument.
// That is not a style choice: UpdateClient takes an UpdateInput and an id, so the argument
// never held the name a revision entry has to record, and ArchiveClient takes an id alone.
// The row is the only place the fields exist, which is why platform returns one.

// CreateClient records a registration, then records that.
func (s *store) CreateClient(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	client *platformoauth2clients.Client,
) (*platformoauth2clients.Client, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	created, err := s.Store.CreateClient(ctx, tx, scope, client)
	if err != nil {
		return nil, err
	}

	if err = s.record(ctx, tx, created,
		audit.AuditLogEventTypeCreated,
		ddboauth.OAuth2ClientCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateClient revises a registration, then records it.
func (s *store) UpdateClient(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	id string,
	input *platformoauth2clients.UpdateInput,
) (*platformoauth2clients.Client, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	revised, err := s.Store.UpdateClient(ctx, tx, scope, id, input)
	if err != nil {
		return nil, err
	}

	if err = s.record(ctx, tx, revised,
		audit.AuditLogEventTypeUpdated,
		ddboauth.OAuth2ClientUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return revised, nil
}

// ArchiveClient withdraws a registration, then records it.
//
// The row comes back from the withdrawal, which is what lets the entry name what was
// withdrawn: every read of this registry but the client_id resolution filters archived
// rows out, so a read afterwards would find nothing.
func (s *store) ArchiveClient(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	id string,
) (*platformoauth2clients.Client, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	withdrawn, err := s.Store.ArchiveClient(ctx, tx, scope, id)
	if err != nil {
		return nil, err
	}

	if err = s.record(ctx, tx, withdrawn,
		audit.AuditLogEventTypeArchived,
		ddboauth.OAuth2ClientArchivedServiceEventType); err != nil {
		return nil, err
	}

	return withdrawn, nil
}

// record writes the audit entry and enqueues the data change event, inside the caller's
// transaction, so they commit with the row they describe or not at all.
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
func (s *store) record(
	ctx context.Context,
	tx database.Tx,
	client *platformoauth2clients.Client,
	auditEventType, changeEventType string,
) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if client == nil {
		// A write that answered with no row and no error. platform's SQL store cannot,
		// but Store is a seam, and an entry naming nothing is worse than none.
		return nil
	}

	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientIDKey, client.ID)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientClientIDKey, client.ClientID)

	logger := s.logger.WithSpan(span).WithValue(oauthkeys.OAuth2ClientIDKey, client.ID)

	return s.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
		ID:            identifiers.New(),
		ResourceType:  resourceTypeOAuth2Clients,
		RelevantID:    client.ID,
		EventType:     auditEventType,
		BelongsToUser: client.BelongsToUser,
	}, changeEventType, "", map[string]any{
		oauthkeys.OAuth2ClientIDKey:       client.ID,
		oauthkeys.OAuth2ClientClientIDKey: client.ClientID,
	})
}
