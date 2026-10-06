package comments

import (
	platformcomments "github.com/primandproper/platform-go/v15/comments"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCommentsRepository registers the comments store with the injector.
func RegisterCommentsRepository(i do.Injector) {
	do.Provide[platformcomments.Store](i, func(i do.Injector) (platformcomments.Store, error) {
		return ProvideCommentsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*platformrecording.Recorder](i),
			do.MustInvoke[platformcomments.Targets](i),
		)
	})
}
