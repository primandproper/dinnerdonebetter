package issue_reports

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	generated "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/issuereports/generated"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/database/filtering"
	platformerrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/identifiers"
	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/tracing"
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

	results, err := r.generatedQuerier.GetIssueReports(ctx, r.readDB, &generated.GetIssueReportsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
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

	results, err := r.generatedQuerier.GetIssueReportsForAccount(ctx, r.readDB, &generated.GetIssueReportsForAccountParams{
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
		BelongsToAccount: accountID,
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
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
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(identitykeys.AccountIDKey, input.BelongsToAccount)

	logger.Debug("CreateIssueReport invoked")

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if err = r.generatedQuerier.CreateIssueReport(ctx, tx, &generated.CreateIssueReportParams{
		ID:               input.ID,
		IssueType:        input.IssueType,
		Details:          input.Details,
		RelevantTable:    database.NullStringFromString(input.RelevantTable),
		RelevantRecordID: database.NullStringFromString(input.RelevantRecordID),
		CreatedByUser:    input.CreatedByUser,
		BelongsToAccount: input.BelongsToAccount,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing issue report creation query")
	}

	x := &types.IssueReport{
		ID:               input.ID,
		IssueType:        input.IssueType,
		Details:          input.Details,
		RelevantTable:    input.RelevantTable,
		RelevantRecordID: input.RelevantRecordID,
		CreatedByUser:    input.CreatedByUser,
		BelongsToAccount: input.BelongsToAccount,
		CreatedAt:        r.CurrentTime(),
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &x.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeIssueReports,
		RelevantID:       x.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing database transaction")
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
		return platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(issuereportkeys.IssueReportIDKey, issueReport.ID)
	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, issueReport.ID)

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	rowsAffected, err := r.generatedQuerier.UpdateIssueReport(ctx, tx, &generated.UpdateIssueReportParams{
		ID:               issueReport.ID,
		IssueType:        issueReport.IssueType,
		Details:          issueReport.Details,
		RelevantTable:    database.NullStringFromString(issueReport.RelevantTable),
		RelevantRecordID: database.NullStringFromString(issueReport.RelevantRecordID),
	})
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "updating issue report")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &issueReport.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeIssueReports,
		RelevantID:       issueReport.ID,
		EventType:        audit.AuditLogEventTypeUpdated,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing database transaction")
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

	results, err := r.generatedQuerier.GetIssueReportsForTable(ctx, r.readDB, &generated.GetIssueReportsForTableParams{
		RelevantTable:   database.NullStringFromString(tableName),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
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

	results, err := r.generatedQuerier.GetIssueReportsForRecord(ctx, r.readDB, &generated.GetIssueReportsForRecordParams{
		RelevantTable:    database.NullStringFromString(tableName),
		RelevantRecordID: database.NullStringFromString(recordID),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
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

	issueReport, err := r.GetIssueReport(ctx, issueReportID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching issue report for archive")
	}

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	rowsAffected, err := r.generatedQuerier.ArchiveIssueReport(ctx, tx, issueReportID)
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "archiving issue report")
	}

	if rowsAffected == 0 {
		r.RollbackTransaction(ctx, tx)
		return sql.ErrNoRows
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &issueReport.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeIssueReports,
		RelevantID:       issueReportID,
		EventType:        audit.AuditLogEventTypeArchived,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	return nil
}
