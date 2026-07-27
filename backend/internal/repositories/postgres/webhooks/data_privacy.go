package webhooks

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/observability"
)

func (r *repository) CollectUserData(ctx context.Context, accountIDs []string) (*webhooks.UserDataCollection, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span)

	x := &webhooks.UserDataCollection{
		Data: make(map[string][]webhooks.Webhook),
	}

	for _, accountID := range accountIDs {
		accountWebhooks, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
			return r.GetWebhooks(ctx, accountID, filter)
		})
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "retrieving webhooks for account")
		}

		for _, hook := range accountWebhooks {
			x.Data[accountID] = append(x.Data[accountID], *hook)
		}
	}

	return x, nil
}
