package grpc

import (
	"context"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	grpctypes "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/notifications/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) GetUserNotification(ctx context.Context, request *notificationssvc.GetUserNotificationRequest) (*notificationssvc.GetUserNotificationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(notificationkeys.UserNotificationIDKey, request.UserNotificationId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	notification, err := s.notificationsManager.GetUserNotification(ctx, sessionContextData.GetUserID(), request.UserNotificationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	x := &notificationssvc.GetUserNotificationResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertUserNotificationToGRPCUserNotification(notification),
	}

	return x, nil
}

func (s *serviceImpl) GetUserNotifications(ctx context.Context, request *notificationssvc.GetUserNotificationsRequest) (*notificationssvc.GetUserNotificationsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	notifs, err := s.notificationsManager.GetUserNotifications(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching user notifs")
	}

	x := &notificationssvc.GetUserNotificationsResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(notifs.Pagination, filter),
	}
	for _, notif := range notifs.Data {
		x.Results = append(x.Results, converters.ConvertUserNotificationToGRPCUserNotification(notif))
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserNotification(ctx context.Context, request *notificationssvc.UpdateUserNotificationRequest) (*notificationssvc.UpdateUserNotificationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "unable to determine authentication")
	}

	logger = logger.WithValue(notificationkeys.UserNotificationIDKey, request.UserNotificationId)

	existing, err := s.notificationsManager.GetUserNotification(ctx, sessionContextData.GetUserID(), request.UserNotificationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching existing notification")
	}

	var newStatus *string
	if request.Input.Status != nil {
		newStatus = new(converters.ConvertUserNotificationStatusToString(*request.Input.Status))
	}

	existing.Update(&notifications.UserNotificationUpdateRequestInput{Status: newStatus})
	if err = s.notificationsManager.UpdateUserNotification(ctx, existing); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating existing notification")
	}

	updated, err := s.notificationsManager.GetUserNotification(ctx, sessionContextData.GetUserID(), request.UserNotificationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching existing notification")
	}

	x := &notificationssvc.UpdateUserNotificationResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertUserNotificationToGRPCUserNotification(updated),
	}

	return x, nil
}
