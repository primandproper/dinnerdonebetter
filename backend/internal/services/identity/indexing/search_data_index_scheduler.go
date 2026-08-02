package indexing

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v9/search/text/indexing"
)

const (
	// IndexTypeUsers represents the users index.
	IndexTypeUsers = "users"
)

func BuildCoreDataIndexingFunctions(dataManager identity.Repository) map[string]indexing.Function {
	return map[string]indexing.Function{
		IndexTypeUsers: dataManager.GetUserIDsThatNeedSearchIndexing,
	}
}
