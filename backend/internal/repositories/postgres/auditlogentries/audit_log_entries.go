package auditlogentries

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/keys"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformaudit "github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
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
		// Translated to the sentinel every other repository here returns for a
		// missing row, so a caller distinguishing "not found" from "broken" does
		// not need to know which package owns this table. An entry can also be
		// absent because retention pruned it or because somebody removed it;
		// VerifyChain is what tells those apart.
		if errors.Is(err, platformaudit.ErrEntryNotFound) {
			return nil, sql.ErrNoRows
		}

		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entry")
	}

	return convertEntry(entry), nil
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

	return q.list(ctx, logger, span, &platformaudit.Query{ActorID: userID}, filter)
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

	return q.list(ctx, logger, span, &platformaudit.Query{ActorID: userID, ResourceTypes: resourceTypes}, filter)
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

	// Scope is a pointer in the platform's Query because the empty scope — the
	// one events belonging to no account land in — is a real chain, so a plain
	// string could not distinguish "only those" from "every account's". The
	// guard above is what keeps a caller from reaching it by passing nothing.
	return q.list(ctx, logger, span, &platformaudit.Query{Scope: &accountID}, filter)
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

	return q.list(ctx, logger, span, &platformaudit.Query{Scope: &accountID, ResourceTypes: resourceTypes}, filter)
}

// list runs a platform query and projects the page onto our read shape.
func (q *repository) list(
	ctx context.Context,
	logger logging.Logger,
	span tracing.Span,
	query *platformaudit.Query,
	filter *filtering.QueryFilter,
) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
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
		data = append(data, convertEntry(result))
	}

	// The pagination is carried over rather than recomputed: the cursor the
	// platform issued refers to its own ordering, and re-deriving one here would
	// hand the caller a token the next page does not understand.
	return &filtering.QueryFilteredResult[audit.AuditLogEntry]{
		Pagination: results.Pagination,
		Data:       data,
	}, nil
}

// Record appends entries to the audit log inside the caller's transaction.
func (q *repository) Record(ctx context.Context, querier database.SQLQueryExecutor, entries ...*audit.Entry) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(entries) == 0 {
		return nil
	}

	logger := q.logger.Clone().WithValue("audit.entry_count", len(entries))

	if err := q.recorder.Record(ctx, querier, entries...); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "recording audit log entries")
	}

	// Attached after the fact because Record is what assigns the IDs.
	tracing.AttachToSpan(span, auditkeys.AuditLogEntryIDKey, entries[0].ID)

	return nil
}

// VerifyChain walks one account's hash chain over a time range.
func (q *repository) VerifyChain(ctx context.Context, accountID string, from, to time.Time) (*audit.VerificationResult, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone().WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := q.reader.Verify(ctx, accountID, from, to)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "verifying audit log chain")
	}

	return result, nil
}

// convertEntry projects a stored entry onto the shape the API returns.
//
// The chain fields are dropped deliberately. A position and a pair of hashes
// mean nothing without the neighbors that give them meaning, and publishing
// them would invite a reader to believe they had checked something; VerifyChain
// is how the chain is asked a question.
func convertEntry(entry *audit.Entry) *audit.AuditLogEntry {
	out := &audit.AuditLogEntry{
		CreatedAt:     entry.RecordedAt,
		ID:            entry.ID,
		ResourceType:  entry.ResourceType,
		RelevantID:    entry.ResourceID,
		EventType:     string(entry.EventType),
		BelongsToUser: entry.Actor.ID,
	}

	if entry.Scope != "" {
		scope := entry.Scope
		out.BelongsToAccount = &scope
	}

	if len(entry.Changes) > 0 {
		out.Changes = make(map[string]*audit.ChangeLog, len(entry.Changes))
		for field := range entry.Changes {
			change := entry.Changes[field]
			out.Changes[field] = &audit.ChangeLog{
				OldValue: renderValue(change.Old),
				NewValue: renderValue(change.New),
			}
		}
	}

	return out
}

// renderValue turns a stored change value into the string the API speaks.
//
// The platform stores typed values, which is the better thing to store — a
// numeric field stays numeric through the round trip. Our gRPC surface has
// always carried strings, so the rendering happens here: strings pass through
// untouched, and anything else is rendered as its JSON encoding, which is both
// unambiguous and what the value was stored as anyway.
func renderValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}

		return string(encoded)
	}
}
