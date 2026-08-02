package grpc

import (
	"context"
	"database/sql"
	"errors"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	notificationkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	notificationssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	grpctypes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/notifications/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) RegisterDeviceToken(ctx context.Context, request *notificationssvc.RegisterDeviceTokenRequest) (*notificationssvc.RegisterDeviceTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	if request == nil || request.Input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input required"), logger, span, codes.InvalidArgument, "input required")
	}

	input := &notifications.UserDeviceTokenDatabaseCreationInput{
		ID:            identifiers.New(),
		DeviceToken:   request.Input.DeviceToken,
		Platform:      request.Input.Platform,
		BelongsToUser: sessionContextData.GetUserID(),
	}

	logger = logger.WithValue(identitykeys.UserIDKey, input.BelongsToUser).WithValue(notificationkeys.UserDeviceTokenIDKey, input.ID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.BelongsToUser)

	created, err := s.notificationsManager.CreateUserDeviceToken(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "registering device token")
	}

	x := &notificationssvc.RegisterDeviceTokenResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertUserDeviceTokenToGRPCUserDeviceToken(created),
	}

	return x, nil
}

func (s *serviceImpl) GetUserDeviceToken(ctx context.Context, request *notificationssvc.GetUserDeviceTokenRequest) (*notificationssvc.GetUserDeviceTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(notificationkeys.UserDeviceTokenIDKey, request.UserDeviceTokenId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	token, err := s.notificationsManager.GetUserDeviceToken(ctx, sessionContextData.GetUserID(), request.UserDeviceTokenId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "device token not found")
		}
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching device token")
	}

	x := &notificationssvc.GetUserDeviceTokenResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertUserDeviceTokenToGRPCUserDeviceToken(token),
	}

	return x, nil
}

func (s *serviceImpl) GetUserDeviceTokens(ctx context.Context, request *notificationssvc.GetUserDeviceTokensRequest) (*notificationssvc.GetUserDeviceTokensResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	var platformFilter *string
	if request.PlatformFilter != nil {
		platformFilter = request.PlatformFilter
	}

	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	tokens, err := s.notificationsManager.GetUserDeviceTokens(ctx, sessionContextData.GetUserID(), filter, platformFilter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching device tokens")
	}

	x := &notificationssvc.GetUserDeviceTokensResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(tokens.Pagination, filter),
	}
	for _, token := range tokens.Data {
		x.Results = append(x.Results, converters.ConvertUserDeviceTokenToGRPCUserDeviceToken(token))
	}

	return x, nil
}

func (s *serviceImpl) ArchiveUserDeviceToken(ctx context.Context, request *notificationssvc.ArchiveUserDeviceTokenRequest) (*notificationssvc.ArchiveUserDeviceTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(notificationkeys.UserDeviceTokenIDKey, request.UserDeviceTokenId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	if err = s.notificationsManager.ArchiveUserDeviceToken(ctx, sessionContextData.GetUserID(), request.UserDeviceTokenId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "device token not found")
		}
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving device token")
	}

	x := &notificationssvc.ArchiveUserDeviceTokenResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
