package datachangemessagehandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncDataChangeMessageHandler_SearchIndexRequestsEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("with invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		rawMsg := []byte("invalid json")

		err := handler.SearchIndexRequestsEventHandler("search_index_requests")(ctx, rawMsg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding JSON body")
	})
}
