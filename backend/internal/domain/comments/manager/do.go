package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"

	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCommentsDataManager registers the comments data manager with the injector.
func RegisterCommentsDataManager(i do.Injector) {
	do.Provide[CommentsDataManager](i, func(i do.Injector) (CommentsDataManager, error) {
		return NewCommentsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[comments.Repository](i),
		)
	})
}
