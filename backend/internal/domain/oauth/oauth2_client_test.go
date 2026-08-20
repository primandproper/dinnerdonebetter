package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2ClientCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &OAuth2ClientCreationRequestInput{
			Name:         t.Name(),
			RedirectURIs: []string{"https://example.com/callback"},
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with no redirect URIs", func(t *testing.T) {
		t.Parallel()

		// Refused at registration rather than at first use. A client with no registered URI
		// can authenticate at the token endpoint and still never complete an authorization
		// request, because redirect_uri is matched against a list it is not in.
		ctx := t.Context()
		x := &OAuth2ClientCreationRequestInput{
			Name: t.Name(),
		}

		assert.Error(t, x.ValidateWithContext(ctx))
	})

	T.Run("with a redirect URI that is not a URI", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &OAuth2ClientCreationRequestInput{
			Name:         t.Name(),
			RedirectURIs: []string{"not a uri"},
		}

		assert.Error(t, x.ValidateWithContext(ctx))
	})

	T.Run("with no name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &OAuth2ClientCreationRequestInput{
			RedirectURIs: []string{"https://example.com/callback"},
		}

		assert.Error(t, x.ValidateWithContext(ctx))
	})
}
