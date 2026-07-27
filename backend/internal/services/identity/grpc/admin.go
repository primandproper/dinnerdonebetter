package grpc

import (
	"context"
	"errors"

	identityconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	errorsgrpc "github.com/primandproper/platform-go/v7/errors/grpc"

	"google.golang.org/grpc/codes"
)

var errNotAuthorizedToUpdateUserStatus = errors.New("not authorized to update user account statuses")

func (s *serviceImpl) AdminSetPasswordChangeRequired(ctx context.Context, request *identitysvc.AdminSetPasswordChangeRequiredRequest) (*identitysvc.AdminSetPasswordChangeRequiredResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	if !sessionContextData.Requester.ServicePermissions.CanUpdateUserAccountStatuses() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedToUpdateUserStatus, logger, span, codes.PermissionDenied, "user does not have permission to set password change required")
	}

	if err = s.identityDataManager.AdminSetPasswordChangeRequired(ctx, request.GetTargetUserId(), request.GetRequiresPasswordChange()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "setting password change required")
	}

	return &identitysvc.AdminSetPasswordChangeRequiredResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}, nil
}

func (s *serviceImpl) AdminUpdateUserStatus(ctx context.Context, request *identitysvc.AdminUpdateUserStatusRequest) (*identitysvc.AdminUpdateUserStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	if !sessionContextData.Requester.ServicePermissions.CanUpdateUserAccountStatuses() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedToUpdateUserStatus, logger, span, codes.PermissionDenied, "user account status update requester does not have permission")
	}

	if err = s.identityDataManager.AdminUpdateUserStatus(ctx, identityconverters.ConvertGRPCAdminUpdateUserStatusRequestToUserAccountStatusUpdateInput(request)); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating user account status")
	}

	x := &identitysvc.AdminUpdateUserStatusResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}
