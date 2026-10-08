package mealplanning

import (
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// idArg is one ID a repository method was handed: the observability key it is logged and traced
// under, and its value.
type idArg struct {
	key, value string
}

// guardIDs is the prologue a repository method that takes IDs opens with. It checks each ID in
// order and, for each one that is present, attaches it to the logger and the span before checking
// the next — so a method refused on its third ID has still traced the first two, exactly as the
// hand-written prologue it replaces did.
//
// The first empty ID stops the walk and returns platformerrors.ErrInvalidIDProvided. Methods whose
// prologue does something else on a missing ID (a different error, no span attribute, a combined
// check) keep their own prologue; this only names the common one.
func guardIDs(logger logging.Logger, span tracing.Span, ids ...idArg) (logging.Logger, error) {
	for _, id := range ids {
		if id.value == "" {
			return logger, platformerrors.ErrInvalidIDProvided
		}
		logger = logger.WithValue(id.key, id.value)
		tracing.AttachToSpan(span, id.key, id.value)
	}

	return logger, nil
}
