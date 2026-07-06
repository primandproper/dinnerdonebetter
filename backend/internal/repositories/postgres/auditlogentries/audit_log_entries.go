package auditlogentries

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	auditkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/keys"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries/generated"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/database/filtering"
	platformerrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

var (
	_ audit.AuditLogEntryDataManager = (*repository)(nil)
)

// GetAuditLogEntry fetches an audit log entry from the database.
func (q *repository) GetAuditLogEntry(ctx context.Context, auditLogEntryID string) (*audit.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if auditLogEntryID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(auditkeys.AuditLogEntryIDKey, auditLogEntryID)
	tracing.AttachToSpan(span, auditkeys.AuditLogEntryIDKey, auditLogEntryID)

	result, err := q.generatedQuerier.GetAuditLogEntry(ctx, q.readDB, auditLogEntryID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entry")
	}

	auditLogEntry := &audit.AuditLogEntry{
		CreatedAt:        result.CreatedAt,
		BelongsToAccount: database.StringPointerFromNullString(result.BelongsToAccount),
		ID:               result.ID,
		ResourceType:     result.ResourceType,
		RelevantID:       result.RelevantID,
		EventType:        string(result.EventType),
		BelongsToUser:    database.StringFromNullString(result.BelongsToUser),
	}

	if err = json.Unmarshal(result.Changes, &auditLogEntry.Changes); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "parsing audit log entry JSON data")
	}

	return auditLogEntry, nil
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetAuditLogEntriesForUser(ctx, q.readDB, &generated.GetAuditLogEntriesForUserParams{
		BelongsToUser: database.NullStringFromString(userID),
		CreatedBefore: database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:  database.NullTimeFromTimePointer(filter.CreatedAfter),
		Cursor:        database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:   database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries from database")
	}

	var (
		data                      []*audit.AuditLogEntry
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		auditLogEntry := &audit.AuditLogEntry{
			CreatedAt:        result.CreatedAt,
			BelongsToAccount: database.StringPointerFromNullString(result.BelongsToAccount),
			ID:               result.ID,
			ResourceType:     result.ResourceType,
			RelevantID:       result.RelevantID,
			EventType:        string(result.EventType),
			BelongsToUser:    database.StringFromNullString(result.BelongsToUser),
		}

		if err = json.Unmarshal(result.Changes, &auditLogEntry.Changes); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "parsing audit log entry JSON data")
		}

		data = append(data, auditLogEntry)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *audit.AuditLogEntry) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetAuditLogEntriesForUserAndResourceTypes fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if len(resourceTypes) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(auditkeys.AuditLogEntryResourceTypesKey, resourceTypes)
	tracing.AttachToSpan(span, auditkeys.AuditLogEntryResourceTypesKey, resourceTypes)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetAuditLogEntriesForUserAndResourceType(ctx, q.readDB, &generated.GetAuditLogEntriesForUserAndResourceTypeParams{
		BelongsToUser: database.NullStringFromString(userID),
		Resources:     resourceTypes,
		CreatedBefore: database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:  database.NullTimeFromTimePointer(filter.CreatedAfter),
		Cursor:        database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:   database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries from database")
	}

	var (
		data                      []*audit.AuditLogEntry
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		auditLogEntry := &audit.AuditLogEntry{
			CreatedAt:        result.CreatedAt,
			BelongsToAccount: database.StringPointerFromNullString(result.BelongsToAccount),
			ID:               result.ID,
			ResourceType:     result.ResourceType,
			RelevantID:       result.RelevantID,
			EventType:        string(result.EventType),
			BelongsToUser:    database.StringFromNullString(result.BelongsToUser),
		}

		if err = json.Unmarshal(result.Changes, &auditLogEntry.Changes); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "parsing audit log entry JSON data")
		}

		data = append(data, auditLogEntry)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *audit.AuditLogEntry) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

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

	results, err := q.generatedQuerier.GetAuditLogEntriesForAccount(ctx, q.readDB, &generated.GetAuditLogEntriesForAccountParams{
		BelongsToAccount: database.NullStringFromString(accountID),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries from database")
	}

	var (
		data                      []*audit.AuditLogEntry
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		auditLogEntry := &audit.AuditLogEntry{
			CreatedAt:        result.CreatedAt,
			BelongsToAccount: database.StringPointerFromNullString(result.BelongsToAccount),
			ID:               result.ID,
			ResourceType:     result.ResourceType,
			RelevantID:       result.RelevantID,
			EventType:        string(result.EventType),
			BelongsToUser:    database.StringFromNullString(result.BelongsToUser),
		}

		if err = json.Unmarshal(result.Changes, &auditLogEntry.Changes); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "parsing audit log entry JSON data")
		}

		data = append(data, auditLogEntry)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *audit.AuditLogEntry) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetAuditLogEntriesForAccountAndResourceTypes fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if len(resourceTypes) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(auditkeys.AuditLogEntryResourceTypesKey, resourceTypes)
	tracing.AttachToSpan(span, auditkeys.AuditLogEntryResourceTypesKey, resourceTypes)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetAuditLogEntriesForAccountAndResourceType(ctx, q.readDB, &generated.GetAuditLogEntriesForAccountAndResourceTypeParams{
		BelongsToAccount: database.NullStringFromString(accountID),
		Resources:        resourceTypes,
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries from database")
	}

	var (
		data                      []*audit.AuditLogEntry
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		auditLogEntry := &audit.AuditLogEntry{
			CreatedAt:        result.CreatedAt,
			BelongsToAccount: database.StringPointerFromNullString(result.BelongsToAccount),
			ID:               result.ID,
			ResourceType:     result.ResourceType,
			RelevantID:       result.RelevantID,
			EventType:        string(result.EventType),
			BelongsToUser:    database.StringFromNullString(result.BelongsToUser),
		}

		if err = json.Unmarshal(result.Changes, &auditLogEntry.Changes); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "parsing audit log entry JSON data")
		}

		data = append(data, auditLogEntry)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *audit.AuditLogEntry) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateAuditLogEntry creates an audit log entry in a database.
func (q *repository) CreateAuditLogEntry(ctx context.Context, querier database.SQLQueryExecutor, input *audit.AuditLogEntryDatabaseCreationInput) (*audit.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(identitykeys.AccountIDKey, input.BelongsToAccount)

	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.BelongsToUser)
	logger = logger.WithValue(identitykeys.UserIDKey, input.BelongsToUser)

	marshaledChanges, err := json.Marshal(input.Changes)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "serializing audit log change list")
	}

	if err = q.generatedQuerier.CreateAuditLogEntry(ctx, querier, &generated.CreateAuditLogEntryParams{
		ID:               input.ID,
		ResourceType:     input.ResourceType,
		RelevantID:       input.RelevantID,
		EventType:        generated.AuditLogEventType(input.EventType),
		Changes:          marshaledChanges,
		BelongsToUser:    sql.NullString{String: input.BelongsToUser, Valid: strings.TrimSpace(input.BelongsToUser) != ""},
		BelongsToAccount: database.NullStringFromStringPointer(input.BelongsToAccount),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing audit log creation query")
	}

	x := &audit.AuditLogEntry{
		ID:               input.ID,
		Changes:          input.Changes,
		BelongsToAccount: input.BelongsToAccount,
		CreatedAt:        q.CurrentTime(),
		ResourceType:     input.ResourceType,
		RelevantID:       input.RelevantID,
		EventType:        input.EventType,
		BelongsToUser:    input.BelongsToUser,
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryIDKey, x.ID)

	return x, nil
}
