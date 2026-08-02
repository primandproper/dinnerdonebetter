package manager

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
)

// SettingsDataManager defines the interface for settings business logic.
// It embeds settings.Repository so the manager provides the full repository surface.
type SettingsDataManager interface {
	settings.Repository
}
