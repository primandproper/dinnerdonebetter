package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

var _ mealplanning.UploadedMediaFetcher = (*repository)(nil)

// GetUploadedMediaWithIDs fetches uploaded media by IDs, in the order the IDs were given.
//
// The registry's batched read answers in id order, one row per distinct id; the order is put
// back here because the callers' order is the bridge rows', which is the order a recipe step's
// images are shown in.
//
// An id with no row is absent from the result rather than failing the read. A bridge row
// pointing at an archived or absent object is a broken reference, not a broken request, and the
// caller asked for the media that is there.
func (q *repository) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*mediaregistry.Object, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithSpan(span)

	if len(ids) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue("id_count", len(ids))

	read, err := mediaregistry.ListObjectsByIDsInBatches(ctx, q.Reader(), tenancy.Global(), q.uploads, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media with IDs")
	}

	byID := make(map[string]*mediaregistry.Object, len(read))
	for _, object := range read {
		byID[object.ID] = object
	}

	objects := make([]*mediaregistry.Object, 0, len(ids))
	for _, id := range ids {
		if object, ok := byID[id]; ok {
			objects = append(objects, object)
		}
	}

	return objects, nil
}
