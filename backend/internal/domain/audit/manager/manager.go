package manager

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const (
	o11yName = "audit_data_manager"
)

var (
	_ audit.Repository = (*auditManager)(nil)
	_ AuditDataManager = (*auditManager)(nil)
)

type auditManager struct {
	tracer tracing.Tracer
	logger logging.Logger
	repo   audit.Repository
}

// NewAuditDataManager returns a new AuditDataManager that wraps the audit repository.
func NewAuditDataManager(
	tracerProvider tracing.Provider,
	logger logging.Logger,
	repo audit.Repository,
) AuditDataManager {
	return &auditManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
	}
}

func (m *auditManager) GetAuditLogEntry(ctx context.Context, auditLogID string) (*audit.AuditLogEntry, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetAuditLogEntry(ctx, auditLogID)
}

func (m *auditManager) GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetAuditLogEntriesForUser(ctx, userID, filter)
}

func (m *auditManager) GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetAuditLogEntriesForUserAndResourceTypes(ctx, userID, resourceTypes, filter)
}

func (m *auditManager) GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetAuditLogEntriesForAccount(ctx, accountID, filter)
}

func (m *auditManager) GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetAuditLogEntriesForAccountAndResourceTypes(ctx, accountID, resourceTypes, filter)
}

func (m *auditManager) Record(ctx context.Context, querier database.SQLQueryExecutor, entries ...*audit.AuditLogEntry) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)
	if len(entries) > 0 {
		logger = logger.WithValue(identitykeys.UserIDKey, entries[0].BelongsToUser)
	}

	if err := m.repo.Record(ctx, querier, entries...); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "recording audit log entries")
	}

	return nil
}

func (m *auditManager) VerifyChain(ctx context.Context, scope string, from, to time.Time) (*audit.VerificationResult, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.VerifyChain(ctx, scope, from, to)
}
