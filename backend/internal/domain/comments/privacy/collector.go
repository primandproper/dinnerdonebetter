// Package privacy is the comments domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v10/dataprivacy"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

const o11yName = "comments_privacy_collector"

// Collector collects the comments a subject authored.
type Collector struct {
	repo   comments.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the comments collector.
func NewCollector(repo comments.Repository, logger logging.Logger, tracerProvider tracing.Provider) *Collector {
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

	authored, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
		return c.repo.GetCommentsForUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, c.logger.WithSpan(span), span, "fetching comments")
	}

	return dataprivacy.Fragment(len(authored) > 0, authored)
}
