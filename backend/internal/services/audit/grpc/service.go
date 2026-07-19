package grpc

import (
	"context"

	auditkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/keys"
	auditmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/manager"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	auditsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	grpctypes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/audit/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v5/errors/grpc"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "audit_service"
)

var _ auditsvc.AuditServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		auditsvc.UnimplementedAuditServiceServer
		tracer       tracing.Tracer
		logger       logging.Logger
		auditManager auditmanager.AuditDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditManager auditmanager.AuditDataManager,
) auditsvc.AuditServiceServer {
	return &serviceImpl{
		logger:       logging.NewNamedLogger(logger, o11yName),
		tracer:       tracing.NewNamedTracer(tracerProvider, o11yName),
		auditManager: auditManager,
	}
}

func (s *serviceImpl) GetAuditLogEntriesForAccount(ctx context.Context, request *auditsvc.GetAuditLogEntriesForAccountRequest) (*auditsvc.GetAuditLogEntriesForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("", "")
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	auditLogEntries, err := s.auditManager.GetAuditLogEntriesForAccount(ctx, request.AccountId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	x := &auditsvc.GetAuditLogEntriesForAccountResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(auditLogEntries.Pagination, filter),
		Results:    nil,
	}

	for _, y := range auditLogEntries.Data {
		x.Results = append(x.Results, converters.ConvertAuditLogEntryToGRPCAuditLogEntry(y))
	}

	return x, nil
}

func (s *serviceImpl) GetAuditLogEntriesForUser(ctx context.Context, request *auditsvc.GetAuditLogEntriesForUserRequest) (*auditsvc.GetAuditLogEntriesForUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("", "")
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	auditLogEntries, err := s.auditManager.GetAuditLogEntriesForUser(ctx, request.UserId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	x := &auditsvc.GetAuditLogEntriesForUserResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(auditLogEntries.Pagination, filter),
		Results:    nil,
	}

	for _, y := range auditLogEntries.Data {
		x.Results = append(x.Results, converters.ConvertAuditLogEntryToGRPCAuditLogEntry(y))
	}

	return x, nil
}

func (s *serviceImpl) GetAuditLogEntryByID(ctx context.Context, request *auditsvc.GetAuditLogEntryByIDRequest) (*auditsvc.GetAuditLogEntryByIDResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(auditkeys.AuditLogEntryIDKey, request.AuditLogEntryId)
	auditLogEntry, err := s.auditManager.GetAuditLogEntry(ctx, request.AuditLogEntryId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	returnValue := converters.ConvertAuditLogEntryToGRPCAuditLogEntry(auditLogEntry)

	x := &auditsvc.GetAuditLogEntryByIDResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: returnValue,
	}

	return x, nil
}
