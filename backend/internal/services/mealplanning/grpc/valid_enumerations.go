package grpc

import (
	"context"

	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	mealplanningconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	errorsgrpc "github.com/primandproper/platform-go/v6/errors/grpc"
	"github.com/primandproper/platform-go/v6/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) ArchiveValidIngredient(ctx context.Context, request *mealplanning.ArchiveValidIngredientRequest) (*mealplanning.ArchiveValidIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)

	if err := s.mealPlanningManager.ArchiveValidIngredient(ctx, request.ValidIngredientId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid ingredient")
	}

	res := &mealplanning.ArchiveValidIngredientResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidIngredientGroup(ctx context.Context, request *mealplanning.ArchiveValidIngredientGroupRequest) (*mealplanning.ArchiveValidIngredientGroupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientGroupIDKey, request.ValidIngredientGroupId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, request.ValidIngredientGroupId)

	if err := s.mealPlanningManager.ArchiveValidIngredientGroup(ctx, request.ValidIngredientGroupId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid ingredient group")
	}

	res := &mealplanning.ArchiveValidIngredientGroupResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidIngredientMeasurementUnit(ctx context.Context, request *mealplanning.ArchiveValidIngredientMeasurementUnitRequest) (*mealplanning.ArchiveValidIngredientMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, request.ValidIngredientMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, request.ValidIngredientMeasurementUnitId)

	if err := s.mealPlanningManager.ArchiveValidIngredientMeasurementUnit(ctx, request.ValidIngredientMeasurementUnitId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid ingredient measurement unit")
	}

	res := &mealplanning.ArchiveValidIngredientMeasurementUnitResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidIngredientPreparation(ctx context.Context, request *mealplanning.ArchiveValidIngredientPreparationRequest) (*mealplanning.ArchiveValidIngredientPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, request.ValidIngredientPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, request.ValidIngredientPreparationId)

	if err := s.mealPlanningManager.ArchiveValidIngredientPreparation(ctx, request.ValidIngredientPreparationId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid ingredient preparation")
	}

	res := &mealplanning.ArchiveValidIngredientPreparationResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidPrepTaskConfig(ctx context.Context, request *mealplanning.ArchiveValidPrepTaskConfigRequest) (*mealplanning.ArchiveValidPrepTaskConfigResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, request.ValidPrepTaskConfigId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, request.ValidPrepTaskConfigId)

	if err := s.mealPlanningManager.ArchiveValidPrepTaskConfig(ctx, request.ValidPrepTaskConfigId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid prep task config")
	}

	res := &mealplanning.ArchiveValidPrepTaskConfigResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidIngredientState(ctx context.Context, request *mealplanning.ArchiveValidIngredientStateRequest) (*mealplanning.ArchiveValidIngredientStateResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)

	if err := s.mealPlanningManager.ArchiveValidIngredientState(ctx, request.ValidIngredientStateId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid ingredient state")
	}

	res := &mealplanning.ArchiveValidIngredientStateResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidIngredientStateIngredient(ctx context.Context, request *mealplanning.ArchiveValidIngredientStateIngredientRequest) (*mealplanning.ArchiveValidIngredientStateIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, request.ValidIngredientStateIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, request.ValidIngredientStateIngredientId)

	if err := s.mealPlanningManager.ArchiveValidIngredientStateIngredient(ctx, request.ValidIngredientStateIngredientId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid ingredient state ingredient")
	}

	res := &mealplanning.ArchiveValidIngredientStateIngredientResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidInstrument(ctx context.Context, request *mealplanning.ArchiveValidInstrumentRequest) (*mealplanning.ArchiveValidInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)

	if err := s.mealPlanningManager.ArchiveValidInstrument(ctx, request.ValidInstrumentId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid instrument")
	}

	res := &mealplanning.ArchiveValidInstrumentResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidMeasurementUnit(ctx context.Context, request *mealplanning.ArchiveValidMeasurementUnitRequest) (*mealplanning.ArchiveValidMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)

	if err := s.mealPlanningManager.ArchiveValidMeasurementUnit(ctx, request.ValidMeasurementUnitId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid measurement unit")
	}

	res := &mealplanning.ArchiveValidMeasurementUnitResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidMeasurementUnitConversion(ctx context.Context, request *mealplanning.ArchiveValidMeasurementUnitConversionRequest) (*mealplanning.ArchiveValidMeasurementUnitConversionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, request.ValidMeasurementUnitConversionId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, request.ValidMeasurementUnitConversionId)

	if err := s.mealPlanningManager.ArchiveValidMeasurementUnitConversion(ctx, request.ValidMeasurementUnitConversionId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid measurement unit conversion")
	}

	res := &mealplanning.ArchiveValidMeasurementUnitConversionResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidPreparation(ctx context.Context, request *mealplanning.ArchiveValidPreparationRequest) (*mealplanning.ArchiveValidPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)

	if err := s.mealPlanningManager.ArchiveValidPreparation(ctx, request.ValidPreparationId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid preparation")
	}

	res := &mealplanning.ArchiveValidPreparationResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidPreparationInstrument(ctx context.Context, request *mealplanning.ArchiveValidPreparationInstrumentRequest) (*mealplanning.ArchiveValidPreparationInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, request.ValidPreparationInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, request.ValidPreparationInstrumentId)

	if err := s.mealPlanningManager.ArchiveValidPreparationInstrument(ctx, request.ValidPreparationInstrumentId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid preparation instrument")
	}

	res := &mealplanning.ArchiveValidPreparationInstrumentResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidPreparationVessel(ctx context.Context, request *mealplanning.ArchiveValidPreparationVesselRequest) (*mealplanning.ArchiveValidPreparationVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationVesselIDKey, request.ValidPreparationVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, request.ValidPreparationVesselId)

	if err := s.mealPlanningManager.ArchiveValidPreparationVessel(ctx, request.ValidPreparationVesselId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid preparation vessel")
	}

	res := &mealplanning.ArchiveValidPreparationVesselResponse{}

	return res, nil
}

func (s *serviceImpl) ArchiveValidVessel(ctx context.Context, request *mealplanning.ArchiveValidVesselRequest) (*mealplanning.ArchiveValidVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)

	if err := s.mealPlanningManager.ArchiveValidVessel(ctx, request.ValidVesselId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving valid vessel")
	}

	res := &mealplanning.ArchiveValidVesselResponse{}

	return res, nil
}

func (s *serviceImpl) CreateValidIngredient(ctx context.Context, request *mealplanning.CreateValidIngredientRequest) (*mealplanning.CreateValidIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidIngredient(ctx, mealplanningconverters.ConvertGRPCCreateValidIngredientRequestToValidIngredientCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid ingredient")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, created.ID)

	result := &mealplanning.CreateValidIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidIngredientGroup(ctx context.Context, request *mealplanning.CreateValidIngredientGroupRequest) (*mealplanning.CreateValidIngredientGroupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidIngredientGroup(ctx, mealplanningconverters.ConvertGRPCValidIngredientGroupCreationRequestInputToValidIngredientGroupCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid ingredient")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, created.ID)

	result := &mealplanning.CreateValidIngredientGroupResponse{
		Result: mealplanningconverters.ConvertValidIngredientGroupToGRPCValidIngredientGroup(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidIngredientMeasurementUnit(ctx context.Context, request *mealplanning.CreateValidIngredientMeasurementUnitRequest) (*mealplanning.CreateValidIngredientMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidIngredientMeasurementUnit(ctx, mealplanningconverters.ConvertGRPCCreateValidIngredientMeasurementUnitRequestToValidIngredientMeasurementUnitCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid ingredient measurement unit")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, created.ID)

	result := &mealplanning.CreateValidIngredientMeasurementUnitResponse{
		Result: mealplanningconverters.ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidIngredientPreparation(ctx context.Context, request *mealplanning.CreateValidIngredientPreparationRequest) (*mealplanning.CreateValidIngredientPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidIngredientPreparation(ctx, mealplanningconverters.ConvertGRPCCreateValidIngredientPreparationRequestToValidIngredientPreparationCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid ingredient preparation")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, created.ID)

	result := &mealplanning.CreateValidIngredientPreparationResponse{
		Result: mealplanningconverters.ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidPrepTaskConfig(ctx context.Context, request *mealplanning.CreateValidPrepTaskConfigRequest) (*mealplanning.CreateValidPrepTaskConfigResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidPrepTaskConfig(ctx, mealplanningconverters.ConvertGRPCValidPrepTaskConfigCreationRequestInputToValidPrepTaskConfigCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid prep task config")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, created.ID)

	result := &mealplanning.CreateValidPrepTaskConfigResponse{
		Result: mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidIngredientState(ctx context.Context, request *mealplanning.CreateValidIngredientStateRequest) (*mealplanning.CreateValidIngredientStateResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	input := mealplanningconverters.ConvertGRPCCreateValidIngredientStateRequestToValidIngredientStateCreationRequestInput(request.Input)
	created, err := s.mealPlanningManager.CreateValidIngredientState(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid ingredient state")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, created.ID)

	result := &mealplanning.CreateValidIngredientStateResponse{
		Result: mealplanningconverters.ConvertValidIngredientStateToGRPCValidIngredientState(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidIngredientStateIngredient(ctx context.Context, request *mealplanning.CreateValidIngredientStateIngredientRequest) (*mealplanning.CreateValidIngredientStateIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidIngredientStateIngredient(ctx, mealplanningconverters.ConvertGRPCCreateValidIngredientStateIngredientRequestToValidIngredientStateIngredientCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid ingredient state ingredient")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, created.ID)

	result := &mealplanning.CreateValidIngredientStateIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidInstrument(ctx context.Context, request *mealplanning.CreateValidInstrumentRequest) (*mealplanning.CreateValidInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidInstrument(ctx, mealplanningconverters.ConvertGRPCCreateValidInstrumentRequestToValidInstrumentCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid instrument")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, created.ID)

	result := &mealplanning.CreateValidInstrumentResponse{
		Result: mealplanningconverters.ConvertValidInstrumentToGRPCValidInstrument(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidMeasurementUnit(ctx context.Context, request *mealplanning.CreateValidMeasurementUnitRequest) (*mealplanning.CreateValidMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidMeasurementUnit(ctx, mealplanningconverters.ConvertGRPCValidMeasurementUnitCreationRequestInputToValidMeasurementUnitCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid measurement unit")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, created.ID)

	result := &mealplanning.CreateValidMeasurementUnitResponse{
		Result: mealplanningconverters.ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidMeasurementUnitConversion(ctx context.Context, request *mealplanning.CreateValidMeasurementUnitConversionRequest) (*mealplanning.CreateValidMeasurementUnitConversionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidMeasurementUnitConversion(ctx, mealplanningconverters.ConvertGRPCCreateValidMeasurementUnitConversionRequestToValidMeasurementUnitConversionCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid measurement unit conversion")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, created.ID)

	result := &mealplanning.CreateValidMeasurementUnitConversionResponse{
		Result: mealplanningconverters.ConvertValidMeasurementUnitConversionToGRPCValidMeasurementUnitConversion(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidPreparation(ctx context.Context, request *mealplanning.CreateValidPreparationRequest) (*mealplanning.CreateValidPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidPreparation(ctx, mealplanningconverters.ConvertGRPCValidPreparationCreationRequestInputToValidPreparationCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid preparation")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, created.ID)

	result := &mealplanning.CreateValidPreparationResponse{
		Result: mealplanningconverters.ConvertValidPreparationToGRPCValidPreparation(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidPreparationInstrument(ctx context.Context, request *mealplanning.CreateValidPreparationInstrumentRequest) (*mealplanning.CreateValidPreparationInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidPreparationInstrument(ctx, mealplanningconverters.ConvertGRPCCreateValidPreparationInstrumentRequestToValidPreparationInstrumentCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid preparation instrument")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, created.ID)

	result := &mealplanning.CreateValidPreparationInstrumentResponse{
		Result: mealplanningconverters.ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidPreparationVessel(ctx context.Context, request *mealplanning.CreateValidPreparationVesselRequest) (*mealplanning.CreateValidPreparationVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidPreparationVessel(ctx, mealplanningconverters.ConvertGRPCCreateValidPreparationVesselRequestToValidPreparationVesselCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid preparation vessel")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, created.ID)

	result := &mealplanning.CreateValidPreparationVesselResponse{
		Result: mealplanningconverters.ConvertValidPreparationVesselToGRPCValidPreparationVessel(created),
	}

	return result, nil
}

func (s *serviceImpl) CreateValidVessel(ctx context.Context, request *mealplanning.CreateValidVesselRequest) (*mealplanning.CreateValidVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	created, err := s.mealPlanningManager.CreateValidVessel(ctx, mealplanningconverters.ConvertGRPCValidVesselCreationRequestInputToValidVesselCreationRequestInput(request.Input))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating valid vessel")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, created.ID)

	result := &mealplanning.CreateValidVesselResponse{
		Result: mealplanningconverters.ConvertValidVesselToGRPCValidVessel(created),
	}

	return result, nil
}

func (s *serviceImpl) GetRandomValidIngredient(ctx context.Context, _ *mealplanning.GetRandomValidIngredientRequest) (*mealplanning.GetRandomValidIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	selected, err := s.mealPlanningManager.RandomValidIngredient(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching random valid ingredient")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, selected.ID)

	res := &mealplanning.GetRandomValidIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(selected),
	}

	return res, nil
}

func (s *serviceImpl) GetRandomValidInstrument(ctx context.Context, _ *mealplanning.GetRandomValidInstrumentRequest) (*mealplanning.GetRandomValidInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	selected, err := s.mealPlanningManager.RandomValidInstrument(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching random valid instrument")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, selected.ID)

	res := &mealplanning.GetRandomValidInstrumentResponse{
		Result: mealplanningconverters.ConvertValidInstrumentToGRPCValidInstrument(selected),
	}

	return res, nil
}

func (s *serviceImpl) GetRandomValidPreparation(ctx context.Context, _ *mealplanning.GetRandomValidPreparationRequest) (*mealplanning.GetRandomValidPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	selected, err := s.mealPlanningManager.RandomValidPreparation(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching random valid preparation")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, selected.ID)

	res := &mealplanning.GetRandomValidPreparationResponse{
		Result: mealplanningconverters.ConvertValidPreparationToGRPCValidPreparation(selected),
	}

	return res, nil
}

func (s *serviceImpl) GetRandomValidVessel(ctx context.Context, _ *mealplanning.GetRandomValidVesselRequest) (*mealplanning.GetRandomValidVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	selected, err := s.mealPlanningManager.RandomValidVessel(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching random valid vessel")
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, selected.ID)

	res := &mealplanning.GetRandomValidVesselResponse{
		Result: mealplanningconverters.ConvertValidVesselToGRPCValidVessel(selected),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredient(ctx context.Context, request *mealplanning.GetValidIngredientRequest) (*mealplanning.GetValidIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)

	x, err := s.mealPlanningManager.ReadValidIngredient(ctx, request.ValidIngredientId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient")
	}

	res := &mealplanning.GetValidIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientGroup(ctx context.Context, request *mealplanning.GetValidIngredientGroupRequest) (*mealplanning.GetValidIngredientGroupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientGroupIDKey, request.ValidIngredientGroupId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, request.ValidIngredientGroupId)

	x, err := s.mealPlanningManager.ReadValidIngredientGroup(ctx, request.ValidIngredientGroupId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient group")
	}

	res := &mealplanning.GetValidIngredientGroupResponse{
		Result: mealplanningconverters.ConvertValidIngredientGroupToGRPCValidIngredientGroup(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientGroups(ctx context.Context, request *mealplanning.GetValidIngredientGroupsRequest) (*mealplanning.GetValidIngredientGroupsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidIngredientGroups(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient groups")
	}

	res := &mealplanning.GetValidIngredientGroupsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}
	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientGroupToGRPCValidIngredientGroup(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientMeasurementUnit(ctx context.Context, request *mealplanning.GetValidIngredientMeasurementUnitRequest) (*mealplanning.GetValidIngredientMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, request.ValidIngredientMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, request.ValidIngredientMeasurementUnitId)

	x, err := s.mealPlanningManager.ReadValidIngredientMeasurementUnit(ctx, request.ValidIngredientMeasurementUnitId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient measurement unit")
	}

	res := &mealplanning.GetValidIngredientMeasurementUnitResponse{
		Result: mealplanningconverters.ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientMeasurementUnits(ctx context.Context, request *mealplanning.GetValidIngredientMeasurementUnitsRequest) (*mealplanning.GetValidIngredientMeasurementUnitsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidIngredientMeasurementUnits(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient measurement units")
	}

	res := &mealplanning.GetValidIngredientMeasurementUnitsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}
	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientMeasurementUnitsByIngredient(ctx context.Context, request *mealplanning.GetValidIngredientMeasurementUnitsByIngredientRequest) (*mealplanning.GetValidIngredientMeasurementUnitsByIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientMeasurementUnitsByIngredient(ctx, request.ValidIngredientId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient measurement units by ingredient")
	}

	res := &mealplanning.GetValidIngredientMeasurementUnitsByIngredientResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientMeasurementUnitsByMeasurementUnit(ctx context.Context, request *mealplanning.GetValidIngredientMeasurementUnitsByMeasurementUnitRequest) (*mealplanning.GetValidIngredientMeasurementUnitsByMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientMeasurementUnitsByMeasurementUnit(ctx, request.ValidMeasurementUnitId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient measurement units by measurement unit")
	}

	res := &mealplanning.GetValidIngredientMeasurementUnitsByMeasurementUnitResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientPreparation(ctx context.Context, request *mealplanning.GetValidIngredientPreparationRequest) (*mealplanning.GetValidIngredientPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, request.ValidIngredientPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, request.ValidIngredientPreparationId)

	x, err := s.mealPlanningManager.ReadValidIngredientPreparation(ctx, request.ValidIngredientPreparationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient preparation")
	}

	res := &mealplanning.GetValidIngredientPreparationResponse{
		Result: mealplanningconverters.ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientPreparations(ctx context.Context, request *mealplanning.GetValidIngredientPreparationsRequest) (*mealplanning.GetValidIngredientPreparationsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidIngredientPreparations(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient preparations")
	}

	res := &mealplanning.GetValidIngredientPreparationsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientPreparationsByIngredient(ctx context.Context, request *mealplanning.GetValidIngredientPreparationsByIngredientRequest) (*mealplanning.GetValidIngredientPreparationsByIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientPreparationsByIngredient(ctx, request.ValidIngredientId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient preparations by ingredient")
	}

	res := &mealplanning.GetValidIngredientPreparationsByIngredientResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientPreparationsByPreparation(ctx context.Context, request *mealplanning.GetValidIngredientPreparationsByPreparationRequest) (*mealplanning.GetValidIngredientPreparationsByPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientPreparationsByPreparation(ctx, request.ValidPreparationId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient preparations by preparation")
	}

	res := &mealplanning.GetValidIngredientPreparationsByPreparationResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPrepTaskConfig(ctx context.Context, request *mealplanning.GetValidPrepTaskConfigRequest) (*mealplanning.GetValidPrepTaskConfigResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, request.ValidPrepTaskConfigId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, request.ValidPrepTaskConfigId)

	x, err := s.mealPlanningManager.ReadValidPrepTaskConfig(ctx, request.ValidPrepTaskConfigId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid prep task config")
	}

	res := &mealplanning.GetValidPrepTaskConfigResponse{
		Result: mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidPrepTaskConfigs(ctx context.Context, request *mealplanning.GetValidPrepTaskConfigsRequest) (*mealplanning.GetValidPrepTaskConfigsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidPrepTaskConfigs(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid prep task configs")
	}

	res := &mealplanning.GetValidPrepTaskConfigsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPrepTaskConfigsByIngredient(ctx context.Context, request *mealplanning.GetValidPrepTaskConfigsByIngredientRequest) (*mealplanning.GetValidPrepTaskConfigsByIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPrepTaskConfigsByIngredient(ctx, request.ValidIngredientId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid prep task configs by ingredient")
	}

	res := &mealplanning.GetValidPrepTaskConfigsByIngredientResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPrepTaskConfigsByPreparation(ctx context.Context, request *mealplanning.GetValidPrepTaskConfigsByPreparationRequest) (*mealplanning.GetValidPrepTaskConfigsByPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPrepTaskConfigsByPreparation(ctx, request.ValidPreparationId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid prep task configs by preparation")
	}

	res := &mealplanning.GetValidPrepTaskConfigsByPreparationResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPrepTaskConfigsByIngredientAndPreparation(ctx context.Context, request *mealplanning.GetValidPrepTaskConfigsByIngredientAndPreparationRequest) (*mealplanning.GetValidPrepTaskConfigsByIngredientAndPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).
		WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId).
		WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPrepTaskConfigsByIngredientAndPreparation(ctx, request.ValidIngredientId, request.ValidPreparationId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid prep task configs by ingredient and preparation")
	}

	res := &mealplanning.GetValidPrepTaskConfigsByIngredientAndPreparationResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientState(ctx context.Context, request *mealplanning.GetValidIngredientStateRequest) (*mealplanning.GetValidIngredientStateResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)

	x, err := s.mealPlanningManager.ReadValidIngredientState(ctx, request.ValidIngredientStateId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient state")
	}

	res := &mealplanning.GetValidIngredientStateResponse{
		Result: mealplanningconverters.ConvertValidIngredientStateToGRPCValidIngredientState(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientStateIngredient(ctx context.Context, request *mealplanning.GetValidIngredientStateIngredientRequest) (*mealplanning.GetValidIngredientStateIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, request.ValidIngredientStateIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, request.ValidIngredientStateIngredientId)

	x, err := s.mealPlanningManager.ReadValidIngredientStateIngredient(ctx, request.ValidIngredientStateIngredientId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient state ingredient")
	}

	res := &mealplanning.GetValidIngredientStateIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientStateIngredients(ctx context.Context, request *mealplanning.GetValidIngredientStateIngredientsRequest) (*mealplanning.GetValidIngredientStateIngredientsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidIngredientStateIngredients(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient state ingredients")
	}

	res := &mealplanning.GetValidIngredientStateIngredientsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientStateIngredientsByIngredient(ctx context.Context, request *mealplanning.GetValidIngredientStateIngredientsByIngredientRequest) (*mealplanning.GetValidIngredientStateIngredientsByIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientStateIngredientsByIngredient(ctx, request.ValidIngredientId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient state ingredients by ingredient")
	}

	res := &mealplanning.GetValidIngredientStateIngredientsByIngredientResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientStateIngredientsByIngredientState(ctx context.Context, request *mealplanning.GetValidIngredientStateIngredientsByIngredientStateRequest) (*mealplanning.GetValidIngredientStateIngredientsByIngredientStateResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientStateIngredientsByIngredientState(ctx, request.ValidIngredientStateId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient state ingredients by ingredient state")
	}

	res := &mealplanning.GetValidIngredientStateIngredientsByIngredientStateResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredientStates(ctx context.Context, request *mealplanning.GetValidIngredientStatesRequest) (*mealplanning.GetValidIngredientStatesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidIngredientStates(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredient states")
	}

	res := &mealplanning.GetValidIngredientStatesResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientStateToGRPCValidIngredientState(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidIngredients(ctx context.Context, request *mealplanning.GetValidIngredientsRequest) (*mealplanning.GetValidIngredientsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidIngredients(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid ingredients")
	}

	logger.WithValue("pagination", x.Pagination).Info("Valid ingredients retrieved")

	res := &mealplanning.GetValidIngredientsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidInstrument(ctx context.Context, request *mealplanning.GetValidInstrumentRequest) (*mealplanning.GetValidInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)

	x, err := s.mealPlanningManager.ReadValidInstrument(ctx, request.ValidInstrumentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid instrument")
	}

	res := &mealplanning.GetValidInstrumentResponse{
		Result: mealplanningconverters.ConvertValidInstrumentToGRPCValidInstrument(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidInstruments(ctx context.Context, request *mealplanning.GetValidInstrumentsRequest) (*mealplanning.GetValidInstrumentsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidInstruments(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid instruments")
	}

	res := &mealplanning.GetValidInstrumentsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidInstrumentToGRPCValidInstrument(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidMeasurementUnit(ctx context.Context, request *mealplanning.GetValidMeasurementUnitRequest) (*mealplanning.GetValidMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)

	x, err := s.mealPlanningManager.ReadValidMeasurementUnit(ctx, request.ValidMeasurementUnitId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid measurement unit")
	}

	res := &mealplanning.GetValidMeasurementUnitResponse{
		Result: mealplanningconverters.ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidMeasurementUnitConversion(ctx context.Context, request *mealplanning.GetValidMeasurementUnitConversionRequest) (*mealplanning.GetValidMeasurementUnitConversionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, request.ValidMeasurementUnitConversionId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, request.ValidMeasurementUnitConversionId)

	x, err := s.mealPlanningManager.ReadValidMeasurementUnitConversion(ctx, request.ValidMeasurementUnitConversionId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid measurement unit conversion")
	}

	res := &mealplanning.GetValidMeasurementUnitConversionResponse{
		Result: mealplanningconverters.ConvertValidMeasurementUnitConversionToGRPCValidMeasurementUnitConversion(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidMeasurementUnitConversionsForUnit(ctx context.Context, request *mealplanning.GetValidMeasurementUnitConversionsForUnitRequest) (*mealplanning.GetValidMeasurementUnitConversionsForUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ValidMeasurementUnitConversionsForMeasurementUnit(ctx, request.ValidMeasurementUnitId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid measurement unit conversions to unit")
	}

	res := &mealplanning.GetValidMeasurementUnitConversionsForUnitResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}
	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidMeasurementUnitConversionToGRPCValidMeasurementUnitConversion(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, request *mealplanning.GetValidMeasurementUnitConversionsForIngredientsRequest) (*mealplanning.GetValidMeasurementUnitConversionsForIngredientsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	x, err := s.mealPlanningManager.GetValidMeasurementUnitConversionsForIngredients(ctx, request.ValidIngredientIds)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid measurement unit conversions for ingredients")
	}

	res := &mealplanning.GetValidMeasurementUnitConversionsForIngredientsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}
	for _, y := range x {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidMeasurementUnitConversionToGRPCValidMeasurementUnitConversion(y))
	}

	return res, nil
}

func (s *serviceImpl) GetMeasurementUnitConversionMismatches(ctx context.Context, request *mealplanning.GetMeasurementUnitConversionMismatchesRequest) (*mealplanning.GetMeasurementUnitConversionMismatchesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	x, err := s.mealPlanningManager.GetMeasurementUnitConversionMismatches(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching measurement unit conversion mismatches")
	}

	res := &mealplanning.GetMeasurementUnitConversionMismatchesResponse{}
	for _, y := range x {
		res.Mismatches = append(res.Mismatches, mealplanningconverters.ConvertMeasurementUnitConversionMismatchToGRPCMeasurementUnitConversionMismatch(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidMeasurementUnits(ctx context.Context, request *mealplanning.GetValidMeasurementUnitsRequest) (*mealplanning.GetValidMeasurementUnitsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidMeasurementUnits(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid measurement units")
	}

	res := &mealplanning.GetValidMeasurementUnitsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparation(ctx context.Context, request *mealplanning.GetValidPreparationRequest) (*mealplanning.GetValidPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)

	x, err := s.mealPlanningManager.ReadValidPreparation(ctx, request.ValidPreparationId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation")
	}

	res := &mealplanning.GetValidPreparationResponse{
		Result: mealplanningconverters.ConvertValidPreparationToGRPCValidPreparation(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationInstrument(ctx context.Context, request *mealplanning.GetValidPreparationInstrumentRequest) (*mealplanning.GetValidPreparationInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, request.ValidPreparationInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, request.ValidPreparationInstrumentId)

	x, err := s.mealPlanningManager.ReadValidPreparationInstrument(ctx, request.ValidPreparationInstrumentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation instrument")
	}

	res := &mealplanning.GetValidPreparationInstrumentResponse{
		Result: mealplanningconverters.ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationInstruments(ctx context.Context, request *mealplanning.GetValidPreparationInstrumentsRequest) (*mealplanning.GetValidPreparationInstrumentsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidPreparationInstruments(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation instruments")
	}

	res := &mealplanning.GetValidPreparationInstrumentsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationInstrumentsByInstrument(ctx context.Context, request *mealplanning.GetValidPreparationInstrumentsByInstrumentRequest) (*mealplanning.GetValidPreparationInstrumentsByInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPreparationInstrumentsByInstrument(ctx, request.ValidInstrumentId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation instruments by instrument")
	}

	res := &mealplanning.GetValidPreparationInstrumentsByInstrumentResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationInstrumentsByPreparation(ctx context.Context, request *mealplanning.GetValidPreparationInstrumentsByPreparationRequest) (*mealplanning.GetValidPreparationInstrumentsByPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPreparationInstrumentsByPreparation(ctx, request.ValidPreparationId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation instruments by preparation")
	}

	res := &mealplanning.GetValidPreparationInstrumentsByPreparationResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationVessel(ctx context.Context, request *mealplanning.GetValidPreparationVesselRequest) (*mealplanning.GetValidPreparationVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationVesselIDKey, request.ValidPreparationVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, request.ValidPreparationVesselId)

	x, err := s.mealPlanningManager.ReadValidPreparationVessel(ctx, request.ValidPreparationVesselId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation vessel")
	}

	res := &mealplanning.GetValidPreparationVesselResponse{
		Result: mealplanningconverters.ConvertValidPreparationVesselToGRPCValidPreparationVessel(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationVessels(ctx context.Context, request *mealplanning.GetValidPreparationVesselsRequest) (*mealplanning.GetValidPreparationVesselsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidPreparationVessels(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation vessels")
	}

	res := &mealplanning.GetValidPreparationVesselsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationVesselToGRPCValidPreparationVessel(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationVesselsByPreparation(ctx context.Context, request *mealplanning.GetValidPreparationVesselsByPreparationRequest) (*mealplanning.GetValidPreparationVesselsByPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPreparationVesselsByPreparation(ctx, request.ValidPreparationId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation vessels by preparation")
	}

	res := &mealplanning.GetValidPreparationVesselsByPreparationResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationVesselToGRPCValidPreparationVessel(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparationVesselsByVessel(ctx context.Context, request *mealplanning.GetValidPreparationVesselsByVesselRequest) (*mealplanning.GetValidPreparationVesselsByVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPreparationVesselsByVessel(ctx, request.ValidVesselId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparation vessels by vessel")
	}

	res := &mealplanning.GetValidPreparationVesselsByVesselResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationVesselToGRPCValidPreparationVessel(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidPreparations(ctx context.Context, request *mealplanning.GetValidPreparationsRequest) (*mealplanning.GetValidPreparationsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidPreparations(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid preparations")
	}

	res := &mealplanning.GetValidPreparationsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationToGRPCValidPreparation(y))
	}

	return res, nil
}

func (s *serviceImpl) GetValidVessel(ctx context.Context, request *mealplanning.GetValidVesselRequest) (*mealplanning.GetValidVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)

	x, err := s.mealPlanningManager.ReadValidVessel(ctx, request.ValidVesselId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid vessel")
	}

	res := &mealplanning.GetValidVesselResponse{
		Result: mealplanningconverters.ConvertValidVesselToGRPCValidVessel(x),
	}

	return res, nil
}

func (s *serviceImpl) GetValidVessels(ctx context.Context, request *mealplanning.GetValidVesselsRequest) (*mealplanning.GetValidVesselsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.ListValidVessels(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching valid vessels")
	}

	res := &mealplanning.GetValidVesselsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidVesselToGRPCValidVessel(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidIngredientGroups(ctx context.Context, request *mealplanning.SearchForValidIngredientGroupsRequest) (*mealplanning.SearchForValidIngredientGroupsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientGroups(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid ingredient groups")
	}

	res := &mealplanning.SearchForValidIngredientGroupsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientGroupToGRPCValidIngredientGroup(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidIngredientStates(ctx context.Context, request *mealplanning.SearchForValidIngredientStatesRequest) (*mealplanning.SearchForValidIngredientStatesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientStates(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid ingredient states")
	}

	res := &mealplanning.SearchForValidIngredientStatesResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientStateToGRPCValidIngredientState(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidIngredients(ctx context.Context, request *mealplanning.SearchForValidIngredientsRequest) (*mealplanning.SearchForValidIngredientsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredients(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid ingredients")
	}

	res := &mealplanning.SearchForValidIngredientsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidInstruments(ctx context.Context, request *mealplanning.SearchForValidInstrumentsRequest) (*mealplanning.SearchForValidInstrumentsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidInstruments(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid instruments")
	}

	res := &mealplanning.SearchForValidInstrumentsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidInstrumentToGRPCValidInstrument(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidMeasurementUnits(ctx context.Context, request *mealplanning.SearchForValidMeasurementUnitsRequest) (*mealplanning.SearchForValidMeasurementUnitsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidMeasurementUnits(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid measurement units")
	}

	res := &mealplanning.SearchForValidMeasurementUnitsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidPreparations(ctx context.Context, request *mealplanning.SearchForValidPreparationsRequest) (*mealplanning.SearchForValidPreparationsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidPreparations(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid preparations")
	}

	res := &mealplanning.SearchForValidPreparationsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidPreparationToGRPCValidPreparation(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchForValidVessels(ctx context.Context, request *mealplanning.SearchForValidVesselsRequest) (*mealplanning.SearchForValidVesselsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidVessels(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid vessels")
	}

	res := &mealplanning.SearchForValidVesselsResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidVesselToGRPCValidVessel(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchValidIngredientsByPreparation(ctx context.Context, request *mealplanning.SearchValidIngredientsByPreparationRequest) (*mealplanning.SearchValidIngredientsByPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidIngredientsByPreparationAndIngredientName(ctx, request.ValidPreparationId, request.Query, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid ingredients by preparation")
	}

	res := &mealplanning.SearchValidIngredientsByPreparationResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(y))
	}

	return res, nil
}

func (s *serviceImpl) SearchValidMeasurementUnitsByIngredient(ctx context.Context, request *mealplanning.SearchValidMeasurementUnitsByIngredientRequest) (*mealplanning.SearchValidMeasurementUnitsByIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	x, err := s.mealPlanningManager.SearchValidMeasurementUnitsByIngredientID(ctx, request.ValidIngredientId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid measurement units by ingredient")
	}

	res := &mealplanning.SearchValidMeasurementUnitsByIngredientResponse{
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(x.Pagination, filter),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, mealplanningconverters.ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(y))
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidIngredient(ctx context.Context, request *mealplanning.UpdateValidIngredientRequest) (*mealplanning.UpdateValidIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, request.ValidIngredientId)

	input := mealplanningconverters.ConvertGRPCValidIngredientUpdateRequestInputToValidIngredientUpdateRequestInput(request.Input)
	updated, err := s.mealPlanningManager.UpdateValidIngredient(ctx, request.ValidIngredientId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid ingredient")
	}

	res := &mealplanning.UpdateValidIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientToGRPCValidIngredient(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidIngredientGroup(ctx context.Context, request *mealplanning.UpdateValidIngredientGroupRequest) (*mealplanning.UpdateValidIngredientGroupResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientGroupIDKey, request.ValidIngredientGroupId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, request.ValidIngredientGroupId)

	input := mealplanningconverters.ConvertGRPCValidIngredientGroupUpdateRequestInputToValidIngredientGroupUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidIngredientGroup(ctx, request.ValidIngredientGroupId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid ingredient group")
	}

	res := &mealplanning.UpdateValidIngredientGroupResponse{
		Result: mealplanningconverters.ConvertValidIngredientGroupToGRPCValidIngredientGroup(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidIngredientMeasurementUnit(ctx context.Context, request *mealplanning.UpdateValidIngredientMeasurementUnitRequest) (*mealplanning.UpdateValidIngredientMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, request.ValidIngredientMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, request.ValidIngredientMeasurementUnitId)

	input := mealplanningconverters.ConvertGRPCValidIngredientMeasurementUnitUpdateRequestInputToValidIngredientMeasurementUnitUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidIngredientMeasurementUnit(ctx, request.ValidIngredientMeasurementUnitId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid ingredient measurement unit")
	}

	res := &mealplanning.UpdateValidIngredientMeasurementUnitResponse{
		Result: mealplanningconverters.ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidIngredientPreparation(ctx context.Context, request *mealplanning.UpdateValidIngredientPreparationRequest) (*mealplanning.UpdateValidIngredientPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, request.ValidIngredientPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, request.ValidIngredientPreparationId)

	input := mealplanningconverters.ConvertGRPCValidIngredientPreparationUpdateRequestInputToValidIngredientPreparationUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidIngredientPreparation(ctx, request.ValidIngredientPreparationId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid ingredient preparation")
	}

	res := &mealplanning.UpdateValidIngredientPreparationResponse{
		Result: mealplanningconverters.ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidPrepTaskConfig(ctx context.Context, request *mealplanning.UpdateValidPrepTaskConfigRequest) (*mealplanning.UpdateValidPrepTaskConfigResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, request.ValidPrepTaskConfigId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, request.ValidPrepTaskConfigId)

	input := mealplanningconverters.ConvertGRPCValidPrepTaskConfigUpdateRequestInputToValidPrepTaskConfigUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidPrepTaskConfig(ctx, request.ValidPrepTaskConfigId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid prep task config")
	}

	res := &mealplanning.UpdateValidPrepTaskConfigResponse{
		Result: mealplanningconverters.ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidIngredientState(ctx context.Context, request *mealplanning.UpdateValidIngredientStateRequest) (*mealplanning.UpdateValidIngredientStateResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, request.ValidIngredientStateId)

	input := mealplanningconverters.ConvertGRPCValidIngredientStateUpdateRequestInputToValidIngredientStateUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidIngredientState(ctx, request.ValidIngredientStateId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid ingredient state")
	}

	res := &mealplanning.UpdateValidIngredientStateResponse{
		Result: mealplanningconverters.ConvertValidIngredientStateToGRPCValidIngredientState(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidIngredientStateIngredient(ctx context.Context, request *mealplanning.UpdateValidIngredientStateIngredientRequest) (*mealplanning.UpdateValidIngredientStateIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, request.ValidIngredientStateIngredientId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, request.ValidIngredientStateIngredientId)

	input := mealplanningconverters.ConvertGRPCValidIngredientStateIngredientUpdateRequestInputToValidIngredientStateIngredientUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidIngredientStateIngredient(ctx, request.ValidIngredientStateIngredientId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid ingredient state ingredient")
	}

	res := &mealplanning.UpdateValidIngredientStateIngredientResponse{
		Result: mealplanningconverters.ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidInstrument(ctx context.Context, request *mealplanning.UpdateValidInstrumentRequest) (*mealplanning.UpdateValidInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, request.ValidInstrumentId)

	input := mealplanningconverters.ConvertGRPCValidInstrumentUpdateRequestInputToValidInstrumentUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidInstrument(ctx, request.ValidInstrumentId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid instrument")
	}

	res := &mealplanning.UpdateValidInstrumentResponse{
		Result: mealplanningconverters.ConvertValidInstrumentToGRPCValidInstrument(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidMeasurementUnit(ctx context.Context, request *mealplanning.UpdateValidMeasurementUnitRequest) (*mealplanning.UpdateValidMeasurementUnitResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, request.ValidMeasurementUnitId)

	input := mealplanningconverters.ConvertGRPCValidMeasurementUnitUpdateRequestInputToValidMeasurementUnitUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidMeasurementUnit(ctx, request.ValidMeasurementUnitId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid measurement unit")
	}

	res := &mealplanning.UpdateValidMeasurementUnitResponse{
		Result: mealplanningconverters.ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidMeasurementUnitConversion(ctx context.Context, request *mealplanning.UpdateValidMeasurementUnitConversionRequest) (*mealplanning.UpdateValidMeasurementUnitConversionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, request.ValidMeasurementUnitConversionId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, request.ValidMeasurementUnitConversionId)

	input := mealplanningconverters.ConvertGRPCValidMeasurementUnitConversionUpdateRequestInputToValidMeasurementUnitConversionUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidMeasurementUnitConversion(ctx, request.ValidMeasurementUnitConversionId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid measurement unit conversion")
	}

	res := &mealplanning.UpdateValidMeasurementUnitConversionResponse{
		Result: mealplanningconverters.ConvertValidMeasurementUnitConversionToGRPCValidMeasurementUnitConversion(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidPreparation(ctx context.Context, request *mealplanning.UpdateValidPreparationRequest) (*mealplanning.UpdateValidPreparationResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, request.ValidPreparationId)

	input := mealplanningconverters.ConvertGRPCValidPreparationUpdateRequestInputToValidPreparationUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidPreparation(ctx, request.ValidPreparationId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid preparation")
	}

	res := &mealplanning.UpdateValidPreparationResponse{
		Result: mealplanningconverters.ConvertValidPreparationToGRPCValidPreparation(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidPreparationInstrument(ctx context.Context, request *mealplanning.UpdateValidPreparationInstrumentRequest) (*mealplanning.UpdateValidPreparationInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, request.ValidPreparationInstrumentId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, request.ValidPreparationInstrumentId)

	input := mealplanningconverters.ConvertGRPCValidPreparationInstrumentUpdateRequestInputToValidPreparationInstrumentUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidPreparationInstrument(ctx, request.ValidPreparationInstrumentId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid preparation instrument")
	}

	res := &mealplanning.UpdateValidPreparationInstrumentResponse{
		Result: mealplanningconverters.ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidPreparationVessel(ctx context.Context, request *mealplanning.UpdateValidPreparationVesselRequest) (*mealplanning.UpdateValidPreparationVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationVesselIDKey, request.ValidPreparationVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, request.ValidPreparationVesselId)

	input := mealplanningconverters.ConvertGRPCValidPreparationVesselUpdateRequestInputToValidPreparationVesselUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidPreparationVessel(ctx, request.ValidPreparationVesselId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid preparation vessel")
	}

	res := &mealplanning.UpdateValidPreparationVesselResponse{
		Result: mealplanningconverters.ConvertValidPreparationVesselToGRPCValidPreparationVessel(updated),
	}

	return res, nil
}

func (s *serviceImpl) UpdateValidVessel(ctx context.Context, request *mealplanning.UpdateValidVesselRequest) (*mealplanning.UpdateValidVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, request.ValidVesselId)

	input := mealplanningconverters.ConvertGRPCValidVesselUpdateRequestInputToValidVesselUpdateRequestInput(request.Input)

	updated, err := s.mealPlanningManager.UpdateValidVessel(ctx, request.ValidVesselId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating valid vessel")
	}

	res := &mealplanning.UpdateValidVesselResponse{
		Result: mealplanningconverters.ConvertValidVesselToGRPCValidVessel(updated),
	}

	return res, nil
}
