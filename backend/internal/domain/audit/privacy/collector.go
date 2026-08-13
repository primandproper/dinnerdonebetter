/*
Package privacy is the audit log's contribution to a subject access request.

The audit log is the section of an export most likely to be misread as an
oversight, so it is worth saying why it is here. An audit entry about a person is
personal data — it says what they did and when — and a subject access request
covers it like anything else. What a subject may not do is have it erased, which
is the asymmetry platform-go models by registering collectors and erasers
separately: this domain exports in full and erases only whole chains it can
remove without making the rest of the log unverifiable. See
internal/domain/audit/privacy/eraser.go for that half.
*/
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v10/dataprivacy"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

const o11yName = "audit_privacy_collector"

// Collector collects the audit entries recorded about a subject.
type Collector struct {
	repo   audit.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the audit log collector.
func NewCollector(repo audit.Repository, logger logging.Logger, tracerProvider tracing.Provider) *Collector {
	return &Collector{
		repo:   repo,
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	entries, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
		return c.repo.GetAuditLogEntriesForUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, c.logger.WithSpan(span), span, "fetching audit log entries")
	}

	return dataprivacy.Fragment(len(entries) > 0, entries)
}
