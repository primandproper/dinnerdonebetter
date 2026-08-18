package fakes

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeOAuth2Client builds a faked OAuth2Client.
func BuildFakeOAuth2Client() *types.OAuth2Client {
	client := fake.BuildFakeRecord[types.OAuth2Client]()

	// A secret that looks like one: it is compared, hashed and stored by code that has
	// opinions about its alphabet and length.
	client.ClientSecret = fake.BuildFakePassword()
	client.Name = gofakeit.Password(true, true, true, false, false, 32)

	return client
}

// BuildFakeOAuth2ClientToken builds a faked OAuth2ClientToken.
func BuildFakeOAuth2ClientToken() *types.OAuth2ClientToken {
	token := fake.BuildFakeRecord[types.OAuth2ClientToken]()

	// The three durations are lifetimes rather than arbitrary quantities, and a token
	// whose lifetime faker picked is one that may already have expired.
	token.CodeExpiresAt = time.Hour
	token.AccessExpiresAt = time.Hour
	token.RefreshExpiresAt = time.Hour

	// The two values the OAuth2 spec constrains: a redirect the authorization path
	// parses as a URL, and the one PKCE challenge method this server supports.
	token.RedirectURI = gofakeit.URL()
	token.CodeChallengeMethod = "S256"

	return token
}

// BuildFakeOAuth2ClientCreationResponse builds a faked OAuth2ClientCreationResponse.
func BuildFakeOAuth2ClientCreationResponse() *types.OAuth2ClientCreationResponse {
	client := BuildFakeOAuth2Client()

	return &types.OAuth2ClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
	}
}

// BuildFakeOAuth2ClientsList builds a faked OAuth2ClientList.
func BuildFakeOAuth2ClientsList() *filtering.QueryFilteredResult[types.OAuth2Client] {
	return fake.BuildFakePage(BuildFakeOAuth2Client)
}

// BuildFakeOAuth2ClientCreationRequestInput builds a faked OAuth2ClientCreationRequestInput.
func BuildFakeOAuth2ClientCreationRequestInput() *types.OAuth2ClientCreationRequestInput {
	client := BuildFakeOAuth2Client()

	return &types.OAuth2ClientCreationRequestInput{
		Name:        client.Name,
		Description: client.Description,
	}
}
