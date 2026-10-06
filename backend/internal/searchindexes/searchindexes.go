/*
Package searchindexes registers the one searchsync.Registry every search index here is built
through.

An index is three things that must be built together — the Syncer the consumer applies events
with, the Reindexer the scheduler rebuilds with, and the buffer that stamps last_indexed_at behind
the Syncer — and the Registry is platform's list of them. The async message handler reads its
pool specs off it, the scheduler rebuilds every index through ReindexAll, and adding an index
touches neither: each domain registers its own indexes, and this package only holds the list and
retires it.

Which writes feed which index is not decided here. That is internal/indexevents, the rule table
registered on the outbox writer.
*/
package searchindexes

import (
	"context"

	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	syncsource "github.com/primandproper/platform-go/v15/searchsync/source"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	textsearch "github.com/primandproper/primitives-go/v2/search/text"

	"github.com/samber/do/v2"
)

// o11yName names the loggers, spans and metrics of every index built here. It keeps the name the
// deleted internal/search/syncsource used, so nothing downstream of a log query has to change.
const o11yName = "search_sync_source"

// Registry is platform's searchsync.Registry with the container's shutdown method on it.
//
// The Registry owns one goroutine per index whose source table carries last_indexed_at — the
// stamp buffer behind its Syncer — and closes them in Close. do calls Shutdown, so this is what
// makes retiring them the container's job, which is the one thing in a process that already
// knows how to discharge it. Whoever shuts the container down must do it after the pools that
// feed the buffers have stopped, or the last flush races the work that fills it.
type Registry struct {
	_ struct{} `json:"-"`

	*searchsync.Registry
}

// Shutdown flushes and closes every stamp buffer the Registry built.
func (r *Registry) Shutdown(ctx context.Context) error {
	return r.Close(ctx)
}

// Registrar adds one domain's indexes to a Registry, resolving what they read from and write to
// out of the injector.
type Registrar func(i do.Injector, registry *searchsync.Registry) error

// Register registers the Registry, built from every registrar handed in. ctx is the process's
// own, and bounds the closing of whatever a failed build had already started.
//
// The registrars are named by the process wiring itself rather than discovered, because they are
// the definition of which indexes it runs: an index whose registrar is missing has a topic nobody
// reads and a rebuild nobody runs, and the only symptom is search results that quietly stop
// moving.
func Register(ctx context.Context, i do.Injector, registrars ...Registrar) {
	do.Provide[*Registry](i, func(i do.Injector) (*Registry, error) {
		registry := searchsync.NewRegistry(
			searchsync.WithRegistryLogger(logging.NewNamedLogger(do.MustInvoke[logging.Logger](i), o11yName)),
			searchsync.WithRegistryTracerProvider(do.MustInvoke[tracing.Provider](i)),
			searchsync.WithRegistryMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)

		for _, register := range registrars {
			if err := register(i, registry); err != nil {
				// Whatever registered before the failure has a running buffer, and this
				// Registry is about to be dropped rather than handed to the container.
				//nolint:errcheck // the registration error is the one worth reporting.
				_ = registry.Close(ctx)

				return nil, err
			}
		}

		return &Registry{Registry: registry}, nil
	})
}

// RegisterTextIndex registers one text index fed from source, on the topic its events arrive on,
// which is the index's own name.
//
// stamp is the bulk write that stamps last_indexed_at for the documents the index accepted —
// conventionally a repository's MarkXAsIndexed. The Registry builds the buffer it runs behind and
// hands it to the Syncer alone: the Reindexer is given none on purpose, because it writes every
// document there is, and stamping it would make the column read as when the last rebuild ran
// rather than how current each document is — which is the question the column exists to answer.
func RegisterTextIndex[E, T any](
	registry *searchsync.Registry,
	source *syncsource.Source[E, T],
	index textsearch.IndexManager,
	stamp func(ctx context.Context, ids []string) error,
) error {
	if source == nil {
		return searchsync.ErrNilSource
	}

	target, err := searchsync.TextTarget[T](index)
	if err != nil {
		return platformerrors.Wrapf(err, "building %s search target", source.Name())
	}

	_, err = searchsync.RegisterIndex(registry, searchsync.IndexSpec[T]{
		Name:   source.Name(),
		Topic:  source.Name(),
		Source: source,
		Target: target,
		Stamp:  stamp,
	})

	return err
}
