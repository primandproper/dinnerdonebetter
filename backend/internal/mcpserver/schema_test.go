package mcpserver

import (
	"reflect"
	"strings"
	"testing"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilterSchemaMatchesQueryFilter asserts that what a tool advertises for its Filter argument
// is named the way the decoder reads it.
//
// A tool's invocation struct holds a *filtering.QueryFilter and the SDK unmarshals into it, so a
// property the schema names something encoding/json does not recognize is not a wrong hint — it
// is a filter that is discarded in full, answered with an unfiltered page that looks like a
// filtered one. The hand-written schema this replaced named all eight of them that way.
func TestFilterSchemaMatchesQueryFilter(t *testing.T) {
	t.Parallel()

	schema := filtering.QueryFilterSchema()

	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema has no properties object")

	filterType := reflect.TypeFor[filtering.QueryFilter]()

	expected := make([]string, 0, filterType.NumField())
	for field := range filterType.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}

		expected = append(expected, name)
	}

	actual := make([]string, 0, len(properties))
	for name := range properties {
		actual = append(actual, name)
	}

	assert.ElementsMatch(t, expected, actual)
}

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
