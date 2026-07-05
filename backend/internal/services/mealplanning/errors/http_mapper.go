package errors

import (
	"errors"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	httperrors "github.com/primandproper/platform-go/v3/errors/http"
)

func init() {
	httperrors.RegisterHTTPErrorMapper(mealPlanningHTTPMapper{})
}

type mealPlanningHTTPMapper struct{}

func (mealPlanningHTTPMapper) Map(err error) (code httperrors.ErrorCode, msg string, ok bool) {
	if err == nil {
		return "", "", false
	}
	switch {
	case errors.Is(err, mealplanning.ErrDuplicateMeal),
		errors.Is(err, mealplanning.ErrDuplicateMealInList),
		errors.Is(err, mealplanning.ErrDuplicateMealPlanOption):
		return httperrors.ErrValidatingRequestInput, "duplicate entry", true
	case errors.Is(err, mealplanningrepo.ErrAlreadyFinalized):
		return httperrors.ErrValidatingRequestInput, "already finalized", true
	default:
		return "", "", false
	}
}
