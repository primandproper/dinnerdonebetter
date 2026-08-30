package grpc

import (
	"context"

	oauthkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/keys"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
	grpctypes "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	oauthgrpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) ArchiveOAuth2Client(ctx context.Context, request *oauthsvc.ArchiveOAuth2ClientRequest) (*oauthsvc.ArchiveOAuth2ClientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.oauthDataManager.ArchiveOAuth2Client(ctx, request.Oauth2ClientId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "archiving oauth2 client")
	}

	x := &oauthsvc.ArchiveOAuth2ClientResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) CreateOAuth2Client(ctx context.Context, request *oauthsvc.CreateOAuth2ClientRequest) (*oauthsvc.CreateOAuth2ClientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	input := oauthgrpcconverters.ConvertGRPCOAuth2ClientCreationRequestInputToOAuth2ClientCreationRequestInput(request.Input)

	created, err := s.oauthDataManager.CreateOAuth2Client(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating oauth2 client")
	}

	x := &oauthsvc.CreateOAuth2ClientResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: oauthgrpcconverters.ConvertOAuth2ClientToGRPCOAuth2Client(created),
	}

	return x, nil
}

func (s *serviceImpl) GetOAuth2Client(ctx context.Context, request *oauthsvc.GetOAuth2ClientRequest) (*oauthsvc.GetOAuth2ClientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(oauthkeys.OAuth2ClientIDKey, request.Oauth2ClientId)

	oauth2Client, err := s.oauthDataManager.GetOAuth2Client(ctx, request.Oauth2ClientId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "getting oauth2 client by database ID")
	}

	x := &oauthsvc.GetOAuth2ClientResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: oauthgrpcconverters.ConvertOAuth2ClientToGRPCOAuth2Client(oauth2Client),
	}

	return x, nil
}

func (s *serviceImpl) GetOAuth2Clients(ctx context.Context, request *oauthsvc.GetOAuth2ClientsRequest) (*oauthsvc.GetOAuth2ClientsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.Clone()
	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	oauth2Clients, err := s.oauthDataManager.GetOAuth2Clients(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "getting oauth2 client by database ID")
	}

	x := &oauthsvc.GetOAuth2ClientsResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(oauth2Clients.Pagination),
	}

	for _, client := range oauth2Clients.Data {
		x.Results = append(x.Results, oauthgrpcconverters.ConvertOAuth2ClientToGRPCOAuth2Client(client))
	}

	return x, nil
}
