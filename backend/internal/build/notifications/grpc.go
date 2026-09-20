/*
Package notifications mounts platform-go's notifications surface.

The service this replaces forwarded seven RPCs and converted between two
spellings of the same fields. Platform's nine are the same reads and writes with
the inbox and the device registry separated, plus the two this application had
no equivalent for: an unread page, and marking a whole inbox read.

The scope is global. A notification is addressed to a person rather than to an
account — see internal/domain/notifications.Scope — so the principal extractor is
the global one.
*/
package notifications

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	notificationsgrpc "github.com/primandproper/platform-go/v14/notifications/grpc"
	"github.com/primandproper/platform-go/v14/notifications/notificationspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsService registers platform's notifications surface.
func RegisterNotificationsService(i do.Injector) {
	do.Provide[notificationspb.NotificationsServiceServer](i, func(i do.Injector) (notificationspb.NotificationsServiceServer, error) {
		return notificationsgrpc.NewServer(
			do.MustInvoke[platformnotifications.Inbox](i),
			do.MustInvoke[platformnotifications.Registry](i),
			do.MustInvoke[database.Client](i),
			sessions.PrincipalFromContext,
			// The grants extractor gates include_archived, which this surface gained
			// with the other two in the include_archived fix. Its absence is
			// fail-closed: a dismissed notification would be unreachable by anybody.
			notificationsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			notificationsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			notificationsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			notificationsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
