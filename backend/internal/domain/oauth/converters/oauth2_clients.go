package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v9/identifiers"
)

func ConvertOAuth2ClientCreationRequestInputToOAuth2ClientDatabaseCreationInput(x *types.OAuth2ClientCreationRequestInput) *types.OAuth2ClientDatabaseCreationInput {
	return &types.OAuth2ClientDatabaseCreationInput{
		ID:           identifiers.New(),
		Name:         x.Name,
		Description:  x.Description,
		ClientID:     "",
		ClientSecret: "",
	}
}

// ConvertOAuth2ClientToOAuth2ClientDatabaseCreationInput builds a faked OAuth2ClientDatabaseCreationInput.
func ConvertOAuth2ClientToOAuth2ClientDatabaseCreationInput(client *types.OAuth2Client) *types.OAuth2ClientDatabaseCreationInput {
	return &types.OAuth2ClientDatabaseCreationInput{
		ID:           client.ID,
		Name:         client.Name,
		Description:  client.Description,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
	}
}

// ConvertOAuth2ClientToOAuth2ClientCreationInput builds a faked OAuth2ClientCreationRequestInput.
func ConvertOAuth2ClientToOAuth2ClientCreationInput(client *types.OAuth2Client) *types.OAuth2ClientCreationRequestInput {
	return &types.OAuth2ClientCreationRequestInput{
		Name:        client.Name,
		Description: client.Description,
	}
}

// ConvertOAuth2ClientToOAuth2ClientCreationResponse builds a faked OAuth2ClientCreationRequestInput.
func ConvertOAuth2ClientToOAuth2ClientCreationResponse(client *types.OAuth2Client) *types.OAuth2ClientCreationResponse {
	return &types.OAuth2ClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Name:         client.Name,
		Description:  client.Description,
	}
}

func ConvertOAuth2ClientTokenToOAuth2ClientTokenDatabaseCreationInput(x *types.OAuth2ClientToken) *types.OAuth2ClientTokenDatabaseCreationInput {
	return &types.OAuth2ClientTokenDatabaseCreationInput{
		RefreshCreatedAt:    x.RefreshCreatedAt,
		AccessCreatedAt:     x.AccessCreatedAt,
		CodeCreatedAt:       x.CodeCreatedAt,
		RedirectURI:         x.RedirectURI,
		Code:                x.Code,
		CodeChallenge:       x.CodeChallenge,
		CodeChallengeMethod: x.CodeChallengeMethod,
		BelongsToUser:       x.BelongsToUser,
		Access:              x.Access,
		ClientID:            x.ClientID,
		Refresh:             x.Refresh,
		ID:                  x.ID,
		CodeExpiresIn:       x.CodeExpiresAt,
		AccessExpiresIn:     x.AccessExpiresAt,
		RefreshExpiresIn:    x.RefreshExpiresAt,
	}
}
