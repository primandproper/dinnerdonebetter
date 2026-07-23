package authentication

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/fakes"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2ClientInfoImpl_GetID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.GetID()
		assert.Equal(t, client.ID, result)
	})
}

func TestOAuth2ClientInfoImpl_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.GetSecret()
		assert.Equal(t, client.ClientSecret, result)
	})
}

func TestOAuth2ClientInfoImpl_GetDomain(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		domain := "example.com"
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: domain,
		}

		result := impl.GetDomain()
		assert.Equal(t, domain, result)
	})
}

func TestOAuth2ClientInfoImpl_IsPublic(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.IsPublic()
		assert.False(t, result)
	})
}

func TestOAuth2ClientInfoImpl_GetUserID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.GetUserID()
		assert.Empty(t, result)
	})
}

func TestOAuth2ClientInfoImpl_VerifyPassword(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		plaintext := fakes.BuildFakeID()
		client := fakes.BuildFakeOAuth2Client()
		client.ClientSecret = types.HashClientSecret(plaintext)

		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		assert.True(t, impl.VerifyPassword(plaintext))
	})

	T.Run("with wrong secret", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		client.ClientSecret = types.HashClientSecret(fakes.BuildFakeID())

		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		assert.False(t, impl.VerifyPassword(fakes.BuildFakeID()))
	})
}
