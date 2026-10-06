package auth

import (
	"net/http"
)

type (
	// AuthDataService describes a structure capable of handling passwords and authorization requests.
	//
	// The four handlers are the OAuth 2.1 authorization server's HTTP surface. Registration
	// is absent on purpose: an OAuth2 client here is created through the permission-gated
	// gRPC surface rather than by an anonymous RFC 7591 POST.
	AuthDataService interface {
		AuthorizeHandler(res http.ResponseWriter, req *http.Request)
		TokenHandler(res http.ResponseWriter, req *http.Request)
		RevokeHandler(res http.ResponseWriter, req *http.Request)
		AuthorizationServerMetadataHandler(res http.ResponseWriter, req *http.Request)
	}
)
