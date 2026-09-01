// Package privacy is the uploaded media domain's contribution to a subject access request.
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

// NewCollector builds the uploaded media collector: every object the subject
// owns, paged to the end and encoded, or nothing if they uploaded none.
//
// The rows, not the bytes. What this section holds is the registry's record of
// each upload — where it lives, what it is, how big it was; the files
// themselves live in object storage and reach the subject, if at all, by the
// keys those rows carry.
func NewCollector(store registry.Store) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[registry.Object], error) {
		return store.ListObjectsByOwner(ctx, uploadedmedia.Scope(), subject.ID, filter)
	})
}
