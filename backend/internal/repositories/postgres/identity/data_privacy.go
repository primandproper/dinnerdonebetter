package identity

import (
	"context"
	"database/sql"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

// DeleteUser hard-deletes a user and all associated data via ON DELETE CASCADE.
func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	changed, err := r.EraseUser(ctx, r.writeDB, userID)
	if err != nil {
		return err
	}

	if changed == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// EraseUser implements identity.UserDataManager.
//
// The single DELETE is the whole erasure for every domain whose rows hang off a
// user or an account by ON DELETE CASCADE, which is all of them bar the audit
// log — see internal/repositories/postgres/dataprivacy/audit_erasure.go for the
// one store the cascade cannot reach, and why.
//
// Zero rows is not an error here, the way it is for DeleteUser. An erasure is a
// statement about the end state, and a subject who is already absent is in it;
// returning sql.ErrNoRows would roll back a transaction with nothing left to do
// and fail a request that had already been satisfied.
func (r *repository) EraseUser(ctx context.Context, q database.SQLQueryExecutor, userID string) (int64, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return 0, platformerrors.ErrInvalidIDProvided
	}

	if q == nil {
		return 0, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil query executor")
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, userID)

	changed, err := r.generatedQuerier.DeleteUser(ctx, q, userID)
	if err != nil {
		return 0, observability.PrepareAndLogError(err, logger, span, "deleting user")
	}

	logger.WithValue("rows_deleted", changed).Info("user deleted")

	return changed, nil
}
