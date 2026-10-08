package comments

import (
	"context"
	"database/sql"
	"errors"

	platformcomments "github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// WithExistenceCheck turns a domain read into the existence hook platform's catalog takes, on
// the definition it is given.
//
// A read that failed because the row is not there is "absent"; any other failure is an error,
// and the two are kept apart deliberately. Platform's hook says an error is not absent, because
// a hook that decided an unreachable table meant a missing target would refuse writes the caller
// should have been told to retry.
//
// The read is handed the target's ID and nothing else. The check platform runs is given the
// comment's scope as well, but every comment here is filed under the global scope, so the scope
// cannot narrow a read and is not passed on.
func WithExistenceCheck(
	definition platformcomments.TargetDefinition,
	read func(ctx context.Context, targetID string) error,
) platformcomments.TargetDefinition {
	definition.Exists = func(ctx context.Context, _ tenancy.Scope, targetID string) (bool, error) {
		if err := read(ctx, targetID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}

			return false, err
		}

		return true, nil
	}

	return definition
}
