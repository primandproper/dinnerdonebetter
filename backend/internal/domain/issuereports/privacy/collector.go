// Package privacy is the issue reports domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const o11yName = "issue_reports_privacy_collector"

// Collector collects issue reports filed from a subject's accounts.
type Collector struct {
	repo            issuereports.Repository
	resolveAccounts dataprivacy.AccountIDResolver
	tracer          tracing.Tracer
	logger          logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the issue reports collector.
func NewCollector(
	repo issuereports.Repository,
	resolveAccounts dataprivacy.AccountIDResolver,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *Collector {
	return &Collector{
		repo:            repo,
		resolveAccounts: resolveAccounts,
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:          logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span)

	accountIDs, err := c.resolveAccounts(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "resolving accounts")
	}

	reports, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetIssueReportsForAccount)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching issue reports")
	}

	return dataprivacy.Fragment(len(reports) > 0, reports)
}
