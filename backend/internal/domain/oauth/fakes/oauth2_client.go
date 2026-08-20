package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeOAuth2Client builds a faked OAuth2Client.
func BuildFakeOAuth2Client() *types.OAuth2Client {
	client := fake.BuildFakeRecord[types.OAuth2Client]()

	// A secret that looks like one: it is compared, hashed and stored by code that has
	// opinions about its alphabet and length.
	client.ClientSecret = fake.BuildFakePassword()
	client.Name = gofakeit.Password(true, true, true, false, false, 32)

	// Redirect URIs are matched byte for byte by the authorization server, so a faked
	// client needs values that parse as absolute URLs rather than faker's arbitrary
	// strings — a client whose registered URI is not a URI can never be authorized
	// against, which makes it useless as a fixture for the flow.
	client.RedirectURIs = []string{gofakeit.URL(), gofakeit.URL()}

	return client
}

// BuildFakeOAuth2ClientCreationResponse builds a faked OAuth2ClientCreationResponse.
func BuildFakeOAuth2ClientCreationResponse() *types.OAuth2ClientCreationResponse {
	client := BuildFakeOAuth2Client()

	return &types.OAuth2ClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		RedirectURIs: client.RedirectURIs,
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
		Name:         client.Name,
		Description:  client.Description,
		RedirectURIs: client.RedirectURIs,
	}
}
