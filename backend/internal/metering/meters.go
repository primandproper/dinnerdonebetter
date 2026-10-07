package metering

import (
	platformmetering "github.com/primandproper/platform-go/v15/metering"
	"github.com/primandproper/primitives-go/v2/errors"
)

const (
	// UploadedMediaBytesMeter counts the bytes an account pushes through the media upload
	// endpoint.
	//
	// It is the first meter because it is the one that maps most directly onto a bill we
	// actually receive: object storage charges for what is held, and this is the only place
	// anything gets held.
	//
	// It counts bytes accepted, not bytes resident. A sum over the period answers "how much
	// did this account add in March", which is what the ingest path can honestly observe —
	// residency is a different question, needs a periodic sweep of the bucket rather than a
	// write-site record, and would be AggregationLast over a gauge. Deleting media does not
	// decrement this meter, and nothing here pretends it does.
	UploadedMediaBytesMeter = "uploaded_media_bytes"
)

// meters is the set this application counts. Adding one here is not enough to make it count:
// something has to record against it, and a meter nothing records against is a row of zeroes.
//
// Names are load-bearing. They travel into metric attributes and into provider-side idempotency
// keys, and renaming one starts a new empty count while the old one stops being flushed.
var meters = []platformmetering.Meter{
	{
		Name:        UploadedMediaBytesMeter,
		Unit:        "bytes",
		Aggregation: platformmetering.AggregationSum,
		Period:      platformmetering.PeriodMonth,
	},
}

// NewRegistry builds the registry every metering component in this application reads.
//
// Each meter gets a static quota alongside it, because Enforcer.Check errors on a meter with no
// quota registered rather than treating the absence as permission. The static quotas are the
// fallback the platform serves when no QuotaSource is configured; the ones that actually answer
// on the read path come from the entitlements catalog, in internal/entitlements.
func NewRegistry() (*platformmetering.Registry, error) {
	registry := platformmetering.NewRegistry()

	for idx := range meters {
		meter := &meters[idx]

		if err := registry.RegisterMeter(*meter); err != nil {
			return nil, errors.Wrapf(err, "registering meter %q", meter.Name)
		}

		// The platform refuses to read an absent quota as an unlimited one — unmetered and
		// unlimited are different facts — and a limit of zero means no usage is allowed. An
		// unlimited quota is stated, and the registry reads its period off the meter.
		if err := registry.RegisterUnlimitedQuota(meter.Name); err != nil {
			return nil, errors.Wrapf(err, "registering quota for meter %q", meter.Name)
		}
	}

	return registry, nil
}
