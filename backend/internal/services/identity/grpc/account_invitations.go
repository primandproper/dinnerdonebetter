package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v11/errors/grpc"
	"github.com/primandproper/platform-go/v11/observability"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) AcceptAccountInvitation(ctx context.Context, request *identitysvc.AcceptAccountInvitationRequest) (*identitysvc.AcceptAccountInvitationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountInvitationIDKey: request.AccountInvitationId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCAccountInvitationUpdateRequestInputToAccountInvitationUpdateRequestInput(request.Input)
	if err = s.identityDataManager.AcceptAccountInvitation(ctx, sessionContextData.GetActiveAccountID(), request.AccountInvitationId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "accepting account invitation")
	}

	x := &identitysvc.AcceptAccountInvitationResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) RejectAccountInvitation(ctx context.Context, request *identitysvc.RejectAccountInvitationRequest) (*identitysvc.RejectAccountInvitationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountInvitationIDKey: request.AccountInvitationId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCAccountInvitationUpdateRequestInputToAccountInvitationUpdateRequestInput(request.Input)
	if err = s.identityDataManager.RejectAccountInvitation(ctx, sessionContextData.GetActiveAccountID(), request.AccountInvitationId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to reject account invitation")
	}

	x := &identitysvc.RejectAccountInvitationResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) CancelAccountInvitation(ctx context.Context, request *identitysvc.CancelAccountInvitationRequest) (*identitysvc.CancelAccountInvitationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountInvitationIDKey: request.AccountInvitationId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.identityDataManager.CancelAccountInvitation(ctx, sessionContextData.GetActiveAccountID(), request.AccountInvitationId, request.GetInput().GetNote()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to cancel account invitation")
	}

	x := &identitysvc.CancelAccountInvitationResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) GetAccountInvitation(ctx context.Context, request *identitysvc.GetAccountInvitationRequest) (*identitysvc.GetAccountInvitationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	accountInvitation, err := s.identityDataManager.GetAccountInvitation(ctx, sessionContextData.GetActiveAccountID(), request.AccountInvitationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to get account invitation")
	}

	x := &identitysvc.GetAccountInvitationResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Result:          converters.ConvertAccountInvitationToGRPCAccountInvitation(accountInvitation),
	}

	return x, nil
}

func (s *serviceImpl) GetReceivedAccountInvitations(ctx context.Context, request *identitysvc.GetReceivedAccountInvitationsRequest) (*identitysvc.GetReceivedAccountInvitationsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	logger := filter.AttachToLogger(s.logger.WithSpan(span))

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	invites, err := s.identityDataManager.GetReceivedAccountInvitations(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get received account invitations")
	}

	x := &identitysvc.GetReceivedAccountInvitationsResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(invites.Pagination, filter),
	}

	for _, invite := range invites.Data {
		x.Results = append(x.Results, converters.ConvertAccountInvitationToGRPCAccountInvitation(invite))
	}

	return x, nil
}

func (s *serviceImpl) GetSentAccountInvitations(ctx context.Context, request *identitysvc.GetSentAccountInvitationsRequest) (*identitysvc.GetSentAccountInvitationsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	logger := filter.AttachToLogger(s.logger.WithSpan(span))

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	invites, err := s.identityDataManager.GetSentAccountInvitations(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get received account invitations")
	}

	x := &identitysvc.GetSentAccountInvitationsResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(invites.Pagination, filter),
	}

	for _, invite := range invites.Data {
		x.Results = append(x.Results, converters.ConvertAccountInvitationToGRPCAccountInvitation(invite))
	}

	return x, nil
}
