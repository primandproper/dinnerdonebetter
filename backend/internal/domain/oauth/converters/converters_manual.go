package converters

// The conversions in this file are hand-written: each does something the generator in
// cmd/tools/codegen/converters does not produce — it fails, it fans one value out into many, it
// defaults something, it needs a second entity to make sense of the first. exceptions.go names
// each one and says why.
//
// Everything else in this package is generated. A conversion that is a field copy with a handful
// of exceptions belongs there, where no destination field can be silently forgotten.

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
)

// ConvertOAuth2ClientToOAuth2ClientCreationResponse builds a faked OAuth2ClientCreationRequestInput.
func ConvertOAuth2ClientToOAuth2ClientCreationResponse(client *oauth.OAuth2Client) *oauth.OAuth2ClientCreationResponse {
	return &oauth.OAuth2ClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Name:         client.Name,
		Description:  client.Description,
	}
}
