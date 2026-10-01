package grpc

import (
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/qrcodes"

	"github.com/samber/do/v2"
)

// RegisterAuthService registers the auth gRPC service with the injector.
func RegisterAuthService(i do.Injector) {
	do.Provide[authsvc.AuthServiceServer](i, func(i do.Injector) (authsvc.AuthServiceServer, error) {
		return NewAuthService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*signin.Service](i),
			do.MustInvoke[qrcodes.Builder](i),
		), nil
	})
}
