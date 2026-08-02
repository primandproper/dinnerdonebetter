package grpc

import (
	"context"

	dataprivacykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	errorsgrpc "github.com/primandproper/platform-go/v9/errors/grpc"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"google.golang.org/grpc/codes"
)

// GetUserDataDisclosure retrieves a single user data disclosure request belonging to the requester.
func (s *serviceImpl) GetUserDataDisclosure(ctx context.Context, request *dataprivacysvc.GetUserDataDisclosureRequest) (*dataprivacysvc.GetUserDataDisclosureResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	disclosureID := request.GetUserDataDisclosureId()
	logger := s.logger.WithValue(dataprivacykeys.UserDataDisclosureIDKey, disclosureID)
	tracing.AttachToSpan(span, dataprivacykeys.UserDataDisclosureIDKey, disclosureID)

	disclosure, err := s.dataPrivacyManager.GetUserDataDisclosure(ctx, disclosureID)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "fetching user data disclosure")
	}

	// Verify the disclosure belongs to the requester before returning it. NotFound avoids leaking the existence of
	// other users' disclosure requests.
	if disclosure.BelongsToUser != sessionContextData.Requester.UserID {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("disclosure does not belong to requester"), logger, span, codes.NotFound, "fetching user data disclosure")
	}

	return &dataprivacysvc.GetUserDataDisclosureResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserDataDisclosure: converters.ConvertUserDataDisclosureToGRPCUserDataDisclosure(disclosure),
	}, nil
}

// ListUserDataDisclosures lists the requester's user data disclosure requests.
func (s *serviceImpl) ListUserDataDisclosures(ctx context.Context, request *dataprivacysvc.ListUserDataDisclosuresRequest) (*dataprivacysvc.ListUserDataDisclosuresResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.Requester.UserID
	logger := s.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.GetFilter())
	tracing.AttachQueryFilterToSpan(span, filter)

	result, err := s.dataPrivacyManager.GetUserDataDisclosuresForUser(ctx, userID, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching user data disclosures")
	}

	response := &dataprivacysvc.ListUserDataDisclosuresResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(result.Pagination, filter),
	}
	for _, d := range result.Data {
		response.Data = append(response.Data, converters.ConvertUserDataDisclosureToGRPCUserDataDisclosure(d))
	}

	return response, nil
}
