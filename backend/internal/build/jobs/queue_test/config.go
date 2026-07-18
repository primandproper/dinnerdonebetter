package queuetest

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	queuetest "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"

	databasecfg "github.com/primandproper/platform-go/v5/database/config"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	"github.com/primandproper/platform-go/v5/observability"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*msgconfig.QueuesConfig](i, func(i do.Injector) (*msgconfig.QueuesConfig, error) {
		cfg := do.MustInvoke[*config.QueueTestJobConfig](i)
		return &cfg.Queues, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.QueueTestJobConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*config.QueueTestJobConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*queuetest.JobParams](i, func(i do.Injector) (*queuetest.JobParams, error) {
		cfg := do.MustInvoke[*config.QueueTestJobConfig](i)
		return ProvideJobParams(cfg), nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		cfg := do.MustInvoke[*config.QueueTestJobConfig](i)
		return ProvideEventsConfig(cfg), nil
	})
}

// ProvideJobParams builds JobParams from the config.
func ProvideJobParams(cfg *config.QueueTestJobConfig) *queuetest.JobParams {
	return &queuetest.JobParams{
		Queues: cfg.Queues,
	}
}

// ProvideEventsConfig returns the message queue config for the publisher.
func ProvideEventsConfig(cfg *config.QueueTestJobConfig) *msgconfig.Config {
	return &cfg.Events
}
