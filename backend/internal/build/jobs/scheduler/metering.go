package scheduler

import (
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	"github.com/primandproper/platform-go/v15/metering"

	"github.com/samber/do/v2"
)

// RegisterMetering registers what platform's metering block cannot build from configuration: the
// meter registry, the period resolver, and the mapping from this application's meters to the
// billing provider's.
//
// The flusher runs here rather than in the API server because a flush is a scheduled pass over a
// backlog under a lease — which is what this process is for — and because the credentials it
// posts usage with are not credentials a request path should hold. platform builds it, with the
// store, from the service.Config Metering block, and the usage reporter from the Capitalism block.
//
// The registry comes along even though only the recorder and the enforcer read it. The flusher
// works off the totals table alone, but registering it in both processes is what keeps the two
// from ever disagreeing about what a meter's period and aggregation are, which is the way a
// total silently starts meaning something else. And platform's block builds the recorder and the
// enforcer here too, whether or not anything records, so both have to be buildable.
//
// The period resolver is the calendar one with no billing cycle behind it, which is what
// platform's recorder and enforcer default to when handed none — the API server's are built that
// way — named here because platform's registrations require one rather than defaulting.
func RegisterMetering(i do.Injector) {
	appmetering.RegisterRegistry(i)

	do.Provide[metering.PeriodResolver](i, func(do.Injector) (metering.PeriodResolver, error) {
		return metering.NewCalendarPeriodResolver(nil), nil
	})

	do.Provide[metering.ProviderMapper](i, func(do.Injector) (metering.ProviderMapper, error) {
		return appmetering.NewProviderMapper(), nil
	})
}
