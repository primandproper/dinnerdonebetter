package webhooks

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/observability"
)

func (r *repository) CollectUserData(ctx context.Context, accountIDs []string) (*webhooks.UserDataCollection, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span)
	filter := filtering.DefaultQueryFilter()

	x := &webhooks.UserDataCollection{
		Data: make(map[string][]webhooks.Webhook),
	}

	for _, accountID := range accountIDs {
		accountWebhooks, err := r.generatedQuerier.GetWebhooksForAccount(ctx, r.readDB, &generated.GetWebhooksForAccountParams{
			CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
			CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
			UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
			UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
			Cursor:           database.NullStringFromStringPointer(filter.Cursor),
			ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
			IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
			BelongsToAccount: accountID,
		})
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "retrieving webhooks for account")
		}

		seen := make(map[string]struct{})
		for _, row := range accountWebhooks {
			if _, ok := seen[row.ID]; ok {
				continue
			}
			seen[row.ID] = struct{}{}
			x.Data[accountID] = append(x.Data[accountID], webhooks.Webhook{
				CreatedAt:        row.CreatedAt_2,
				ArchivedAt:       database.TimePointerFromNullTime(row.ArchivedAt_2),
				LastUpdatedAt:    database.TimePointerFromNullTime(row.LastUpdatedAt),
				Name:             row.Name,
				URL:              row.URL,
				Method:           string(row.Method),
				ID:               row.ID,
				BelongsToAccount: row.BelongsToAccount,
				CreatedByUser:    row.CreatedByUser,
				ContentType:      string(row.ContentType),
				TriggerConfigs:   nil,
			})
		}
	}

	return x, nil
}
