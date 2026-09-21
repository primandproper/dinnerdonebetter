package push

import (
	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	"github.com/primandproper/platform-go/v14/notifications/push"
	mobile "github.com/primandproper/primitives-go/v2/notifications/mobile"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterFanout registers platform's push fanout with the injector.
//
// Prerequisites: the notifications registry, a push sender, and the observability trio.
// Every process that pushes has all of them already — the fanout is the part they were each
// about to write.
//
// What this package used to hold was that part, written here: read the recipients' device
// tokens, send to each, archive the ones the provider rejects. platform's does the same
// three things and does the first of them in one query, where this read once per recipient
// and took only the first page of each — a user with more handsets than a default page
// silently missed the rest. It is the same shape as the account roster this port already
// fixed once, and the reason to take somebody else's is that theirs does not have it.
func RegisterFanout(i do.Injector) {
	do.Provide[*push.Fanout](i, func(i do.Injector) (*push.Fanout, error) {
		return push.NewFanout(
			do.MustInvoke[platformnotifications.Registry](i),
			do.MustInvoke[mobile.PushNotificationSender](i),
			push.WithLogger(do.MustInvoke[logging.Logger](i)),
			push.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			push.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
