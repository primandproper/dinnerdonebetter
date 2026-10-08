package mcptools

import (
	"reflect"
	"strings"
	"testing"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestObjectType(T *testing.T) {
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
		actual := ObjectType(map[string]any{"things": "stuff"}, required...)

		assert.Equal(t, expected, actual)
	})

	T.Run("without required fields", func(t *testing.T) {
		t.Parallel()

		actual := ObjectType(map[string]any{"things": "stuff"})

		assert.NotContains(t, actual, "required")
	})
}

func TestAccountFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("reads the account off the token", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()
		req := &mcp.CallToolRequest{Extra: &mcp.RequestExtra{TokenInfo: &auth.TokenInfo{
			Extra: map[string]any{ClaimAccountID: accountID},
		}}}

		actual, err := AccountFromRequest(req)
		require.NoError(t, err)
		assert.Equal(t, accountID, actual)
	})

	T.Run("refuses a call with no token", func(t *testing.T) {
		t.Parallel()

		_, err := AccountFromRequest(&mcp.CallToolRequest{})
		require.ErrorIs(t, err, errNotAuthenticated)

		_, err = AccountFromRequest(nil)
		require.ErrorIs(t, err, errNotAuthenticated)
	})

	T.Run("refuses a token naming no account", func(t *testing.T) {
		t.Parallel()

		req := &mcp.CallToolRequest{Extra: &mcp.RequestExtra{TokenInfo: &auth.TokenInfo{UserID: fake.BuildFakeID()}}}

		_, err := AccountFromRequest(req)
		require.ErrorIs(t, err, errNoAccountOnToken)
	})
}
