package grpc

import (
	"context"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"

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
		Created: converters.ConvertWebhookToGRPCWebhook(created.Webhook),
		// The only time this value leaves the server. It is not logged and no read path can
		// produce it; a caller who loses it calls RotateWebhookSecret.
		Secret: created.Secret,
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
		EventType:        request.Input.GetEventType(),
	}
	if err = requestInput.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate webhook trigger config request")
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

	if err = s.webhookManager.ArchiveWebhookTriggerConfig(ctx, request.WebhookId, sessionContextData.ActiveAccountID, request.WebhookTriggerConfigId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive webhook trigger config")
	}

	return &webhookssvc.ArchiveWebhookTriggerConfigResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

// RotateWebhookSecret mints a new signing secret for a webhook and returns it, once.
//
// Deliveries are signed under both the new key and the outgoing one until this is called again,
// so a subscriber accepts either signature for as long as it needs to switch over.
func (s *serviceImpl) RotateWebhookSecret(ctx context.Context, request *webhookssvc.RotateWebhookSecretRequest) (*webhookssvc.RotateWebhookSecretResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(webhookkeys.WebhookIDKey, request.WebhookId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	secret, err := s.webhookManager.RotateWebhookSecret(ctx, request.WebhookId, sessionContextData.ActiveAccountID)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to rotate webhook signing secret")
	}

	return &webhookssvc.RotateWebhookSecretResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Secret: secret,
	}, nil
}

// GetWebhookEventTypes lists the events a webhook may subscribe to.
//
// The list is generated Go rather than a table, so this reads no database and takes no filter:
// it is a constant for the lifetime of the deployment.
func (s *serviceImpl) GetWebhookEventTypes(ctx context.Context, _ *webhookssvc.GetWebhookEventTypesRequest) (*webhookssvc.GetWebhookEventTypesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if _, err := s.sessionContextDataFetcher(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Unauthenticated, "failed to fetch session context data")
	}

	x := &webhookssvc.GetWebhookEventTypesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, eventType := range s.webhookManager.GetWebhookEventTypes(ctx) {
		x.Results = append(x.Results, converters.ConvertWebhookEventTypeToGRPCWebhookEventType(eventType))
	}

	return x, nil
}
