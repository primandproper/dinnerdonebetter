package grpc

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mockmanagers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanninggrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/primandproper/platform-go/v9/fake"
	"github.com/primandproper/platform-go/v9/filtering"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/stretchr/testify/assert"
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidIngredientFunc: func(_ context.Context, validIngredientID string) error {
				assert.Equal(t, exampleValidIngredientID, validIngredientID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredient(ctx, &mealplanninggrpc.ArchiveValidIngredientRequest{ValidIngredientId: exampleValidIngredientID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidIngredientCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientGroupID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidIngredientGroupFunc: func(_ context.Context, validIngredientGroupID string) error {
				assert.Equal(t, exampleValidIngredientGroupID, validIngredientGroupID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientGroup(ctx, &mealplanninggrpc.ArchiveValidIngredientGroupRequest{ValidIngredientGroupId: exampleValidIngredientGroupID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidIngredientGroupCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientMeasurementUnitID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) error {
				assert.Equal(t, exampleValidIngredientMeasurementUnitID, validIngredientMeasurementUnitID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientMeasurementUnit(ctx, &mealplanninggrpc.ArchiveValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: exampleValidIngredientMeasurementUnitID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientPreparationID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) error {
				assert.Equal(t, exampleValidIngredientPreparationID, validIngredientPreparationID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientPreparation(ctx, &mealplanninggrpc.ArchiveValidIngredientPreparationRequest{ValidIngredientPreparationId: exampleValidIngredientPreparationID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidIngredientPreparationCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPrepTaskConfigID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidPrepTaskConfigFunc: func(_ context.Context, validPrepTaskConfigID string) error {
				assert.Equal(t, exampleValidPrepTaskConfigID, validPrepTaskConfigID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPrepTaskConfig(ctx, &mealplanninggrpc.ArchiveValidPrepTaskConfigRequest{ValidPrepTaskConfigId: exampleValidPrepTaskConfigID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidPrepTaskConfigCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientStateID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidIngredientStateFunc: func(_ context.Context, validIngredientStateID string) error {
				assert.Equal(t, exampleValidIngredientStateID, validIngredientStateID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientState(ctx, &mealplanninggrpc.ArchiveValidIngredientStateRequest{ValidIngredientStateId: exampleValidIngredientStateID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidIngredientStateCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientStateIngredientID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidIngredientStateIngredientFunc: func(_ context.Context, validIngredientStateIngredientID string) error {
				assert.Equal(t, exampleValidIngredientStateIngredientID, validIngredientStateIngredientID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidIngredientStateIngredient(ctx, &mealplanninggrpc.ArchiveValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: exampleValidIngredientStateIngredientID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidIngredientStateIngredientCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidInstrumentID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidInstrumentFunc: func(_ context.Context, validInstrumentID string) error {
				assert.Equal(t, exampleValidInstrumentID, validInstrumentID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidInstrument(ctx, &mealplanninggrpc.ArchiveValidInstrumentRequest{ValidInstrumentId: exampleValidInstrumentID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidInstrumentCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnitID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string) error {
				assert.Equal(t, exampleValidMeasurementUnitID, validMeasurementUnitID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidMeasurementUnit(ctx, &mealplanninggrpc.ArchiveValidMeasurementUnitRequest{ValidMeasurementUnitId: exampleValidMeasurementUnitID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnitConversionID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidMeasurementUnitConversionFunc: func(_ context.Context, validMeasurementUnitConversionID string) error {
				assert.Equal(t, exampleValidMeasurementUnitConversionID, validMeasurementUnitConversionID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidMeasurementUnitConversion(ctx, &mealplanninggrpc.ArchiveValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: exampleValidMeasurementUnitConversionID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidMeasurementUnitConversionCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidPreparationFunc: func(_ context.Context, validPreparationID string) error {
				assert.Equal(t, exampleValidPreparationID, validPreparationID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPreparation(ctx, &mealplanninggrpc.ArchiveValidPreparationRequest{ValidPreparationId: exampleValidPreparationID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidPreparationCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationInstrumentID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) error {
				assert.Equal(t, exampleValidPreparationInstrumentID, validPreparationInstrumentID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPreparationInstrument(ctx, &mealplanninggrpc.ArchiveValidPreparationInstrumentRequest{ValidPreparationInstrumentId: exampleValidPreparationInstrumentID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidPreparationInstrumentCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationVesselID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) error {
				assert.Equal(t, exampleValidPreparationVesselID, validPreparationVesselID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidPreparationVessel(ctx, &mealplanninggrpc.ArchiveValidPreparationVesselRequest{ValidPreparationVesselId: exampleValidPreparationVesselID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidPreparationVesselCalls(), 1)
	})
}

func TestServiceImpl_ArchiveValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidVesselID := mealplanningfakes.BuildFakeID()

		mvem := &mockmanagers.MealPlanningManagerMock{
			ArchiveValidVesselFunc: func(_ context.Context, validVesselID string) error {
				assert.Equal(t, exampleValidVesselID, validVesselID)

				return nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.ArchiveValidVessel(ctx, &mealplanninggrpc.ArchiveValidVesselRequest{ValidVesselId: exampleValidVesselID})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		assert.Len(t, mvem.ArchiveValidVesselCalls(), 1)
	})
}

func TestServiceImpl_CreateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredient := mealplanningfakes.BuildFakeValidIngredient()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidIngredientFunc: func(_ context.Context, _ *mealplanning.ValidIngredientCreationRequestInput) (*mealplanning.ValidIngredient, error) {
				return exampleValidIngredient, nil
			},
		}
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidIngredientGroupFunc: func(_ context.Context, _ *mealplanning.ValidIngredientGroupCreationRequestInput) (*mealplanning.ValidIngredientGroup, error) {
				return exampleValidIngredientGroup, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientGroupRequest](t)

		actual, err := s.CreateValidIngredientGroup(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidIngredientGroupCalls(), 1)
	})
}

func TestServiceImpl_CreateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientMeasurementUnit := mealplanningfakes.BuildFakeValidIngredientMeasurementUnit()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidIngredientMeasurementUnitFunc: func(_ context.Context, _ *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput) (*mealplanning.ValidIngredientMeasurementUnit, error) {
				return exampleValidIngredientMeasurementUnit, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientMeasurementUnitRequest](t)

		actual, err := s.CreateValidIngredientMeasurementUnit(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_CreateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientPreparation := mealplanningfakes.BuildFakeValidIngredientPreparation()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidIngredientPreparationFunc: func(_ context.Context, _ *mealplanning.ValidIngredientPreparationCreationRequestInput) (*mealplanning.ValidIngredientPreparation, error) {
				return exampleValidIngredientPreparation, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientPreparationRequest](t)

		actual, err := s.CreateValidIngredientPreparation(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidIngredientPreparationCalls(), 1)
	})
}

func TestServiceImpl_CreateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPrepTaskConfig := mealplanningfakes.BuildFakeValidPrepTaskConfig()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidPrepTaskConfigFunc: func(_ context.Context, _ *mealplanning.ValidPrepTaskConfigCreationRequestInput) (*mealplanning.ValidPrepTaskConfig, error) {
				return exampleValidPrepTaskConfig, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPrepTaskConfigRequest](t)

		actual, err := s.CreateValidPrepTaskConfig(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidPrepTaskConfigCalls(), 1)
	})
}

func TestServiceImpl_CreateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientState := mealplanningfakes.BuildFakeValidIngredientState()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidIngredientStateFunc: func(_ context.Context, _ *mealplanning.ValidIngredientStateCreationRequestInput) (*mealplanning.ValidIngredientState, error) {
				return exampleValidIngredientState, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientStateRequest](t)

		actual, err := s.CreateValidIngredientState(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidIngredientStateCalls(), 1)
	})
}
func TestServiceImpl_CreateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidIngredientStateIngredient := mealplanningfakes.BuildFakeValidIngredientStateIngredient()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidIngredientStateIngredientFunc: func(_ context.Context, _ *mealplanning.ValidIngredientStateIngredientCreationRequestInput) (*mealplanning.ValidIngredientStateIngredient, error) {
				return exampleValidIngredientStateIngredient, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidIngredientStateIngredientRequest](t)

		actual, err := s.CreateValidIngredientStateIngredient(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidIngredientStateIngredientCalls(), 1)
	})
}

func TestServiceImpl_CreateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidInstrument := mealplanningfakes.BuildFakeValidInstrument()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidInstrumentFunc: func(_ context.Context, _ *mealplanning.ValidInstrumentCreationRequestInput) (*mealplanning.ValidInstrument, error) {
				return exampleValidInstrument, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidInstrumentRequest](t)

		actual, err := s.CreateValidInstrument(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidInstrumentCalls(), 1)
	})
}

func TestServiceImpl_CreateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnit := mealplanningfakes.BuildFakeValidMeasurementUnit()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidMeasurementUnitFunc: func(_ context.Context, _ *mealplanning.ValidMeasurementUnitCreationRequestInput) (*mealplanning.ValidMeasurementUnit, error) {
				return exampleValidMeasurementUnit, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidMeasurementUnitRequest](t)

		actual, err := s.CreateValidMeasurementUnit(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_CreateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidMeasurementUnitConversion := mealplanningfakes.BuildFakeValidMeasurementUnitConversion()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidMeasurementUnitConversionFunc: func(_ context.Context, _ *mealplanning.ValidMeasurementUnitConversionCreationRequestInput) (*mealplanning.ValidMeasurementUnitConversion, error) {
				return exampleValidMeasurementUnitConversion, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidMeasurementUnitConversionRequest](t)

		actual, err := s.CreateValidMeasurementUnitConversion(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidMeasurementUnitConversionCalls(), 1)
	})
}

func TestServiceImpl_CreateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparation := mealplanningfakes.BuildFakeValidPreparation()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidPreparationFunc: func(_ context.Context, _ *mealplanning.ValidPreparationCreationRequestInput) (*mealplanning.ValidPreparation, error) {
				return exampleValidPreparation, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPreparationRequest](t)

		actual, err := s.CreateValidPreparation(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidPreparationCalls(), 1)
	})
}

func TestServiceImpl_CreateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationInstrument := mealplanningfakes.BuildFakeValidPreparationInstrument()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidPreparationInstrumentFunc: func(_ context.Context, _ *mealplanning.ValidPreparationInstrumentCreationRequestInput) (*mealplanning.ValidPreparationInstrument, error) {
				return exampleValidPreparationInstrument, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPreparationInstrumentRequest](t)

		actual, err := s.CreateValidPreparationInstrument(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidPreparationInstrumentCalls(), 1)
	})
}

func TestServiceImpl_CreateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidPreparationVessel := mealplanningfakes.BuildFakeValidPreparationVessel()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidPreparationVesselFunc: func(_ context.Context, _ *mealplanning.ValidPreparationVesselCreationRequestInput) (*mealplanning.ValidPreparationVessel, error) {
				return exampleValidPreparationVessel, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidPreparationVesselRequest](t)

		actual, err := s.CreateValidPreparationVessel(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidPreparationVesselCalls(), 1)
	})
}

func TestServiceImpl_CreateValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		exampleValidVessel := mealplanningfakes.BuildFakeValidVessel()

		mvem := &mockmanagers.MealPlanningManagerMock{
			CreateValidVesselFunc: func(_ context.Context, _ *mealplanning.ValidVesselCreationRequestInput) (*mealplanning.ValidVessel, error) {
				return exampleValidVessel, nil
			},
		}
		s.mealPlanningManager = mvem

		exampleInput := fake.BuildFakeForTest[mealplanninggrpc.CreateValidVesselRequest](t)

		actual, err := s.CreateValidVessel(ctx, exampleInput)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Len(t, mvem.CreateValidVesselCalls(), 1)
	})
}

func TestServiceImpl_GetRandomValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredient()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			RandomValidIngredientFunc: func(_ context.Context) (*mealplanning.ValidIngredient, error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidIngredient(ctx, &mealplanninggrpc.GetRandomValidIngredientRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.RandomValidIngredientCalls(), 1)
	})
}

func TestServiceImpl_GetRandomValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrument()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			RandomValidInstrumentFunc: func(_ context.Context) (*mealplanning.ValidInstrument, error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidInstrument(ctx, &mealplanninggrpc.GetRandomValidInstrumentRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.RandomValidInstrumentCalls(), 1)
	})
}

func TestServiceImpl_GetRandomValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparation()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			RandomValidPreparationFunc: func(_ context.Context) (*mealplanning.ValidPreparation, error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidPreparation(ctx, &mealplanninggrpc.GetRandomValidPreparationRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.RandomValidPreparationCalls(), 1)
	})
}

func TestServiceImpl_GetRandomValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVessel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			RandomValidVesselFunc: func(_ context.Context) (*mealplanning.ValidVessel, error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetRandomValidVessel(ctx, &mealplanninggrpc.GetRandomValidVesselRequest{})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.RandomValidVesselCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredient()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidIngredientFunc: func(_ context.Context, validIngredientID string) (*mealplanning.ValidIngredient, error) {
				assert.Equal(t, exampleResult.ID, validIngredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredient(ctx, &mealplanninggrpc.GetValidIngredientRequest{ValidIngredientId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidIngredientCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientGroup()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidIngredientGroupFunc: func(_ context.Context, validIngredientGroupID string) (*mealplanning.ValidIngredientGroup, error) {
				assert.Equal(t, exampleResult.ID, validIngredientGroupID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientGroup(ctx, &mealplanninggrpc.GetValidIngredientGroupRequest{ValidIngredientGroupId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidIngredientGroupCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientGroups(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientGroupsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidIngredientGroupsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientGroups(ctx, &mealplanninggrpc.GetValidIngredientGroupsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidIngredientGroupsCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientMeasurementUnit()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) (*mealplanning.ValidIngredientMeasurementUnit, error) {
				assert.Equal(t, exampleResult.ID, validIngredientMeasurementUnitID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnit(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientMeasurementUnitsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidIngredientMeasurementUnitsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnits(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidIngredientMeasurementUnitsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientMeasurementUnitsByIngredientFunc: func(_ context.Context, validIngredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
				assert.Equal(t, exampleID, validIngredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnitsByIngredient(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientMeasurementUnitsByIngredientCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientMeasurementUnitsByMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
				assert.Equal(t, exampleID, validMeasurementUnitID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientMeasurementUnitsByMeasurementUnit(ctx, &mealplanninggrpc.GetValidIngredientMeasurementUnitsByMeasurementUnitRequest{
			ValidMeasurementUnitId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientMeasurementUnitsByMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientPreparation()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) (*mealplanning.ValidIngredientPreparation, error) {
				assert.Equal(t, exampleResult.ID, validIngredientPreparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparation(ctx, &mealplanninggrpc.GetValidIngredientPreparationRequest{ValidIngredientPreparationId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidIngredientPreparationCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientPreparationsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidIngredientPreparationsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparations(ctx, &mealplanninggrpc.GetValidIngredientPreparationsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidIngredientPreparationsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientPreparationsByIngredientFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
				assert.Equal(t, exampleID, ingredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparationsByIngredient(ctx, &mealplanninggrpc.GetValidIngredientPreparationsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientPreparationsByIngredientCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientPreparationsByPreparationFunc: func(_ context.Context, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
				assert.Equal(t, exampleID, preparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientPreparationsByPreparation(ctx, &mealplanninggrpc.GetValidIngredientPreparationsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientPreparationsByPreparationCalls(), 1)
	})
}

func TestServiceImpl_GetValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfig()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidPrepTaskConfigFunc: func(_ context.Context, validPrepTaskConfigID string) (*mealplanning.ValidPrepTaskConfig, error) {
				assert.Equal(t, exampleResult.ID, validPrepTaskConfigID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfig(ctx, &mealplanninggrpc.GetValidPrepTaskConfigRequest{ValidPrepTaskConfigId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidPrepTaskConfigCalls(), 1)
	})
}

func TestServiceImpl_GetValidPrepTaskConfigs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPrepTaskConfigsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidPrepTaskConfigsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigs(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidPrepTaskConfigsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPrepTaskConfigsByIngredientFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
				assert.Equal(t, exampleID, ingredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigsByIngredient(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPrepTaskConfigsByIngredientCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPrepTaskConfigsByPreparationFunc: func(_ context.Context, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
				assert.Equal(t, exampleID, preparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigsByPreparation(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPrepTaskConfigsByPreparationCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPrepTaskConfigsByIngredientAndPreparationFunc: func(_ context.Context, ingredientID string, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
				assert.Equal(t, exampleIngredientID, ingredientID)
				assert.Equal(t, examplePreparationID, preparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPrepTaskConfigsByIngredientAndPreparation(ctx, &mealplanninggrpc.GetValidPrepTaskConfigsByIngredientAndPreparationRequest{
			ValidIngredientId:  exampleIngredientID,
			ValidPreparationId: examplePreparationID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPrepTaskConfigsByIngredientAndPreparationCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientState()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidIngredientStateFunc: func(_ context.Context, validIngredientStateID string) (*mealplanning.ValidIngredientState, error) {
				assert.Equal(t, exampleResult.ID, validIngredientStateID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientState(ctx, &mealplanninggrpc.GetValidIngredientStateRequest{ValidIngredientStateId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidIngredientStateCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStateIngredient()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidIngredientStateIngredientFunc: func(_ context.Context, validIngredientStateIngredientID string) (*mealplanning.ValidIngredientStateIngredient, error) {
				assert.Equal(t, exampleResult.ID, validIngredientStateIngredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredient(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidIngredientStateIngredientCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientStateIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStateIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidIngredientStateIngredientsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredients(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidIngredientStateIngredientsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientStateIngredientsByIngredientFunc: func(_ context.Context, validIngredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
				assert.Equal(t, exampleID, validIngredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredientsByIngredient(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientsByIngredientRequest{
			ValidIngredientId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientStateIngredientsByIngredientCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientStateIngredientsByIngredientStateFunc: func(_ context.Context, validIngredientStateID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
				assert.Equal(t, exampleID, validIngredientStateID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStateIngredientsByIngredientState(ctx, &mealplanninggrpc.GetValidIngredientStateIngredientsByIngredientStateRequest{
			ValidIngredientStateId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientStateIngredientsByIngredientStateCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredientStates(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientStatesList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidIngredientStatesFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredientStates(ctx, &mealplanninggrpc.GetValidIngredientStatesRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidIngredientStatesCalls(), 1)
	})
}

func TestServiceImpl_GetValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidIngredientsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidIngredientsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidIngredients(ctx, &mealplanninggrpc.GetValidIngredientsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidIngredientsCalls(), 1)
	})
}

func TestServiceImpl_GetValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrument()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidInstrumentFunc: func(_ context.Context, validInstrumentID string) (*mealplanning.ValidInstrument, error) {
				assert.Equal(t, exampleResult.ID, validInstrumentID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidInstrument(ctx, &mealplanninggrpc.GetValidInstrumentRequest{ValidInstrumentId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidInstrumentCalls(), 1)
	})
}

func TestServiceImpl_GetValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidInstrumentsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidInstruments(ctx, &mealplanninggrpc.GetValidInstrumentsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidInstrumentsCalls(), 1)
	})
}

func TestServiceImpl_GetValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnit()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string) (*mealplanning.ValidMeasurementUnit, error) {
				assert.Equal(t, exampleResult.ID, validMeasurementUnitID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnit(ctx, &mealplanninggrpc.GetValidMeasurementUnitRequest{ValidMeasurementUnitId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_GetValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitConversion()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidMeasurementUnitConversionFunc: func(_ context.Context, validMeasurementUnitConversionID string) (*mealplanning.ValidMeasurementUnitConversion, error) {
				assert.Equal(t, exampleResult.ID, validMeasurementUnitConversionID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnitConversion(ctx, &mealplanninggrpc.GetValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidMeasurementUnitConversionCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			ValidMeasurementUnitConversionsForMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion], error) {
				assert.Equal(t, exampleID, validMeasurementUnitID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnitConversionsForUnit(ctx, &mealplanninggrpc.GetValidMeasurementUnitConversionsForUnitRequest{
			ValidMeasurementUnitId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ValidMeasurementUnitConversionsForMeasurementUnitCalls(), 1)
	})
}

func TestServiceImpl_GetValidMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidMeasurementUnitsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidMeasurementUnitsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidMeasurementUnits(ctx, &mealplanninggrpc.GetValidMeasurementUnitsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidMeasurementUnitsCalls(), 1)
	})
}

func TestServiceImpl_GetValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparation()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidPreparationFunc: func(_ context.Context, validPreparationID string) (*mealplanning.ValidPreparation, error) {
				assert.Equal(t, exampleResult.ID, validPreparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparation(ctx, &mealplanninggrpc.GetValidPreparationRequest{ValidPreparationId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidPreparationCalls(), 1)
	})
}

func TestServiceImpl_GetValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationInstrument()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) (*mealplanning.ValidPreparationInstrument, error) {
				assert.Equal(t, exampleResult.ID, validPreparationInstrumentID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstrument(ctx, &mealplanninggrpc.GetValidPreparationInstrumentRequest{ValidPreparationInstrumentId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidPreparationInstrumentCalls(), 1)
	})
}

func TestServiceImpl_GetValidPreparationInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationInstrumentsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidPreparationInstrumentsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstruments(ctx, &mealplanninggrpc.GetValidPreparationInstrumentsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidPreparationInstrumentsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPreparationInstrumentsByInstrumentFunc: func(_ context.Context, validInstrumentID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
				assert.Equal(t, exampleID, validInstrumentID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstrumentsByInstrument(ctx, &mealplanninggrpc.GetValidPreparationInstrumentsByInstrumentRequest{
			ValidInstrumentId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPreparationInstrumentsByInstrumentCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPreparationInstrumentsByPreparationFunc: func(_ context.Context, validPreparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
				assert.Equal(t, exampleID, validPreparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationInstrumentsByPreparation(ctx, &mealplanninggrpc.GetValidPreparationInstrumentsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPreparationInstrumentsByPreparationCalls(), 1)
	})
}

func TestServiceImpl_GetValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationVessel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) (*mealplanning.ValidPreparationVessel, error) {
				assert.Equal(t, exampleResult.ID, validPreparationVesselID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVessel(ctx, &mealplanninggrpc.GetValidPreparationVesselRequest{ValidPreparationVesselId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidPreparationVesselCalls(), 1)
	})
}

func TestServiceImpl_GetValidPreparationVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationVesselsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidPreparationVesselsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVessels(ctx, &mealplanninggrpc.GetValidPreparationVesselsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidPreparationVesselsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPreparationVesselsByPreparationFunc: func(_ context.Context, validPreparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
				assert.Equal(t, exampleID, validPreparationID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVesselsByPreparation(ctx, &mealplanninggrpc.GetValidPreparationVesselsByPreparationRequest{
			ValidPreparationId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPreparationVesselsByPreparationCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPreparationVesselsByVesselFunc: func(_ context.Context, validVesselID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
				assert.Equal(t, exampleID, validVesselID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparationVesselsByVessel(ctx, &mealplanninggrpc.GetValidPreparationVesselsByVesselRequest{
			ValidVesselId: exampleID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPreparationVesselsByVesselCalls(), 1)
	})
}

func TestServiceImpl_GetValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidPreparationsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidPreparationsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidPreparations(ctx, &mealplanninggrpc.GetValidPreparationsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidPreparationsCalls(), 1)
	})
}

func TestServiceImpl_GetValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVessel()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ReadValidVesselFunc: func(_ context.Context, validVesselID string) (*mealplanning.ValidVessel, error) {
				assert.Equal(t, exampleResult.ID, validVesselID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidVessel(ctx, &mealplanninggrpc.GetValidVesselRequest{ValidVesselId: exampleResult.ID})
		assert.Equal(t, exampleResult.ID, result.Result.Id)
		assert.NoError(t, err)

		assert.Len(t, mvem.ReadValidVesselCalls(), 1)
	})
}

func TestServiceImpl_GetValidVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleResult := mealplanningfakes.BuildFakeValidVesselsList()

		ctx := t.Context()
		s := buildServiceImplForTest(t)

		mvem := &mockmanagers.MealPlanningManagerMock{
			ListValidVesselsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.GetValidVessels(ctx, &mealplanninggrpc.GetValidVesselsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.ListValidVesselsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientGroupsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidIngredientGroups(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientGroupsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientStatesFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidIngredientStates(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientStatesCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidIngredients(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidInstrumentsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidInstruments(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidInstrumentsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidMeasurementUnitsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidMeasurementUnits(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidMeasurementUnitsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidPreparationsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidPreparations(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidPreparationsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidVesselsFunc: func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
				assert.Equal(t, exampleRequest.Query, query)
				assert.Equal(t, exampleRequest.UseSearchService, useSearchService)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchForValidVessels(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidVesselsCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidIngredientsByPreparationAndIngredientNameFunc: func(_ context.Context, preparationID string, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
				assert.Equal(t, exampleRequest.ValidPreparationId, preparationID)
				assert.Equal(t, exampleRequest.Query, query)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchValidIngredientsByPreparation(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidIngredientsByPreparationAndIngredientNameCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			SearchValidMeasurementUnitsByIngredientIDFunc: func(_ context.Context, validIngredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
				assert.Equal(t, exampleRequest.ValidIngredientId, validIngredientID)

				return exampleResult, nil
			},
		}
		s.mealPlanningManager = mvem

		result, err := s.SearchValidMeasurementUnitsByIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, len(exampleResult.Data))

		assert.Len(t, mvem.SearchValidMeasurementUnitsByIngredientIDCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidIngredientFunc: func(_ context.Context, validIngredientID string, _ *mealplanning.ValidIngredientUpdateRequestInput) (*mealplanning.ValidIngredient, error) {
				assert.Equal(t, exampleRequest.ValidIngredientId, validIngredientID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidIngredientCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidIngredientGroupFunc: func(_ context.Context, validIngredientGroupID string, _ *mealplanning.ValidIngredientGroupUpdateRequestInput) (*mealplanning.ValidIngredientGroup, error) {
				assert.Equal(t, exampleRequest.ValidIngredientGroupId, validIngredientGroupID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientGroup(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidIngredientGroupCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string, _ *mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput) (*mealplanning.ValidIngredientMeasurementUnit, error) {
				assert.Equal(t, exampleRequest.ValidIngredientMeasurementUnitId, validIngredientMeasurementUnitID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientMeasurementUnit(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidIngredientMeasurementUnitCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string, _ *mealplanning.ValidIngredientPreparationUpdateRequestInput) (*mealplanning.ValidIngredientPreparation, error) {
				assert.Equal(t, exampleRequest.ValidIngredientPreparationId, validIngredientPreparationID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientPreparation(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidIngredientPreparationCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidPrepTaskConfigFunc: func(_ context.Context, validPrepTaskConfigID string, _ *mealplanning.ValidPrepTaskConfigUpdateRequestInput) (*mealplanning.ValidPrepTaskConfig, error) {
				assert.Equal(t, exampleRequest.ValidPrepTaskConfigId, validPrepTaskConfigID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPrepTaskConfig(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidPrepTaskConfigCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidIngredientStateFunc: func(_ context.Context, validIngredientStateID string, _ *mealplanning.ValidIngredientStateUpdateRequestInput) (*mealplanning.ValidIngredientState, error) {
				assert.Equal(t, exampleRequest.ValidIngredientStateId, validIngredientStateID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientState(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidIngredientStateCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidIngredientStateIngredientFunc: func(_ context.Context, validIngredientStateIngredientID string, _ *mealplanning.ValidIngredientStateIngredientUpdateRequestInput) (*mealplanning.ValidIngredientStateIngredient, error) {
				assert.Equal(t, exampleRequest.ValidIngredientStateIngredientId, validIngredientStateIngredientID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidIngredientStateIngredient(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidIngredientStateIngredientCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidInstrumentFunc: func(_ context.Context, validInstrumentID string, _ *mealplanning.ValidInstrumentUpdateRequestInput) (*mealplanning.ValidInstrument, error) {
				assert.Equal(t, exampleRequest.ValidInstrumentId, validInstrumentID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidInstrument(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidInstrumentCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string, _ *mealplanning.ValidMeasurementUnitUpdateRequestInput) (*mealplanning.ValidMeasurementUnit, error) {
				assert.Equal(t, exampleRequest.ValidMeasurementUnitId, validMeasurementUnitID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidMeasurementUnit(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidMeasurementUnitCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidMeasurementUnitConversionFunc: func(_ context.Context, validMeasurementUnitConversionID string, _ *mealplanning.ValidMeasurementUnitConversionUpdateRequestInput) (*mealplanning.ValidMeasurementUnitConversion, error) {
				assert.Equal(t, exampleRequest.ValidMeasurementUnitConversionId, validMeasurementUnitConversionID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidMeasurementUnitConversion(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidMeasurementUnitConversionCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidPreparationFunc: func(_ context.Context, validPreparationID string, _ *mealplanning.ValidPreparationUpdateRequestInput) (*mealplanning.ValidPreparation, error) {
				assert.Equal(t, exampleRequest.ValidPreparationId, validPreparationID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPreparation(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidPreparationCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string, _ *mealplanning.ValidPreparationInstrumentUpdateRequestInput) (*mealplanning.ValidPreparationInstrument, error) {
				assert.Equal(t, exampleRequest.ValidPreparationInstrumentId, validPreparationInstrumentID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPreparationInstrument(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidPreparationInstrumentCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string, _ *mealplanning.ValidPreparationVesselUpdateRequestInput) (*mealplanning.ValidPreparationVessel, error) {
				assert.Equal(t, exampleRequest.ValidPreparationVesselId, validPreparationVesselID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidPreparationVessel(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidPreparationVesselCalls(), 1)
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

		mvem := &mockmanagers.MealPlanningManagerMock{
			UpdateValidVesselFunc: func(_ context.Context, validVesselID string, _ *mealplanning.ValidVesselUpdateRequestInput) (*mealplanning.ValidVessel, error) {
				assert.Equal(t, exampleRequest.ValidVesselId, validVesselID)

				return exampleResponse, nil
			},
		}
		s.mealPlanningManager = mvem

		res, err := s.UpdateValidVessel(ctx, exampleRequest)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse.ID, res.Result.Id)

		assert.Len(t, mvem.UpdateValidVesselCalls(), 1)
	})
}
