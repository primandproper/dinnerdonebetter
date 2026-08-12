package auditlogentries

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/keys"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformaudit "github.com/primandproper/platform-go/v10/audit"
	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	"github.com/primandproper/platform-go/v10/observability/tracing"
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

	entry, err := q.reader.Get(ctx, auditLogEntryID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entry")
	}

	return fromPlatformEntry(entry), nil
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	return q.list(ctx, span, &platformaudit.Query{ActorID: userID}, filter,
		identitykeys.UserIDKey, userID)
}

// GetAuditLogEntriesForUserAndResourceTypes fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	if len(resourceTypes) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryResourceTypesKey, resourceTypes)

	return q.list(ctx, span, &platformaudit.Query{ActorID: userID, ResourceTypes: resourceTypes}, filter,
		identitykeys.UserIDKey, userID)
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	// Scope is a pointer in the platform query because the empty string is a real
	// scope. Taking the address of the parameter rather than passing nil is what
	// keeps this from reading every tenant's entries.
	return q.list(ctx, span, &platformaudit.Query{Scope: &accountID}, filter,
		identitykeys.AccountIDKey, accountID)
}

// GetAuditLogEntriesForAccountAndResourceTypes fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	if len(resourceTypes) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryResourceTypesKey, resourceTypes)

	return q.list(ctx, span, &platformaudit.Query{Scope: &accountID, ResourceTypes: resourceTypes}, filter,
		identitykeys.AccountIDKey, accountID)
}

// list runs one platform query and converts the page it returns.
//
// The five read methods differ only in the query they build and the identifier
// they log, so everything after that lives here — including the conversion,
// which is the part that would otherwise be copied five times and drift.
func (q *repository) list(
	ctx context.Context,
	span tracing.Span,
	query *platformaudit.Query,
	filter *filtering.QueryFilter,
	logKey, logValue string,
) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	logger := q.logger.Clone().WithValue(logKey, logValue)
	tracing.AttachToSpan(span, logKey, logValue)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.reader.List(ctx, query, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries from database")
	}

	data := make([]*audit.AuditLogEntry, 0, len(results.Data))
	for _, result := range results.Data {
		data = append(data, fromPlatformEntry(result))
	}

	return filtering.NewQueryFilteredResult(
		data,
		results.FilteredCount,
		results.TotalCount,
		func(t *audit.AuditLogEntry) string {
			return t.ID
		},
		filter,
	), nil
}

// Record appends audit log entries inside the caller's transaction.
func (q *repository) Record(ctx context.Context, querier database.SQLQueryExecutor, entries ...*audit.AuditLogEntry) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(entries) == 0 {
		return nil
	}

	converted := make([]*platformaudit.Entry, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			return observability.PrepareAndLogError(platformerrors.ErrNilInputParameter, logger, span, "recording audit log entries")
		}

		converted = append(converted, toPlatformEntry(entry))
	}

	if err := q.recorder.Record(ctx, querier, converted...); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "recording audit log entries")
	}

	// Record assigns the ID, timestamp, and chain fields, and applies redaction to
	// the changes. Copying them back means a caller that logs or returns the entry
	// it just wrote describes the row that actually landed, rather than the value
	// it hoped to write.
	for i, entry := range entries {
		applyRecorded(entry, converted[i])
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryIDKey, entries[0].ID)

	return nil
}

// VerifyChain walks one scope's hash chain and reports the first break.
func (q *repository) VerifyChain(ctx context.Context, scope string, from, to time.Time) (*audit.VerificationResult, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	result, err := q.reader.Verify(ctx, scope, from, to)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, q.logger.Clone(), span, "verifying audit log chain")
	}

	return result, nil
}
