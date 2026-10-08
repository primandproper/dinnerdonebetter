package indexing

import (
	"context"

	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterSearchers registers the text index client behind the users index.
//
// It is what RegisterIndexes resolves, so a process that registers the index registers this
// beside it. The client is built from the one *textsearchcfg.Config the process provides.
func RegisterSearchers(i do.Injector) {
	do.Provide(i, func(i do.Injector) (UserTextSearcher, error) {
		index, err := textsearchcfg.NewIndex[UserSearchSubset](
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*textsearchcfg.Config](i),
			IndexTypeUsers,
			textsearchcfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			textsearchcfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			textsearchcfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)

		return UserTextSearcher(index), err
	})
}
