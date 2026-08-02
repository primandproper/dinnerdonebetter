package manager

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
)

// AuditDataManager defines the interface for audit log business logic.
// It embeds audit.Repository so the manager provides the full repository surface.
type AuditDataManager interface {
	audit.Repository
}
