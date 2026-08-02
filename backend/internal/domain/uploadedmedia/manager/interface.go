package manager

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
)

// UploadedMediaManager defines the interface for uploaded media business logic.
// It embeds uploadedmedia.Repository so the manager provides the full repository surface.
type UploadedMediaManager interface {
	uploadedmedia.Repository
}
