package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backendRoot is where this tool runs from in anger — four directories up from
// cmd/tools/codegen/webhook_catalog.
const backendRoot = "../../../.."

func TestCatalogIsUpToDate(T *testing.T) {
	T.Parallel()

	// The catalog gates both subscription and dispatch, so a stale one is not a cosmetic
	// problem: an event whose constant exists but whose catalog entry does not cannot be
	// subscribed to and cannot be dispatched, and the only symptom is an absence. This is the
	// check that turns forgetting `make webhook_catalog` into a failing build.
	T.Run("generated file matches the domain constants", func(t *testing.T) {
		t.Parallel()

		events, err := collect(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)
		require.NotEmpty(t, events)

		rendered, err := render(events)
		require.NoError(t, err)

		onDisk, err := os.ReadFile(filepath.Join(backendRoot, outputPath))
		require.NoError(t, err)

		assert.Equal(t, string(rendered), string(onDisk),
			"webhook catalog is stale; run `make webhook_catalog`")
	})
}

func TestCollect(T *testing.T) {
	T.Parallel()

	T.Run("descriptions carry no linter directives", func(t *testing.T) {
		t.Parallel()

		events, err := collect(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)

		for _, e := range events {
			// A description is rendered to whoever is choosing events in a subscription
			// UI. "#nosec G101" reaching them means the doc comment was shipped raw.
			assert.NotContains(t, e.description, "#nosec", "event type %q", e.eventType)
			assert.NotEmpty(t, e.eventType)
		}
	})

	T.Run("event types are unique", func(t *testing.T) {
		t.Parallel()

		events, err := collect(filepath.Join(backendRoot, domainDir))
		require.NoError(t, err)

		seen := map[string]string{}
		for _, e := range events {
			// collect already refuses a collision between two differently named
			// constants; this guards the sort/dedupe rather than the source.
			previous, ok := seen[e.eventType]
			assert.False(t, ok, "event type %q declared by both %s and %s", e.eventType, previous, e.constName)
			seen[e.eventType] = e.constName
		}
	})
}
