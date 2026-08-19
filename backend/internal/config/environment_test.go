package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringOrDefault(T *testing.T) {
	T.Parallel()

	T.Run("with empty string", func(t *testing.T) {
		t.Parallel()

		result := stringOrDefault("", "default")
		assert.Equal(t, "default", result)
	})

	T.Run("with non-empty string", func(t *testing.T) {
		t.Parallel()

		result := stringOrDefault("value", "default")
		assert.Equal(t, "value", result)
	})
}
