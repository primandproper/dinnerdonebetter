package payments

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeSubscriptions = "subscriptions"
)

func (r *repository) CreateSubscription(ctx context.Context, input *payments.SubscriptionDatabaseCreationInput) (*payments.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := r.logger.Clone()
	logger = logger.WithValue(paymentskeys.SubscriptionIDKey, input.ID)
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, input.ID)

	arg := &generated.CreateSubscriptionParams{
		ID:                     input.ID,
		BelongsToAccount:       input.BelongsToAccount,
		ProductID:              input.ProductID,
		ExternalSubscriptionID: database.NullStringFromString(input.ExternalSubscriptionID),
		Status:                 generated.SubscriptionStatus(input.Status),
		CurrentPeriodStart:     input.CurrentPeriodStart,
		CurrentPeriodEnd:       input.CurrentPeriodEnd,
	}

	if err := r.withEvent(ctx, logger, payments.SubscriptionCreatedServiceEventType, "", map[string]any{
		paymentskeys.SubscriptionIDKey: input.ID,
	}, func(tx database.Tx) error {
		if err := r.generatedQuerier.CreateSubscription(ctx, tx, arg); err != nil {
			return err
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &input.BelongsToAccount,
			ResourceType:     resourceTypeSubscriptions,
			RelevantID:       input.ID,
			EventType:        audit.AuditLogEventTypeCreated,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating subscription")
	}

	return r.GetSubscription(ctx, input.ID)
}

func (r *repository) GetSubscription(ctx context.Context, id string) (*payments.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if id == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(paymentskeys.SubscriptionIDKey, id)
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, id)

	result, err := r.generatedQuerier.GetSubscription(ctx, r.readDB, id)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching subscription")
	}

	return convertSubscriptionFromGenerated(result), nil
}

func (r *repository) GetSubscriptionByExternalID(ctx context.Context, externalID string) (*payments.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	logger = logger.WithValue("external_subscription_id", externalID)
	tracing.AttachToSpan(span, "external_subscription_id", externalID)

	arg := sql.NullString{String: externalID, Valid: externalID != ""}
	result, err := r.generatedQuerier.GetSubscriptionByExternalID(ctx, r.readDB, arg)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
	}

	return convertSubscriptionFromGenerated(result), nil
}

func (r *repository) GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.Clone()
	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	params := &generated.GetSubscriptionsForAccountParams{
		BelongsToAccount: accountID,
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		PageCursor:       database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
	}

	results, err := r.generatedQuerier.GetSubscriptionsForAccount(ctx, r.readDB, params)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching subscriptions for account")
	}

	return convertSubscriptionsResult(results, filter), nil
}

func (r *repository) UpdateSubscription(ctx context.Context, sub *payments.Subscription) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if sub == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := r.logger.Clone()
	logger = logger.WithValue(paymentskeys.SubscriptionIDKey, sub.ID)
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, sub.ID)

	arg := &generated.UpdateSubscriptionParams{
		ID:                     sub.ID,
		ExternalSubscriptionID: database.NullStringFromString(sub.ExternalSubscriptionID),
		Status:                 generated.SubscriptionStatus(sub.Status),
		CurrentPeriodStart:     sub.CurrentPeriodStart,
		CurrentPeriodEnd:       sub.CurrentPeriodEnd,
	}

	return r.withEvent(ctx, logger, payments.SubscriptionUpdatedServiceEventType, "", map[string]any{
		paymentskeys.SubscriptionIDKey: sub.ID,
	}, func(tx database.Tx) error {
		_, err := r.generatedQuerier.UpdateSubscription(ctx, tx, arg)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &sub.BelongsToAccount,
			ResourceType:     resourceTypeSubscriptions,
			RelevantID:       sub.ID,
			EventType:        audit.AuditLogEventTypeUpdated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	})
}

func (r *repository) UpdateSubscriptionStatus(ctx context.Context, id, status string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(paymentskeys.SubscriptionIDKey, id)
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, id)

	arg := &generated.UpdateSubscriptionStatusParams{
		ID:     id,
		Status: generated.SubscriptionStatus(status),
	}

	if err := r.WithTransaction(ctx, func(tx database.Tx) error {
		if _, updateErr := r.generatedQuerier.UpdateSubscriptionStatus(ctx, tx, arg); updateErr != nil {
			return updateErr
		}

		// This path has no account ID in scope, so the entry chains under the empty
		// scope rather than the subscription's account.
		return r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeSubscriptions,
			RelevantID:   id,
			EventType:    audit.AuditLogEventTypeUpdated,
		})
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
	}

	return nil
}

func (r *repository) ArchiveSubscription(ctx context.Context, id string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(paymentskeys.SubscriptionIDKey, id)
	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, id)

	return r.withEvent(ctx, logger, payments.SubscriptionArchivedServiceEventType, "", map[string]any{
		paymentskeys.SubscriptionIDKey: id,
	}, func(tx database.Tx) error {
		_, err := r.generatedQuerier.ArchiveSubscription(ctx, tx, id)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving subscription")
		}

		// ArchiveSubscription does not have account ID; create audit entry without it
		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeSubscriptions,
			RelevantID:   id,
			EventType:    audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	})
}

func convertSubscriptionFromGenerated(m *generated.Subscriptions) *payments.Subscription {
	if m == nil {
		return nil
	}
	return &payments.Subscription{
		ID:                     m.ID,
		BelongsToAccount:       m.BelongsToAccount,
		ProductID:              m.ProductID,
		ExternalSubscriptionID: database.StringFromNullString(m.ExternalSubscriptionID),
		Status:                 string(m.Status),
		CurrentPeriodStart:     m.CurrentPeriodStart,
		CurrentPeriodEnd:       m.CurrentPeriodEnd,
		CreatedAt:              m.CreatedAt,
		LastUpdatedAt:          database.TimePointerFromNullTime(m.LastUpdatedAt),
		ArchivedAt:             database.TimePointerFromNullTime(m.ArchivedAt),
	}
}

func convertSubscriptionsResult(rows []*generated.GetSubscriptionsForAccountRow, filter *filtering.QueryFilter) *filtering.QueryFilteredResult[payments.Subscription] {
	data := make([]*payments.Subscription, 0, len(rows))
	var filteredCount, totalCount uint64
	for _, row := range rows {
		data = append(data, &payments.Subscription{
			ID:                     row.ID,
			BelongsToAccount:       row.BelongsToAccount,
			ProductID:              row.ProductID,
			ExternalSubscriptionID: database.StringFromNullString(row.ExternalSubscriptionID),
			Status:                 string(row.Status),
			CurrentPeriodStart:     row.CurrentPeriodStart,
			CurrentPeriodEnd:       row.CurrentPeriodEnd,
			CreatedAt:              row.CreatedAt,
			LastUpdatedAt:          database.TimePointerFromNullTime(row.LastUpdatedAt),
			ArchivedAt:             database.TimePointerFromNullTime(row.ArchivedAt),
		})
		filteredCount = uint64(row.FilteredCount)
		totalCount = uint64(row.TotalCount)
	}
	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(p *payments.Subscription) string { return p.ID },
		filter,
	)
}
