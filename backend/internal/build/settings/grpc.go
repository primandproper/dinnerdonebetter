/*
Package settings mounts platform-go's settings surface.

There is no service of this application's own any more. The one that used to
live in internal/services/settings forwarded thirteen RPCs to the store,
converted between two spellings of the same fields, and enforced AdminOnly —
and platform-go v14 ships all three. The last of those arrived late: until
platform asked PermissionWriteAdminValues inside the handler, mounting this
surface would have let any member write a setting the catalog had reserved. See
internal/repositories/postgres/settingsspike for the regression test.

What this application still owns is the store it is given: the repository in
internal/repositories/postgres/settings, which is platform's SQL store with an
audit entry and a data change event wrapped around every write.
*/
package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v14/callers"
	platformsettings "github.com/primandproper/platform-go/v14/settings"
	settingsgrpc "github.com/primandproper/platform-go/v14/settings/grpc"
	"github.com/primandproper/platform-go/v14/settings/settingspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// selfServiceOnly is this deployment's answer to whose settings a caller may
// reach: their own, and nobody else's.
//
// It is the closure platform's own documentation gives for a deployment with one
// subject type, and this is that deployment — see internal/domain/settings for
// why there is only SubjectUser.
//
// It is not where AdminOnly is decided. Whose settings these are and whether
// this setting is one an ordinary member may answer are different questions; the
// second is the PermissionWriteAdminValues grant, asked by the handler against
// the definition it read.
func selfServiceOnly() settingsgrpc.SubjectAuthorizer {
	return settingsgrpc.SubjectAuthorizerFunc(
		func(_ context.Context, caller callers.Principal, subject platformsettings.Subject) error {
			if subject.Type == platformsettings.SubjectUser && subject.ID == caller.UserID() {
				return nil
			}

			return callers.ErrTargetNotPermitted
		})
}

// RegisterSettingsService registers platform's settings surface with the injector.
func RegisterSettingsService(i do.Injector) {
	do.Provide[settingspb.SettingsServiceServer](i, func(i do.Injector) (settingspb.SettingsServiceServer, error) {
		return settingsgrpc.NewServer(
			do.MustInvoke[platformsettings.Store](i),
			do.MustInvoke[database.Client](i),
			sessions.PrincipalFromContext,
			selfServiceOnly(),
			// The grants extractor is what carries PermissionWriteAdminValues, so a
			// server built without one refuses every reserved setting. Safe, and not
			// what this deployment wants: an administrator has to be able to write
			// them.
			settingsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			settingsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			settingsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			settingsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
