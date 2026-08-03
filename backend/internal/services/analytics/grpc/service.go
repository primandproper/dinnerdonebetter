package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	analyticspb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"
	grpctypes "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"

	"github.com/primandproper/platform-go/v9/analytics/multisource"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const o11yName = "analytics_service"

var _ analyticspb.AnalyticsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		analyticspb.UnimplementedAnalyticsServiceServer
		tracer              tracing.Tracer
		logger              logging.Logger
		multiSourceReporter *multisource.MultiSourceEventReporter
	}
)

// NewService returns a new AnalyticsServiceServer.
func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	multiSourceReporter *multisource.MultiSourceEventReporter,
) analyticspb.AnalyticsServiceServer {
	return &serviceImpl{
		logger:              logging.NewNamedLogger(logger, o11yName),
		tracer:              tracing.NewNamedTracer(tracerProvider, o11yName),
		multiSourceReporter: multiSourceReporter,
	}
}

func (s *serviceImpl) TrackEvent(ctx context.Context, req *analyticspb.TrackEventRequest) (*analyticspb.TrackEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("event", req.Event)

	sessionCtxData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionCtxData.GetUserID()
	if userID == "" {
		return nil, observability.PrepareAndLogGRPCStatus(sessions.ErrAuthenticationNotFound, logger, span, codes.Unauthenticated, "user ID missing from session context")
	}

	properties := stringMapToAnyMap(req.GetProperties())
	if err = s.multiSourceReporter.TrackEvent(ctx, req.GetSource(), req.GetEvent(), userID, properties); err != nil {
		return nil, err
	}

	x := &analyticspb.TrackEventResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) TrackAnonymousEvent(ctx context.Context, req *analyticspb.TrackAnonymousEventRequest) (*analyticspb.TrackAnonymousEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("event", req.Event)

	anonymousID := req.GetAnonymousId()
	if anonymousID == "" {
		return nil, status.Error(codes.InvalidArgument, "anonymous_id is required")
	}

	properties := stringMapToAnyMap(req.GetProperties())
	if err := s.multiSourceReporter.TrackAnonymousEvent(ctx, req.GetSource(), req.GetEvent(), anonymousID, properties); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unavailable, "anonymous")
	}

	x := &analyticspb.TrackAnonymousEventResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func stringMapToAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
