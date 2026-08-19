package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
)

func ConvertGRPCOAuth2ClientCreationRequestInputToOAuth2ClientCreationRequestInput(input *oauthsvc.OAuth2ClientCreationRequestInput) *oauth.OAuth2ClientCreationRequestInput {
	return &oauth.OAuth2ClientCreationRequestInput{
		Name:         input.Name,
		Description:  input.Description,
		RedirectURIs: input.RedirectUris,
	}
}

func ConvertOAuth2ClientCreationRequestInputToGRPCOAuth2ClientCreationRequestInput(input *oauth.OAuth2ClientCreationRequestInput) *oauthsvc.OAuth2ClientCreationRequestInput {
	return &oauthsvc.OAuth2ClientCreationRequestInput{
		Name:         input.Name,
		Description:  input.Description,
		RedirectUris: input.RedirectURIs,
	}
}

func ConvertOAuth2ClientToGRPCOAuth2Client(client *oauth.OAuth2Client) *oauthsvc.OAuth2Client {
	return &oauthsvc.OAuth2Client{
		CreatedAt:    grpcconverters.ConvertTimeToPBTimestamp(client.CreatedAt),
		ArchivedAt:   grpcconverters.ConvertTimePointerToPBTimestamp(client.ArchivedAt),
		Name:         client.Name,
		Description:  client.Description,
		ClientId:     client.ClientID,
		Id:           client.ID,
		ClientSecret: client.ClientSecret,
		RedirectUris: client.RedirectURIs,
	}
}

func ConvertGRPCOAuth2ClientToOAuth2Client(client *oauthsvc.OAuth2Client) *oauth.OAuth2Client {
	return &oauth.OAuth2Client{
		CreatedAt:    grpcconverters.ConvertPBTimestampToTime(client.CreatedAt),
		ArchivedAt:   grpcconverters.ConvertPBTimestampToTimePointer(client.ArchivedAt),
		Name:         client.Name,
		Description:  client.Description,
		ClientID:     client.ClientId,
		ID:           client.Id,
		ClientSecret: client.ClientSecret,
		RedirectURIs: client.RedirectUris,
	}
}
