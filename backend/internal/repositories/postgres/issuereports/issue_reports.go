package issue_reports

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeIssueReports = "issue_reports"
)

var (
	_ types.IssueReportDataManager = (*repository)(nil)
)

// GetIssueReport fetches an issue report from the database.
func (r *repository) GetIssueReport(ctx context.Context, issueReportID string) (*types.IssueReport, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if issueReportID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(issuereportkeys.IssueReportIDKey, issueReportID)
	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, issueReportID)

	result, err := r.generatedQuerier.GetIssueReport(ctx, r.readDB, issueReportID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching issue report")
	}

	issueReport := &types.IssueReport{
		ID:               result.ID,
		IssueType:        result.IssueType,
		Details:          result.Details,
		RelevantTable:    database.StringFromNullString(result.RelevantTable),
		RelevantRecordID: database.StringFromNullString(result.RelevantRecordID),
		CreatedAt:        result.CreatedAt,
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		CreatedByUser:    result.CreatedByUser,
		BelongsToAccount: result.BelongsToAccount,
	}

	return issueReport, nil
}

// GetIssueReports fetches a list of issue reports from the database that meet a particular filter.
func (r *repository) GetIssueReports(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.IssueReport], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := r.generatedQuerier.GetIssueReports(ctx, r.readDB, &generated.GetIssueReportsParams{
		CreatedAfter:    filterArgs.CreatedAfter,
		CreatedBefore:   filterArgs.CreatedBefore,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		IncludeArchived: filterArgs.IncludeArchived,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching issue reports from database")
	}

	var (
		data                      []*types.IssueReport
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.IssueReport{
			ID:               result.ID,
			IssueType:        result.IssueType,
			Details:          result.Details,
			RelevantTable:    database.StringFromNullString(result.RelevantTable),
			RelevantRecordID: database.StringFromNullString(result.RelevantRecordID),
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.IssueReport) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetIssueReportsForAccount fetches a list of issue reports for a specific account from the database that meet a particular filter.
func (r *repository) GetIssueReportsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.IssueReport], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := r.generatedQuerier.GetIssueReportsForAccount(ctx, r.readDB, &generated.GetIssueReportsForAccountParams{
		CreatedAfter:     filterArgs.CreatedAfter,
		CreatedBefore:    filterArgs.CreatedBefore,
		UpdatedBefore:    filterArgs.UpdatedBefore,
		UpdatedAfter:     filterArgs.UpdatedAfter,
		IncludeArchived:  filterArgs.IncludeArchived,
		BelongsToAccount: accountID,
		PageCursor:       filterArgs.Cursor,
		ResultLimit:      filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching issue reports from database")
	}

	var (
		data                      []*types.IssueReport
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.IssueReport{
			ID:               result.ID,
			IssueType:        result.IssueType,
			Details:          result.Details,
			RelevantTable:    database.StringFromNullString(result.RelevantTable),
			RelevantRecordID: database.StringFromNullString(result.RelevantRecordID),
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.IssueReport) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateIssueReport creates an issue report in the database.
func (r *repository) CreateIssueReport(ctx context.Context, input *types.IssueReportDatabaseCreationInput) (*types.IssueReport, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(identitykeys.AccountIDKey, input.BelongsToAccount)

	logger.Debug("CreateIssueReport invoked")

	var err error
	var x *types.IssueReport
	if err = r.WithTransaction(ctx, func(tx database.Tx) error {
		if err = r.generatedQuerier.CreateIssueReport(ctx, tx, &generated.CreateIssueReportParams{
			ID:               input.ID,
			IssueType:        input.IssueType,
			Details:          input.Details,
			RelevantTable:    database.NullStringFromString(input.RelevantTable),
			RelevantRecordID: database.NullStringFromString(input.RelevantRecordID),
			CreatedByUser:    input.CreatedByUser,
			BelongsToAccount: input.BelongsToAccount,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing issue report creation query")
		}

		x = &types.IssueReport{
			ID:               input.ID,
			IssueType:        input.IssueType,
			Details:          input.Details,
			RelevantTable:    input.RelevantTable,
			RelevantRecordID: input.RelevantRecordID,
			CreatedByUser:    input.CreatedByUser,
			BelongsToAccount: input.BelongsToAccount,
			CreatedAt:        r.CurrentTime(),
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &x.BelongsToAccount,
			BelongsToUser:    x.CreatedByUser,
			ResourceType:     resourceTypeIssueReports,
			RelevantID:       x.ID,
			EventType:        audit.AuditLogEventTypeCreated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.IssueReportCreatedServiceEventType, "", map[string]any{
			issuereportkeys.IssueReportIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, x.ID)

	return x, nil
}

// UpdateIssueReport updates an issue report in the database.
func (r *repository) UpdateIssueReport(ctx context.Context, issueReport *types.IssueReport) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if issueReport == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger = logger.WithValue(issuereportkeys.IssueReportIDKey, issueReport.ID)
	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, issueReport.ID)

	if err := r.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, err := r.generatedQuerier.UpdateIssueReport(ctx, tx, &generated.UpdateIssueReportParams{
			ID:               issueReport.ID,
			IssueType:        issueReport.IssueType,
			Details:          issueReport.Details,
			RelevantTable:    database.NullStringFromString(issueReport.RelevantTable),
			RelevantRecordID: database.NullStringFromString(issueReport.RelevantRecordID),
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating issue report")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &issueReport.BelongsToAccount,
			ResourceType:     resourceTypeIssueReports,
			RelevantID:       issueReport.ID,
			EventType:        audit.AuditLogEventTypeUpdated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.IssueReportUpdatedServiceEventType, "", map[string]any{
			issuereportkeys.IssueReportIDKey: issueReport.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// GetIssueReportsForTable fetches a list of issue reports for a specific table from the database that meet a particular filter.
func (r *repository) GetIssueReportsForTable(ctx context.Context, tableName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.IssueReport], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if tableName == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("relevant_table", tableName)
	tracing.AttachToSpan(span, "relevant_table", tableName)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := r.generatedQuerier.GetIssueReportsForTable(ctx, r.readDB, &generated.GetIssueReportsForTableParams{
		RelevantTable:   database.NullStringFromString(tableName),
		CreatedAfter:    filterArgs.CreatedAfter,
		CreatedBefore:   filterArgs.CreatedBefore,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		IncludeArchived: filterArgs.IncludeArchived,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching issue reports from database")
	}

	var (
		data                      []*types.IssueReport
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.IssueReport{
			ID:               result.ID,
			IssueType:        result.IssueType,
			Details:          result.Details,
			RelevantTable:    database.StringFromNullString(result.RelevantTable),
			RelevantRecordID: database.StringFromNullString(result.RelevantRecordID),
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.IssueReport) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetIssueReportsForRecord fetches a list of issue reports for a specific table+record combination from the database that meet a particular filter.
func (r *repository) GetIssueReportsForRecord(ctx context.Context, tableName, recordID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.IssueReport], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if tableName == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("relevant_table", tableName)
	tracing.AttachToSpan(span, "relevant_table", tableName)

	if recordID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("relevant_record_id", recordID)
	tracing.AttachToSpan(span, "relevant_record_id", recordID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := r.generatedQuerier.GetIssueReportsForRecord(ctx, r.readDB, &generated.GetIssueReportsForRecordParams{
		RelevantTable:    database.NullStringFromString(tableName),
		RelevantRecordID: database.NullStringFromString(recordID),
		CreatedAfter:     filterArgs.CreatedAfter,
		CreatedBefore:    filterArgs.CreatedBefore,
		UpdatedBefore:    filterArgs.UpdatedBefore,
		UpdatedAfter:     filterArgs.UpdatedAfter,
		IncludeArchived:  filterArgs.IncludeArchived,
		PageCursor:       filterArgs.Cursor,
		ResultLimit:      filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching issue reports from database")
	}

	var (
		data                      []*types.IssueReport
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.IssueReport{
			ID:               result.ID,
			IssueType:        result.IssueType,
			Details:          result.Details,
			RelevantTable:    database.StringFromNullString(result.RelevantTable),
			RelevantRecordID: database.StringFromNullString(result.RelevantRecordID),
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.IssueReport) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// ArchiveIssueReport archives an issue report from the database.
func (r *repository) ArchiveIssueReport(ctx context.Context, issueReportID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if issueReportID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, issueReportID)

	logger := r.logger.WithValue(issuereportkeys.IssueReportIDKey, issueReportID)

	issueReport, getErr := r.GetIssueReport(ctx, issueReportID)
	if getErr != nil {
		return observability.PrepareAndLogError(getErr, logger, span, "fetching issue report for archive")
	}

	if err := r.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, err := r.generatedQuerier.ArchiveIssueReport(ctx, tx, issueReportID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving issue report")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &issueReport.BelongsToAccount,
			ResourceType:     resourceTypeIssueReports,
			RelevantID:       issueReportID,
			EventType:        audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.IssueReportArchivedServiceEventType, "", map[string]any{
			issuereportkeys.IssueReportIDKey: issueReportID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
