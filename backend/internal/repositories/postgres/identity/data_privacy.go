package identity

import (
	"context"
	"database/sql"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

// DeleteUser hard-deletes a user and all associated data via ON DELETE CASCADE.
func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	changed, err := r.generatedQuerier.DeleteUser(ctx, r.writeDB, userID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "deleting user")
	}

	if changed == 0 {
		return sql.ErrNoRows
	}

	logger.Info("user deleted")

	return nil
}
