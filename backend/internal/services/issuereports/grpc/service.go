package grpc

import (
	"context"

	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/errors"

	comments "github.com/primandproper/platform-go/v14/comments"
	issuereports "github.com/primandproper/platform-go/v14/issuereports"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "issue_reports_service"
)

var _ issuereportssvc.IssueReportsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		issuereportssvc.UnimplementedIssueReportsServiceServer
		tracer       tracing.Tracer
		db           database.Client
		logger       logging.Logger
		issueReports issuereports.Store
		comments     comments.Store
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	db database.Client,
	issueReports issuereports.Store,
	commentStore comments.Store,
) issuereportssvc.IssueReportsServiceServer {
	return &serviceImpl{
		logger:       logging.NewNamedLogger(logger, o11yName),
		tracer:       tracing.NewNamedTracer(tracerProvider, o11yName),
		db:           db,
		issueReports: issueReports,
		comments:     commentStore,
	}
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
