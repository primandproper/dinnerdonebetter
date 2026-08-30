// Package privacy is the waitlists domain's contribution to a subject access request.
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
)

// NewCollector builds the waitlists collector: every waitlist signup belonging
// to the subject, paged to the end and encoded, or nothing if they signed up for
// none.
func NewCollector(repo waitlists.Repository) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
		return repo.GetWaitlistSignupsForUser(ctx, subject.ID, filter)
	})
}
