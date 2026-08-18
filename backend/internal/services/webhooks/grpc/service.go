package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/manager"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

const (
	o11yName = "webhooks_service"
)

var _ webhookssvc.WebhooksServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		webhookssvc.UnimplementedWebhooksServiceServer
		tracer         tracing.Tracer
		logger         logging.Logger
		webhookManager manager.WebhookDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	webhookManager manager.WebhookDataManager,
) webhookssvc.WebhooksServiceServer {
	return &serviceImpl{
		logger:         logging.NewNamedLogger(logger, o11yName),
		tracer:         tracing.NewNamedTracer(tracerProvider, o11yName),
		webhookManager: webhookManager,
	}
}
