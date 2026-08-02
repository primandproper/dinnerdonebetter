package payments

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	resourceTypePaymentTransactions = "payment_transactions"
)

func (r *repository) CreatePaymentTransaction(ctx context.Context, input *payments.PaymentTransactionDatabaseCreationInput) (*payments.PaymentTransaction, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := r.logger.Clone()
	logger = logger.WithValue(paymentskeys.PaymentTransactionIDKey, input.ID)
	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, input.ID)

	arg := &generated.CreatePaymentTransactionParams{
		ID:                    input.ID,
		BelongsToAccount:      input.BelongsToAccount,
		SubscriptionID:        database.NullStringFromStringPointer(input.SubscriptionID),
		PurchaseID:            database.NullStringFromStringPointer(input.PurchaseID),
		ExternalTransactionID: input.ExternalTransactionID,
		AmountCents:           input.AmountCents,
		Currency:              input.Currency,
		Status:                generated.PaymentTransactionStatus(input.Status),
	}

	if err := r.generatedQuerier.CreatePaymentTransaction(ctx, r.writeDB, arg); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating payment transaction")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, r.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &input.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypePaymentTransactions,
		RelevantID:       input.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	return &payments.PaymentTransaction{
		ID:                    input.ID,
		BelongsToAccount:      input.BelongsToAccount,
		SubscriptionID:        input.SubscriptionID,
		PurchaseID:            input.PurchaseID,
		ExternalTransactionID: input.ExternalTransactionID,
		AmountCents:           input.AmountCents,
		Currency:              input.Currency,
		Status:                input.Status,
		CreatedAt:             time.Now().UTC(),
	}, nil
}

func (r *repository) GetPaymentTransactionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.PaymentTransaction], error) {
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

	params := &generated.GetPaymentTransactionsForAccountParams{
		BelongsToAccount: accountID,
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	}

	results, err := r.generatedQuerier.GetPaymentTransactionsForAccount(ctx, r.readDB, params)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching payment transactions for account")
	}

	return convertPaymentTransactionsResult(results, filter), nil
}

func convertPaymentTransactionsResult(rows []*generated.GetPaymentTransactionsForAccountRow, filter *filtering.QueryFilter) *filtering.QueryFilteredResult[payments.PaymentTransaction] {
	data := make([]*payments.PaymentTransaction, 0, len(rows))
	var filteredCount, totalCount uint64
	for _, row := range rows {
		data = append(data, &payments.PaymentTransaction{
			ID:                    row.ID,
			BelongsToAccount:      row.BelongsToAccount,
			SubscriptionID:        database.StringPointerFromNullString(row.SubscriptionID),
			PurchaseID:            database.StringPointerFromNullString(row.PurchaseID),
			ExternalTransactionID: row.ExternalTransactionID,
			AmountCents:           row.AmountCents,
			Currency:              row.Currency,
			Status:                string(row.Status),
			CreatedAt:             row.CreatedAt,
		})
		filteredCount = uint64(row.FilteredCount)
		totalCount = uint64(row.TotalCount)
	}
	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(p *payments.PaymentTransaction) string { return p.ID },
		filter,
	)
}
