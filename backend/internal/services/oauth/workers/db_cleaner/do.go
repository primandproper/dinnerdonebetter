package dbcleaner

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/platform-go/v14/authentication/signin/refreshtokens"
	sessionsdatabase "github.com/primandproper/platform-go/v14/sessions/database"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterDBCleaner registers the DB cleaner with the injector.
func RegisterDBCleaner(i do.Injector) {
	do.Provide[*Job](i, func(i do.Injector) (*Job, error) {
		return NewDBCleaner(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[oauth2server.Store](i),
			do.MustInvoke[*passwordreset.SQLStore](i),
			do.MustInvoke[*refreshtokens.SQLStore](i),
			do.MustInvoke[*sessionsdatabase.Backend[auth.SessionPayload]](i),
		)
	})
}
