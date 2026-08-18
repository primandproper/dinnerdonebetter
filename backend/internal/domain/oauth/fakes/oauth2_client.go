package fakes

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v11/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeOAuth2Client builds a faked OAuth2Client.
func BuildFakeOAuth2Client() *types.OAuth2Client {
	return &types.OAuth2Client{
		ID:           BuildFakeID(),
		Name:         fake.Password(true, true, true, false, false, 32),
		ClientID:     BuildFakeID(),
		ClientSecret: buildFakePassword(),
		CreatedAt:    BuildFakeTime(),
	}
}

// BuildFakeOAuth2ClientToken builds a faked OAuth2ClientToken.
func BuildFakeOAuth2ClientToken() *types.OAuth2ClientToken {
	return &types.OAuth2ClientToken{
		RefreshCreatedAt:    BuildFakeTime(),
		AccessCreatedAt:     BuildFakeTime(),
		CodeCreatedAt:       BuildFakeTime(),
		RedirectURI:         fake.URL(),
		Code:                buildUniqueString(),
		CodeChallenge:       buildUniqueString(),
		CodeChallengeMethod: "S256",
		BelongsToUser:       BuildFakeID(),
		Access:              buildUniqueString(),
		ClientID:            BuildFakeID(),
		Refresh:             buildUniqueString(),
		ID:                  BuildFakeID(),
		CodeExpiresAt:       time.Hour,
		AccessExpiresAt:     time.Hour,
		RefreshExpiresAt:    time.Hour,
	}
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
	var examples []*types.OAuth2Client
	for range exampleQuantity {
		examples = append(examples, BuildFakeOAuth2Client())
	}

	return &filtering.QueryFilteredResult[types.OAuth2Client]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeOAuth2ClientCreationRequestInput builds a faked OAuth2ClientCreationRequestInput.
func BuildFakeOAuth2ClientCreationRequestInput() *types.OAuth2ClientCreationRequestInput {
	client := BuildFakeOAuth2Client()

	return &types.OAuth2ClientCreationRequestInput{
		Name:        client.Name,
		Description: client.Description,
	}
}
