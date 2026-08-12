package authentication

import (
	"net/http"

	"github.com/primandproper/platform-go/v10/observability"
)

// AuthorizeHandler is our oauth2 auth route.
func (s *service) AuthorizeHandler(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	if err := s.oauth2Server.HandleAuthorizeRequest(res, req); err != nil {
		observability.AcknowledgeError(err, logger, span, "handling authorization request")
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
}

func (s *service) TokenHandler(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	if err := s.oauth2Server.HandleTokenRequest(res, req); err != nil {
		observability.AcknowledgeError(err, logger, span, "handling token request")
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
