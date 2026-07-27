package internalops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuerier_Integration_DeleteExpiredOAuth2ClientTokens(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	// fetch as list
	count, err := dbc.DeleteExpiredOAuth2ClientTokens(ctx)
	assert.Zero(t, count)
	assert.NoError(t, err)
}
