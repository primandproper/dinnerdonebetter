package emaildeliverabilitytest

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	emaildeliverabilitytest "github.com/primandproper/dinnerdonebetter/backend/internal/services/email/workers/email_deliverability_test"

	emailcfg "github.com/primandproper/platform-go/v10/email/config"
	httpclientcfg "github.com/primandproper/platform-go/v10/httpclient"
	"github.com/primandproper/platform-go/v10/observability"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.EmailDeliverabilityTestConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[*emailcfg.Config](i, func(i do.Injector) (*emailcfg.Config, error) {
		cfg := do.MustInvoke[*config.EmailDeliverabilityTestConfig](i)
		return &cfg.Email, nil
	})
	do.Provide[*httpclientcfg.Config](i, func(i do.Injector) (*httpclientcfg.Config, error) {
		cfg := do.MustInvoke[*config.EmailDeliverabilityTestConfig](i)
		return cfg.HTTPClient, nil
	})
	do.Provide[*emaildeliverabilitytest.JobParams](i, func(i do.Injector) (*emaildeliverabilitytest.JobParams, error) {
		cfg := do.MustInvoke[*config.EmailDeliverabilityTestConfig](i)
		return ProvideJobParams(cfg), nil
	})
}

// ProvideJobParams builds JobParams from the config.
func ProvideJobParams(cfg *config.EmailDeliverabilityTestConfig) *emaildeliverabilitytest.JobParams {
	return &emaildeliverabilitytest.JobParams{
		RecipientEmailAddress: cfg.RecipientEmailAddress,
		ServiceEnvironment:    cfg.ServiceEnvironment,
	}
}
