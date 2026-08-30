package authentication

import (
	"encoding/json"
	"net/http"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
)

// The authorization server's HTTP surface is the platform's, and these methods exist only to
// carry it across the auth.AuthDataService interface the router builds against.
//
// They are handlers rather than a Mount call because the router owns the route table: which
// paths exist, what middleware bounds them, and — for /register — that it is not served at
// all. See internal/build/services/api/http/http_routes.go.

// AuthorizeHandler serves GET and POST /authorize.
func (s *service) AuthorizeHandler(res http.ResponseWriter, req *http.Request) {
	s.oauth2Server.AuthorizeHandler().ServeHTTP(res, req)
}

// TokenHandler serves POST /token.
func (s *service) TokenHandler(res http.ResponseWriter, req *http.Request) {
	s.oauth2Server.TokenHandler().ServeHTTP(res, req)
}

// RevokeHandler serves POST /revoke, RFC 7009 token revocation.
func (s *service) RevokeHandler(res http.ResponseWriter, req *http.Request) {
	s.oauth2Server.RevokeHandler().ServeHTTP(res, req)
}

// AuthorizationServerMetadataHandler serves the RFC 8414 discovery document, with the
// registration endpoint removed.
//
// The platform derives that document from the server so that what it advertises and what the
// endpoints do cannot disagree — but it always names a registration endpoint, and this server
// does not serve one: a client registration here is created through the permission-gated gRPC
// surface, not by an anonymous POST. Advertising it would be the exact failure the platform's
// own metadata doc warns about, with the sign flipped: a client believes the document, and
// finds a 404.
//
// Upstream ticket: oauth2server should let a deployment declare that dynamic registration is
// not served, and omit the field itself.
func (s *service) AuthorizationServerMetadataHandler(res http.ResponseWriter, req *http.Request) {
	metadata := s.oauth2Server.Metadata()

	document, err := metadataWithoutRegistration(&metadata)
	if err != nil {
		s.logger.WithRequest(req).Error("rendering authorization server metadata", err)
		res.WriteHeader(http.StatusInternalServerError)

		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err = res.Write(document); err != nil {
		s.logger.WithRequest(req).Error("writing authorization server metadata", err)
	}
}

// metadataWithoutRegistration renders the discovery document with registration_endpoint absent.
//
// Absent, not empty. The platform's AuthorizationServerMetadata declares that field without
// omitempty — it is always set there, because that server always mounts /register — so clearing
// it would advertise an empty string, which is a worse answer than no answer: a client that
// resolves it gets this server's own root. Round-tripping through a map is what removes the key
// rather than the value.
func metadataWithoutRegistration(metadata *oauth2server.AuthorizationServerMetadata) ([]byte, error) {
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	fields := map[string]any{}
	if err = json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}

	delete(fields, "registration_endpoint")

	return json.Marshal(fields)
}
