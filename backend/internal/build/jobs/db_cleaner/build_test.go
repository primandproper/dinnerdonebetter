package dbcleaner

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestBuildInjector_RegistersAllProviders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &config.DBCleanerConfig{}

	i := BuildInjector(ctx, cfg)

	services := i.ListProvidedServices()
	assert.NotEmpty(t, services, "expected providers to be registered")
	assert.Greater(t, len(services), 5, "expected many providers to be registered")
}
