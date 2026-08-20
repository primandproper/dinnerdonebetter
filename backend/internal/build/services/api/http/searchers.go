package api

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	textsearchcfg "github.com/primandproper/platform-go/v12/search/text/config"

	"github.com/samber/do/v2"
)

// ProvideTextSearchConfig provides a pointer to the text search config for dependency injection.
func ProvideTextSearchConfig(cfg *config.APIServiceConfig) *textsearchcfg.Config {
	return &cfg.TextSearch
}

// ProvideUserTextSearcher provides a user text searcher for the identity manager.
func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.NewIndex[identityindexing.UserSearchSubset](
		ctx,
		cfg,
		identityindexing.IndexTypeUsers,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

// RegisterSearchers registers text search providers with the injector.
func RegisterSearchers(i do.Injector) {
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		return ProvideTextSearchConfig(do.MustInvoke[*config.APIServiceConfig](i)), nil
	})

	do.Provide[identityindexing.UserTextSearcher](i, func(i do.Injector) (identityindexing.UserTextSearcher, error) {
		return ProvideUserTextSearcher(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*textsearchcfg.Config](i),
		)
	})
}
