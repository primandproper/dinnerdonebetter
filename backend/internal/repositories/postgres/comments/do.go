package comments

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCommentsRepository registers the comments store with the injector.
//
// It invokes the target catalog rather than building one, so a process that
// registered no catalog fails at startup rather than at the first write. See
// internal/build/comments.
func RegisterCommentsRepository(i do.Injector) {
	do.Provide[platformcomments.Store](i, func(i do.Injector) (platformcomments.Store, error) {
		return ProvideCommentsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
			do.MustInvoke[platformcomments.Targets](i),
		)
	})
}
