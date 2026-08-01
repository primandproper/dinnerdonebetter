package comments

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	domaincomments "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v8/database"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCommentsRepository registers the comments repository with the injector.
func RegisterCommentsRepository(i do.Injector) {
	do.Provide[domaincomments.Repository](i, func(i do.Injector) (domaincomments.Repository, error) {
		return ProvideCommentsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		), nil
	})
}
