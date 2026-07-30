package mobilenotificationscheduler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	"github.com/primandproper/platform-go/v8/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*msgconfig.QueuesConfig](i, func(i do.Injector) (*msgconfig.QueuesConfig, error) {
		cfg := do.MustInvoke[*config.MobileNotificationSchedulerConfig](i)
		return &cfg.Queues, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*config.MobileNotificationSchedulerConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		return ProvideObservabilityConfig(do.MustInvoke[*config.MobileNotificationSchedulerConfig](i)), nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		return ProvideEventsConfig(do.MustInvoke[*config.MobileNotificationSchedulerConfig](i)), nil
	})
	do.Provide[messagequeue.Publisher](i, func(i do.Injector) (messagequeue.Publisher, error) {
		ctx := do.MustInvoke[context.Context](i)
		publisherProvider := do.MustInvoke[messagequeue.PublisherProvider](i)
		queues := do.MustInvoke[*msgconfig.QueuesConfig](i)
		return ProvideMobileNotificationsPublisher(ctx, publisherProvider, queues)
	})
	do.Provide[*Scheduler](i, func(i do.Injector) (*Scheduler, error) {
		return NewScheduler(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[messagequeue.Publisher](i),
		), nil
	})
}

// ProvideObservabilityConfig provides the observability config.
func ProvideObservabilityConfig(cfg *config.MobileNotificationSchedulerConfig) *observability.Config {
	return &cfg.Observability
}

// ProvideEventsConfig provides the message queue events config.
func ProvideEventsConfig(cfg *config.MobileNotificationSchedulerConfig) *msgconfig.Config {
	return &cfg.Events
}

// ProvideMobileNotificationsPublisher provides a publisher for the mobile_notifications topic.
func ProvideMobileNotificationsPublisher(
	ctx context.Context,
	messageQueuePublisherProvider messagequeue.PublisherProvider,
	queues *msgconfig.QueuesConfig,
) (messagequeue.Publisher, error) {
	return messageQueuePublisherProvider.NewPublisher(ctx, queues.MobileNotificationsTopicName)
}
