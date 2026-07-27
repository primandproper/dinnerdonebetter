package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	dataprivacymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/manager"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/dataprivacy/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	errorsgrpc "github.com/primandproper/platform-go/v7/errors/grpc"
	"github.com/primandproper/platform-go/v7/identifiers"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	"github.com/primandproper/platform-go/v7/observability/logging"
	metricsnoop "github.com/primandproper/platform-go/v7/observability/metrics/noop"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"
	"github.com/primandproper/platform-go/v7/uploads"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "data_privacy_service"

	// userDataDisclosureTTL is how long a generated user data report remains available before it expires.
	userDataDisclosureTTL = 7 * 24 * time.Hour
)

var _ dataprivacysvc.DataPrivacyServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		dataprivacysvc.UnimplementedDataPrivacyServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		dataPrivacyManager        dataprivacymanager.DataPrivacyManager
		uploadManager             uploads.UploadManager
		msgConfig                 *msgconfig.Config
		queuesConfig              *msgconfig.QueuesConfig
	}
)

// NewDataPrivacyService creates a new data privacy gRPC service.
func NewDataPrivacyService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error),
	dataPrivacyManager dataprivacymanager.DataPrivacyManager,
	uploadManager uploads.UploadManager,
	msgConfig *msgconfig.Config,
	queuesConfig *msgconfig.QueuesConfig,
) dataprivacysvc.DataPrivacyServiceServer {
	return &serviceImpl{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		sessionContextDataFetcher: sessionContextDataFetcher,
		dataPrivacyManager:        dataPrivacyManager,
		uploadManager:             uploadManager,
		msgConfig:                 msgConfig,
		queuesConfig:              queuesConfig,
	}
}

// AggregateUserDataReport records a user data disclosure request and enqueues the aggregation work for GDPR/CCPA
// disclosure. The heavy cross-domain gather is performed asynchronously by the user data aggregation worker, which
// writes the report to object storage and marks the disclosure completed or failed. The report is available via
// FetchUserDataReport once the disclosure reports as completed.
func (s *serviceImpl) AggregateUserDataReport(ctx context.Context, _ *dataprivacysvc.AggregateUserDataReportRequest) (*dataprivacysvc.AggregateUserDataReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.Requester.UserID
	logger := s.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	// Generate a unique report ID
	reportID := identifiers.New()
	logger = logger.WithValue(dataprivacykeys.UserDataAggregationReportIDKey, reportID)
	tracing.AttachToSpan(span, dataprivacykeys.UserDataAggregationReportIDKey, reportID)

	logger.Info("recording user data disclosure request")

	// Record the disclosure request so its status can be tracked and polled.
	disclosure, err := s.dataPrivacyManager.CreateUserDataDisclosure(ctx, &dataprivacy.UserDataDisclosureCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: userID,
		ExpiresAt:     time.Now().Add(userDataDisclosureTTL).UTC(),
	})
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating user data disclosure")
	}

	// Enqueue the aggregation work for the async worker to process off the request path.
	if err = s.publishAggregationRequest(ctx, disclosure.ID, reportID, userID); err != nil {
		// The disclosure was created but could not be enqueued; mark it failed so it does not linger as pending.
		if markErr := s.dataPrivacyManager.MarkUserDataDisclosureFailed(ctx, disclosure.ID); markErr != nil {
			logger.Error("marking disclosure failed after publish failure", markErr)
		}
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "publishing user data aggregation request")
	}

	logger.Info("user data aggregation request enqueued")

	return &dataprivacysvc.AggregateUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		ReportId: reportID,
	}, nil
}

// publishAggregationRequest publishes a UserDataAggregationRequest to the user data aggregation topic.
func (s *serviceImpl) publishAggregationRequest(ctx context.Context, disclosureID, reportID, userID string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	pp, err := msgconfig.NewPublisherProvider(ctx, s.logger, tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), s.msgConfig)
	if err != nil {
		return fmt.Errorf("establishing publisher provider: %w", err)
	}

	publisher, err := pp.NewPublisher(ctx, s.queuesConfig.UserDataAggregationTopicName)
	if err != nil {
		return fmt.Errorf("initializing publisher: %w", err)
	}

	if err = publisher.Publish(ctx, &dataprivacy.UserDataAggregationRequest{
		RequestID: disclosureID,
		ReportID:  reportID,
		UserID:    userID,
	}); err != nil {
		return fmt.Errorf("publishing user data aggregation request: %w", err)
	}

	return nil
}

// DestroyAllUserData permanently deletes a user and all associated data.
func (s *serviceImpl) DestroyAllUserData(ctx context.Context, _ *dataprivacysvc.DestroyAllUserDataRequest) (*dataprivacysvc.DestroyAllUserDataResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.Requester.UserID
	logger := s.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	logger.Info("destroying all user data")

	if err = s.dataPrivacyManager.DeleteUser(ctx, userID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "deleting user")
	}

	logger.Info("user data destroyed successfully")

	return &dataprivacysvc.DestroyAllUserDataResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Successful: true,
	}, nil
}

// FetchUserDataReport retrieves a previously generated user data report from object storage.
func (s *serviceImpl) FetchUserDataReport(ctx context.Context, request *dataprivacysvc.FetchUserDataReportRequest) (*dataprivacysvc.FetchUserDataReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	reportID := request.GetUserDataAggregationReportId()
	logger := s.logger.WithValue(dataprivacykeys.UserDataAggregationReportIDKey, reportID)
	tracing.AttachToSpan(span, dataprivacykeys.UserDataAggregationReportIDKey, reportID)

	logger.Info("fetching user data report")

	// Read the report from object storage
	reportBytes, err := uploads.ReadFile(ctx, s.uploadManager, fmt.Sprintf("%s.json", reportID))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "reading report from storage")
	}

	// Unmarshal the report
	var collection dataprivacy.UserDataCollection
	if err = json.Unmarshal(reportBytes, &collection); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "unmarshaling report")
	}

	// Verify the report belongs to the requester before disclosing it.
	if collection.Identity.User.ID != sessionContextData.Requester.UserID {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("report does not belong to requester"), logger, span, codes.PermissionDenied, "report does not belong to requester")
	}

	logger.Info("user data report fetched successfully")

	// Convert to proto type
	return &dataprivacysvc.FetchUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserDataCollection: converters.ConvertUserDataCollectionToGRPCUserDataCollection(&collection, reportID),
	}, nil
}
