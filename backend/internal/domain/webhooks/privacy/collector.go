// Package privacy is the webhooks domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const o11yName = "webhooks_privacy_collector"

// Collector collects webhook data about a subject.
type Collector struct {
	repo            webhooks.Repository
	resolveAccounts dataprivacy.AccountIDResolver
	tracer          tracing.Tracer
	logger          logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the webhooks collector.
func NewCollector(
	repo webhooks.Repository,
	resolveAccounts dataprivacy.AccountIDResolver,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) *Collector {
	return &Collector{
		repo:            repo,
		resolveAccounts: resolveAccounts,
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:          logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
//
// Webhooks belong to an account rather than a person, so they are keyed by
// account in the fragment. A subject in three accounts gets three groups, which
// is the only rendering that lets them tell which endpoint belongs to which.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span)

	accountIDs, err := c.resolveAccounts(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "resolving accounts")
	}

	collection := webhooks.UserDataCollection{Data: map[string][]webhooks.Webhook{}}

	for _, accountID := range accountIDs {
		hooks, hooksErr := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
			return c.repo.GetWebhooks(ctx, accountID, filter)
		})
		if hooksErr != nil {
			return nil, observability.PrepareAndLogError(hooksErr, logger, span, "fetching webhooks for account")
		}

		if len(hooks) > 0 {
			collection.Data[accountID] = hooks
		}
	}

	return dataprivacy.Fragment(len(collection.Data) > 0, &collection)
}
