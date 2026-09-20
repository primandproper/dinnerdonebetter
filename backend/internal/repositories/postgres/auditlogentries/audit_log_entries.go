package auditlogentries

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/keys"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
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

	entry, err := q.reader.Get(ctx, q.db.Reader(), nil, auditLogEntryID)
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
func (q *repository) GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID, resourceType string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	if resourceType == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryResourceTypesKey, resourceType)

	return q.list(ctx, span, &platformaudit.Query{ActorID: userID, ResourceType: resourceType}, filter,
		identitykeys.UserIDKey, userID)
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	// Scope is a pointer in the platform query because the global scope is a real
	// scope. Naming the account rather than passing nil is what keeps this from
	// reading every tenant's entries.
	scope := tenancy.Of(accountID)

	return q.list(ctx, span, &platformaudit.Query{Scope: &scope}, filter,
		identitykeys.AccountIDKey, accountID)
}

// GetAuditLogEntriesForAccountAndResourceTypes fetches a list of audit log entries from the database that meet a particular filter.
func (q *repository) GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID, resourceType string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	if resourceType == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryResourceTypesKey, resourceType)

	scope := tenancy.Of(accountID)

	return q.list(ctx, span, &platformaudit.Query{Scope: &scope, ResourceType: resourceType}, filter,
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

	results, err := q.reader.List(ctx, q.db.Reader(), query, filter)
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
func (q *repository) Record(ctx context.Context, querier database.Tx, entries ...*audit.AuditLogEntry) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(entries) == 0 {
		return nil
	}

	// The platform Recorder appends to one chain per call, and a chain is
	// identified by scope. A batch reaching this method may span scopes — an
	// account-scoped change and the user-scoped login that authorized it are
	// legitimately recorded together — so the batch is split by scope and each
	// group is appended to its own chain. Order within a scope is preserved,
	// which is the only order the chain defines.
	converted := make([]*platformaudit.Entry, len(entries))
	order := make([]tenancy.Scope, 0, len(entries))
	groups := map[tenancy.Scope][]*platformaudit.Entry{}
	for i, entry := range entries {
		if entry == nil {
			return observability.PrepareAndLogError(platformerrors.ErrNilInputParameter, logger, span, "recording audit log entries")
		}

		converted[i] = toPlatformEntry(entry)

		scope := converted[i].Scope
		if _, ok := groups[scope]; !ok {
			order = append(order, scope)
		}
		groups[scope] = append(groups[scope], converted[i])
	}

	for _, scope := range order {
		if err := q.recorder.Record(ctx, querier, scope, groups[scope]...); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "recording audit log entries")
		}
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
func (q *repository) VerifyChain(ctx context.Context, scope tenancy.Scope, from, to time.Time) (*audit.VerificationResult, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	// ChainStart walks the range from its beginning. Resuming from a previous
	// result's LastSeq is the platform's affordance for paging a long
	// verification; this application verifies a range in one call.
	result, err := q.reader.Verify(ctx, q.db.Reader(), scope, from, to, platformaudit.ChainStart)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, q.logger.Clone(), span, "verifying audit log chain")
	}

	return result, nil
}
