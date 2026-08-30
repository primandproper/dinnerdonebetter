// Package privacy is the payments domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const o11yName = "payments_privacy_collector"

// Collector collects payment records about a subject.
type Collector struct {
	repo            payments.Repository
	resolveAccounts dataprivacy.AccountIDResolver
	tracer          tracing.Tracer
	logger          logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the payments collector.
func NewCollector(
	repo payments.Repository,
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
// Payments is the domain whose export and erasure disagree most sharply: a
// subject is entitled to see every transaction, and financial records are
// generally the ones that must be retained rather than deleted. That asymmetry
// is why platform-go registers collectors and erasers separately, and it is the
// reason this file has no counterpart under an eraser key.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span)

	accountIDs, err := c.resolveAccounts(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "resolving accounts")
	}

	subscriptions, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetSubscriptionsForAccount)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching subscriptions")
	}

	purchases, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetPurchasesForAccount)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching purchases")
	}

	transactions, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetPaymentTransactionsForAccount)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching payment transactions")
	}

	held := len(subscriptions) > 0 || len(purchases) > 0 || len(transactions) > 0

	return dataprivacy.Fragment(held, &payments.UserDataCollection{
		Subscriptions:       subscriptions,
		Purchases:           purchases,
		PaymentTransactions: transactions,
	})
}
