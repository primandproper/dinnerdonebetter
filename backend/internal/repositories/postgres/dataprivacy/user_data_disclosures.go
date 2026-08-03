package dataprivacy

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/dataprivacy/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	disclosureIDKey                = "disclosure_id"
	resourceTypeUserDataDisclosure = "user_data_disclosure"
)

// CreateUserDataDisclosure creates a new user data disclosure record.
func (r *disclosureRepository) CreateUserDataDisclosure(ctx context.Context, input *dataprivacy.UserDataDisclosureCreationInput) (*dataprivacy.UserDataDisclosure, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	if input.ID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, disclosureIDKey, input.ID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.BelongsToUser)

	logger := r.logger.WithValue(disclosureIDKey, input.ID).WithValue(identitykeys.UserIDKey, input.BelongsToUser)
	logger.Info("creating user data disclosure")

	if err := r.generatedQuerier.CreateUserDataDisclosure(ctx, r.writeDB, &generated.CreateUserDataDisclosureParams{
		ID:            input.ID,
		BelongsToUser: input.BelongsToUser,
		ExpiresAt:     input.ExpiresAt,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating user data disclosure")
	}

	if _, err := r.auditLogRepo.CreateAuditLogEntry(ctx, r.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		ID:            identifiers.New(),
		ResourceType:  resourceTypeUserDataDisclosure,
		RelevantID:    input.ID,
		EventType:     audit.AuditLogEventTypeCreated,
		BelongsToUser: input.BelongsToUser,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
	}

	disclosure, err := r.GetUserDataDisclosure(ctx, input.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching created disclosure")
	}

	return disclosure, nil
}

// GetUserDataDisclosure fetches a user data disclosure by ID.
func (r *disclosureRepository) GetUserDataDisclosure(ctx context.Context, disclosureID string) (*dataprivacy.UserDataDisclosure, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if disclosureID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, disclosureIDKey, disclosureID)
	logger := r.logger.WithValue(disclosureIDKey, disclosureID)

	result, err := r.generatedQuerier.GetUserDataDisclosure(ctx, r.readDB, disclosureID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user data disclosure")
	}

	disclosure := &dataprivacy.UserDataDisclosure{
		ID:            result.ID,
		BelongsToUser: result.BelongsToUser,
		Status:        dataprivacy.UserDataDisclosureStatus(result.Status),
		ExpiresAt:     result.ExpiresAt,
		CreatedAt:     result.CreatedAt,
	}

	if result.LastUpdatedAt.Valid {
		disclosure.LastUpdatedAt = &result.LastUpdatedAt.Time
	}
	if result.CompletedAt.Valid {
		disclosure.CompletedAt = &result.CompletedAt.Time
	}
	if result.ArchivedAt.Valid {
		disclosure.ArchivedAt = &result.ArchivedAt.Time
	}
	if result.ReportID.Valid {
		disclosure.ReportID = result.ReportID.String
	}

	return disclosure, nil
}

// GetUserDataDisclosuresForUser fetches user data disclosures for a user.
func (r *disclosureRepository) GetUserDataDisclosuresForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	args := &generated.GetUserDataDisclosuresForUserParams{
		UserID: userID,
	}

	if filter.CreatedAfter != nil {
		args.CreatedAfter = sql.NullTime{Time: *filter.CreatedAfter, Valid: true}
	}
	if filter.CreatedBefore != nil {
		args.CreatedBefore = sql.NullTime{Time: *filter.CreatedBefore, Valid: true}
	}
	if filter.Cursor != nil {
		args.Cursor = sql.NullString{String: *filter.Cursor, Valid: true}
	}
	if filter.MaxResponseSize != nil {
		args.ResultLimit = sql.NullInt32{Int32: int32(*filter.MaxResponseSize), Valid: true}
	}

	results, err := r.generatedQuerier.GetUserDataDisclosuresForUser(ctx, r.readDB, args)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user data disclosures")
	}

	disclosures := make([]*dataprivacy.UserDataDisclosure, 0, len(results))
	var filteredCount, totalCount int64

	for _, result := range results {
		disclosure := &dataprivacy.UserDataDisclosure{
			ID:            result.ID,
			BelongsToUser: result.BelongsToUser,
			Status:        dataprivacy.UserDataDisclosureStatus(result.Status),
			ExpiresAt:     result.ExpiresAt,
			CreatedAt:     result.CreatedAt,
		}

		if result.LastUpdatedAt.Valid {
			disclosure.LastUpdatedAt = &result.LastUpdatedAt.Time
		}
		if result.CompletedAt.Valid {
			disclosure.CompletedAt = &result.CompletedAt.Time
		}
		if result.ArchivedAt.Valid {
			disclosure.ArchivedAt = &result.ArchivedAt.Time
		}
		if result.ReportID.Valid {
			disclosure.ReportID = result.ReportID.String
		}

		disclosures = append(disclosures, disclosure)
		filteredCount = result.FilteredCount
		totalCount = result.TotalCount
	}

	return &filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure]{
		Data: disclosures,
		Pagination: filtering.Pagination{
			FilteredCount: uint64(filteredCount),
			TotalCount:    uint64(totalCount),
		},
	}, nil
}

// GetExpiredUserDataDisclosures fetches disclosures that are past their expiry and have not yet
// been reaped, oldest first, up to dataprivacy.ExpiredUserDataDisclosureBatchSize of them. A full
// batch means there may be more waiting.
func (r *disclosureRepository) GetExpiredUserDataDisclosures(ctx context.Context) ([]*dataprivacy.UserDataDisclosure, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	results, err := r.generatedQuerier.GetExpiredUserDataDisclosures(ctx, r.readDB)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "fetching expired user data disclosures")
	}

	disclosures := make([]*dataprivacy.UserDataDisclosure, 0, len(results))
	for _, result := range results {
		disclosure := &dataprivacy.UserDataDisclosure{
			ID:            result.ID,
			BelongsToUser: result.BelongsToUser,
			Status:        dataprivacy.UserDataDisclosureStatus(result.Status),
			ExpiresAt:     result.ExpiresAt,
			CreatedAt:     result.CreatedAt,
		}

		if result.LastUpdatedAt.Valid {
			disclosure.LastUpdatedAt = &result.LastUpdatedAt.Time
		}
		if result.CompletedAt.Valid {
			disclosure.CompletedAt = &result.CompletedAt.Time
		}
		if result.ArchivedAt.Valid {
			disclosure.ArchivedAt = &result.ArchivedAt.Time
		}
		if result.ReportID.Valid {
			disclosure.ReportID = result.ReportID.String
		}

		disclosures = append(disclosures, disclosure)
	}

	return disclosures, nil
}

// MarkUserDataDisclosureExpired marks a disclosure as expired.
//
// This is the record that the artifact is gone, so it must be written after the artifact is
// actually destroyed and never before: a row claiming a report has expired while the object is
// still in the bucket is worse than no row at all, because nothing will come back for it.
func (r *disclosureRepository) MarkUserDataDisclosureExpired(ctx context.Context, disclosureID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if disclosureID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, disclosureIDKey, disclosureID)
	logger := r.logger.WithValue(disclosureIDKey, disclosureID)

	if err := r.generatedQuerier.MarkUserDataDisclosureExpired(ctx, r.writeDB, disclosureID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking disclosure expired")
	}

	return nil
}

// MarkUserDataDisclosureCompleted marks a disclosure as completed.
func (r *disclosureRepository) MarkUserDataDisclosureCompleted(ctx context.Context, disclosureID, reportID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if disclosureID == "" || reportID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, disclosureIDKey, disclosureID)
	logger := r.logger.WithValue(disclosureIDKey, disclosureID)

	if err := r.generatedQuerier.MarkUserDataDisclosureCompleted(ctx, r.writeDB, &generated.MarkUserDataDisclosureCompletedParams{
		ID:       disclosureID,
		ReportID: sql.NullString{String: reportID, Valid: true},
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking disclosure completed")
	}

	return nil
}

// MarkUserDataDisclosureFailed marks a disclosure as failed.
func (r *disclosureRepository) MarkUserDataDisclosureFailed(ctx context.Context, disclosureID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if disclosureID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, disclosureIDKey, disclosureID)
	logger := r.logger.WithValue(disclosureIDKey, disclosureID)

	if err := r.generatedQuerier.MarkUserDataDisclosureFailed(ctx, r.writeDB, disclosureID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking disclosure failed")
	}

	return nil
}

// ArchiveUserDataDisclosure archives a disclosure.
func (r *disclosureRepository) ArchiveUserDataDisclosure(ctx context.Context, disclosureID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if disclosureID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	tracing.AttachToSpan(span, disclosureIDKey, disclosureID)
	logger := r.logger.WithValue(disclosureIDKey, disclosureID)

	disclosure, err := r.GetUserDataDisclosure(ctx, disclosureID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching disclosure for archive")
	}

	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.ArchiveUserDataDisclosure(ctx, tx, disclosureID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving disclosure")
		}

		if _, err = r.auditLogRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			ID:            identifiers.New(),
			ResourceType:  resourceTypeUserDataDisclosure,
			RelevantID:    disclosureID,
			EventType:     audit.AuditLogEventTypeArchived,
			BelongsToUser: disclosure.BelongsToUser,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
