package grpc

import (
	"context"
	"errors"
	"io"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	dataprivacykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/grpc/converters"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

const o11yName = "data_privacy_service"

var _ dataprivacysvc.DataPrivacyServiceServer = (*serviceImpl)(nil)

type serviceImpl struct {
	dataprivacysvc.UnimplementedDataPrivacyServiceServer
	tracer   tracing.Tracer
	logger   logging.Logger
	requests platformdataprivacy.Service
}

// NewDataPrivacyService creates a new data privacy gRPC service.
//
// It is a thin shell over platform-go's Service, and deliberately so. Every method
// here does three things: resolve who is asking, hand the question to the platform,
// and refuse to answer about anybody else. The state machine, the durable queue, the
// artifact packaging, and the expiry all live behind that interface — this file's
// only real job is the authorization the platform cannot do, because it has no
// notion of a session.
func NewDataPrivacyService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	requests platformdataprivacy.Service,
) dataprivacysvc.DataPrivacyServiceServer {
	return &serviceImpl{
		logger:   logging.NewNamedLogger(logger, o11yName),
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		requests: requests,
	}
}

// AggregateUserDataReport submits a subject access request for the requester's data.
//
// It returns as soon as the request is recorded. The cross-domain gather happens in
// the fulfillment worker, which is the point: a subject access request is a legal
// obligation, and the one guarantee it needs is that it survives the process that
// accepted it.
func (s *serviceImpl) AggregateUserDataReport(ctx context.Context, _ *dataprivacysvc.AggregateUserDataReportRequest) (*dataprivacysvc.AggregateUserDataReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	request, err := s.submit(ctx, span, platformdataprivacy.RequestExport)
	if err != nil {
		return nil, err
	}

	return &dataprivacysvc.AggregateUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Request:         converters.ConvertRequestToGRPCRequest(request),
	}, nil
}

// DestroyAllUserData submits a right-to-be-forgotten request for the requester.
//
// It queues rather than deletes, and returns a pending request rather than a
// success. The confirmation window is zero, so nothing further is needed from the
// subject; what changed is that the deletion now happens inside one transaction
// shared by every domain's eraser and by the record that the erasure occurred,
// instead of on the request path where a timeout halfway through left a subject in
// a state no status could describe.
func (s *serviceImpl) DestroyAllUserData(ctx context.Context, _ *dataprivacysvc.DestroyAllUserDataRequest) (*dataprivacysvc.DestroyAllUserDataResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	request, err := s.submit(ctx, span, platformdataprivacy.RequestErasure)
	if err != nil {
		return nil, err
	}

	return &dataprivacysvc.DestroyAllUserDataResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Request:         converters.ConvertRequestToGRPCRequest(request),
	}, nil
}

// FetchUserDataReport returns a completed export's artifact to its subject.
//
// Open, not Download. Artifacts are encrypted at rest, so a signed URL would hand
// the subject the stored object — ciphertext they cannot open, discovered some days
// into a statutory window. platform-go refuses Download outright once an encryptor
// is configured; this is the path that works, at the cost of the bytes passing
// through here.
func (s *serviceImpl) FetchUserDataReport(ctx context.Context, request *dataprivacysvc.FetchUserDataReportRequest) (*dataprivacysvc.FetchUserDataReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	requestID := request.GetDataPrivacyRequestId()

	// Ownership is checked against the request row before a single byte is read, not
	// against something inside the artifact. The code this replaces unmarshaled the
	// report and compared the user ID it found there, which made the file its own
	// access control.
	stored, err := s.ownedRequest(ctx, span, requestID)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(dataprivacykeys.RequestIDKey, requestID)

	artifact, err := s.requests.Open(ctx, stored.ID)
	if err != nil {
		// Unavailable covers an erasure, an export that has not completed, and one whose
		// artifact has expired and been deleted. All three are "there is nothing here to
		// give you", which is NotFound rather than an error about our storage.
		if errors.Is(err, platformdataprivacy.ErrArtifactUnavailable) {
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "artifact is unavailable")
		}

		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "opening artifact")
	}

	defer func() {
		if closeErr := artifact.Close(); closeErr != nil {
			observability.AcknowledgeError(closeErr, logger, span, "closing artifact")
		}
	}()

	body, err := io.ReadAll(artifact)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "reading artifact")
	}

	return &dataprivacysvc.FetchUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Artifact:        body,
	}, nil
}

// GetDataPrivacyRequest reads one of the requester's own privacy requests.
func (s *serviceImpl) GetDataPrivacyRequest(ctx context.Context, request *dataprivacysvc.GetDataPrivacyRequestRequest) (*dataprivacysvc.GetDataPrivacyRequestResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stored, err := s.ownedRequest(ctx, span, request.GetDataPrivacyRequestId())
	if err != nil {
		return nil, err
	}

	return &dataprivacysvc.GetDataPrivacyRequestResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Request:         converters.ConvertRequestToGRPCRequest(stored),
	}, nil
}

// ListDataPrivacyRequests pages through the requester's privacy requests.
//
// A subject is entitled to know what has been asked in their name, which is why the
// platform scopes List to a subject rather than offering a global one.
func (s *serviceImpl) ListDataPrivacyRequests(ctx context.Context, request *dataprivacysvc.ListDataPrivacyRequestsRequest) (*dataprivacysvc.ListDataPrivacyRequestsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.GetUserID()
	logger := s.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.GetFilter())
	tracing.AttachQueryFilterToSpan(span, filter)

	result, err := s.requests.List(ctx, subjectFor(userID), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "listing data privacy requests")
	}

	response := &dataprivacysvc.ListDataPrivacyRequestsResponse{
		ResponseDetails: &types.ResponseDetails{TraceId: span.SpanContext().TraceID().String()},
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(result.Pagination, filter),
	}

	for _, stored := range result.Data {
		response.Data = append(response.Data, converters.ConvertRequestToGRPCRequest(stored))
	}

	return response, nil
}

// submit records a request of the given type for the authenticated user.
func (s *serviceImpl) submit(ctx context.Context, span tracing.Span, requestType platformdataprivacy.RequestType) (*platformdataprivacy.Request, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.GetUserID()
	logger := s.logger.WithSpan(span).
		WithValue(identitykeys.UserIDKey, userID).
		WithValue(dataprivacykeys.RequestTypeKey, string(requestType))
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, dataprivacykeys.RequestTypeKey, string(requestType))

	request, err := s.requests.Submit(ctx, subjectFor(userID), requestType)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "submitting data privacy request")
	}

	logger.WithValue(dataprivacykeys.RequestIDKey, request.ID).Info("data privacy request submitted")

	return request, nil
}

// ownedRequest reads a request and refuses it if it belongs to somebody else.
//
// NotFound rather than PermissionDenied, in both the missing and the not-yours case.
// A distinct denial would confirm that a given request ID exists, and "is there a
// privacy request with this ID" is a question about somebody else's data.
func (s *serviceImpl) ownedRequest(ctx context.Context, span tracing.Span, requestID string) (*platformdataprivacy.Request, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	logger := s.logger.WithSpan(span).WithValue(dataprivacykeys.RequestIDKey, requestID)
	tracing.AttachToSpan(span, dataprivacykeys.RequestIDKey, requestID)

	request, err := s.requests.Get(ctx, requestID)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "fetching data privacy request")
	}

	if request.Subject.ID != sessionContextData.GetUserID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("data privacy request does not belong to requester"),
			logger, span, codes.NotFound, "fetching data privacy request",
		)
	}

	return request, nil
}

// subjectFor names a user as the subject of a request.
//
// Scope is left empty on purpose: an empty scope means every scope the subject
// appears in, which is what a plain "give me my data" asks for. A scoped request is
// the one that arrives when a business customer leaves one of several accounts, and
// this application has no route that expresses that yet.
func subjectFor(userID string) platformdataprivacy.Subject {
	return platformdataprivacy.Subject{
		ID:   userID,
		Type: platformdataprivacy.SubjectUser,
	}
}
