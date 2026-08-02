package config

import (
	"testing"

	queuescfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/queues/config"

	"github.com/stretchr/testify/assert"
)

func TestQueueSettings_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := queuescfg.Config{}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
