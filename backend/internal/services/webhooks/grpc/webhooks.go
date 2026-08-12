package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v10/errors/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateWebhook(ctx context.Context, request *webhookssvc.CreateWebhookRequest) (*webhookssvc.CreateWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	requestInput := converters.ConvertGRPCWebhookCreationRequestInputToWebhookCreationRequestInput(request.Input)
	if err = requestInput.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate webhook creation request")
	}

	created, err := s.webhookManager.CreateWebhook(ctx, sessionContextData.GetUserID(), sessionContextData.GetActiveAccountID(), requestInput)
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

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	requestInput := &webhooks.WebhookTriggerConfigCreationRequestInput{
		BelongsToWebhook: request.WebhookId,
		EventType:        request.Input.GetEventType(),
	}
	if err = requestInput.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate webhook trigger config request")
	}

	created, err := s.webhookManager.AddWebhookTriggerConfig(ctx, sessionContextData.GetActiveAccountID(), requestInput)
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

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

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

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.webhookManager.GetWebhooks(ctx, sessionContextData.GetActiveAccountID(), filter)
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

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	if err = s.webhookManager.ArchiveWebhook(ctx, request.WebhookId, sessionContextData.GetActiveAccountID()); err != nil {
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

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	// verify the webhook belongs to the caller's active account before mutating its trigger configs.
	if _, err = s.webhookManager.GetWebhook(ctx, request.WebhookId, sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch webhook")
	}

	if err = s.webhookManager.ArchiveWebhookTriggerConfig(ctx, request.WebhookId, sessionContextData.GetActiveAccountID(), request.WebhookTriggerConfigId); err != nil {
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

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	secret, err := s.webhookManager.RotateWebhookSecret(ctx, request.WebhookId, sessionContextData.GetActiveAccountID())
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

	if _, err := sessions.RequireFromContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Unauthenticated, "fetching session context data")
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
