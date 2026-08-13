package scheduler

import (
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	capitalismcfg "github.com/primandproper/platform-go/v10/capitalism/config"

	"github.com/samber/do/v2"
)

// RegisterMeteringFlusher registers the usage flusher and everything it stands on.
//
// The flusher runs here rather than in the API server because a flush is a scheduled pass over a
// backlog under a lease — which is what this process is for — and because the credentials it
// posts usage with are not credentials a request path should hold.
//
// The registry comes along even though only the recorder and the enforcer read it. The flusher
// works off the totals table alone, but registering it in both processes is what keeps the two
// from ever disagreeing about what a meter's period and aggregation are, which is the way a
// total silently starts meaning something else.
func RegisterMeteringFlusher(i do.Injector) {
	capitalismcfg.RegisterUsageReporter(i)

	appmetering.RegisterRegistry(i)
	appmetering.RegisterStore(i)
	appmetering.RegisterFlusher(i)
}
