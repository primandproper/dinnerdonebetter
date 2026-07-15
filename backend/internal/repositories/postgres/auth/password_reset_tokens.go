package auth

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	authkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auth/generated"

	"github.com/primandproper/platform-go/v4/database"
	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/identifiers"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

const (
	resourceTypePasswordResetTokens = "password_reset_tokens"
)

var (
	_ auth.PasswordResetTokenDataManager = (*repository)(nil)
)

// GetPasswordResetTokenByID fetches a password reset token from the database by its ID.
func (r *repository) GetPasswordResetTokenByID(ctx context.Context, passwordResetTokenID string) (*auth.PasswordResetToken, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if passwordResetTokenID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, passwordResetTokenID)

	result, err := r.generatedQuerier.GetPasswordResetTokenByID(ctx, r.readDB, passwordResetTokenID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "getting password reset token by ID")
	}

	return &auth.PasswordResetToken{
		CreatedAt:     result.CreatedAt,
		ExpiresAt:     result.ExpiresAt,
		RedeemedAt:    database.TimePointerFromNullTime(result.RedeemedAt),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:            result.ID,
		Token:         result.Token,
		BelongsToUser: result.BelongsToUser,
	}, nil
}

// GetPasswordResetTokenByToken fetches a password reset token from the database by its token.
func (r *repository) GetPasswordResetTokenByToken(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if token == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, token)

	result, err := r.generatedQuerier.GetPasswordResetToken(ctx, r.readDB, token)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "getting password reset token")
	}

	passwordResetToken := &auth.PasswordResetToken{
		CreatedAt:     result.CreatedAt,
		ExpiresAt:     result.ExpiresAt,
		RedeemedAt:    database.TimePointerFromNullTime(result.RedeemedAt),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:            result.ID,
		Token:         result.Token,
		BelongsToUser: result.BelongsToUser,
	}

	return passwordResetToken, nil
}

// CreatePasswordResetToken creates a password reset token in the database.
func (r *repository) CreatePasswordResetToken(ctx context.Context, input *auth.PasswordResetTokenDatabaseCreationInput) (*auth.PasswordResetToken, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger := r.logger.WithValue(authkeys.PasswordResetTokenIDKey, input.ID)
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, input.ID)

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	// create the password reset token.
	if err = r.generatedQuerier.CreatePasswordResetToken(ctx, tx, &generated.CreatePasswordResetTokenParams{
		ID:            input.ID,
		Token:         input.Token,
		BelongsToUser: input.BelongsToUser,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing password reset token creation query")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:            identifiers.New(),
		ResourceType:  resourceTypePasswordResetTokens,
		RelevantID:    input.ID,
		EventType:     audit.AuditLogEventTypeCreated,
		BelongsToUser: input.BelongsToUser,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	x := &auth.PasswordResetToken{
		ID:            input.ID,
		Token:         input.Token,
		ExpiresAt:     input.ExpiresAt,
		CreatedAt:     r.CurrentTime(),
		BelongsToUser: input.BelongsToUser,
	}

	logger.Info("password reset token created")

	return x, nil
}

// RedeemPasswordResetToken redeems a password reset token from the database by its ID.
func (r *repository) RedeemPasswordResetToken(ctx context.Context, passwordResetTokenID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if passwordResetTokenID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(authkeys.PasswordResetTokenIDKey, passwordResetTokenID)
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, passwordResetTokenID)

	token, err := r.generatedQuerier.GetPasswordResetTokenByID(ctx, r.readDB, passwordResetTokenID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching password reset token for redeem")
	}

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if err = r.generatedQuerier.RedeemPasswordResetToken(ctx, tx, passwordResetTokenID); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "redeeming password reset token")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:            identifiers.New(),
		ResourceType:  resourceTypePasswordResetTokens,
		RelevantID:    passwordResetTokenID,
		EventType:     audit.AuditLogEventTypeUpdated,
		BelongsToUser: token.BelongsToUser,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	return nil
}
