package grpc

import (
	"context"

	settingskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v11/errors/grpc"
	platformkeys "github.com/primandproper/platform-go/v11/observability/keys"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateServiceSetting(ctx context.Context, request *settingssvc.CreateServiceSettingRequest) (*settingssvc.CreateServiceSettingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	input := converters.ConvertGPRCServiceSettingCreationRequestInputToServiceSettingDatabaseCreationInput(request.Input)
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid service setting")
	}

	created, err := s.settingsManager.CreateServiceSetting(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create service setting")
	}

	x := &settingssvc.CreateServiceSettingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertServiceSettingToGRPCServiceSetting(created),
	}

	return x, nil
}

func (s *serviceImpl) GetServiceSetting(ctx context.Context, request *settingssvc.GetServiceSettingRequest) (*settingssvc.GetServiceSettingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingIDKey, request.ServiceSettingId)

	serviceSetting, err := s.settingsManager.GetServiceSetting(ctx, request.ServiceSettingId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service setting")
	}

	x := &settingssvc.GetServiceSettingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertServiceSettingToGRPCServiceSetting(serviceSetting),
	}

	return x, nil
}

func (s *serviceImpl) GetServiceSettings(ctx context.Context, request *settingssvc.GetServiceSettingsRequest) (*settingssvc.GetServiceSettingsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	serviceSettings, err := s.settingsManager.GetServiceSettings(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service settings")
	}

	x := &settingssvc.GetServiceSettingsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(serviceSettings.Pagination, filter),
	}

	for _, serviceSetting := range serviceSettings.Data {
		x.Results = append(x.Results, converters.ConvertServiceSettingToGRPCServiceSetting(serviceSetting))
	}

	return x, nil
}

func (s *serviceImpl) SearchForServiceSettings(ctx context.Context, request *settingssvc.SearchForServiceSettingsRequest) (*settingssvc.SearchForServiceSettingsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, request.Query)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	serviceSettings, err := s.settingsManager.SearchForServiceSettings(ctx, request.Query, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service settings")
	}

	x := &settingssvc.SearchForServiceSettingsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(serviceSettings.Pagination, filter),
	}

	for _, serviceSetting := range serviceSettings.Data {
		x.Results = append(x.Results, converters.ConvertServiceSettingToGRPCServiceSetting(serviceSetting))
	}

	return x, nil
}

func (s *serviceImpl) ArchiveServiceSetting(ctx context.Context, request *settingssvc.ArchiveServiceSettingRequest) (*settingssvc.ArchiveServiceSettingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingIDKey, request.ServiceSettingId)

	if err := s.settingsManager.ArchiveServiceSetting(ctx, request.ServiceSettingId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive service setting")
	}

	x := &settingssvc.ArchiveServiceSettingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
