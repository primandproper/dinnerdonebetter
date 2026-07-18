package grpc

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mockmanagers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanninggrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/fake"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildServiceImplForTest(t *testing.T) *serviceImpl {
	t.Helper()

	return &serviceImpl{
		tracer: tracing.NewTracerForTest(t.Name()),
		logger: loggingnoop.NewLogger(),
	}
}

func TestServiceImpl_ArchiveValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidIngredient), testutils.ContextMatcher, exampleValidIngredientID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredient(ctx, &mealplanninggrpc.ArchiveValidIngredientRequest{ValidIngredientId: exampleValidIngredientID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientGroupID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidIngredientGroup), testutils.ContextMatcher, exampleValidIngredientGroupID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientGroup(ctx, &mealplanninggrpc.ArchiveValidIngredientGroupRequest{ValidIngredientGroupId: exampleValidIngredientGroupID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientMeasurementUnitID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidIngredientMeasurementUnit), testutils.ContextMatcher, exampleValidIngredientMeasurementUnitID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientMeasurementUnit(ctx, &mealplanninggrpc.ArchiveValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: exampleValidIngredientMeasurementUnitID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientPreparationID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidIngredientPreparation), testutils.ContextMatcher, exampleValidIngredientPreparationID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientPreparation(ctx, &mealplanninggrpc.ArchiveValidIngredientPreparationRequest{ValidIngredientPreparationId: exampleValidIngredientPreparationID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPrepTaskConfigID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidPrepTaskConfig), testutils.ContextMatcher, exampleValidPrepTaskConfigID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPrepTaskConfig(ctx, &mealplanninggrpc.ArchiveValidPrepTaskConfigRequest{ValidPrepTaskConfigId: exampleValidPrepTaskConfigID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientStateID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidIngredientState), testutils.ContextMatcher, exampleValidIngredientStateID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientState(ctx, &mealplanninggrpc.ArchiveValidIngredientStateRequest{ValidIngredientStateId: exampleValidIngredientStateID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientStateIngredientID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidIngredientStateIngredient), testutils.ContextMatcher, exampleValidIngredientStateIngredientID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientStateIngredient(ctx, &mealplanninggrpc.ArchiveValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: exampleValidIngredientStateIngredientID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidInstrumentID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidInstrument), testutils.ContextMatcher, exampleValidInstrumentID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidInstrument(ctx, &mealplanninggrpc.ArchiveValidInstrumentRequest{ValidInstrumentId: exampleValidInstrumentID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnitID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidMeasurementUnit), testutils.ContextMatcher, exampleValidMeasurementUnitID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidMeasurementUnit(ctx, &mealplanninggrpc.ArchiveValidMeasurementUnitRequest{ValidMeasurementUnitId: exampleValidMeasurementUnitID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnitConversionID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidMeasurementUnitConversion), testutils.ContextMatcher, exampleValidMeasurementUnitConversionID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidMeasurementUnitConversion(ctx, &mealplanninggrpc.ArchiveValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: exampleValidMeasurementUnitConversionID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidPreparation), testutils.ContextMatcher, exampleValidPreparationID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPreparation(ctx, &mealplanninggrpc.ArchiveValidPreparationRequest{ValidPreparationId: exampleValidPreparationID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationInstrumentID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidPreparationInstrument), testutils.ContextMatcher, exampleValidPreparationInstrumentID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPreparationInstrument(ctx, &mealplanninggrpc.ArchiveValidPreparationInstrumentRequest{ValidPreparationInstrumentId: exampleValidPreparationInstrumentID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationVesselID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidPreparationVessel), testutils.ContextMatcher, exampleValidPreparationVesselID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPreparationVessel(ctx, &mealplanninggrpc.ArchiveValidPreparationVesselRequest{ValidPreparationVesselId: exampleValidPreparationVesselID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_ArchiveValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidVesselID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ArchiveValidVessel), testutils.ContextMatcher, exampleValidVesselID).Return(nil)
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidVessel(ctx, &mealplanninggrpc.ArchiveValidVesselRequest{ValidVesselId: exampleValidVesselID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredient := mealplanningfakes.BuildFakeValidIngredient()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidIngredient), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidIngredientCreationRequestInput]()).Return(exampleValidIngredient, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientRequest](t)

		actual, err := s.CreateValidIngredient(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestServiceImpl_CreateValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientGroup := mealplanningfakes.BuildFakeValidIngredientGroup()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidIngredientGroup), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidIngredientGroupCreationRequestInput]()).Return(exampleValidIngredientGroup, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientGroupRequest](t)

		actual, err := s.CreateValidIngredientGroup(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientMeasurementUnit := mealplanningfakes.BuildFakeValidIngredientMeasurementUnit()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidIngredientMeasurementUnit), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidIngredientMeasurementUnitCreationRequestInput]()).Return(exampleValidIngredientMeasurementUnit, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientMeasurementUnitRequest](t)

		actual, err := s.CreateValidIngredientMeasurementUnit(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientPreparation := mealplanningfakes.BuildFakeValidIngredientPreparation()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidIngredientPreparation), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidIngredientPreparationCreationRequestInput]()).Return(exampleValidIngredientPreparation, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientPreparationRequest](t)

		actual, err := s.CreateValidIngredientPreparation(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPrepTaskConfig := mealplanningfakes.BuildFakeValidPrepTaskConfig()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidPrepTaskConfig), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidPrepTaskConfigCreationRequestInput]()).Return(exampleValidPrepTaskConfig, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPrepTaskConfigRequest](t)

		actual, err := s.CreateValidPrepTaskConfig(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientState := mealplanningfakes.BuildFakeValidIngredientState()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidIngredientState), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidIngredientStateCreationRequestInput]()).Return(exampleValidIngredientState, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientStateRequest](t)

		actual, err := s.CreateValidIngredientState(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}
func TestServiceImpl_CreateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientStateIngredient := mealplanningfakes.BuildFakeValidIngredientStateIngredient()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidIngredientStateIngredient), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidIngredientStateIngredientCreationRequestInput]()).Return(exampleValidIngredientStateIngredient, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientStateIngredientRequest](t)

		actual, err := s.CreateValidIngredientStateIngredient(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidInstrument := mealplanningfakes.BuildFakeValidInstrument()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidInstrument), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidInstrumentCreationRequestInput]()).Return(exampleValidInstrument, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidInstrumentRequest](t)

		actual, err := s.CreateValidInstrument(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnit := mealplanningfakes.BuildFakeValidMeasurementUnit()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidMeasurementUnit), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidMeasurementUnitCreationRequestInput]()).Return(exampleValidMeasurementUnit, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidMeasurementUnitRequest](t)

		actual, err := s.CreateValidMeasurementUnit(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnitConversion := mealplanningfakes.BuildFakeValidMeasurementUnitConversion()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidMeasurementUnitConversion), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidMeasurementUnitConversionCreationRequestInput]()).Return(exampleValidMeasurementUnitConversion, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidMeasurementUnitConversionRequest](t)

		actual, err := s.CreateValidMeasurementUnitConversion(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparation := mealplanningfakes.BuildFakeValidPreparation()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidPreparation), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidPreparationCreationRequestInput]()).Return(exampleValidPreparation, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPreparationRequest](t)

		actual, err := s.CreateValidPreparation(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationInstrument := mealplanningfakes.BuildFakeValidPreparationInstrument()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidPreparationInstrument), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidPreparationInstrumentCreationRequestInput]()).Return(exampleValidPreparationInstrument, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPreparationInstrumentRequest](t)

		actual, err := s.CreateValidPreparationInstrument(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationVessel := mealplanningfakes.BuildFakeValidPreparationVessel()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidPreparationVessel), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidPreparationVesselCreationRequestInput]()).Return(exampleValidPreparationVessel, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPreparationVesselRequest](t)

		actual, err := s.CreateValidPreparationVessel(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_CreateValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidVessel := mealplanningfakes.BuildFakeValidVessel()

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.CreateValidVessel), testutils.ContextMatcher, testutils.MatchType[*mealplanning.ValidVesselCreationRequestInput]()).Return(exampleValidVessel, nil)
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidVesselRequest](t)

		actual, err := s.CreateValidVessel(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetRandomValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredient()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.RandomValidIngredient), testutils.ContextMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidIngredient(ctx, &mealplanninggrpc.GetRandomValidIngredientRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetRandomValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrument()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.RandomValidInstrument), testutils.ContextMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidInstrument(ctx, &mealplanninggrpc.GetRandomValidInstrumentRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetRandomValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparation()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.RandomValidPreparation), testutils.ContextMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidPreparation(ctx, &mealplanninggrpc.GetRandomValidPreparationRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetRandomValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVessel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.RandomValidVessel), testutils.ContextMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidVessel(ctx, &mealplanninggrpc.GetRandomValidVesselRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredient()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidIngredient), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredient(ctx, &mealplanninggrpc.GetValidIngredientRequest{ValidIngredientId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientGroup()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidIngredientGroup), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientGroup(ctx, &mealplanninggrpc.GetValidIngredientGroupRequest{ValidIngredientGroupId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientGroups(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientGroupsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidIngredientGroups), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientGroups(ctx, &mealplanninggrpc.GetValidIngredientGroupsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientMeasurementUnit()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidIngredientMeasurementUnit), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnit(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientMeasurementUnitsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidIngredientMeasurementUnits), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnits(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientMeasurementUnitsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidIngredientMeasurementUnitsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientMeasurementUnitsByIngredient), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnitsByIngredient(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientMeasurementUnitsByMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidIngredientMeasurementUnitsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientMeasurementUnitsByMeasurementUnit), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnitsByMeasurementUnit(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitsByMeasurementUnitRequest{
			ValidMeasurementUnitId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientPreparation()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidIngredientPreparation), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparation(ctx, &mealplanninggrpc.GetValidIngredientPreparationRequest{ValidIngredientPreparationId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientPreparationsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidIngredientPreparations), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparations(ctx, &mealplanninggrpc.GetValidIngredientPreparationsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientPreparationsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidIngredientPreparationsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientPreparationsByIngredient), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparationsByIngredient(ctx, &mealplanninggrpc.GetValidIngredientPreparationsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientPreparationsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidIngredientPreparationsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientPreparationsByPreparation), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparationsByPreparation(ctx, &mealplanninggrpc.GetValidIngredientPreparationsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfig()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidPrepTaskConfig), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfig(ctx, &mealplanninggrpc.GetValidPrepTaskConfigRequest{ValidPrepTaskConfigId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPrepTaskConfigs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfigsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidPrepTaskConfigs), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigs(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPrepTaskConfigsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfigsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPrepTaskConfigsByIngredient), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigsByIngredient(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPrepTaskConfigsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfigsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPrepTaskConfigsByPreparation), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigsByPreparation(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPrepTaskConfigsByIngredientAndPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleIngredientID := mealplanningfakes.BuildFakeID()
		examplePreparationID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfigsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPrepTaskConfigsByIngredientAndPreparation), testutils.ContextMatcher, exampleIngredientID, examplePreparationID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigsByIngredientAndPreparation(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsByIngredientAndPreparationRequest{
			ValidIngredientId:  exampleIngredientID,
			ValidPreparationId: examplePreparationID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientState()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidIngredientState), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientState(ctx, &mealplanninggrpc.GetValidIngredientStateRequest{ValidIngredientStateId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStateIngredient()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidIngredientStateIngredient), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredient(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientStateIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStateIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidIngredientStateIngredients), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredients(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientStateIngredientsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidIngredientStateIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientStateIngredientsByIngredient), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredientsByIngredient(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientStateIngredientsByIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidIngredientStateIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientStateIngredientsByIngredientState), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredientsByIngredientState(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientsByIngredientStateRequest{
			ValidIngredientStateId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredientStates(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStatesList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidIngredientStates), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStates(ctx, &mealplanninggrpc.GetValidIngredientStatesRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidIngredients), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredients(ctx, &mealplanninggrpc.GetValidIngredientsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrument()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidInstrument), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidInstrument(ctx, &mealplanninggrpc.GetValidInstrumentRequest{ValidInstrumentId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidInstruments), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidInstruments(ctx, &mealplanninggrpc.GetValidInstrumentsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnit()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidMeasurementUnit), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnit(ctx, &mealplanninggrpc.GetValidMeasurementUnitRequest{ValidMeasurementUnitId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitConversion()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidMeasurementUnitConversion), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnitConversion(ctx, &mealplanninggrpc.GetValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidMeasurementUnitConversionsFromUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitConversionsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ValidMeasurementUnitConversionsForMeasurementUnit), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnitConversionsForUnit(ctx, &mealplanninggrpc.GetValidMeasurementUnitConversionsForUnitRequest{
			ValidMeasurementUnitId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidMeasurementUnits), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnits(ctx, &mealplanninggrpc.GetValidMeasurementUnitsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparation()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidPreparation), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparation(ctx, &mealplanninggrpc.GetValidPreparationRequest{ValidPreparationId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationInstrument()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidPreparationInstrument), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstrument(ctx, &mealplanninggrpc.GetValidPreparationInstrumentRequest{ValidPreparationInstrumentId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidPreparationInstruments), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstruments(ctx, &mealplanninggrpc.GetValidPreparationInstrumentsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationInstrumentsByInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPreparationInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPreparationInstrumentsByInstrument), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstrumentsByInstrument(ctx, &mealplanninggrpc.GetValidPreparationInstrumentsByInstrumentRequest{
			ValidInstrumentId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationInstrumentsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPreparationInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPreparationInstrumentsByPreparation), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstrumentsByPreparation(ctx, &mealplanninggrpc.GetValidPreparationInstrumentsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationVessel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidPreparationVessel), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVessel(ctx, &mealplanninggrpc.GetValidPreparationVesselRequest{ValidPreparationVesselId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationVesselsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidPreparationVessels), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVessels(ctx, &mealplanninggrpc.GetValidPreparationVesselsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationVesselsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPreparationVesselsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPreparationVesselsByPreparation), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVesselsByPreparation(ctx, &mealplanninggrpc.GetValidPreparationVesselsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparationVesselsByVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleID := mealplanningfakes.BuildFakeID()
		exampleResult := mealplanningfakes.BuildFakeValidPreparationVesselsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPreparationVesselsByVessel), testutils.ContextMatcher, exampleID, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVesselsByVessel(ctx, &mealplanninggrpc.GetValidPreparationVesselsByVesselRequest{
			ValidVesselId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidPreparations), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparations(ctx, &mealplanninggrpc.GetValidPreparationsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVessel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ReadValidVessel), testutils.ContextMatcher, exampleResult.ID).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidVessel(ctx, &mealplanninggrpc.GetValidVesselRequest{ValidVesselId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_GetValidVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVesselsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.ListValidVessels), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.GetValidVessels(ctx, &mealplanninggrpc.GetValidVesselsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidIngredientGroups(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientGroupsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidIngredientGroupsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientGroups), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidIngredientGroups(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidIngredientStates(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStatesList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidIngredientStatesRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientStates), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidIngredientStates(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidIngredientsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredients), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidIngredients(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrumentsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidInstrumentsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidInstruments), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidInstruments(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidMeasurementUnitsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidMeasurementUnits), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidMeasurementUnits(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidPreparationsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidPreparations), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidPreparations(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchForValidVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVesselsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchForValidVesselsRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidVessels), testutils.ContextMatcher, exampleRequest.Query, exampleRequest.UseSearchService, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidVessels(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchValidIngredientsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchValidIngredientsByPreparationRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidIngredientsByPreparationAndIngredientName), testutils.ContextMatcher, exampleRequest.ValidPreparationId, exampleRequest.Query, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchValidIngredientsByPreparation(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_SearchValidMeasurementUnitsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitsList()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.SearchValidMeasurementUnitsByIngredientRequest](t)

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.SearchValidMeasurementUnitsByIngredientID), testutils.ContextMatcher, exampleRequest.ValidIngredientId, testutils.QueryFilterMatcher).Return(exampleResult, nil)
		s.mealPlanningManager = mvem

		result, err := s.SearchValidMeasurementUnitsByIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidIngredientRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidIngredient()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidIngredient), testutils.ContextMatcher, exampleRequest.ValidIngredientId, testutils.MatchType[*mealplanning.ValidIngredientUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidIngredientGroupRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidIngredientGroup()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidIngredientGroup), testutils.ContextMatcher, exampleRequest.ValidIngredientGroupId, testutils.MatchType[*mealplanning.ValidIngredientGroupUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientGroup(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidIngredientMeasurementUnitRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidIngredientMeasurementUnit()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidIngredientMeasurementUnit), testutils.ContextMatcher, exampleRequest.ValidIngredientMeasurementUnitId, testutils.MatchType[*mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientMeasurementUnit(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidIngredientPreparationRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidIngredientPreparation()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidIngredientPreparation), testutils.ContextMatcher, exampleRequest.ValidIngredientPreparationId, testutils.MatchType[*mealplanning.ValidIngredientPreparationUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientPreparation(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidPrepTaskConfigRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidPrepTaskConfig()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidPrepTaskConfig), testutils.ContextMatcher, exampleRequest.ValidPrepTaskConfigId, testutils.MatchType[*mealplanning.ValidPrepTaskConfigUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPrepTaskConfig(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidIngredientStateRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidIngredientState()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidIngredientState), testutils.ContextMatcher, exampleRequest.ValidIngredientStateId, testutils.MatchType[*mealplanning.ValidIngredientStateUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientState(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidIngredientStateIngredientRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidIngredientStateIngredient()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidIngredientStateIngredient), testutils.ContextMatcher, exampleRequest.ValidIngredientStateIngredientId, testutils.MatchType[*mealplanning.ValidIngredientStateIngredientUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientStateIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidInstrumentRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidInstrument()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidInstrument), testutils.ContextMatcher, exampleRequest.ValidInstrumentId, testutils.MatchType[*mealplanning.ValidInstrumentUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidInstrument(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidMeasurementUnitRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidMeasurementUnit()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidMeasurementUnit), testutils.ContextMatcher, exampleRequest.ValidMeasurementUnitId, testutils.MatchType[*mealplanning.ValidMeasurementUnitUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidMeasurementUnit(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidMeasurementUnitConversionRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidMeasurementUnitConversion()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidMeasurementUnitConversion), testutils.ContextMatcher, exampleRequest.ValidMeasurementUnitConversionId, testutils.MatchType[*mealplanning.ValidMeasurementUnitConversionUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidMeasurementUnitConversion(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidPreparationRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidPreparation()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidPreparation), testutils.ContextMatcher, exampleRequest.ValidPreparationId, testutils.MatchType[*mealplanning.ValidPreparationUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPreparation(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidPreparationInstrumentRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidPreparationInstrument()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidPreparationInstrument), testutils.ContextMatcher, exampleRequest.ValidPreparationInstrumentId, testutils.MatchType[*mealplanning.ValidPreparationInstrumentUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPreparationInstrument(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidPreparationVesselRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidPreparationVessel()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidPreparationVessel), testutils.ContextMatcher, exampleRequest.ValidPreparationVesselId, testutils.MatchType[*mealplanning.ValidPreparationVesselUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPreparationVessel(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}

func TestServiceImpl_UpdateValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRequest := fake.BuildFakeForTest[mealplanninggrpc.UpdateValidVesselRequest](t)
		exampleResponse := mealplanningfakes.BuildFakeValidVessel()

		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MockMealPlanningManager{}
		mvem.On(reflection.GetMethodName(mvem.UpdateValidVessel), testutils.ContextMatcher, exampleRequest.ValidVesselId, testutils.MatchType[*mealplanning.ValidVesselUpdateRequestInput]()).Return(exampleResponse, nil)
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidVessel(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		mock.AssertExpectationsForObjects(t, mvem)
	})
}
