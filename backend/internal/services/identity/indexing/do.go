package indexing

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCoreDataIndexer registers the core data indexer with the injector.
func RegisterCoreDataIndexer(i do.Injector) {
	do.Provide[*UserDataIndexer](i, func(i do.Injector) (*UserDataIndexer, error) {
		return NewCoreDataIndexer(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[UserTextSearcher](i),
		), nil
	})
}
