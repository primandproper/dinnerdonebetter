package manager

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
)

// DataPrivacyManager defines the interface for data privacy business logic.
// It embeds dataprivacy.Repository so the manager provides the full repository surface.
type DataPrivacyManager interface {
	dataprivacy.Repository
}
