package authentication

import (
	"net/http"
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

// AuthorizationServerMetadataHandler serves the RFC 8414 discovery document.
//
// The document omits registration_endpoint because the server is built with dynamic registration
// off (see ProvideOAuth2Server): a client registration here is created through the
// permission-gated gRPC surface, not by an anonymous POST.
func (s *service) AuthorizationServerMetadataHandler(res http.ResponseWriter, req *http.Request) {
	s.oauth2Server.MetadataHandler().ServeHTTP(res, req)
}
