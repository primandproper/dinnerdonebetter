package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	dataprivacymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/manager"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/dataprivacy/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	errorsgrpc "github.com/primandproper/platform-go/v5/errors/grpc"
	"github.com/primandproper/platform-go/v5/identifiers"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/uploads"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "data_privacy_service"
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
	}
)

// NewDataPrivacyService creates a new data privacy gRPC service.
func NewDataPrivacyService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error),
	dataPrivacyManager dataprivacymanager.DataPrivacyManager,
	uploadManager uploads.UploadManager,
) dataprivacysvc.DataPrivacyServiceServer {
	return &serviceImpl{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		sessionContextDataFetcher: sessionContextDataFetcher,
		dataPrivacyManager:        dataPrivacyManager,
		uploadManager:             uploadManager,
	}
}

// AggregateUserDataReport collects all user data and saves it to object storage for GDPR/CCPA disclosure.
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

	logger.Info("aggregating user data")

	// Fetch all user data
	collection, err := s.dataPrivacyManager.FetchUserDataCollection(ctx, userID)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching user data collection")
	}

	// Marshal and save to object storage
	collectionBytes, err := json.Marshal(collection)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "marshaling user data collection")
	}

	if err = uploads.SaveFile(ctx, s.uploadManager, fmt.Sprintf("%s.json", reportID), collectionBytes); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "saving user data report")
	}

	logger.Info("user data aggregation complete")

	return &dataprivacysvc.AggregateUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		ReportId: reportID,
	}, nil
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
