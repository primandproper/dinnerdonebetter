package mealplanning

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

var _ mealplanning.UploadedMediaFetcher = (*repository)(nil)

// GetUploadedMediaWithIDs fetches uploaded media by IDs.
//
// It is a read per id rather than one statement over the set, because the
// registry ships no bulk read and this repository cannot write one: the table is
// platform-go's, created by a generated migration rather than by a file in
// migration_files, so sqlc — whose schema is migration_files — has never seen
// it. Rolling a statement by hand against another package's schema is how a
// column rename becomes a runtime error in the one place nothing regenerates.
//
// The cost is bounded by what actually calls this: the media on one ingredient
// or preparation, the images on one recipe step. Each is a handful of rows on a
// primary key. A caller looking to hydrate a whole page's worth of media should
// not reach for this — it would multiply, and the fix is a bulk read upstream
// rather than a wider loop here.
//
// An id with no row is skipped rather than failing the read. A bridge row
// pointing at an archived or absent object is a broken reference, not a broken
// request, and the caller asked for the media that is there.
func (q *repository) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*registry.Object, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithSpan(span)

	if len(ids) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue("id_count", len(ids))

	objects := make([]*registry.Object, 0, len(ids))
	for _, id := range ids {
		object, err := q.uploads.GetObject(ctx, ddbuploadedmedia.Scope(), id)
		if err != nil {
			if errors.Is(err, registry.ErrObjectNotFound) {
				continue
			}

			return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media with IDs")
		}

		objects = append(objects, object)
	}

	return objects, nil
}
