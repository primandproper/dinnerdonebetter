package disclosureartifactreaper

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/reportartifacts"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterDisclosureArtifactReaper registers the disclosure artifact reaper with the injector.
func RegisterDisclosureArtifactReaper(i do.Injector) {
	do.Provide[*Worker](i, func(i do.Injector) (*Worker, error) {
		return NewDisclosureArtifactReaper(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[dataprivacy.UserDataDisclosureDataManager](i),
			do.MustInvoke[reportartifacts.Store](i),
		)
	})
}
