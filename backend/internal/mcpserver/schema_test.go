package mcpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_objectType(T *testing.T) {
	T.Parallel()

	T.Run("with required fields", func(t *testing.T) {
		t.Parallel()

		required := []string{"one", "two", "three"}
		expected := map[string]any{
			"type": objType,
			"properties": map[string]any{
				"things": "stuff",
			},
			"required": required,
		}
		actual := objectType(map[string]any{"things": "stuff"}, required...)

		assert.Equal(t, expected, actual)
	})
}
