package localdev

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

// UploadsRegistry builds the upload registry store the identity and meal
// planning repositories read media through.
//
// It is the platform's store directly rather than
// internal/repositories/postgres/uploadedmedia's, because that one adds an audit
// entry and an outbox event to every write and nothing in these local processes
// writes an object — they only read the rows a request already created.
//
// A failure is returned rather than swallowed: the only way NewSQLStore fails is
// a nil client or a prefix that is not a legal identifier, both of which are
// programming errors that would otherwise surface as a nil store panicking on
// the first user with an avatar.
func UploadsRegistry(logger logging.Logger, tracerProvider tracing.Provider, client database.Client) (registry.Store, error) {
	return registry.NewSQLStore(
		client,
		registry.WithTablePrefix(uploadedmedia.TablePrefix),
		registry.WithStoreLogger(logger),
		registry.WithStoreTracerProvider(tracerProvider),
	)
}
