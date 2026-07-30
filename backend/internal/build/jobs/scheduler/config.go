package scheduler

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	analyticscfg "github.com/primandproper/platform-go/v8/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability"
	textsearchcfg "github.com/primandproper/platform-go/v8/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*msgconfig.QueuesConfig](i, func(i do.Injector) (*msgconfig.QueuesConfig, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Queues, nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Events, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Observability, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Database, nil
	})
	do.Provide[*analyticscfg.Config](i, func(i do.Injector) (*analyticscfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Analytics, nil
	})
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Search, nil
	})
	do.Provide[*config.ScheduledJobsConfig](i, func(i do.Injector) (*config.ScheduledJobsConfig, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Jobs, nil
	})
}
