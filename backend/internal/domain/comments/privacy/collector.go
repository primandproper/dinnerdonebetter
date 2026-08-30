// Package privacy is the comments domain's contribution to a subject access request.
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
)

// NewCollector builds the comments collector: every comment the subject
// authored, paged to the end and encoded, or nothing if they authored none.
func NewCollector(repo comments.Repository) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
		return repo.GetCommentsForUser(ctx, subject.ID, filter)
	})
}
