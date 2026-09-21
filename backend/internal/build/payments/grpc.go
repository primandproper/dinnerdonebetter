/*
Package payments mounts platform-go's billing surface.

There is no service of this application's own any more, and two of its RPCs are
gone rather than moved. CreateSubscription and UpdateSubscription had no
counterpart on platform's surface and were not supposed to: a subscription is
what a payment provider reports, and billing/grpc's writes are the archives and
the product catalog. This application already worked that way in practice — the
RevenueCat webhook is what opens and moves a subscription, through
internal/domain/payments/manager — so the two RPCs were a second door onto rows
the provider owns. The repository methods behind them stay, because that is the
door the provider comes through.

The scope is global rather than per-account: this application keeps one billing
ledger and distinguishes accounts by the belongs_to_account column on the rows,
not by the tenant they are filed under. Which account a caller may see is
therefore the AccountAuthorizer's question rather than the scope's — the same
split platform's billing/privacy draws when it takes both a scope and an account
id and says neither is inferable from the other.
*/
package payments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/platform-go/v14/billing/billingpb"
	billinggrpc "github.com/primandproper/platform-go/v14/billing/grpc"
	"github.com/primandproper/platform-go/v14/callers"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// ownAccountOrAdmin is this deployment's rule: a caller sees the billing of the
// account they have active, and a service administrator sees any.
//
// It is the whole of the tenancy check on this surface, because the scope is not
// carrying it — see the package comment.
func ownAccountOrAdmin() billinggrpc.AccountAuthorizer {
	return billinggrpc.AccountAuthorizerFunc(
		func(ctx context.Context, caller callers.Principal, accountID string) error {
			if caller != nil && accountID != "" && caller.ActiveAccountID() == accountID {
				return nil
			}

			if sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin() {
				return nil
			}

			return callers.ErrTargetNotPermitted
		})
}

// RegisterPaymentsService registers platform's billing surface with the injector.
func RegisterPaymentsService(i do.Injector) {
	do.Provide[billingpb.BillingServiceServer](i, func(i do.Injector) (billingpb.BillingServiceServer, error) {
		return billinggrpc.NewServer(
			do.MustInvoke[billing.Store](i),
			do.MustInvoke[database.Client](i),
			sessions.PrincipalFromContext,
			ownAccountOrAdmin(),
			// Gates include_archived, which this surface gained with notifications
			// and webhooks. Its absence is fail-closed, and a withdrawn product or a
			// cancelled subscription is exactly what an operator needs to page.
			billinggrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			billinggrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			billinggrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			billinggrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
