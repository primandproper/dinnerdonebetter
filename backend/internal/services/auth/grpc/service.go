package grpc

import (
	"context"

	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/qrcodes"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	o11yName = "auth_service"
)

var _ authsvc.AuthServiceServer = (*serviceImpl)(nil)

type (
	// SignIns is the part of signin.Service this service's two doors go through.
	SignIns interface {
		Register(ctx context.Context, scope tenancy.Scope, registration *signin.Registration) (*signin.Registered, error)
		ExchangeRefreshToken(ctx context.Context, scope tenancy.Scope, refreshToken string) (*signin.SignIn, error)
		IssueForPrincipal(ctx context.Context, scope tenancy.Scope, userID, activeAccountID string, opts ...signin.IssueOption) (*signin.SignIn, error)
		SignOut(ctx context.Context, scope tenancy.Scope, refreshToken string) error
	}

	serviceImpl struct {
		authsvc.UnimplementedAuthServiceServer
		tracer  tracing.Tracer
		logger  logging.Logger
		signIns SignIns
		qrCodes qrcodes.Builder
	}
)

var _ SignIns = (*signin.Service)(nil)

// NewAuthService builds the two doors platform's SignInService does not have yet. Both are
// signin.Service calls; what this adds is the wire each answers on.
func NewAuthService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	signIns SignIns,
	qrCodes qrcodes.Builder,
) authsvc.AuthServiceServer {
	return &serviceImpl{
		logger:  logging.NewNamedLogger(logger, o11yName),
		tracer:  tracing.NewNamedTracer(tracerProvider, o11yName),
		signIns: signIns,
		qrCodes: qrCodes,
	}
}
