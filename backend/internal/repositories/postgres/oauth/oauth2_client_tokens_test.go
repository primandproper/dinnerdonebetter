package oauth

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createOAuth2ClientTokenForTest(t *testing.T, ctx context.Context, exampleOAuth2ClientToken *types.OAuth2ClientToken, dbc *repository) *types.OAuth2ClientToken {
	t.Helper()

	// create
	if exampleOAuth2ClientToken == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		oauth2Client := createOAuth2ClientForTest(t, ctx, nil, dbc)
		exampleOAuth2ClientToken = fakes.BuildFakeOAuth2ClientToken()
		exampleOAuth2ClientToken.BelongsToUser = user.ID
		exampleOAuth2ClientToken.ClientID = oauth2Client.ID
	}
	dbInput := converters.ConvertOAuth2ClientTokenToOAuth2ClientTokenDatabaseCreationInput(exampleOAuth2ClientToken)

	created, err := dbc.CreateOAuth2ClientToken(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleOAuth2ClientToken.CodeCreatedAt = created.CodeCreatedAt
	exampleOAuth2ClientToken.AccessCreatedAt = created.AccessCreatedAt
	exampleOAuth2ClientToken.RefreshCreatedAt = created.RefreshCreatedAt
	assert.Equal(t, exampleOAuth2ClientToken, created)

	oauth2ClientToken, err := dbc.GetOAuth2ClientTokenByAccess(ctx, created.Access)
	require.NoError(t, err)
	exampleOAuth2ClientToken.CodeCreatedAt = oauth2ClientToken.CodeCreatedAt
	exampleOAuth2ClientToken.AccessCreatedAt = oauth2ClientToken.AccessCreatedAt
	exampleOAuth2ClientToken.RefreshCreatedAt = oauth2ClientToken.RefreshCreatedAt
	assert.Equal(t, oauth2ClientToken, exampleOAuth2ClientToken)

	return created
}

func TestQuerier_Integration_OAuth2ClientTokens(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	oauth2Client := createOAuth2ClientForTest(t, ctx, nil, dbc)
	exampleOAuth2ClientToken := fakes.BuildFakeOAuth2ClientToken()
	exampleOAuth2ClientToken.BelongsToUser = user.ID
	exampleOAuth2ClientToken.ClientID = oauth2Client.ClientID

	// create
	createdOAuth2ClientToken := createOAuth2ClientTokenForTest(t, ctx, exampleOAuth2ClientToken, dbc)

	// get
	byCode, err := dbc.GetOAuth2ClientTokenByCode(ctx, exampleOAuth2ClientToken.Code)
	require.NoError(t, err)
	assert.NotNil(t, byCode)

	// delete
	require.NoError(t, dbc.DeleteOAuth2ClientTokenByCode(ctx, createdOAuth2ClientToken.Code))

	// create
	createdOAuth2ClientToken = createOAuth2ClientTokenForTest(t, ctx, exampleOAuth2ClientToken, dbc)

	// get
	byAccess, err := dbc.GetOAuth2ClientTokenByAccess(ctx, exampleOAuth2ClientToken.Access)
	require.NoError(t, err)
	assert.NotNil(t, byAccess)

	// delete
	require.NoError(t, dbc.DeleteOAuth2ClientTokenByAccess(ctx, createdOAuth2ClientToken.Access))

	// create
	createdOAuth2ClientToken = createOAuth2ClientTokenForTest(t, ctx, exampleOAuth2ClientToken, dbc)

	// get
	byRefresh, err := dbc.GetOAuth2ClientTokenByRefresh(ctx, exampleOAuth2ClientToken.Refresh)
	require.NoError(t, err)
	assert.NotNil(t, byRefresh)

	// delete
	assert.NoError(t, dbc.DeleteOAuth2ClientTokenByRefresh(ctx, createdOAuth2ClientToken.Refresh))
}

func TestQuerier_ArchiveOAuth2ClientTokenByAccess(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.DeleteOAuth2ClientTokenByAccess(ctx, ""))
	})
}

func TestQuerier_ArchiveOAuth2ClientTokenByCode(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.DeleteOAuth2ClientTokenByCode(ctx, ""))
	})
}

func TestQuerier_ArchiveOAuth2ClientTokenByRefresh(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.DeleteOAuth2ClientTokenByRefresh(ctx, ""))
	})
}
