package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	"github.com/primandproper/primitives-go/v2/errors/grpc"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(mealPlanningGRPCMapper{})
}

type mealPlanningGRPCMapper struct{}

func (mealPlanningGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}
	switch {
	case errors.Is(err, mealplanning.ErrDuplicateMeal),
		errors.Is(err, mealplanning.ErrDuplicateMealInList),
		errors.Is(err, mealplanning.ErrDuplicateMealPlanOption):
		return codes.AlreadyExists, true
	case errors.Is(err, mealplanningrepo.ErrAlreadyFinalized):
		return codes.FailedPrecondition, true
	// A recipe whose bridge-table references disagree with its steps is a malformed
	// request, not a broken server. Without this it reached a client as Internal.
	case errors.Is(err, mealplanning.ErrInvalidRecipeInput):
		return codes.InvalidArgument, true
	default:
		return codes.Unknown, false
	}
}
