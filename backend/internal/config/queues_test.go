package config

import (
	"testing"

	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"

	"github.com/stretchr/testify/assert"
)

func TestQueueSettings_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := msgconfig.QueuesConfig{}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
