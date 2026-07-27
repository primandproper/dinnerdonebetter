package grpc

import (
	"context"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	webhookssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v7/errors/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateWebhook(ctx context.Context, request *webhookssvc.CreateWebhookRequest) (*webhookssvc.CreateWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	requestInput := converters.ConvertGRPCWebhookCreationRequestInputToWebhookCreationRequestInput(request.Input)
	if err = requestInput.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate webhook creation request")
	}

	created, err := s.webhookManager.CreateWebhook(ctx, sessionContextData.GetUserID(), sessionContextData.ActiveAccountID, requestInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create webhook")
	}

	x := &webhookssvc.CreateWebhookResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertWebhookToGRPCWebhook(created),
	}

	return x, nil
}

func (s *serviceImpl) AddWebhookTriggerConfig(ctx context.Context, request *webhookssvc.AddWebhookTriggerConfigRequest) (*webhookssvc.AddWebhookTriggerConfigResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	requestInput := &webhooks.WebhookTriggerConfigCreationRequestInput{
		BelongsToWebhook: request.WebhookId,
		TriggerEventID:   request.Input.GetTriggerEventId(),
	}
	created, err := s.webhookManager.AddWebhookTriggerConfig(ctx, sessionContextData.ActiveAccountID, requestInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add webhook trigger config")
	}

	x := &webhookssvc.AddWebhookTriggerConfigResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertWebhookTriggerConfigToGRPCWebhookTriggerConfig(created),
	}

	return x, nil
}

func (s *serviceImpl) GetWebhook(ctx context.Context, request *webhookssvc.GetWebhookRequest) (*webhookssvc.GetWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(webhookkeys.WebhookIDKey, request.WebhookId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	webhook, err := s.webhookManager.GetWebhook(ctx, request.WebhookId, sessionContextData.GetActiveAccountID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch webhook")
	}

	x := &webhookssvc.GetWebhookResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertWebhookToGRPCWebhook(webhook),
	}

	return x, nil
}

func (s *serviceImpl) GetWebhooks(ctx context.Context, request *webhookssvc.GetWebhooksRequest) (*webhookssvc.GetWebhooksResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.webhookManager.GetWebhooks(ctx, sessionContextData.ActiveAccountID, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch webhooks")
	}

	x := &webhookssvc.GetWebhooksResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(retrieved.Pagination, filter),
	}

	for _, webhook := range retrieved.Data {
		x.Results = append(x.Results, converters.ConvertWebhookToGRPCWebhook(webhook))
	}

	return x, nil
}

func (s *serviceImpl) ArchiveWebhook(ctx context.Context, request *webhookssvc.ArchiveWebhookRequest) (*webhookssvc.ArchiveWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	if err = s.webhookManager.ArchiveWebhook(ctx, request.WebhookId, sessionContextData.ActiveAccountID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive webhook")
	}

	x := &webhookssvc.ArchiveWebhookResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveWebhookTriggerConfig(ctx context.Context, request *webhookssvc.ArchiveWebhookTriggerConfigRequest) (*webhookssvc.ArchiveWebhookTriggerConfigResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(webhookkeys.WebhookIDKey, request.WebhookId).WithValue(webhookkeys.WebhookTriggerConfigIDKey, request.WebhookTriggerConfigId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	// verify the webhook belongs to the caller's active account before mutating its trigger configs.
	if _, err = s.webhookManager.GetWebhook(ctx, request.WebhookId, sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch webhook")
	}

	if err = s.webhookManager.ArchiveWebhookTriggerConfig(ctx, request.WebhookId, request.WebhookTriggerConfigId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive webhook trigger config")
	}

	return &webhookssvc.ArchiveWebhookTriggerConfigResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) ArchiveWebhookTriggerEvent(ctx context.Context, request *webhookssvc.ArchiveWebhookTriggerEventRequest) (*webhookssvc.ArchiveWebhookTriggerEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(webhookkeys.WebhookTriggerEventIDKey, request.Id)

	if err := s.webhookManager.ArchiveWebhookTriggerEvent(ctx, request.Id); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive webhook trigger event")
	}

	return &webhookssvc.ArchiveWebhookTriggerEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

func (s *serviceImpl) CreateWebhookTriggerEvent(ctx context.Context, request *webhookssvc.CreateWebhookTriggerEventRequest) (*webhookssvc.CreateWebhookTriggerEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if _, err := s.sessionContextDataFetcher(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Unauthenticated, "failed to fetch session context data")
	}

	requestInput := converters.ConvertGRPCWebhookTriggerEventCreationRequestInputToWebhookTriggerEventCreationRequestInput(request.Input)
	created, err := s.webhookManager.CreateWebhookTriggerEvent(ctx, requestInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "failed to create webhook trigger event")
	}
	return &webhookssvc.CreateWebhookTriggerEventResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Created:         converters.ConvertWebhookTriggerEventCatalogToGRPCWebhookTriggerEvent(created),
	}, nil
}

func (s *serviceImpl) GetWebhookTriggerEvent(ctx context.Context, request *webhookssvc.GetWebhookTriggerEventRequest) (*webhookssvc.GetWebhookTriggerEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	result, err := s.webhookManager.GetWebhookTriggerEvent(ctx, request.Id)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "failed to get webhook trigger event")
	}
	return &webhookssvc.GetWebhookTriggerEventResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Result:          converters.ConvertWebhookTriggerEventCatalogToGRPCWebhookTriggerEvent(result),
	}, nil
}

func (s *serviceImpl) GetWebhookTriggerEvents(ctx context.Context, request *webhookssvc.GetWebhookTriggerEventsRequest) (*webhookssvc.GetWebhookTriggerEventsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.webhookManager.GetWebhookTriggerEvents(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "failed to get webhook trigger events")
	}
	x := &webhookssvc.GetWebhookTriggerEventsResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(retrieved.Pagination, filter),
	}
	for _, ev := range retrieved.Data {
		x.Results = append(x.Results, converters.ConvertWebhookTriggerEventCatalogToGRPCWebhookTriggerEvent(ev))
	}
	return x, nil
}

func (s *serviceImpl) UpdateWebhookTriggerEvent(ctx context.Context, request *webhookssvc.UpdateWebhookTriggerEventRequest) (*webhookssvc.UpdateWebhookTriggerEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertGRPCWebhookTriggerEventUpdateRequestInputToWebhookTriggerEventUpdateRequestInput(request.Input)
	if err := s.webhookManager.UpdateWebhookTriggerEvent(ctx, request.Id, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "failed to update webhook trigger event")
	}
	return &webhookssvc.UpdateWebhookTriggerEventResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
	}, nil
}
