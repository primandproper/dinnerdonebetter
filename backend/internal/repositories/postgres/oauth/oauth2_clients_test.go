package oauth

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createOAuth2ClientForTest(t *testing.T, ctx context.Context, exampleOAuth2Client *types.OAuth2Client, dbc *repository) *types.OAuth2Client {
	t.Helper()

	// create
	if exampleOAuth2Client == nil {
		exampleOAuth2Client = fakes.BuildFakeOAuth2Client()
	}
	dbInput := converters.ConvertOAuth2ClientToOAuth2ClientDatabaseCreationInput(exampleOAuth2Client)

	created, err := dbc.CreateOAuth2Client(ctx, dbInput)
	exampleOAuth2Client.CreatedAt = created.CreatedAt
	require.NoError(t, err)
	assert.Equal(t, exampleOAuth2Client, created)

	oauth2Client, err := dbc.GetOAuth2ClientByDatabaseID(ctx, created.ID)
	exampleOAuth2Client.CreatedAt = oauth2Client.CreatedAt

	require.NoError(t, err)
	assert.Equal(t, oauth2Client, exampleOAuth2Client)

	return created
}

func TestQuerier_Integration_OAuth2Clients(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
	createdOAuth2Clients := []*types.OAuth2Client{}

	// create
	createdOAuth2Clients = append(createdOAuth2Clients, createOAuth2ClientForTest(t, ctx, exampleOAuth2Client, dbc))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeOAuth2Client()
		input.Name = fmt.Sprintf("%s %d", exampleOAuth2Client.Name, i)
		createdOAuth2Clients = append(createdOAuth2Clients, createOAuth2ClientForTest(t, ctx, input, dbc))
	}

	// fetch as list
	oauth2Clients, err := dbc.GetOAuth2Clients(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, oauth2Clients.Data)
	assert.Len(t, oauth2Clients.Data, len(createdOAuth2Clients))

	// delete
	for _, oauth2Client := range createdOAuth2Clients {
		require.NoError(t, dbc.ArchiveOAuth2Client(ctx, oauth2Client.ID))

		var y *types.OAuth2Client
		y, err = dbc.GetOAuth2ClientByClientID(ctx, oauth2Client.ClientID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetOAuth2ClientByClientID(T *testing.T) {
	T.Parallel()

	T.Run("with empty client ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetOAuth2ClientByClientID(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetOAuth2ClientByDatabaseID(T *testing.T) {
	T.Parallel()

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetOAuth2ClientByDatabaseID(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateOAuth2Client(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("with invalid client ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveOAuth2Client(ctx, ""))
	})
}
