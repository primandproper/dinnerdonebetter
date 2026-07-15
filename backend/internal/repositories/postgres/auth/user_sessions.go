package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	authkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auth/generated"

	"github.com/primandproper/platform-go/v4/database"
	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/identifiers"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

const (
	resourceTypeUserSessions = "user_sessions"
)

var _ auth.UserSessionDataManager = (*repository)(nil)

func convertUserSession(row *generated.UserSessions) *auth.UserSession {
	return &auth.UserSession{
		ID:             row.ID,
		BelongsToUser:  row.BelongsToUser,
		SessionTokenID: row.SessionTokenID,
		RefreshTokenID: row.RefreshTokenID,
		ClientIP:       row.ClientIp,
		UserAgent:      row.UserAgent,
		DeviceName:     row.DeviceName,
		LoginMethod:    row.LoginMethod,
		CreatedAt:      row.CreatedAt,
		LastActiveAt:   row.LastActiveAt,
		ExpiresAt:      row.ExpiresAt,
		RevokedAt:      database.TimePointerFromNullTime(row.RevokedAt),
	}
}

// CreateUserSession creates a user session in the database.
func (r *repository) CreateUserSession(ctx context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger := r.logger.WithValue(authkeys.UserSessionIDKey, input.ID)
	tracing.AttachToSpan(span, authkeys.UserSessionIDKey, input.ID)

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if err = r.generatedQuerier.CreateUserSession(ctx, tx, &generated.CreateUserSessionParams{
		ID:             input.ID,
		BelongsToUser:  input.BelongsToUser,
		SessionTokenID: input.SessionTokenID,
		RefreshTokenID: input.RefreshTokenID,
		ClientIp:       input.ClientIP,
		UserAgent:      input.UserAgent,
		DeviceName:     input.DeviceName,
		LoginMethod:    input.LoginMethod,
		ExpiresAt:      input.ExpiresAt,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "creating user session")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:            identifiers.New(),
		ResourceType:  resourceTypeUserSessions,
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

	session := &auth.UserSession{
		ID:             input.ID,
		BelongsToUser:  input.BelongsToUser,
		SessionTokenID: input.SessionTokenID,
		RefreshTokenID: input.RefreshTokenID,
		ClientIP:       input.ClientIP,
		UserAgent:      input.UserAgent,
		DeviceName:     input.DeviceName,
		LoginMethod:    input.LoginMethod,
		ExpiresAt:      input.ExpiresAt,
		CreatedAt:      r.CurrentTime(),
		LastActiveAt:   r.CurrentTime(),
	}

	logger.Info("user session created")

	return session, nil
}

// GetUserSessionBySessionTokenID fetches a user session by its access token JTI.
func (r *repository) GetUserSessionBySessionTokenID(ctx context.Context, sessionTokenID string) (*auth.UserSession, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if sessionTokenID == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	result, err := r.generatedQuerier.GetUserSessionBySessionTokenID(ctx, r.readDB, sessionTokenID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "getting user session by session token ID")
	}

	return convertUserSession(result), nil
}

// GetUserSessionByRefreshTokenID fetches a user session by its refresh token JTI.
func (r *repository) GetUserSessionByRefreshTokenID(ctx context.Context, refreshTokenID string) (*auth.UserSession, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if refreshTokenID == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	result, err := r.generatedQuerier.GetUserSessionByRefreshTokenID(ctx, r.readDB, refreshTokenID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "getting user session by refresh token ID")
	}

	return convertUserSession(result), nil
}

// GetActiveSessionsForUser fetches all active sessions for a user.
func (r *repository) GetActiveSessionsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetActiveSessionsForUser(ctx, r.readDB, &generated.GetActiveSessionsForUserParams{
		BelongsToUser: userID,
		CreatedAfter:  database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore: database.NullTimeFromTimePointer(filter.CreatedBefore),
		Cursor:        database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:   database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "getting active sessions for user")
	}

	var (
		data                      = []*auth.UserSession{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		s := &auth.UserSession{
			ID:             result.ID,
			BelongsToUser:  result.BelongsToUser,
			SessionTokenID: result.SessionTokenID,
			RefreshTokenID: result.RefreshTokenID,
			ClientIP:       result.ClientIp,
			UserAgent:      result.UserAgent,
			DeviceName:     result.DeviceName,
			LoginMethod:    result.LoginMethod,
			CreatedAt:      result.CreatedAt,
			LastActiveAt:   result.LastActiveAt,
			ExpiresAt:      result.ExpiresAt,
			RevokedAt:      database.TimePointerFromNullTime(result.RevokedAt),
		}
		data = append(data, s)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	return filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(t *auth.UserSession) string {
		return t.ID
	}, filter), nil
}

// RevokeUserSession revokes a specific user session.
func (r *repository) RevokeUserSession(ctx context.Context, sessionID, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if sessionID == "" || userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(authkeys.UserSessionIDKey, sessionID)
	tracing.AttachToSpan(span, authkeys.UserSessionIDKey, sessionID)

	rowsAffected, err := r.generatedQuerier.RevokeUserSession(ctx, r.writeDB, &generated.RevokeUserSessionParams{
		ID:            sessionID,
		BelongsToUser: userID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "revoking user session")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("user session revoked")

	return nil
}

// RevokeAllSessionsForUser revokes all sessions for a user.
func (r *repository) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if _, err := r.generatedQuerier.RevokeAllSessionsForUser(ctx, r.writeDB, userID); err != nil {
		return observability.PrepareAndLogError(err, r.logger, span, "revoking all sessions for user")
	}

	r.logger.Info("all sessions revoked for user")

	return nil
}

// RevokeAllSessionsForUserExcept revokes all sessions for a user except the specified one.
func (r *repository) RevokeAllSessionsForUserExcept(ctx context.Context, userID, sessionID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || sessionID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if _, err := r.generatedQuerier.RevokeAllSessionsForUserExcept(ctx, r.writeDB, &generated.RevokeAllSessionsForUserExceptParams{
		BelongsToUser: userID,
		SessionID:     sessionID,
	}); err != nil {
		return observability.PrepareAndLogError(err, r.logger, span, "revoking all sessions for user except current")
	}

	r.logger.Info("all other sessions revoked for user")

	return nil
}

// UpdateSessionTokenIDs updates the token IDs for a session after a token refresh.
func (r *repository) UpdateSessionTokenIDs(ctx context.Context, sessionID, newSessionTokenID, newRefreshTokenID string, newExpiresAt time.Time) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if sessionID == "" || newSessionTokenID == "" || newRefreshTokenID == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	tracing.AttachToSpan(span, authkeys.UserSessionIDKey, sessionID)

	rowsAffected, err := r.generatedQuerier.UpdateSessionTokenIDs(ctx, r.writeDB, &generated.UpdateSessionTokenIDsParams{
		ID:             sessionID,
		SessionTokenID: newSessionTokenID,
		RefreshTokenID: newRefreshTokenID,
		ExpiresAt:      newExpiresAt,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, r.logger, span, "updating session token IDs")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// TouchSessionLastActive updates the last active timestamp for a session.
func (r *repository) TouchSessionLastActive(ctx context.Context, sessionTokenID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if sessionTokenID == "" {
		return platformerrors.ErrEmptyInputProvided
	}

	if _, err := r.generatedQuerier.TouchSessionLastActive(ctx, r.writeDB, sessionTokenID); err != nil {
		return observability.PrepareAndLogError(err, r.logger, span, "touching session last active")
	}

	return nil
}
