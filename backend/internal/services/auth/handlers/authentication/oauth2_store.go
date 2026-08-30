package authentication

import (
	"context"
	"database/sql"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
)

// clientRegistryStore is the platform authorization server's Store with its client half
// redirected at this repository's own registry.
//
// The split is not arbitrary. Codes, access tokens and refresh tokens are protocol records:
// nothing outside the authorization server reads them, they are written and consumed by it
// alone, and the platform's tables model them better than ours did — digests rather than
// reversible ciphertext, a family identifier, a redemption stamp. Those move wholesale.
//
// A client registration here is not a protocol record. It is an administered object with a
// listing endpoint, three RBAC permissions, an archival lifecycle and an audit trail, none of
// which oauth2server.Store models — its client half is Create, Get and Delete, sized for the
// anonymous RFC 7591 registration that the MCP server serves and this one does not. Moving
// clients onto the platform's table would mean reimplementing all four of those against a
// table the platform owns, in exchange for nothing: GetClient is the only client method the
// authorization path calls, and it is four field copies away from what we already store.
//
// So: one table set for the protocol, one for the registry, one Store that reads whichever
// is the source of truth for what it was asked.
type clientRegistryStore struct {
	// Store carries every method not overridden below — the codes, the tokens, Sweep and
	// Close. Embedded rather than delegated field by field so that a Store method added
	// upstream reaches this one without a compile error that says nothing.
	oauth2server.Store

	clients oauth.OAuth2ClientDataManager
}

var _ oauth2server.Store = (*clientRegistryStore)(nil)

// errRegistrationEndpointNotServed is what the two write paths report.
//
// The API server does not mount /register: a client here is created through the
// permission-gated gRPC surface, and an anonymous endpoint that could write to the same
// registry would be a way around those permissions rather than a second way into it. These
// methods exist to satisfy the interface, and reaching one means a caller mounted an endpoint
// this store cannot back.
var errRegistrationEndpointNotServed = errors.New("this authorization server does not serve dynamic client registration")

// GetClient reads a registration from oauth2_clients.
//
// The lookup already excludes archived rows, so an archived client is a miss — which is the
// answer that should be given anyway: a registration that has been revoked is one the server
// has never heard of, as far as anything an anonymous caller can observe.
func (s *clientRegistryStore) GetClient(ctx context.Context, clientID string) (*oauth2server.Client, error) {
	if clientID == "" {
		return nil, oauth2server.ErrEmptyIdentifier
	}

	client, err := s.clients.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		// The repository reports a miss as a wrapped sql.ErrNoRows. Translating it here is
		// what stops it reaching the wire: without this the authorization server has no
		// protocol error to map, answers 500, and echoes "sql: no rows in result set" to an
		// unauthenticated caller — which is exactly what TestAuth_OAuth2AuthorizationCodeFlow
		// pinned as a defect before this change.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, platformerrors.Wrap(oauth2server.ErrNotFound, "oauth2 client")
		}

		return nil, err
	}

	if client == nil {
		return nil, platformerrors.Wrap(oauth2server.ErrNotFound, "oauth2 client")
	}

	return &oauth2server.Client{
		CreatedAt: client.CreatedAt,
		ID:        client.ClientID,
		Name:      client.Name,
		// Already the hex SHA-256 digest the platform expects: oauth.HashClientSecret and
		// oauth2server.Hash are the same function under two names, so every client registered
		// against the old server authenticates against this one without being re-issued.
		SecretHash:              client.ClientSecret,
		TokenEndpointAuthMethod: oauth2server.AuthMethodClientSecret,
		RedirectURIs:            client.RedirectURIs,
		GrantTypes:              []string{oauth2server.GrantTypeAuthorizationCode, oauth2server.GrantTypeRefreshToken},
		ResponseTypes:           []string{oauth2server.ResponseTypeCode},
		// No ExpiresAt: an administered registration does not lapse on a timer, it is
		// archived. The TTL the platform's Client carries is there to bound a table an
		// anonymous caller writes to, and this is not that table.
	}, nil
}

// CreateClient is not served. See errRegistrationEndpointNotServed.
func (s *clientRegistryStore) CreateClient(context.Context, *oauth2server.Client) error {
	return errRegistrationEndpointNotServed
}

// DeleteClient is not served. Archiving a client is the gRPC surface's job, and it records an
// audit entry while doing it.
func (s *clientRegistryStore) DeleteClient(context.Context, string) error {
	return errRegistrationEndpointNotServed
}
