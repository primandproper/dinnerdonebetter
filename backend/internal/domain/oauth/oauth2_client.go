package oauth

import (
	"context"
	"time"

	"github.com/primandproper/platform-go/v11/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// TablePrefix namespaces the platform authorization server's four tables,
// rendering ddb_oauth2_clients, ddb_oauth2_authorization_codes,
// ddb_oauth2_access_tokens, and ddb_oauth2_refresh_tokens.
//
// A prefix rather than the platform's empty default, and here the collision is
// not hypothetical: the platform's first table is named oauth2_clients, which is
// exactly the name 00004_oauth.sql already created for the go-oauth2 server this
// repository still runs. Its DDL says CREATE TABLE IF NOT EXISTS, so against our
// database the platform's schema would be a silent no-op followed by a store
// reading columns that are not there.
//
// The two servers therefore coexist during the migration rather than by accident,
// and the prefix stays right afterwards: when the API server moves onto these
// tables too, ddb_oauth2_clients is still the name, and the unprefixed pair goes
// away with the go-oauth2 dependency.
const TablePrefix = "ddb"

const (
	ClientIDSize     = 16
	ClientSecretSize = 16

	// OAuth2ClientCreatedServiceEventType indicates an OAuth2 client was created.
	OAuth2ClientCreatedServiceEventType = "oauth2_client_created"
	// OAuth2ClientArchivedServiceEventType indicates an OAuth2 client was archived.
	OAuth2ClientArchivedServiceEventType = "oauth2_client_archived"
)

type (
	// OAuth2Client represents a user-authorized OAuth2 client.
	OAuth2Client struct {
		_ struct{} `json:"-"`

		CreatedAt    time.Time  `json:"createdAt"`
		ArchivedAt   *time.Time `json:"archivedAt"`
		Name         string     `json:"name"`
		Description  string     `json:"description"`
		ClientID     string     `json:"clientID"`
		ID           string     `json:"id"`
		ClientSecret string     `json:"clientSecret"`
	}

	// OAuth2ClientCreationRequestInput is a struct for use when creating OAuth2 clients.
	OAuth2ClientCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// OAuth2ClientDatabaseCreationInput is a struct for use when creating OAuth2 clients.
	OAuth2ClientDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID           string `json:"-"`
		Name         string `json:"-"`
		Description  string `json:"-"`
		ClientID     string `json:"-"`
		ClientSecret string `json:"-"`
	}

	// OAuth2ClientCreationResponse is a struct for informing users of what their OAuth2 client's secret key is.
	OAuth2ClientCreationResponse struct {
		_ struct{} `json:"-"`

		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		ID           string `json:"id"`
	}

	// OAuth2ClientDataManager handles OAuth2 clients.
	OAuth2ClientDataManager interface {
		GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*OAuth2Client, error)
		GetOAuth2ClientByDatabaseID(ctx context.Context, id string) (*OAuth2Client, error)
		GetOAuth2Clients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[OAuth2Client], error)
		CreateOAuth2Client(ctx context.Context, input *OAuth2ClientDatabaseCreationInput) (*OAuth2Client, error)
		ArchiveOAuth2Client(ctx context.Context, clientID string) error
	}
)

// ValidateWithContext validates an APICreationInput.
func (x *OAuth2ClientCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
	)
}
