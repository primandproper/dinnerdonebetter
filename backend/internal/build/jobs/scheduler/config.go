package scheduler

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	analyticscfg "github.com/primandproper/platform-go/v13/analytics/config"
	capitalismcfg "github.com/primandproper/platform-go/v13/capitalism/config"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	msgconfig "github.com/primandproper/platform-go/v13/messagequeue/config"
	meteringcfg "github.com/primandproper/platform-go/v13/metering/config"
	notificationscfg "github.com/primandproper/platform-go/v13/notifications/mobile/config"
	"github.com/primandproper/platform-go/v13/observability"
	operationscfg "github.com/primandproper/platform-go/v13/operations/config"
	"github.com/primandproper/platform-go/v13/saga"
	textsearchcfg "github.com/primandproper/platform-go/v13/search/text/config"
	webhookscfg "github.com/primandproper/platform-go/v13/webhooks/config"
	"github.com/primandproper/platform-go/v13/workqueue"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*queuescfg.Config](i, func(i do.Injector) (*queuescfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Queues, nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Events, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Observability, nil
	})
	do.Provide[*dbcfg.Config](i, func(i do.Injector) (*dbcfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Database, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		return &do.MustInvoke[*dbcfg.Config](i).Config, nil
	})
	do.Provide[*analyticscfg.Config](i, func(i do.Injector) (*analyticscfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Analytics, nil
	})
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Search, nil
	})
	do.Provide[*dataprivacycfg.Config](i, func(i do.Injector) (*dataprivacycfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).DataPrivacy, nil
	})
	do.Provide[*saga.WorkerConfig](i, func(i do.Injector) (*saga.WorkerConfig, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Sagas, nil
	})
	do.Provide[*meteringcfg.Config](i, func(i do.Injector) (*meteringcfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Metering, nil
	})
	do.Provide[*capitalismcfg.Config](i, func(i do.Injector) (*capitalismcfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Capitalism, nil
	})
	do.Provide[*operationscfg.Config](i, func(i do.Injector) (*operationscfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Operations, nil
	})
	do.Provide[*config.ScheduledJobsConfig](i, func(i do.Injector) (*config.ScheduledJobsConfig, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Jobs, nil
	})
	do.Provide[*webhookscfg.Config](i, func(i do.Injector) (*webhookscfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Webhooks, nil
	})
	do.Provide[notificationscfg.Config](i, func(i do.Injector) (notificationscfg.Config, error) {
		return do.MustInvoke[*config.SchedulerConfig](i).PushNotifications, nil
	})
	// The one work queue this process runs. It is provided as the bare *workqueue.Config the
	// platform's constructor takes, because there is exactly one — a second would need a name
	// to tell them apart in the container, which is the point at which this stops being a
	// single unnamed provider.
	do.Provide[*workqueue.Config](i, func(i do.Injector) (*workqueue.Config, error) {
		return &do.MustInvoke[*config.ScheduledJobsConfig](i).MealPlanning.MealPlanTaskNotificationQueue, nil
	})
}
