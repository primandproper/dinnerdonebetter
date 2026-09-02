/*
Package privacy is the settings domain's contribution to a subject access
request: the answers somebody has stored about themselves.

There is no eraser here, unlike waitlists and comments, and its absence is a
property of the schema rather than an omission. Every setting value in this
deployment belongs to a user, so ddb_settings_values carries a foreign key to
users with ON DELETE CASCADE — see internal/repositories/postgres/migrations —
and the single identity eraser takes the rows with the user. See
internal/domain/settings for why there is only one subject type.
*/
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
	platformsettings "github.com/primandproper/platform-go/v13/settings"
)

// NewCollector builds the settings collector: every value the subject has
// stored, paged to the end and encoded, or nothing if they have answered no
// setting.
//
// What it does not export is the catalog. A definition is an administrative row
// that says what a setting means for everybody, and a copy of it in one person's
// export would describe this deployment rather than them. The value names the
// definition it answers, which is what makes the export readable against the
// catalog somebody can already list.
func NewCollector(store platformsettings.ValueStore) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[platformsettings.Value], error) {
		return store.ListValuesForSubject(ctx, settings.Scope(), settings.SubjectFor(subject.ID), filter)
	})
}
