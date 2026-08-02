package config

import (
	"context"
	"fmt"
	"time"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	distributedlockcfg "github.com/primandproper/platform-go/v9/distributedlock/config"
	"github.com/primandproper/platform-go/v9/jobs"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/outbox"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

type (
	// SchedulerConfig configures the long-lived worker that runs every interval-shaped periodic
	// job. It replaces one Kubernetes CronJob per job: each execution is held under a
	// distributedlock lease, so every replica ticks and only the one that wins the lock runs.
	//
	// Calendar-shaped work ("at midnight on the 1st") is deliberately not here. jobs.Scheduler
	// takes intervals, not cron expressions, and an interval that starts whenever the pod last
	// restarted is not the same thing as a wall-clock time. Those jobs stay CronJobs.
	SchedulerConfig struct {
		_ struct{} `json:"-"`

		Queues        queuescfg.Config     `envPrefix:"QUEUES_"        json:"queues"`
		Events        msgconfig.Config     `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability"`
		Analytics     analyticscfg.Config  `envPrefix:"ANALYTICS_"     json:"analytics"`
		Search        textsearchcfg.Config `envPrefix:"SEARCH_"        json:"search"`
		Database      dbcfg.Config         `envPrefix:"DATABASE_"      json:"database"`

		// Outbox moves events written inside a caller's transaction onto the broker. It
		// lives here because it is a background loop, which is what this process is for,
		// and because it needs exactly what this process already has: the database and a
		// publisher provider.
		Outbox outbox.RelayConfig  `envPrefix:"OUTBOX_" json:"outbox"`
		Jobs   ScheduledJobsConfig `envPrefix:"JOBS_"   json:"jobs"`
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

		// Domain: mealplanning — swapping the domain replaces this field and the type it
		// names, and touches nothing else in this struct.
		MealPlanning MealPlanningScheduledJobsConfig `envPrefix:"MEAL_PLANNING_" json:"mealPlanning"`
	}

	// ScheduledJobConfig is one job's schedule.
	ScheduledJobConfig struct {
		_ struct{} `json:"-"`

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

		// RunOnStart fires the job once at startup instead of waiting a full interval.
		RunOnStart bool `env:"RUN_ON_START" json:"runOnStart"`
	}
)

var _ validation.ValidatableWithContext = (*ScheduledJobConfig)(nil)

// ValidateWithContext validates a ScheduledJobConfig struct. A disabled job is not validated:
// it is never registered, so its schedule is inert.
func (cfg *ScheduledJobConfig) ValidateWithContext(ctx context.Context) error {
	if !cfg.Enabled {
		return nil
	}

	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Interval, validation.Required, validation.Min(time.Second)),
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
		"Jobs":          cfg.Jobs.ValidateWithContext,
		"Outbox":        cfg.Outbox.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
