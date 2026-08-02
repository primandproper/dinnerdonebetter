package grpc

import (
	"context"
	"errors"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	waitlistkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"

	"google.golang.org/grpc/codes"
)

// errNotAuthorizedForWaitlistSignup is returned when a caller targets a signup that isn't theirs.
var errNotAuthorizedForWaitlistSignup = errors.New("not authorized to access waitlist signup")

func (s *serviceImpl) CreateWaitlist(ctx context.Context, request *waitlistssvc.CreateWaitlistRequest) (*waitlistssvc.CreateWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	input := converters.ConvertGRPCWaitlistCreationRequestInputToWaitlistDatabaseCreationInput(request.Input)
	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate waitlist creation request")
	}

	created, err := s.waitlistsManager.CreateWaitlist(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create waitlist")
	}

	x := &waitlistssvc.CreateWaitlistResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertWaitlistToGRPCWaitlist(created),
	}

	return x, nil
}

func (s *serviceImpl) GetWaitlist(ctx context.Context, request *waitlistssvc.GetWaitlistRequest) (*waitlistssvc.GetWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	waitlist, err := s.waitlistsManager.GetWaitlist(ctx, request.WaitlistId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlist")
	}

	x := &waitlistssvc.GetWaitlistResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertWaitlistToGRPCWaitlist(waitlist),
	}

	return x, nil
}

func (s *serviceImpl) GetWaitlists(ctx context.Context, request *waitlistssvc.GetWaitlistsRequest) (*waitlistssvc.GetWaitlistsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.waitlistsManager.GetWaitlists(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlists")
	}

	x := &waitlistssvc.GetWaitlistsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(retrieved.Pagination, filter),
	}

	for _, waitlist := range retrieved.Data {
		x.Results = append(x.Results, converters.ConvertWaitlistToGRPCWaitlist(waitlist))
	}

	return x, nil
}

func (s *serviceImpl) GetActiveWaitlists(ctx context.Context, request *waitlistssvc.GetActiveWaitlistsRequest) (*waitlistssvc.GetActiveWaitlistsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.waitlistsManager.GetActiveWaitlists(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch active waitlists")
	}

	x := &waitlistssvc.GetActiveWaitlistsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(retrieved.Pagination, filter),
	}

	for _, waitlist := range retrieved.Data {
		x.Results = append(x.Results, converters.ConvertWaitlistToGRPCWaitlist(waitlist))
	}

	return x, nil
}

func (s *serviceImpl) UpdateWaitlist(ctx context.Context, request *waitlistssvc.UpdateWaitlistRequest) (*waitlistssvc.UpdateWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	waitlist, err := s.waitlistsManager.GetWaitlist(ctx, request.WaitlistId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlist for update")
	}

	updateInput := converters.ConvertGRPCWaitlistUpdateRequestInputToWaitlistUpdateRequestInput(request.Input)
	waitlist.Update(updateInput)

	if err = s.waitlistsManager.UpdateWaitlist(ctx, waitlist); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update waitlist")
	}

	x := &waitlistssvc.UpdateWaitlistResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertWaitlistToGRPCWaitlist(waitlist),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveWaitlist(ctx context.Context, request *waitlistssvc.ArchiveWaitlistRequest) (*waitlistssvc.ArchiveWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	if err := s.waitlistsManager.ArchiveWaitlist(ctx, request.WaitlistId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive waitlist")
	}

	x := &waitlistssvc.ArchiveWaitlistResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) WaitlistIsNotExpired(ctx context.Context, request *waitlistssvc.WaitlistIsNotExpiredRequest) (*waitlistssvc.WaitlistIsNotExpiredResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	isNotExpired, err := s.waitlistsManager.WaitlistIsNotExpired(ctx, request.WaitlistId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to check waitlist expiration status")
	}

	x := &waitlistssvc.WaitlistIsNotExpiredResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		IsNotExpired: isNotExpired,
	}

	return x, nil
}

func (s *serviceImpl) CreateWaitlistSignup(ctx context.Context, request *waitlistssvc.CreateWaitlistSignupRequest) (*waitlistssvc.CreateWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCWaitlistSignupCreationRequestInputToWaitlistSignupDatabaseCreationInput(request.Input)
	input.BelongsToUser = sessionContextData.GetUserID()
	input.BelongsToAccount = sessionContextData.ActiveAccountID

	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate waitlist signup creation request")
	}

	created, err := s.waitlistsManager.CreateWaitlistSignup(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create waitlist signup")
	}

	x := &waitlistssvc.CreateWaitlistSignupResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertWaitlistSignupToGRPCWaitlistSignup(created),
	}

	return x, nil
}

func (s *serviceImpl) GetWaitlistSignup(ctx context.Context, request *waitlistssvc.GetWaitlistSignupRequest) (*waitlistssvc.GetWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, request.WaitlistSignupId).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	signup, err := s.waitlistsManager.GetWaitlistSignup(ctx, request.WaitlistSignupId, request.WaitlistId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlist signup")
	}

	// signups are user-owned: only the owner (or a service admin) may read one.
	if signup.BelongsToUser != sessionContextData.GetUserID() && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForWaitlistSignup, logger, span, codes.PermissionDenied, "not authorized to access waitlist signup")
	}

	x := &waitlistssvc.GetWaitlistSignupResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup),
	}

	return x, nil
}

func (s *serviceImpl) GetWaitlistSignupsForWaitlist(ctx context.Context, request *waitlistssvc.GetWaitlistSignupsForWaitlistRequest) (*waitlistssvc.GetWaitlistSignupsForWaitlistResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	// the waitlist-wide signup listing exposes every user's signup, so it is reserved for service admins.
	if !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForWaitlistSignup, logger, span, codes.PermissionDenied, "not authorized to list waitlist signups")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.waitlistsManager.GetWaitlistSignupsForWaitlist(ctx, request.WaitlistId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlist signups")
	}

	x := &waitlistssvc.GetWaitlistSignupsForWaitlistResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(retrieved.Pagination, filter),
	}

	for _, signup := range retrieved.Data {
		x.Results = append(x.Results, converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup))
	}

	return x, nil
}

func (s *serviceImpl) UpdateWaitlistSignup(ctx context.Context, request *waitlistssvc.UpdateWaitlistSignupRequest) (*waitlistssvc.UpdateWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, request.WaitlistSignupId).WithValue(waitlistkeys.WaitlistIDKey, request.WaitlistId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	signup, err := s.waitlistsManager.GetWaitlistSignup(ctx, request.WaitlistSignupId, request.WaitlistId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlist signup for update")
	}

	// signups are user-owned: only the owner (or a service admin) may update one.
	if signup.BelongsToUser != sessionContextData.GetUserID() && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForWaitlistSignup, logger, span, codes.PermissionDenied, "not authorized to update waitlist signup")
	}

	updateInput := converters.ConvertGRPCWaitlistSignupUpdateRequestInputToWaitlistSignupUpdateRequestInput(request.Input)
	signup.Update(updateInput)

	if err = s.waitlistsManager.UpdateWaitlistSignup(ctx, signup); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update waitlist signup")
	}

	x := &waitlistssvc.UpdateWaitlistSignupResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertWaitlistSignupToGRPCWaitlistSignup(signup),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveWaitlistSignup(ctx context.Context, request *waitlistssvc.ArchiveWaitlistSignupRequest) (*waitlistssvc.ArchiveWaitlistSignupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, request.WaitlistSignupId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	signup, err := s.waitlistsManager.GetWaitlistSignupByID(ctx, request.WaitlistSignupId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch waitlist signup for archival")
	}

	// signups are user-owned: only the owner (or a service admin) may archive one.
	if signup.BelongsToUser != sessionContextData.GetUserID() && !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForWaitlistSignup, logger, span, codes.PermissionDenied, "not authorized to archive waitlist signup")
	}

	if err = s.waitlistsManager.ArchiveWaitlistSignup(ctx, request.WaitlistSignupId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive waitlist signup")
	}

	x := &waitlistssvc.ArchiveWaitlistSignupResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
