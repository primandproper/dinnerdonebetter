package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	"github.com/primandproper/platform-go/v9/audit"
	capitalismcfg "github.com/primandproper/platform-go/v9/capitalism/config"
	distributedlockcfg "github.com/primandproper/platform-go/v9/distributedlock/config"
	"github.com/primandproper/platform-go/v9/jobs"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	meteringcfg "github.com/primandproper/platform-go/v9/metering/config"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/outbox"
	"github.com/primandproper/platform-go/v9/saga"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"
	webhookscfg "github.com/primandproper/platform-go/v9/webhooks/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

type (
	// SchedulerConfig configures the long-lived worker that runs every periodic job. It
	// replaces one Kubernetes CronJob per job: each execution is held under a distributedlock
	// lease, so every replica ticks and only the one that wins the lock runs.
	//
	// Both shapes of periodic work live here. A job that wants a frequency takes an Interval;
	// a job that wants a wall-clock time takes a cron Schedule, which every replica reads the
	// same way rather than phasing off whenever its pod last restarted.
	SchedulerConfig struct {
		_ struct{} `json:"-"`

		Queues queuescfg.Config `envPrefix:"QUEUES_" json:"queues"`

		// Capitalism is where the flusher's usage reporter comes from. It lives in this
		// process rather than the API server's because usage reporting happens on a
		// scheduler tick, and a request path has no business holding the credentials for
		// it. A provider of "noop" is a supported deployment and the current one: usage
		// accumulates durably and nothing reaches a billing provider.
		Capitalism    capitalismcfg.Config `envPrefix:"CAPITALISM_"    json:"capitalism"`
		Events        msgconfig.Config     `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability"`
		Analytics     analyticscfg.Config  `envPrefix:"ANALYTICS_"     json:"analytics"`
		Search        textsearchcfg.Config `envPrefix:"SEARCH_"        json:"search"`

		Jobs ScheduledJobsConfig `envPrefix:"JOBS_" json:"jobs"`

		// Audit carries the retention window for the audit log. It lives here for the
		// same reason the outbox does — the sweeper is a background loop over the
		// database — and it runs in exactly one process, unlike the Recorder, which runs
		// wherever a mutation does.
		Audit    audit.SweeperConfig `envPrefix:"AUDIT_"    json:"audit"`
		Database dbcfg.Config        `envPrefix:"DATABASE_" json:"database"`

		// Outbox moves events written inside a caller's transaction onto the broker. It
		// lives here because it is a background loop, which is what this process is for,
		// and because it needs exactly what this process already has: the database and a
		// publisher provider.
		Outbox outbox.RelayConfig `envPrefix:"OUTBOX_" json:"outbox"`

		// Sagas advances every durable saga instance this build knows how to run. It is the
		// other half of the scheduled jobs that start them: a job writes an instance, this
		// loop steps it through, and it polls in seconds rather than minutes because the
		// poll interval is the floor on how long a step's delay costs.
		Sagas saga.WorkerConfig `envPrefix:"SAGAS_" json:"sagas"`

		// Metering is the same struct the API server carries, because the flusher has to
		// read the tables the API server's recorder wrote. Only the flusher half is used
		// here; the recorder and enforcer knobs are carried anyway so the table prefix
		// cannot drift between the process that counts and the process that bills.
		Metering meteringcfg.Config `envPrefix:"METERING_" json:"metering"`

		// Webhooks configures the outbound webhook delivery worker, which lives here for
		// the same reasons the outbox relay does: it is a polling loop that must not be
		// tied to a request, and it needs exactly what this process already has.
		//
		// Its own tick also reaps delivered dispatches and their attempts past the
		// retention window, so retention needs no separate scheduled job.
		Webhooks webhookscfg.Config `envPrefix:"WEBHOOKS_" json:"webhooks"`

		// DataPrivacy configures the fulfillment worker and the expiry sweep, both of
		// which run here: the request table, the artifact bucket, and the cipher. It is
		// the same struct the API server is configured with, because the two have to agree
		// on all three or an artifact this process writes is not one the API can read —
		// and a sweep pointed at the wrong bucket deletes nothing and reports success.
		//
		// The async message handler no longer carries it. It stopped touching artifacts
		// when aggregation moved off the queue, and a process holding an encryption key it
		// has no use for is exposure with nothing on the other side of it.
		DataPrivacy dataprivacycfg.Config `envPrefix:"DATA_PRIVACY_" json:"dataPrivacy"`
	}

	// ScheduledJobsConfig carries the scheduler's own knobs, the lock backend that serializes
	// executions across replicas, and the schedule for each registered job.
	ScheduledJobsConfig struct {
		_ struct{} `json:"-"`

		Scheduler jobs.SchedulerConfig `envPrefix:"SCHEDULER_" json:"scheduler"`

		// Lock decides which replica runs a given tick. The noop locker acquires
		// unconditionally, which means every replica runs every job — right for a
		// single-replica deployment, wrong the moment it scales.
		Lock distributedlockcfg.Config `envPrefix:"LOCK_" json:"lock"`

		SearchDataIndexScheduler    ScheduledJobConfig `envPrefix:"SEARCH_DATA_INDEX_SCHEDULER_"   json:"searchDataIndexScheduler"`
		MobileNotificationScheduler ScheduledJobConfig `envPrefix:"MOBILE_NOTIFICATION_SCHEDULER_" json:"mobileNotificationScheduler"`
		QueueTest                   ScheduledJobConfig `envPrefix:"QUEUE_TEST_"                    json:"queueTest"`

		// DataPrivacySweep expires export artifacts, lapses unconfirmed erasures, and
		// samples the overdue gauge. Disabling it does not pause expiry so much as
		// abandon it: every artifact ever written — each one everything the system knows
		// about one person — stays in the bucket and nothing else will ever delete it.
		DataPrivacySweep ScheduledJobConfig `envPrefix:"DATA_PRIVACY_SWEEP_" json:"dataPrivacySweep"`

		// AuditRetentionSweeper prunes audit entries past the retention window in
		// SchedulerConfig.Audit. Disabling it does not pause retention so much as
		// abandon it: the log grows without bound and nothing else will trim it.
		//
		// It is a scheduled job rather than the Sweeper's own Run loop so that one
		// replica prunes per tick, by construction rather than by convention. The
		// Sweeper is safe to run concurrently — it prunes a prefix of a chain inside a
		// transaction — but every replica sweeping every hour is the same work done
		// several times for one result, and it is work that deletes.
		AuditRetentionSweeper ScheduledJobConfig `envPrefix:"AUDIT_RETENTION_SWEEPER_" json:"auditRetentionSweeper"`
		// MeteringFlusher posts accumulated usage to the billing provider and reaps the
		// usage event ledger past its retention. Disabling it stops neither the counting
		// nor the totals it feeds — the recorder is in the API server — but the event
		// ledger then grows without bound.
		MeteringFlusher ScheduledJobConfig `envPrefix:"METERING_FLUSHER_" json:"meteringFlusher"`

		// Domain: mealplanning — swapping the domain replaces this field and the type it
		// names, and touches nothing else in this struct.
		MealPlanning MealPlanningScheduledJobsConfig `envPrefix:"MEAL_PLANNING_" json:"mealPlanning"`
	}

	// ScheduledJobConfig is one job's schedule. Exactly one of Schedule and Interval is set:
	// a job either belongs at an hour or at a frequency, and a job carrying both is rejected
	// rather than resolved by precedence.
	ScheduledJobConfig struct {
		_ struct{} `json:"-"`

		// Schedule is a five-field crontab expression — minute, hour, day of month, month,
		// day of week — for work that belongs at a wall-clock time rather than at a
		// frequency. The usual descriptors (@daily, @hourly) are accepted too.
		//
		// The zone is the scheduler's Timezone unless the expression names its own with a
		// CRON_TZ= prefix, which is how one job opts into a calendar the rest do not share.
		// Anything but UTC needs the zoneinfo database; cmd/workers/scheduler embeds it.
		//
		// There is no catch-up. A fire time that passes while the process is down, or while
		// the previous run is still going, is skipped rather than queued.
		Schedule string `env:"SCHEDULE" json:"schedule"`

		// Interval is how often the job fires. Ticks are not queued: a job that overruns its
		// interval fires again as soon as it finishes rather than accumulating a backlog.
		Interval time.Duration `env:"INTERVAL" json:"interval"`

		// Timeout bounds one execution.
		Timeout time.Duration `env:"TIMEOUT" json:"timeout"`

		// LeaseTTL is how long the lock is held. It is not renewed while the job runs, so it
		// must comfortably exceed the job's worst-case duration — past it, a second replica
		// may start the same job.
		LeaseTTL time.Duration `env:"LEASE_TTL" json:"leaseTTL"`

		// Enabled registers the job. A disabled job is not registered at all rather than
		// registered and skipped, so it costs nothing and reports nothing.
		Enabled bool `env:"ENABLED" json:"enabled"`

		// RunOnStart fires the job once at startup instead of waiting a full interval or for
		// the schedule's next fire time.
		RunOnStart bool `env:"RUN_ON_START" json:"runOnStart"`
	}
)

// Job renders the config as the jobs.Job the Scheduler registers, under the given name and
// running the given work.
//
// Which of Interval and Schedule a job is shaped by stays in here rather than at the call site:
// jobs.Job takes the two as separate fields and rejects a job that sets both, so the mapping is
// the other half of the invariant ValidateWithContext enforces.
func (cfg *ScheduledJobConfig) Job(name string, run func(context.Context) error) (jobs.Job, error) {
	job := jobs.Job{
		Name:       name,
		Interval:   cfg.Interval,
		Timeout:    cfg.Timeout,
		LeaseTTL:   cfg.LeaseTTL,
		RunOnStart: cfg.RunOnStart,
		Run:        run,
	}

	if cfg.Schedule != "" {
		schedule, err := jobs.Cron(cfg.Schedule)
		if err != nil {
			return jobs.Job{}, fmt.Errorf("parsing cron schedule for job %q: %w", name, err)
		}

		job.Schedule = schedule
	}

	return job, nil
}

var _ validation.ValidatableWithContext = (*ScheduledJobConfig)(nil)

// ValidateWithContext validates a ScheduledJobConfig struct. A disabled job is not validated:
// it is never registered, so its schedule is inert.
func (cfg *ScheduledJobConfig) ValidateWithContext(ctx context.Context) error {
	if !cfg.Enabled {
		return nil
	}

	// Checked here rather than left to jobs.Job.validate so that a bad schedule fails config
	// rendering in CI, where it is a red build, instead of scheduler startup, where it is a
	// crash loop.
	switch {
	case cfg.Schedule != "" && cfg.Interval > 0:
		return fmt.Errorf("scheduled job sets both an interval and a cron schedule %q", cfg.Schedule)
	case cfg.Schedule == "" && cfg.Interval <= 0:
		return errors.New("scheduled job sets neither an interval nor a cron schedule")
	}

	if cfg.Schedule != "" {
		if _, err := jobs.Cron(cfg.Schedule); err != nil {
			return fmt.Errorf("parsing cron schedule: %w", err)
		}
	}

	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Interval, validation.When(cfg.Schedule == "", validation.Min(time.Second))),
		validation.Field(&cfg.LeaseTTL, validation.Required, validation.Min(time.Second)),
		validation.Field(&cfg.Timeout, validation.Min(time.Duration(0))),
	)
}

var _ validation.ValidatableWithContext = (*ScheduledJobsConfig)(nil)

// ValidateWithContext validates a ScheduledJobsConfig struct.
func (cfg *ScheduledJobsConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Scheduler":                   cfg.Scheduler.ValidateWithContext,
		"Lock":                        cfg.Lock.ValidateWithContext,
		"SearchDataIndexScheduler":    cfg.SearchDataIndexScheduler.ValidateWithContext,
		"MobileNotificationScheduler": cfg.MobileNotificationScheduler.ValidateWithContext,
		"QueueTest":                   cfg.QueueTest.ValidateWithContext,
		"DataPrivacySweep":            cfg.DataPrivacySweep.ValidateWithContext,
		"AuditRetentionSweeper":       cfg.AuditRetentionSweeper.ValidateWithContext,
		"MeteringFlusher":             cfg.MeteringFlusher.ValidateWithContext,
		"MealPlanning":                cfg.MealPlanning.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*SchedulerConfig)(nil)

// ValidateWithContext validates a SchedulerConfig struct.
func (cfg *SchedulerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Queues":        cfg.Queues.ValidateWithContext,
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Search":        cfg.Search.ValidateWithContext,
		"DataPrivacy":   cfg.DataPrivacy.ValidateWithContext,
		"Jobs":          cfg.Jobs.ValidateWithContext,
		"Outbox":        cfg.Outbox.ValidateWithContext,
		"Audit":         cfg.Audit.ValidateWithContext,
		"Webhooks":      cfg.Webhooks.ValidateWithContext,
		"Sagas":         cfg.Sagas.ValidateWithContext,
		"Metering":      cfg.Metering.ValidateWithContext,
		"Capitalism":    cfg.Capitalism.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
