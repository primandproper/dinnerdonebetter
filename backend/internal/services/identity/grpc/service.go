package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/errors"

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"
)

const (
	o11yName = "identity_service"
)

var _ identitysvc.IdentityServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		identitysvc.UnimplementedIdentityServiceServer
		tracer              tracing.Tracer
		db                  database.Client
		logger              logging.Logger
		identityDataManager manager.IdentityDataManager
		uploadManager       uploads.UploadManager

		// registry holds the avatar rows, uploadManager holds the bytes. See
		// UploadUserAvatar for why the two stay separate.
		registry mediaregistry.Store
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	db database.Client,
	identityDataManager manager.IdentityDataManager,
	registryStore mediaregistry.Store,
	uploadManager uploads.UploadManager,
) identitysvc.IdentityServiceServer {
	return &serviceImpl{
		logger:              logging.NewNamedLogger(logger, o11yName),
		tracer:              tracing.NewNamedTracer(tracerProvider, o11yName),
		db:                  db,
		identityDataManager: identityDataManager,
		registry:            registryStore,
		uploadManager:       uploadManager,
	}
}

func (s *serviceImpl) buildResponseDetails(ctx context.Context, span tracing.Span) *types.ResponseDetails {
	out := &types.ResponseDetails{}
	if span != nil {
		out.TraceId = span.SpanContext().TraceID().String()
	}

	// Response details are built for unauthenticated routes too, so absence is expected here.
	out.CurrentAccountId = sessions.FromContext(ctx).GetActiveAccountID()

	return out
}

// inTransaction runs one store write on a transaction of its own.
//
// As of platform-go v14 a store holds no database handle: a write takes the
// caller's database.Tx. Every write in this service is a single store call, so
// each gets one transaction — which is exactly what the store opened for itself
// before the caller was required to supply it. A handler that ever writes twice
// should take one transaction across both rather than call this twice.
func inTransaction[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
