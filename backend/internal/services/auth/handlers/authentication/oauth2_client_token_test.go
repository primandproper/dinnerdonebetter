package authentication

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/stretchr/testify/assert"
)

func TestTokenImpl_New(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		result := token.New()

		assert.NotNil(t, result)
		_, ok := result.(*tokenImpl)
		assert.True(t, ok)
	})
}

func TestTokenImpl_ClientID(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		clientID := "test-client-id"

		token.SetClientID(clientID)
		result := token.GetClientID()

		assert.Equal(t, clientID, result)
	})
}

func TestTokenImpl_UserID(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		userID := "test-user-id"

		token.SetUserID(userID)
		result := token.GetUserID()

		assert.Equal(t, userID, result)
	})
}

func TestTokenImpl_RedirectURI(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		redirectURI := "https://example.com/callback"

		token.SetRedirectURI(redirectURI)
		result := token.GetRedirectURI()

		assert.Equal(t, redirectURI, result)
	})
}

func TestTokenImpl_Scope(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}

		// SetScope should be a no-op
		token.SetScope("test-scope")
		result := token.GetScope()

		assert.Equal(t, "n/a", result)
	})
}

func TestTokenImpl_Code(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		code := "test-code"

		token.SetCode(code)
		result := token.GetCode()

		assert.Equal(t, code, result)
	})
}

func TestTokenImpl_CodeCreateAt(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		createAt := time.Now()

		token.SetCodeCreateAt(createAt)
		result := token.GetCodeCreateAt()

		assert.Equal(t, createAt, result)
	})
}

func TestTokenImpl_CodeExpiresIn(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		expiresIn := time.Hour

		token.SetCodeExpiresIn(expiresIn)
		result := token.GetCodeExpiresIn()

		assert.Equal(t, expiresIn, result)
	})
}

func TestTokenImpl_CodeChallenge(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		challenge := "test-challenge"

		token.SetCodeChallenge(challenge)
		result := token.GetCodeChallenge()

		assert.Equal(t, challenge, result)
	})
}

func TestTokenImpl_CodeChallengeMethod(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		method := oauth2.CodeChallengePlain

		token.SetCodeChallengeMethod(method)
		result := token.GetCodeChallengeMethod()

		assert.Equal(t, method, result)
	})
}

func TestTokenImpl_Access(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		access := "test-access-token"

		token.SetAccess(access)
		result := token.GetAccess()

		assert.Equal(t, access, result)
	})
}

func TestTokenImpl_AccessCreateAt(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		createAt := time.Now()

		token.SetAccessCreateAt(createAt)
		result := token.GetAccessCreateAt()

		assert.Equal(t, createAt, result)
	})
}

func TestTokenImpl_AccessExpiresIn(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		expiresIn := time.Hour

		token.SetAccessExpiresIn(expiresIn)
		result := token.GetAccessExpiresIn()

		assert.Equal(t, expiresIn, result)
	})
}

func TestTokenImpl_Refresh(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		refresh := "test-refresh-token"

		token.SetRefresh(refresh)
		result := token.GetRefresh()

		assert.Equal(t, refresh, result)
	})
}

func TestTokenImpl_RefreshCreateAt(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		createAt := time.Now()

		token.SetRefreshCreateAt(createAt)
		result := token.GetRefreshCreateAt()

		assert.Equal(t, createAt, result)
	})
}

func TestTokenImpl_RefreshExpiresIn(T *testing.T) {
	T.Parallel()

	T.Run("get and set", func(t *testing.T) {
		t.Parallel()

		token := &tokenImpl{}
		expiresIn := 24 * time.Hour

		token.SetRefreshExpiresIn(expiresIn)
		result := token.GetRefreshExpiresIn()

		assert.Equal(t, expiresIn, result)
	})
}

func TestConvertTokenToImpl(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		originalToken := fakes.BuildFakeOAuth2ClientToken()
		result := convertTokenToImpl(originalToken)

		assert.NotNil(t, result)
		impl, ok := result.(*tokenImpl)
		assert.True(t, ok)
		assert.Equal(t, originalToken.ClientID, impl.Token.ClientID)
		assert.Equal(t, originalToken.BelongsToUser, impl.Token.BelongsToUser)
		assert.Equal(t, originalToken.RedirectURI, impl.Token.RedirectURI)
		assert.Equal(t, originalToken.Code, impl.Token.Code)
		assert.Equal(t, originalToken.Access, impl.Token.Access)
		assert.Equal(t, originalToken.Refresh, impl.Token.Refresh)
	})
}
