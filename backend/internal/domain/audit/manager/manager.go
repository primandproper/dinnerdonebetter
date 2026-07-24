package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v6/database"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"
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
	tracerProvider tracing.TracerProvider,
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

func (m *auditManager) CreateAuditLogEntry(ctx context.Context, querier database.SQLQueryExecutor, input *audit.AuditLogEntryDatabaseCreationInput) (*audit.AuditLogEntry, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, input.BelongsToUser)

	created, err := m.repo.CreateAuditLogEntry(ctx, querier, input)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
	}

	return created, nil
}
