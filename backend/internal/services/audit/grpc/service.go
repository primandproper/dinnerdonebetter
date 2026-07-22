package grpc

import (
	"context"
	"errors"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	auditkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/keys"
	auditmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/manager"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
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

// errNotAuthorizedToViewAuditLogEntry is returned when a requester is not permitted to view a given audit log entry.
var errNotAuthorizedToViewAuditLogEntry = errors.New("not authorized to view audit log entry")

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

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	sessionContextData, err := sessions.FetchContextDataFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	accountID := sessionContextData.GetActiveAccountID()
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)

	auditLogEntries, err := s.auditManager.GetAuditLogEntriesForAccount(ctx, accountID, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get audit log entries for account")
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

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	sessionContextData, err := sessions.FetchContextDataFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	userID := sessionContextData.GetUserID()
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	auditLogEntries, err := s.auditManager.GetAuditLogEntriesForUser(ctx, userID, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get audit log entries for user")
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

	sessionContextData, err := sessions.FetchContextDataFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	auditLogEntry, err := s.auditManager.GetAuditLogEntry(ctx, request.AuditLogEntryId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get audit log entry")
	}

	// verify the requester is entitled to view this entry: it must belong to them or to their active account (service admins may view any).
	belongsToUser := auditLogEntry.BelongsToUser == sessionContextData.GetUserID()
	belongsToAccount := auditLogEntry.BelongsToAccount != nil && *auditLogEntry.BelongsToAccount == sessionContextData.GetActiveAccountID()
	if !belongsToUser && !belongsToAccount && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedToViewAuditLogEntry, logger, span, codes.PermissionDenied, "not authorized to view audit log entry")
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
