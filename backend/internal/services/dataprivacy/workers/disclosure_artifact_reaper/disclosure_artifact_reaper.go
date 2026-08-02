/*
Package disclosureartifactreaper destroys the object behind a user data disclosure once the
disclosure has expired.

Without it, `expires_at` on a disclosure is decorative: the row says the report is no longer
available while the object holding everything the system knows about that person stays in the
bucket indefinitely. This job is what makes the column mean something.
*/
package disclosureartifactreaper

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/reportartifacts"

	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/hashicorp/go-multierror"
)

const (
	o11yName = "disclosure_artifact_reaper"

	// maxBatchesPerRun bounds one execution. A full batch means there is probably more waiting,
	// so the job keeps going rather than leaving a backlog to drain one batch per interval —
	// which matters on the first run after deployment, when every artifact ever written is
	// already past its expiry. The bound is what keeps that first run from outliving its lease.
	maxBatchesPerRun = 20
)

type (
	// Worker reaps expired disclosure artifacts.
	Worker struct {
		logger          logging.Logger
		tracer          tracing.Tracer
		disclosureRepo  dataprivacy.UserDataDisclosureDataManager
		reportArtifacts reportartifacts.Store
		reapedCounter   metrics.Int64Counter
	}
)

// NewDisclosureArtifactReaper builds the reaper.
func NewDisclosureArtifactReaper(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	disclosureRepo dataprivacy.UserDataDisclosureDataManager,
	reportArtifacts reportartifacts.Store,
) (*Worker, error) {
	reapedCounter, err := metricsProvider.NewInt64Counter(o11yName + ".artifacts_reaped")
	if err != nil {
		return nil, err
	}

	return &Worker{
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		disclosureRepo:  disclosureRepo,
		reportArtifacts: reportArtifacts,
		reapedCounter:   reapedCounter,
	}, nil
}

// Work destroys the artifact behind every expired disclosure and records each one as expired.
//
// A disclosure that fails to reap is left untouched — not marked expired, not skipped
// permanently — so the next run picks it up again. The batch loop stops as soon as a pass reaps
// nothing new, because a batch that is full of rows the reaper cannot destroy would otherwise be
// re-fetched until the job hit its timeout.
func (w *Worker) Work(ctx context.Context) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	errs := &multierror.Error{}

	for batch := range maxBatchesPerRun {
		disclosures, err := w.disclosureRepo.GetExpiredUserDataDisclosures(ctx)
		if err != nil {
			return observability.PrepareAndLogError(err, w.logger, span, "fetching expired user data disclosures")
		}

		if len(disclosures) == 0 {
			return errs.ErrorOrNil()
		}

		reaped := 0
		for _, disclosure := range disclosures {
			if err = w.reap(ctx, span, disclosure); err != nil {
				errs = multierror.Append(errs, err)
				continue
			}

			reaped++
		}

		w.reapedCounter.Add(ctx, int64(reaped))

		// Nothing moved, so asking again returns the same rows. Stop rather than spin.
		if reaped == 0 {
			w.logger.WithValue("batch", batch).WithValue("stuck_qty", len(disclosures)).
				Info("no expired disclosures could be reaped, stopping early")
			return errs.ErrorOrNil()
		}

		// A short batch is the last one.
		if len(disclosures) < dataprivacy.ExpiredUserDataDisclosureBatchSize {
			return errs.ErrorOrNil()
		}
	}

	w.logger.Info("reached the per-run batch limit, remaining artifacts will be reaped on the next run")

	return errs.ErrorOrNil()
}

// reap destroys one disclosure's artifact and records the disclosure as expired.
func (w *Worker) reap(ctx context.Context, span tracing.Span, disclosure *dataprivacy.UserDataDisclosure) error {
	logger := w.logger.
		WithValue(dataprivacykeys.UserDataDisclosureIDKey, disclosure.ID).
		WithValue(dataprivacykeys.UserDataAggregationReportIDKey, disclosure.ReportID)

	// The object goes first. Flipping the row to expired before the object is gone would leave
	// an artifact that nothing ever comes back for, since the query that finds work excludes
	// rows already marked expired. A disclosure that never produced a report has nothing to
	// destroy and only needs the status.
	if disclosure.ReportID != "" {
		if err := w.reportArtifacts.Delete(ctx, disclosure.ReportID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "deleting expired disclosure artifact")
		}
	}

	if err := w.disclosureRepo.MarkUserDataDisclosureExpired(ctx, disclosure.ID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking disclosure expired")
	}

	return nil
}
