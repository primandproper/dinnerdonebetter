package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWaitlistDataManager registers the waitlist data manager with the injector.
func RegisterWaitlistDataManager(i do.Injector) {
	// Register the repo provider (was included in wire.NewSet)
	waitlistsrepo.RegisterWaitlistsRepository(i)

	// Bind *waitlistsrepo.Repository to the waitlistRepository interface
	do.Provide[waitlistRepository](i, func(i do.Injector) (waitlistRepository, error) {
		return do.MustInvoke[*waitlistsrepo.Repository](i), nil
	})

	do.Provide[WaitlistsDataManager](i, func(i do.Injector) (WaitlistsDataManager, error) {
		return NewWaitlistDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[waitlistRepository](i),
		)
	})

	// Bind WaitlistsDataManager to waitlists.Repository
	do.Provide[waitlists.Repository](i, func(i do.Injector) (waitlists.Repository, error) {
		return do.MustInvoke[WaitlistsDataManager](i), nil
	})
}
