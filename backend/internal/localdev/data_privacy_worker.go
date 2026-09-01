package localdev

import (
	"context"
	"fmt"
	"time"

	schedulerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	"github.com/primandproper/platform-go/v13/clock"
	"github.com/primandproper/platform-go/v13/database"
	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	platformdataprivacycfg "github.com/primandproper/platform-go/v13/dataprivacy/config"
	"github.com/primandproper/platform-go/v13/operations"

	"github.com/samber/do/v2"
)

// DataPrivacyFulfillment is the half of data privacy that does not run in the API server: the
// operations worker that fulfills a request, and the sweeper that expires what it produced.
//
// It exists for the reason StartSagaWorker does. The API server records a subject access
// request and returns; the gather, the artifact, and the erasure all happen in the scheduler,
// so an in-process harness that wants to assert an export was produced has to stand that half
// up itself. What it stands up is the scheduler's own container — the same registry, the same
// collectors, the same bucket and cipher — rather than a hand-assembled approximation of it,
// because the failure this is here to catch is a domain missing from an export, and a
// hand-assembled registry is one that cannot miss anything.
//
// Nothing here is a fake. The queue is the operations queue over the real table, the claim and
// the lease are the ones production runs, and the artifact is written to the configured bucket
// and read back through the API server's own Open. The only thing a caller supplies is when to
// run: Worker.Run blocks and is stopped by cancelling its context, which is what lets a test
// submit a request and then wait for an outcome rather than for a poll interval.
type DataPrivacyFulfillment struct {
	// Worker claims privacy operations and runs them. Run blocks until its context is done.
	Worker *operations.Worker

	// Sweeper expires artifacts past their window, lapses unconfirmed erasures, and reaps
	// terminal request records. It is driven by a scheduled job in production, so a caller
	// advances it by calling Sweep.
	Sweeper *platformdataprivacy.Sweeper

	// Registry is the set of collectors and erasers the worker fulfills through. An export is
	// exactly what this holds, so a caller asserting on which domains an export covers asks it
	// rather than reading a document that omits the domains a subject happens to be absent
	// from.
	Registry *platformdataprivacy.Registry

	// Store is the request table. It is exposed because there are requests the API cannot be
	// asked about: an erasure ends by deleting the subject, so the one principal entitled to
	// read that request's outcome no longer exists. An operator reads the row; so does a test.
	Store platformdataprivacy.Store

	// Shutdown releases the container: the connection pool this half opened, and the queue's
	// goroutine.
	Shutdown func(context.Context) error

	artifacts dataprivacycfg.ArtifactUploadManager
	cfg       *platformdataprivacycfg.Config
}

// NewDataPrivacyFulfillment builds the fulfillment worker and the sweeper from a scheduler
// configuration.
//
// It takes the scheduler's own configuration rather than deriving one from the API server's,
// which makes the agreement between the two processes something a caller can test rather than
// something this function arranges. They have to agree on the bucket, the cipher, the request
// table and the operations queue — an artifact written under one key is not one the other can
// open, and a request submitted under one queue name is not one anything else claims — and a
// harness that fed both halves the same struct could never notice if they stopped.
func NewDataPrivacyFulfillment(ctx context.Context, cfg *config.SchedulerConfig) (*DataPrivacyFulfillment, error) {
	i := schedulerbuild.BuildInjector(ctx, cfg)

	worker, err := do.Invoke[*operations.Worker](i)
	if err != nil {
		return nil, fmt.Errorf("building the operations worker: %w", err)
	}

	sweeper, err := do.Invoke[*platformdataprivacy.Sweeper](i)
	if err != nil {
		return nil, fmt.Errorf("building the data privacy sweeper: %w", err)
	}

	store, err := do.Invoke[platformdataprivacy.Store](i)
	if err != nil {
		return nil, fmt.Errorf("building the data privacy store: %w", err)
	}

	artifacts, err := do.Invoke[dataprivacycfg.ArtifactUploadManager](i)
	if err != nil {
		return nil, fmt.Errorf("building the artifact upload manager: %w", err)
	}

	registry, err := do.Invoke[*platformdataprivacy.Registry](i)
	if err != nil {
		return nil, fmt.Errorf("building the data privacy registry: %w", err)
	}

	return &DataPrivacyFulfillment{
		Worker:   worker,
		Sweeper:  sweeper,
		Registry: registry,
		Store:    store,
		Shutdown: func(ctx context.Context) error {
			if report := i.ShutdownWithContext(ctx); report != nil && !report.Succeed {
				return report
			}

			return nil
		},
		artifacts: artifacts,
		cfg: dataprivacycfg.PlatformConfig(
			&cfg.DataPrivacy,
			do.MustInvoke[database.Client](i),
		),
	}, nil
}

// SweeperAt returns a second Sweeper over the same store and bucket that believes the time is
// now.
//
// An export artifact survives for a week, and a sweep is the thing that ends that. Neither
// number is something a test can wait out, and neither is something to shorten in configuration
// either: the TTL is stamped into the request row by the worker that completed it, so lowering
// it would expire every artifact the rest of the suite produced. Moving one sweeper's clock
// instead leaves the rows alone and asks the question the sweep actually answers — what is past
// its window as of this instant.
func (f *DataPrivacyFulfillment) SweeperAt(ctx context.Context, now time.Time) (*platformdataprivacy.Sweeper, error) {
	return platformdataprivacycfg.NewSweeper(
		ctx,
		f.cfg,
		f.Store,
		f.artifacts.UploadManager,
		platformdataprivacycfg.WithSweeperOptions(platformdataprivacy.WithSweeperClock(fixedClock{now: now})),
	)
}

// fixedClock is a clock stopped at an instant. Only Now and Since are meaningful; the sweep
// neither sleeps nor ticks.
type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

func (c fixedClock) Since(t time.Time) time.Duration { return c.now.Sub(t) }

func (c fixedClock) Sleep(context.Context, time.Duration) error { return nil }

func (c fixedClock) NewTicker(d time.Duration) clock.Ticker { return clock.NewClock().NewTicker(d) }
