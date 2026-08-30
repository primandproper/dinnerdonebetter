// Package privacy is the uploaded media domain's contribution to a subject access request.
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
)

// NewCollector builds the uploaded media collector: every media record belonging
// to the subject, paged to the end and encoded, or nothing if they uploaded
// none.
//
// The records, not the bytes. What this section holds is the metadata stored
// about each upload; the files themselves live in object storage and reach the
// subject, if at all, by the URLs those records carry.
func NewCollector(repo uploadedmedia.Repository) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
		return repo.GetUploadedMediaForUser(ctx, subject.ID, filter)
	})
}
