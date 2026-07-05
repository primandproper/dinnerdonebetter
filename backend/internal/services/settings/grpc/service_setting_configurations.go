package grpc

import (
	"context"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	settingskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	settingssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/settings/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v3/errors/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateServiceSettingConfiguration(ctx context.Context, request *settingssvc.CreateServiceSettingConfigurationRequest) (*settingssvc.CreateServiceSettingConfigurationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingIDKey, request.Input.ServiceSettingId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	belongsToUser := sessionContextData.GetUserID()
	belongsToAccount := sessionContextData.GetActiveAccountID()
	input := converters.ConvertGRPCServiceSettingConfigurationCreationRequestInputToServiceSettingConfigurationDatabaseCreationInput(request.Input, belongsToUser, belongsToAccount)

	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid service setting configuration")
	}

	created, err := s.settingsManager.CreateServiceSettingConfiguration(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create service setting")
	}

	x := &settingssvc.CreateServiceSettingConfigurationResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(created),
	}

	return x, nil
}

func (s *serviceImpl) GetServiceSettingConfigurationByName(ctx context.Context, request *settingssvc.GetServiceSettingConfigurationByNameRequest) (*settingssvc.GetServiceSettingConfigurationByNameResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	serviceSettingConfig, err := s.settingsManager.GetServiceSettingConfigurationForAccountByName(ctx, sessionContextData.ActiveAccountID, request.ServiceSettingConfigurationName)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service setting configuration for account by name")
	}

	x := &settingssvc.GetServiceSettingConfigurationByNameResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(serviceSettingConfig),
	}

	return x, nil
}

func (s *serviceImpl) GetServiceSettingConfigurationsForAccount(ctx context.Context, request *settingssvc.GetServiceSettingConfigurationsForAccountRequest) (*settingssvc.GetServiceSettingConfigurationsForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	serviceSettingConfigs, err := s.settingsManager.GetServiceSettingConfigurationsForAccount(ctx, sessionContextData.ActiveAccountID, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service setting configurations for account")
	}

	x := &settingssvc.GetServiceSettingConfigurationsForAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(serviceSettingConfigs.Pagination, filter),
	}

	for _, cfg := range serviceSettingConfigs.Data {
		x.Results = append(x.Results, converters.ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(cfg))
	}

	return x, nil
}

func (s *serviceImpl) GetServiceSettingConfigurationsForUser(ctx context.Context, request *settingssvc.GetServiceSettingConfigurationsForUserRequest) (*settingssvc.GetServiceSettingConfigurationsForUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	serviceSettingConfigs, err := s.settingsManager.GetServiceSettingConfigurationsForUser(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service setting configurations for user")
	}

	x := &settingssvc.GetServiceSettingConfigurationsForUserResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(serviceSettingConfigs.Pagination, filter),
	}

	for _, cfg := range serviceSettingConfigs.Data {
		x.Results = append(x.Results, converters.ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(cfg))
	}

	return x, nil
}

func (s *serviceImpl) UpdateServiceSettingConfiguration(ctx context.Context, request *settingssvc.UpdateServiceSettingConfigurationRequest) (*settingssvc.UpdateServiceSettingConfigurationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingConfigurationIDKey, request.ServiceSettingConfigurationId)

	existing, err := s.settingsManager.GetServiceSettingConfiguration(ctx, request.ServiceSettingConfigurationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to retrieve service setting configuration")
	}

	existing.Update(converters.ConvertGRPCServiceSettingConfigurationUpdateRequestInputToServiceSettingConfigurationUpdateRequestInputTo(request.Input))

	if err = s.settingsManager.UpdateServiceSettingConfiguration(ctx, existing); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update service setting configuration")
	}

	x := &settingssvc.UpdateServiceSettingConfigurationResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(existing),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveServiceSettingConfiguration(ctx context.Context, request *settingssvc.ArchiveServiceSettingConfigurationRequest) (*settingssvc.ArchiveServiceSettingConfigurationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingConfigurationIDKey, request.ServiceSettingConfigurationId)

	if err := s.settingsManager.ArchiveServiceSettingConfiguration(ctx, request.ServiceSettingConfigurationId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive service setting configuration")
	}

	x := &settingssvc.ArchiveServiceSettingConfigurationResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
