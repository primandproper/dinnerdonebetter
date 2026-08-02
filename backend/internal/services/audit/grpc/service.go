package grpc

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	auditkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/keys"
	auditmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/manager"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	auditsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	grpctypes "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/audit/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

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

	// verify the requester is a member of the requested account (service admins may read any account).
	// This mirrors the GetAccount handler: reads are scoped to the requested account ID, but only for
	// members of that account, so there is no cross-account leak to non-members.
	if _, isMember := sessionContextData.AccountPermissions[request.AccountId]; !isMember && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedToViewAuditLogEntry, logger, span, codes.PermissionDenied, "not authorized to view audit log entries for account")
	}

	accountID := request.AccountId
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

	// a requester may only read their own user audit log; service admins may read any user's.
	// This mirrors the account handler: the query is scoped to the requested user ID, but non-admins
	// are limited to themselves, so there is no cross-user leak.
	if request.UserId != sessionContextData.GetUserID() && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedToViewAuditLogEntry, logger, span, codes.PermissionDenied, "not authorized to view audit log entries for user")
	}

	userID := request.UserId
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
