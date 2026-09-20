package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/errors"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/errors"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	comments "github.com/primandproper/platform-go/v14/comments"
	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"
)

var _ mealplanningsvc.MealPlanningServiceServer = (*serviceImpl)(nil)

const (
	o11yName = "mealplanning_service"
)

type (
	serviceImpl struct {
		mealplanningsvc.UnimplementedMealPlanningServiceServer
		tracer              tracing.Tracer
		db                  database.Client
		logger              logging.Logger
		mealPlanningManager managers.MealPlanningManager
		// One starter, where there used to be three workers. All three of the admin RPCs that
		// ran those on demand reach this: finalizing, creating tasks, and building the grocery
		// list are one saga now, and the only part of it left to run on demand is entering
		// plans into it.
		mealPlanFinalizationStarter *mealplanfinalization.Starter
		comments                    comments.Store
		registry                    mediaregistry.Store
		uploadManager               uploads.UploadManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	db database.Client,
	mealPlanningManager managers.MealPlanningManager,
	mealPlanFinalizationStarter *mealplanfinalization.Starter,
	commentStore comments.Store,
	registryStore mediaregistry.Store,
	uploadManager uploads.UploadManager,
) mealplanningsvc.MealPlanningServiceServer {
	return &serviceImpl{
		logger:                      logging.NewNamedLogger(logger, o11yName),
		tracer:                      tracing.NewNamedTracer(tracerProvider, o11yName),
		db:                          db,
		mealPlanningManager:         mealPlanningManager,
		mealPlanFinalizationStarter: mealPlanFinalizationStarter,
		comments:                    commentStore,
		registry:                    registryStore,
		uploadManager:               uploadManager,
	}
}

// inTransaction runs one store write on a transaction of its own.
//
// As of platform-go v14 a store holds no database handle: a write takes the
// caller's database.Tx. Every write in this service is a single store call, so
// each gets one transaction — which is exactly what the store opened for itself
// before the caller was required to supply it. A handler that ever writes twice
// should take one transaction across both rather than call this twice.
func inTransaction[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
