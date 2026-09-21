// Package privacy is the uploaded media domain's contribution to a subject access request.
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// NewCollector builds the uploaded media collector: every object the subject
// owns, paged to the end and encoded, or nothing if they uploaded none.
//
// The rows, not the bytes. What this section holds is the registry's record of
// each upload — where it lives, what it is, how big it was; the files
// themselves live in object storage and reach the subject, if at all, by the
// keys those rows carry.
//
// The reader is taken at construction because dataprivacy.Collector.Collect is
// handed no executor: an export is a read, and it runs outside the erasure
// transaction by design. The request scope is not consulted either — every
// object this deployment registers is in uploadedmedia.Scope.
func NewCollector(store mediaregistry.Store, reader database.SQLQueryExecutor) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mediaregistry.Object], error) {
		return store.ListObjectsByOwner(ctx, reader, uploadedmedia.Scope(), subject.ID, filter)
	})
}
