package main

import (
	"testing"

	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInjector(T *testing.T) {
	T.Parallel()

	// Registration is where a duplicate declaration panics — a component registered by this
	// tool and again by a domain's registration — so building the container with nothing
	// resolved is the test that catches it, and it needs neither a database nor a search
	// backend: no provider runs until something is invoked.
	T.Run("registers every provider once", func(t *testing.T) {
		t.Parallel()

		var i *do.RootScope

		require.NotPanics(t, func() {
			i = buildInjector(t.Context(), &databasecfg.Config{Provider: databasecfg.ProviderPostgres}, &textsearchcfg.Config{})
		})

		services := i.ListProvidedServices()
		assert.NotEmpty(t, services, "expected providers to be registered")
		assert.Greater(t, len(services), 10, "expected many providers to be registered")
	})
}
