package grpc

import (
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterCommentsService registers the comments gRPC service with the injector.
func RegisterCommentsService(i do.Injector) {
	do.Provide[CommentsMethodPermissions](i, func(i do.Injector) (CommentsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[commentssvc.CommentsServiceServer](i, func(i do.Injector) (commentssvc.CommentsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[commentsmanager.CommentsDataManager](i),
		), nil
	})
}
