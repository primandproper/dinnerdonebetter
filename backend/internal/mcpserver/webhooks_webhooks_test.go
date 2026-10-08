package mcpserver

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPToolManager_GetWebhookEventTypes(T *testing.T) {
	T.Parallel()

	T.Run("lists every subscribable event type, with its description", func(t *testing.T) {
		t.Parallel()

		_, actual, err := (&mcpToolManager{}).GetWebhookEventTypes()(t.Context(), nil, &GetWebhookEventTypesInvocation{})
		require.NoError(t, err)
		require.NotNil(t, actual)

		known := catalog.Catalog()

		listed := make([]platformwebhooks.EventType, 0, len(actual.Results))
		for _, definition := range actual.Results {
			listed = append(listed, definition.EventType)
			assert.Equal(t, known[definition.EventType].Description, definition.Description, "event type %q", definition.EventType)
		}

		assert.Equal(t, known.SubscribableEventTypes(), listed)
	})

	T.Run("leaves the internal event types out", func(t *testing.T) {
		t.Parallel()

		// An internal event is in the catalog so that it can be published, and offering it here
		// would be offering a subscription Subscribe refuses.
		known := catalog.Catalog()

		var internal []platformwebhooks.EventType
		for eventType, definition := range known {
			if definition.Internal {
				internal = append(internal, eventType)
			}
		}
		require.NotEmpty(t, internal, "the catalog marks nothing internal, so this asserts nothing")

		_, actual, err := (&mcpToolManager{}).GetWebhookEventTypes()(t.Context(), nil, &GetWebhookEventTypesInvocation{})
		require.NoError(t, err)

		for _, definition := range actual.Results {
			assert.NotContains(t, internal, definition.EventType)
		}
	})
}
