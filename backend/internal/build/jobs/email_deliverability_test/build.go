package emaildeliverabilitytest

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	emaildeliverabilitytest "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/email/workers/email_deliverability_test"

	emailcfg "github.com/primandproper/platform-go/v6/email/config"
	"github.com/primandproper/platform-go/v6/httpclient"
	"github.com/primandproper/platform-go/v6/observability"
	loggingcfg "github.com/primandproper/platform-go/v6/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v6/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v6/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.EmailDeliverabilityTestConfig,
) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	RegisterConfigs(i)

	observability.RegisterO11yConfigs(i)
	tracingcfg.RegisterTracerProvider(i)
	loggingcfg.RegisterLogger(i)
	metricscfg.RegisterMetricsProvider(i)
	httpclient.RegisterHTTPClient(i)
	emailcfg.RegisterEmailer(i)
	emaildeliverabilitytest.RegisterEmailDeliverabilityTest(i)

	return i
}

// Build builds the email deliverability test job.
func Build(
	ctx context.Context,
	cfg *config.EmailDeliverabilityTestConfig,
) (*emaildeliverabilitytest.Job, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*emaildeliverabilitytest.Job](i), nil
}
