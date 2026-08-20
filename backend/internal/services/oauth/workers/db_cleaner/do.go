package dbcleaner

import (
	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"

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
		)
	})
}
