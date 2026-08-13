package metering

import (
	"context"

	"github.com/primandproper/platform-go/v10/errors"
	platformmetering "github.com/primandproper/platform-go/v10/metering"
)

// NewProviderMapper returns the ProviderMapper the flusher runs with until this service bills
// for usage: one that reports no provider reference for any account or meter.
//
// The flusher reads that as "nothing to post" rather than as a failure, and settles each total
// instead of re-claiming it every interval forever. Two consequences worth being explicit about:
//
// The flush pass still does real work. It settles the backlog, keeps the backlog gauge honest,
// and reaps event-ledger rows past their retention, none of which depend on a provider.
//
// Usage settled this way is not retro-billable. The durable quantity is untouched — the totals
// this exists to accumulate stay complete, and the dashboards over them see every byte — but
// flushed_quantity advances with it, so a real mapper wired in later starts from the totals as
// they stand rather than replaying the months before it existed. That is the intended shape of
// "count usage now, bill later"; it is only a surprise if nobody said it.
//
// A real one needs two things this service has only half of: the provider-side customer, which
// is identity.Account.PaymentProcessorCustomerID, and the provider-side meter name, which is
// whoever owns pricing naming a billing meter at the provider. The second does not exist yet,
// and inventing a name here would post usage against a meter the provider has never heard of.
func NewProviderMapper() platformmetering.ProviderMapper {
	return platformmetering.ProviderMapperFunc(
		func(_ context.Context, subject, meter string) (platformmetering.ProviderRef, error) {
			return platformmetering.ProviderRef{}, errors.Wrapf(
				platformmetering.ErrNoProviderRef,
				"usage billing is not configured: account %q meter %q",
				subject, meter,
			)
		},
	)
}
