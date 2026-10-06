package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/primitives-go/v2/database"

	"github.com/samber/do/v2"
)

var _ searchindexes.Registrar = RegisterIndexes

// RegisterIndexes adds the users index to registry: its Syncer, its Reindexer, and the stamp
// buffer behind the Syncer. It is a searchindexes.Registrar.
func RegisterIndexes(i do.Injector, registry *searchsync.Registry) error {
	client := do.MustInvoke[database.Client](i)
	store := do.MustInvoke[platformidentity.Store](i)

	source, err := UserSource(client, store)
	if err != nil {
		return err
	}

	return searchindexes.RegisterTextIndex(registry, source, do.MustInvoke[UserTextSearcher](i), UserStamps(client, store))
}
