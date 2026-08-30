package authentication

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/mock"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRegistryStore_GetClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := oauthfakes.BuildFakeOAuth2Client()

		repo := &oauthmock.RepositoryMock{
			GetOAuth2ClientByClientIDFunc: func(_ context.Context, clientID string) (*oauth.OAuth2Client, error) {
				assert.Equal(t, expected.ClientID, clientID)
				return expected, nil
			},
		}

		actual, err := (&clientRegistryStore{clients: repo}).GetClient(ctx, expected.ClientID)
		require.NoError(t, err)

		assert.Equal(t, expected.ClientID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		// The stored column is already the digest form the authorization server compares
		// against, which is what lets a client registered before this change keep working.
		assert.Equal(t, expected.ClientSecret, actual.SecretHash)
		assert.Equal(t, expected.RedirectURIs, actual.RedirectURIs)
		assert.Equal(t, []string{oauth2server.GrantTypeAuthorizationCode, oauth2server.GrantTypeRefreshToken}, actual.GrantTypes)
		assert.False(t, actual.Public())
		// Never expires: an administered registration is archived, not aged out.
		assert.True(t, actual.ExpiresAt.IsZero())
	})

	T.Run("with a lookup miss", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		// The repository reports "no such client" as a wrapped sql.ErrNoRows. It has to arrive
		// at the authorization server as ErrNotFound, or the endpoint has no protocol error to
		// map, answers 500, and echoes the repository's own error text to an unauthenticated
		// caller.
		repo := &oauthmock.RepositoryMock{
			GetOAuth2ClientByClientIDFunc: func(context.Context, string) (*oauth.OAuth2Client, error) {
				return nil, errors.Join(errors.New("fetching oauth2 client"), sql.ErrNoRows)
			},
		}

		actual, err := (&clientRegistryStore{clients: repo}).GetClient(ctx, "nonexistent")

		assert.Nil(t, actual)
		assert.ErrorIs(t, err, oauth2server.ErrNotFound)
	})

	T.Run("with an empty client identifier", func(t *testing.T) {
		t.Parallel()

		actual, err := (&clientRegistryStore{clients: &oauthmock.RepositoryMock{}}).GetClient(t.Context(), "")

		assert.Nil(t, actual)
		assert.ErrorIs(t, err, oauth2server.ErrEmptyIdentifier)
	})

	T.Run("with a broken repository", func(t *testing.T) {
		t.Parallel()

		// Not a miss. A database that is down must not read as "no such client": that answer
		// tells the caller the registration does not exist, which is a different and wrong
		// thing to say.
		expected := errors.New("connection refused")
		repo := &oauthmock.RepositoryMock{
			GetOAuth2ClientByClientIDFunc: func(context.Context, string) (*oauth.OAuth2Client, error) {
				return nil, expected
			},
		}

		actual, err := (&clientRegistryStore{clients: repo}).GetClient(t.Context(), "whatever")

		assert.Nil(t, actual)
		require.ErrorIs(t, err, expected)
		assert.NotErrorIs(t, err, oauth2server.ErrNotFound)
	})
}

func TestClientRegistryStore_RegistrationIsNotServed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		store := &clientRegistryStore{clients: &oauthmock.RepositoryMock{}}

		require.ErrorIs(t, store.CreateClient(t.Context(), &oauth2server.Client{}), errRegistrationEndpointNotServed)
		require.ErrorIs(t, store.DeleteClient(t.Context(), "whatever"), errRegistrationEndpointNotServed)
	})
}
