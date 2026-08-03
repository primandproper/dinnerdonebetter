package config

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/reportartifacts"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterReportArtifactStore registers the disclosure artifact store with the injector.
//
// Every process that touches artifacts calls this, so the bucket and the cipher are chosen in
// one place rather than three. Prerequisite: *Config must be registered in the injector.
func RegisterReportArtifactStore(i do.Injector) {
	do.Provide[reportartifacts.Store](i, func(i do.Injector) (reportartifacts.Store, error) {
		cfg := do.MustInvoke[*Config](i)

		return reportartifacts.ProvideStore(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			&cfg.Uploads.Storage,
			&cfg.Encryption,
			cfg.ArtifactEncryptionKey,
		)
	})
}
