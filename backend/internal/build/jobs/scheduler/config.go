package scheduler

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"

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
	do.Provide[*config.ScheduledJobsConfig](i, func(i do.Injector) (*config.ScheduledJobsConfig, error) {
		return &do.MustInvoke[*config.SchedulerConfig](i).Jobs, nil
	})
}
