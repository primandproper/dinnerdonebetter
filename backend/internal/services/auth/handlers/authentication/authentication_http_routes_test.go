package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationService_AuthorizeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		req := helper.req
		res := helper.res

		// The AuthorizeHandler delegates to the oauth2Server, so we just test that it calls it
		// Since we can't easily mock the oauth2Server.HandleAuthorizeRequest, we'll test error handling
		helper.service.AuthorizeHandler(res, req)

		// The response code will depend on the oauth2 library's behavior
		// We're mainly testing that the method doesn't panic
		assert.GreaterOrEqual(t, res.Code, 400) // Expect some error since we don't have a real OAuth2 setup
	})
}

func TestAuthenticationService_TokenHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		req := helper.req
		res := helper.res

		// The TokenHandler delegates to the oauth2Server, so we just test that it calls it
		// Since we can't easily mock the oauth2Server.HandleTokenRequest, we'll test error handling
		helper.service.TokenHandler(res, req)

		// The response code will depend on the oauth2 library's behavior
		// We're mainly testing that the method doesn't panic
		assert.GreaterOrEqual(t, res.Code, 400) // Expect some error since we don't have a real OAuth2 setup
	})
}
