package config

import (
	"context"
	"errors"
	"fmt"

	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	auditcfg "github.com/primandproper/platform-go/v15/audit/config"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/service"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"
	notificationscfg "github.com/primandproper/primitives-go/v2/notifications/mobile/config"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

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

		Queues queuescfg.Config `envPrefix:"QUEUES_" json:"queues,omitzero"`

		// PushNotifications is the sender the meal plan task notification worker delivers
		// through. It lives here because that worker sends under its queue lease rather
		// than by publishing a message — see internal/services/mealplanning/workers/
		// meal_plan_task_notifications for why the send cannot be somebody else's job.
		//
		// It is not Service.MobileNotifications, because service.Config validates every block
		// it holds and this one cannot pass at render time: its APNs credentials arrive as
		// environment variables at startup, so a rendered config file has none of them. The
		// async message handler carries the same struct for the same reason, so both
		// processes push through one configuration.
		PushNotifications notificationscfg.Config `envPrefix:"PUSH_NOTIFICATIONS_" json:"pushNotifications,omitzero"`

		// DataPrivacyArtifactEncryptionKey is the key export artifacts are sealed under,
		// filed under Service.DataPrivacy.Artifacts.Encryption's CurrentKeyID. It is not part
		// of that block because platform does not take keys from configuration — a keyring
		// is built over an encryption.Keyset the container supplies — so this is where the
		// one key this deployment has comes from.
		//
		// The API server carries the same value, because it opens what this process seals.
		DataPrivacyArtifactEncryptionKey string `env:"DATA_PRIVACY_ARTIFACT_ENCRYPTION_KEY" json:"dataPrivacyArtifactEncryptionKey,omitempty"`

		// Service is everything platform composes for this process: the database, the
		// broker, the pillars, the job scheduler and its lock, and every platform-owned
		// loop this process exists to run — operations, sagas, webhook delivery, the
		// retention sweep, the metering flusher, and data privacy fulfillment. It is read
		// by service.Register, which registers what is present and nothing else.
		//
		// It carries no envPrefix, so its blocks keep the names they always had here —
		// DATABASE_, OBSERVABILITY_, OPERATIONS_ — rather than growing a second prefix in
		// front of them.
		//
		// Operations and Saga each bring the reapers their store owns, and service.New
		// hands those to the scheduler beside this application's own jobs: operations'
		// recovery and reap, and saga's retention. Recovery is the one that is easy to
		// miss — an operation whose worker died between its insert and its enqueue sits
		// pending until a recovery pass re-offers it, and before this process was composed
		// from a service.Config nothing ever ran one.
		Service service.Config `json:"service,omitzero"`

		Search textsearchcfg.Config `envPrefix:"SEARCH_" json:"search,omitzero"`

		// AuditLog carries the retention window for the audit log, which the retention sweep
		// prunes under. It is not Service.Audit for the same kind of reason OutboxRelay is not
		// Service.Outbox: this application's recorder is platform's with the impersonating
		// administrator attached to each entry, and Service.Audit would register platform's
		// bare one beside it. AUDIT_LOG_ rather than AUDIT_, so nothing set here configures
		// Service.Audit by accident.
		AuditLog auditcfg.Config `envPrefix:"AUDIT_LOG_" json:"auditLog,omitzero"`

		Jobs ScheduledJobsConfig `envPrefix:"JOBS_" json:"jobs,omitzero"`

		// OutboxRelay moves events written inside a caller's transaction onto the broker.
		//
		// It is not Service.Outbox, and the difference is one option. Every outbox row this
		// application writes goes through a writer carrying the search index side effect
		// (internal/indexevents), and platform's outbox/config builds its writer from
		// configuration alone, with no way to hand it one. Configuring Service.Outbox would
		// register that bare writer beside this application's own, which samber/do refuses.
		// So the outbox — and with it the recording spine, which service.Register builds
		// only when Audit, Webhooks and Outbox are all present — stays wired by hand. The
		// prefix is OUTBOX_RELAY_ rather than OUTBOX_ so that nothing set for this block can
		// switch Service.Outbox on.
		OutboxRelay outbox.RelayConfig `envPrefix:"OUTBOX_RELAY_" json:"outboxRelay,omitzero"`
	}

	// ScheduledJobsConfig carries the schedule for each of this application's own jobs. The
	// scheduler that runs them, and the lock that serializes them across replicas, are
	// Service.JobsScheduler; the jobs platform schedules for itself carry their own schedules
	// inside their own blocks.
	ScheduledJobsConfig struct {
		_ struct{} `json:"-"`

		SearchDataIndexScheduler jobscfg.JobConfig `envPrefix:"SEARCH_DATA_INDEX_SCHEDULER_" json:"searchDataIndexScheduler,omitzero"`
		QueueTest                jobscfg.JobConfig `envPrefix:"QUEUE_TEST_"                  json:"queueTest,omitzero"`

		// DataPrivacySweep expires export artifacts, lapses unconfirmed erasures, and
		// samples the overdue gauge. Disabling it does not pause expiry so much as
		// abandon it: every artifact ever written — each one everything the system knows
		// about one person — stays in the bucket and nothing else will ever delete it.
		DataPrivacySweep jobscfg.JobConfig `envPrefix:"DATA_PRIVACY_SWEEP_" json:"dataPrivacySweep,omitzero"`

		// AuditRetentionSweeper prunes audit entries past the retention window in
		// SchedulerConfig.AuditLog. Disabling it does not pause retention so much as
		// abandon it: the log grows without bound and nothing else will trim it.
		//
		// It is a scheduled job rather than the Sweeper's own Run loop so that one
		// replica prunes per tick, by construction rather than by convention. The
		// Sweeper is safe to run concurrently — it prunes a prefix of a chain inside a
		// transaction — but every replica sweeping every hour is the same work done
		// several times for one result, and it is work that deletes.
		AuditRetentionSweeper jobscfg.JobConfig `envPrefix:"AUDIT_RETENTION_SWEEPER_" json:"auditRetentionSweeper,omitzero"`
		// MeteringFlusher posts accumulated usage to the billing provider and reaps the
		// usage event ledger past its retention. Disabling it stops neither the counting
		// nor the totals it feeds — the recorder is in the API server — but the event
		// ledger then grows without bound.
		MeteringFlusher jobscfg.JobConfig `envPrefix:"METERING_FLUSHER_" json:"meteringFlusher,omitzero"`

		// Domain: mealplanning — swapping the domain replaces this field and the type it
		// names, and touches nothing else in this struct.
		MealPlanning MealPlanningScheduledJobsConfig `envPrefix:"MEAL_PLANNING_" json:"mealPlanning,omitzero"`
	}
)

var _ validation.ValidatableWithContext = (*ScheduledJobsConfig)(nil)

// ValidateWithContext validates a ScheduledJobsConfig struct.
func (cfg *ScheduledJobsConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"SearchDataIndexScheduler": cfg.SearchDataIndexScheduler.ValidateWithContext,
		"QueueTest":                cfg.QueueTest.ValidateWithContext,
		"DataPrivacySweep":         cfg.DataPrivacySweep.ValidateWithContext,
		"AuditRetentionSweeper":    cfg.AuditRetentionSweeper.ValidateWithContext,
		"MeteringFlusher":          cfg.MeteringFlusher.ValidateWithContext,
		"MealPlanning":             cfg.MealPlanning.ValidateWithContext,
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
//
// Service is validated first and on its own, because its validation is not only a check: it
// releases the blocks env parsing allocated and nobody filled in, and until it has run every
// platform subsystem looks configured. Everything that reads Service afterwards — service.Register
// above all — has to see what is left.
func (cfg *SchedulerConfig) ValidateWithContext(ctx context.Context) error {
	if err := cfg.Service.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Service config: %w", err)
	}

	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		sectionQueues: cfg.Queues.ValidateWithContext,
		"Search":      cfg.Search.ValidateWithContext,
		"Jobs":        cfg.Jobs.ValidateWithContext,
		"OutboxRelay": cfg.OutboxRelay.ValidateWithContext,
		"AuditLog":    cfg.AuditLog.ValidateWithContext,
		sectionService: func(context.Context) error {
			return requireBlocks(map[string]bool{
				sectionDatabase:     cfg.Service.Database != nil,
				sectionMessageQueue: cfg.Service.MessageQueue != nil,
				"JobsScheduler":     cfg.Service.JobsScheduler != nil,
				// The saga worker's per-instance lock is a scoped locker over this one.
				"DistributedLock": cfg.Service.DistributedLock != nil,
				"Operations":      cfg.Service.Operations != nil,
				"DataPrivacy":     cfg.Service.DataPrivacy != nil,
				"Saga":            cfg.Service.Saga != nil,
				"Webhooks":        cfg.Service.Webhooks != nil,
				"Metering":        cfg.Service.Metering != nil,
				"Retention":       cfg.Service.Retention != nil,
				"Capitalism":      cfg.Service.Capitalism != nil,
			})
		},
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

// requireBlocks reports every service.Config block a process cannot run without and that
// normalization left nil.
//
// A process composed from a service.Config treats a missing block as a subsystem nobody asked
// for, which is right for a library and wrong for a process whose whole purpose is that
// subsystem: a scheduler with no Operations block is a scheduler that boots, reports healthy,
// and never fulfills a privacy request. So the blocks a process exists to run are asserted by
// name here, where a missing one is a red render rather than a quiet deployment.
func requireBlocks(present map[string]bool) error {
	var errs []error

	for name, ok := range present {
		if !ok {
			errs = append(errs, fmt.Errorf("%s is required", name))
		}
	}

	return errors.Join(errs...)
}
