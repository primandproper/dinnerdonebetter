package grpc

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"
	"github.com/primandproper/platform-go/v9/observability"

	"google.golang.org/grpc/codes"
)

var (
	// errNotAuthorizedForAccount is returned when a requester is not a member of the requested account.
	errNotAuthorizedForAccount = errors.New("not authorized for account")
	// errAccountIDMismatch is returned when a request's account ID does not match the session's active account.
	errAccountIDMismatch = errors.New("request account ID does not match active account")
)

func (s *serviceImpl) ArchiveAccount(ctx context.Context, request *identitysvc.ArchiveAccountRequest) (*identitysvc.ArchiveAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: request.AccountId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.identityDataManager.ArchiveAccount(ctx, request.AccountId, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive account")
	}

	x := &identitysvc.ArchiveAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) CreateAccount(ctx context.Context, request *identitysvc.CreateAccountRequest) (*identitysvc.CreateAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	belongsToUser := sessionContextData.GetUserID()
	input := converters.ConvertGRPCAccountCreationRequestInputToAccountCreationRequestInput(request.Input, belongsToUser)

	created, err := s.identityDataManager.CreateAccount(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to create account")
	}

	x := &identitysvc.CreateAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Created:         converters.ConvertAccountToGRPCAccount(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateAccountInvitation(ctx context.Context, request *identitysvc.CreateAccountInvitationRequest) (*identitysvc.CreateAccountInvitationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	// TODO: more fields here, probably
	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCAccountInvitationCreationRequestInputToAccountInvitationCreationRequestInput(request.Input)
	created, err := s.identityDataManager.CreateAccountInvitation(ctx, sessionContextData.GetUserID(), sessionContextData.GetActiveAccountID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to create account invitation")
	}

	x := &identitysvc.CreateAccountInvitationResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Created:         converters.ConvertAccountInvitationToGRPCAccountInvitation(created),
	}

	return x, nil
}

func (s *serviceImpl) GetAccount(ctx context.Context, request *identitysvc.GetAccountRequest) (*identitysvc.GetAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: request.AccountId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	// verify the requester is a member of the requested account (service admins may read any account).
	if _, isMember := sessionContextData.AccountPermissions[request.AccountId]; !isMember && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForAccount, logger, span, codes.PermissionDenied, "not authorized for account")
	}

	account, err := s.identityDataManager.GetAccount(ctx, request.AccountId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get account")
	}

	x := &identitysvc.GetAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Result:          converters.ConvertAccountToGRPCAccount(account),
	}

	return x, nil
}

func (s *serviceImpl) GetAccounts(ctx context.Context, request *identitysvc.GetAccountsRequest) (*identitysvc.GetAccountsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	accounts, err := s.identityDataManager.GetAccounts(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to get accounts")
	}

	x := &identitysvc.GetAccountsResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(accounts.Pagination, filter),
	}

	for _, account := range accounts.Data {
		x.Results = append(x.Results, converters.ConvertAccountToGRPCAccount(account))
	}

	return x, nil
}

func (s *serviceImpl) GetAccountsForUser(ctx context.Context, request *identitysvc.GetAccountsForUserRequest) (*identitysvc.GetAccountsForUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	accounts, err := s.identityDataManager.GetAccounts(ctx, request.UserId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to get accounts")
	}

	x := &identitysvc.GetAccountsForUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(accounts.Pagination, filter),
	}

	for _, account := range accounts.Data {
		x.Results = append(x.Results, converters.ConvertAccountToGRPCAccount(account))
	}

	return x, nil
}

func (s *serviceImpl) SetDefaultAccount(ctx context.Context, request *identitysvc.SetDefaultAccountRequest) (*identitysvc.SetDefaultAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: request.AccountId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.identityDataManager.SetDefaultAccount(ctx, sessionContextData.GetUserID(), request.AccountId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to set default account")
	}

	x := &identitysvc.SetDefaultAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Success:         true,
	}

	return x, nil
}

func (s *serviceImpl) TransferAccountOwnership(ctx context.Context, request *identitysvc.TransferAccountOwnershipRequest) (*identitysvc.TransferAccountOwnershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: request.AccountId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if request.AccountId != "" && request.AccountId != sessionContextData.GetActiveAccountID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errAccountIDMismatch, logger, span, codes.InvalidArgument, "request account ID does not match active account")
	}

	input := converters.ConvertGRPCAccountOwnershipTransferInputToAccountOwnershipTransferInput(request.Input)

	if err = s.identityDataManager.TransferAccountOwnership(ctx, sessionContextData.GetActiveAccountID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to transfer account ownership")
	}

	x := &identitysvc.TransferAccountOwnershipResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Success:         true,
	}

	return x, nil
}

func (s *serviceImpl) UpdateAccount(ctx context.Context, request *identitysvc.UpdateAccountRequest) (*identitysvc.UpdateAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: request.AccountId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if request.AccountId != "" && request.AccountId != sessionContextData.GetActiveAccountID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errAccountIDMismatch, logger, span, codes.InvalidArgument, "request account ID does not match active account")
	}

	input := converters.ConvertGRPCAccountUpdateRequestInputToAccountUpdateRequestInput(request.Input)

	if err = s.identityDataManager.UpdateAccount(ctx, sessionContextData.GetActiveAccountID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update account")
	}

	x := &identitysvc.UpdateAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UpdateAccountMemberPermissions(ctx context.Context, request *identitysvc.UpdateAccountMemberPermissionsRequest) (*identitysvc.UpdateAccountMemberPermissionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: request.UserId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCModifyUserPermissionsInputToModifyUserPermissionsInput(request.Input)
	if err = s.identityDataManager.UpdateAccountMemberPermissions(ctx, sessionContextData.GetActiveAccountID(), request.UserId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update account member permissions")
	}

	x := &identitysvc.UpdateAccountMemberPermissionsResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveUserMembership(ctx context.Context, request *identitysvc.ArchiveUserMembershipRequest) (*identitysvc.ArchiveUserMembershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: request.UserId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	// remove the member from the caller's active account only; the request's account ID is not trusted.
	if err = s.identityDataManager.ArchiveUserMembership(ctx, request.UserId, sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive user membership")
	}

	x := &identitysvc.ArchiveUserMembershipResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}
