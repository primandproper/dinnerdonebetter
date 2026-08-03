package grpc

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	domaininternalops "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"

	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"
	"github.com/primandproper/platform-go/v9/identifiers"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability/logging"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "internalops_service"

	testQueueMessagePollInterval = 500 * time.Millisecond
	testQueueMessageTimeout      = 30 * time.Second
)

var _ settingssvc.InternalOperationsServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		settingssvc.UnimplementedInternalOperationsServer
		tracer          tracing.Tracer
		logger          logging.Logger
		msgConfig       *msgconfig.Config
		internalOpsRepo domaininternalops.InternalOpsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	msgConfig *msgconfig.Config,
	repo domaininternalops.InternalOpsDataManager,
) settingssvc.InternalOperationsServer {
	return &serviceImpl{
		msgConfig:       msgConfig,
		internalOpsRepo: repo,
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
	}
}

func (s *serviceImpl) TestQueueMessage(ctx context.Context, request *settingssvc.TestQueueMessageRequest) (*settingssvc.TestQueueMessageResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue("queue_name", request.QueueName)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if !sessionContextData.ServiceRolePermissionChecker().HasPermission(authorization.PublishArbitraryQueueMessagePermission) {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.PermissionDenied, "user is not permitted")
	}

	if request.QueueName == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(nil, logger, span, codes.InvalidArgument, "queue_name is required")
	}

	testID := identifiers.New()
	start := time.Now()

	if err = s.internalOpsRepo.CreateQueueTestMessage(ctx, testID, request.QueueName); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating queue test message record")
	}

	msg, err := domaininternalops.BuildQueueTestMessage(request.QueueName, testID, sessionContextData.GetUserID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "building queue test message")
	}

	pp, err := msgconfig.NewPublisherProvider(ctx, s.msgConfig,
		msgconfig.WithLogger(s.logger),
		msgconfig.WithTracerProvider(tracingnoop.NewTracerProvider()),
		msgconfig.WithMetricsProvider(metricsnoop.NewMetricsProvider()),
	)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "establishing publisher provider")
	}

	publisher, err := pp.NewPublisher(ctx, request.QueueName)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "initializing publisher")
	}

	if err = publisher.Publish(ctx, msg); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "publishing test message")
	}

	ticker := time.NewTicker(testQueueMessagePollInterval)
	defer ticker.Stop()

	deadline := time.After(testQueueMessageTimeout)

	for {
		select {
		case <-deadline:
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(nil, logger, span, codes.DeadlineExceeded, "timed out waiting for queue test message acknowledgment")
		case <-ctx.Done():
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(ctx.Err(), logger, span, codes.Canceled, "context canceled while waiting for acknowledgment")
		case <-ticker.C:
			record, pollErr := s.internalOpsRepo.GetQueueTestMessage(ctx, testID)
			if pollErr != nil {
				logger.Error("polling for queue test message acknowledgment", pollErr)
				continue
			}

			if record.AcknowledgedAt != nil {
				roundTripMs := time.Since(start).Milliseconds()
				return &settingssvc.TestQueueMessageResponse{
					ResponseDetails: &types.ResponseDetails{
						TraceId: span.SpanContext().TraceID().String(),
					},
					Success:     true,
					TestId:      testID,
					RoundTripMs: roundTripMs,
				}, nil
			}
		}
	}
}
