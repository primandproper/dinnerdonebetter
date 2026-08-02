package identity

import (
	"context"
	"database/sql"
	"errors"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

var (
	_ identity.WebAuthnCredentialDataManager = (*repository)(nil)
)

// CreateWebAuthnCredential creates a new WebAuthn credential for a user.
func (r *repository) CreateWebAuthnCredential(ctx context.Context, input *identity.WebAuthnCredentialCreationInput) (*identity.WebAuthnCredential, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	if input.BelongsToUser == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	if len(input.CredentialID) == 0 || len(input.PublicKey) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	logger = logger.WithValue(identitykeys.UserIDKey, input.BelongsToUser)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.BelongsToUser)

	id := input.ID
	if id == "" {
		id = identifiers.New()
	}

	err := r.generatedQuerier.CreateWebAuthnCredential(ctx, r.writeDB, &generated.CreateWebAuthnCredentialParams{
		ID:            id,
		BelongsToUser: input.BelongsToUser,
		CredentialID:  input.CredentialID,
		PublicKey:     input.PublicKey,
		SignCount:     int32(input.SignCount),
		Transports:    input.Transports,
		FriendlyName:  input.FriendlyName,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating webauthn credential")
	}

	return &identity.WebAuthnCredential{
		ID:            id,
		BelongsToUser: input.BelongsToUser,
		CredentialID:  input.CredentialID,
		PublicKey:     input.PublicKey,
		SignCount:     input.SignCount,
		Transports:    input.Transports,
		FriendlyName:  input.FriendlyName,
		CreatedAt:     r.CurrentTime(),
	}, nil
}

// GetWebAuthnCredentialsForUser fetches all active WebAuthn credentials for a user.
func (r *repository) GetWebAuthnCredentialsForUser(ctx context.Context, userID string) ([]*identity.WebAuthnCredential, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	rows, err := r.generatedQuerier.GetWebAuthnCredentialsForUser(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting webauthn credentials for user")
	}

	creds := make([]*identity.WebAuthnCredential, 0, len(rows))
	for _, row := range rows {
		creds = append(creds, webauthnCredentialFromRow(row))
	}
	return creds, nil
}

// GetWebAuthnCredentialByCredentialID fetches a WebAuthn credential by its credential ID.
func (r *repository) GetWebAuthnCredentialByCredentialID(ctx context.Context, credentialID []byte) (*identity.WebAuthnCredential, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if len(credentialID) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	row, err := r.generatedQuerier.GetWebAuthnCredentialByCredentialID(ctx, r.readDB, credentialID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, observability.PrepareError(err, span, "getting webauthn credential by credential id")
	}

	return webauthnCredentialFromGetByIDRow(row), nil
}

// UpdateWebAuthnCredentialSignCount updates the sign count for a WebAuthn credential.
func (r *repository) UpdateWebAuthnCredentialSignCount(ctx context.Context, id string, signCount uint32) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	_, err := r.generatedQuerier.UpdateWebAuthnCredentialSignCount(ctx, r.writeDB, &generated.UpdateWebAuthnCredentialSignCountParams{
		ID:        id,
		SignCount: int32(signCount),
	})
	return observability.PrepareError(err, span, "updating webauthn credential sign count")
}

// ArchiveWebAuthnCredential archives a WebAuthn credential by its internal ID.
func (r *repository) ArchiveWebAuthnCredential(ctx context.Context, id string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	_, err := r.generatedQuerier.ArchiveWebAuthnCredential(ctx, r.writeDB, id)
	return observability.PrepareError(err, span, "archiving webauthn credential")
}

// ArchiveWebAuthnCredentialForUser archives a WebAuthn credential only if it belongs to the given user.
// Returns nil if the credential was archived or if it did not exist / did not belong to the user.
func (r *repository) ArchiveWebAuthnCredentialForUser(ctx context.Context, id, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" || userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	_, err := r.generatedQuerier.ArchiveWebAuthnCredentialForUser(ctx, r.writeDB, &generated.ArchiveWebAuthnCredentialForUserParams{
		ID:            id,
		BelongsToUser: userID,
	})
	return observability.PrepareError(err, span, "archiving webauthn credential for user")
}

func webauthnCredentialFromRow(row *generated.GetWebAuthnCredentialsForUserRow) *identity.WebAuthnCredential {
	return &identity.WebAuthnCredential{
		ID:            row.ID,
		BelongsToUser: row.BelongsToUser,
		CredentialID:  row.CredentialID,
		PublicKey:     row.PublicKey,
		SignCount:     uint32(row.SignCount),
		Transports:    row.Transports,
		FriendlyName:  row.FriendlyName,
		CreatedAt:     row.CreatedAt,
		LastUsedAt:    database.TimePointerFromNullTime(row.LastUsedAt),
	}
}

func webauthnCredentialFromGetByIDRow(row *generated.GetWebAuthnCredentialByCredentialIDRow) *identity.WebAuthnCredential {
	return &identity.WebAuthnCredential{
		ID:            row.ID,
		BelongsToUser: row.BelongsToUser,
		CredentialID:  row.CredentialID,
		PublicKey:     row.PublicKey,
		SignCount:     uint32(row.SignCount),
		Transports:    row.Transports,
		FriendlyName:  row.FriendlyName,
		CreatedAt:     row.CreatedAt,
		LastUsedAt:    database.TimePointerFromNullTime(row.LastUsedAt),
	}
}
