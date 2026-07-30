package manager

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	o11yName = "issue_reports_data_manager"
)

var (
	_ issuereports.Repository = (*issueReportsManager)(nil)
	_ IssueReportsDataManager = (*issueReportsManager)(nil)
)

type issueReportsManager struct {
	tracer               tracing.Tracer
	logger               logging.Logger
	repo                 issuereports.Repository
	dataChangesPublisher messagequeue.Publisher
}

// NewIssueReportsDataManager returns a new IssueReportsDataManager that wraps the issue reports repository and emits data change events.
func NewIssueReportsDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo issuereports.Repository,
	cfg *msgconfig.QueuesConfig,
	publisherProvider messagequeue.PublisherProvider,
) (IssueReportsDataManager, error) {
	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide publisher for data changes topic: %w", err)
	}

	return &issueReportsManager{
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:               logging.NewNamedLogger(logger, o11yName),
		repo:                 repo,
		dataChangesPublisher: dataChangesPublisher,
	}, nil
}

func (m *issueReportsManager) GetIssueReport(ctx context.Context, issueReportID string) (*issuereports.IssueReport, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetIssueReport(ctx, issueReportID)
}

func (m *issueReportsManager) GetIssueReports(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetIssueReports(ctx, filter)
}

func (m *issueReportsManager) GetIssueReportsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetIssueReportsForAccount(ctx, accountID, filter)
}

func (m *issueReportsManager) GetIssueReportsForTable(ctx context.Context, tableName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetIssueReportsForTable(ctx, tableName, filter)
}

func (m *issueReportsManager) GetIssueReportsForRecord(ctx context.Context, tableName, recordID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetIssueReportsForRecord(ctx, tableName, recordID, filter)
}

func (m *issueReportsManager) CreateIssueReport(ctx context.Context, input *issuereports.IssueReportDatabaseCreationInput) (*issuereports.IssueReport, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	created, err := m.repo.CreateIssueReport(ctx, input)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, created.ID)
	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, issuereports.IssueReportCreatedServiceEventType, map[string]any{
		issuereportkeys.IssueReportIDKey: created.ID,
	}))

	return created, nil
}

func (m *issueReportsManager) UpdateIssueReport(ctx context.Context, issueReport *issuereports.IssueReport) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if issueReport == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, issueReport.ID)
	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, issueReport.ID)

	if err := m.repo.UpdateIssueReport(ctx, issueReport); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, issuereports.IssueReportUpdatedServiceEventType, map[string]any{
		issuereportkeys.IssueReportIDKey: issueReport.ID,
	}))

	return nil
}

func (m *issueReportsManager) ArchiveIssueReport(ctx context.Context, issueReportID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, issueReportID)
	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, issueReportID)

	if err := m.repo.ArchiveIssueReport(ctx, issueReportID); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, issuereports.IssueReportArchivedServiceEventType, map[string]any{
		issuereportkeys.IssueReportIDKey: issueReportID,
	}))

	return nil
}
