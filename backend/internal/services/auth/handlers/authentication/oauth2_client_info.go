package authentication

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/go-oauth2/oauth2/v4"
)

var (
	_ oauth2.ClientInfo             = (*oauth2ClientInfoImpl)(nil)
	_ oauth2.ClientPasswordVerifier = (*oauth2ClientInfoImpl)(nil)
)

type oauth2ClientInfoImpl struct {
	client *types.OAuth2Client
	domain string
}

func (i *oauth2ClientInfoImpl) GetID() string {
	return i.client.ID
}

func (i *oauth2ClientInfoImpl) GetSecret() string {
	return i.client.ClientSecret
}

// VerifyPassword implements oauth2.ClientPasswordVerifier so the oauth2 library
// compares presented secrets against the stored SHA-256 digest instead of doing
// a plaintext equality check on GetSecret.
func (i *oauth2ClientInfoImpl) VerifyPassword(secret string) bool {
	return types.ClientSecretMatches(i.client.ClientSecret, secret)
}

func (i *oauth2ClientInfoImpl) GetDomain() string {
	return i.domain
}

func (i *oauth2ClientInfoImpl) IsPublic() bool {
	return false
}

func (i *oauth2ClientInfoImpl) GetUserID() string {
	// AFAICT this isn't used anywhere
	return ""
}
