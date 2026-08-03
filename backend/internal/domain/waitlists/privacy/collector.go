// Package privacy is the waitlists domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"

	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const o11yName = "waitlists_privacy_collector"

// Collector collects a subject's waitlist signups.
type Collector struct {
	repo   waitlists.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the waitlists collector.
func NewCollector(repo waitlists.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) *Collector {
	return &Collector{
		repo:   repo,
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
//
// The signups, not the waitlists. A waitlist is global and admin-managed, so it
// is not data held about the subject; the signup is theirs.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	signups, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
		return c.repo.GetWaitlistSignupsForUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, c.logger.WithSpan(span), span, "fetching waitlist signups")
	}

	return dataprivacy.Fragment(len(signups) > 0, signups)
}
