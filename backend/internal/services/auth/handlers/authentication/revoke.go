package authentication

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	oauthkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/keys"

	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

const (
	tokenTypeHintAccess  = "access_token"
	tokenTypeHintRefresh = "refresh_token"
)

// RevokeHandler implements RFC 7009 OAuth 2.0 Token Revocation.
// It revokes an access or refresh token when the client presents valid credentials
// and the token belongs to that client. Per RFC 7009, always returns 200 for valid
// requests (even if the token was invalid or already revoked) to prevent token enumeration.
func (s *service) RevokeHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req).WithSpan(span)
	tracing.AttachRequestToSpan(span, req)

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := req.ParseForm(); err != nil {
		logger.Error("parsing revoke form", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	token := req.Form.Get("token")
	if token == "" {
		// RFC 7009: respond 200 even for invalid requests when token is missing?
		// Actually RFC says "invalid token" returns 200 - missing token might be 400.
		// RFC 7009 section 2.1: token (REQUIRED). So we can return 400 for missing token.
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	clientID, clientSecret := req.Form.Get("client_id"), req.Form.Get("client_secret")
	if clientID == "" || clientSecret == "" {
		username, password, ok := req.BasicAuth()
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		clientID, clientSecret = username, password
	}

	oauthClient, err := s.oauthRepo.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching oauth2 client for revoke")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	if oauthClient.ArchivedAt != nil {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	if oauthClient.ClientSecret != clientSecret {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenHint := req.Form.Get("token_type_hint")

	var tokenInfo *types.OAuth2ClientToken
	switch tokenHint {
	case tokenTypeHintRefresh:
		tokenInfo, err = s.oauthRepo.GetOAuth2ClientTokenByRefresh(ctx, token)
	case tokenTypeHintAccess:
		tokenInfo, err = s.oauthRepo.GetOAuth2ClientTokenByAccess(ctx, token)
	default:
		// No hint or unknown hint: try access first, then refresh per RFC 7009
		tokenInfo, err = s.oauthRepo.GetOAuth2ClientTokenByAccess(ctx, token)
		if err != nil {
			tokenInfo, err = s.oauthRepo.GetOAuth2ClientTokenByRefresh(ctx, token)
		}
	}

	if err != nil || tokenInfo == nil {
		// RFC 7009: return 200 for invalid/unknown token (avoids enumeration)
		res.WriteHeader(http.StatusOK)
		return
	}

	if tokenInfo.ClientID != clientID {
		// Token was not issued to this client; do not revoke, return 200
		res.WriteHeader(http.StatusOK)
		return
	}

	if token == tokenInfo.Refresh {
		err = s.oauthRepo.DeleteOAuth2ClientTokenByRefresh(ctx, token)
	} else {
		err = s.oauthRepo.DeleteOAuth2ClientTokenByAccess(ctx, token)
	}
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "revoking oauth2 token")
		// Still return 200 per RFC 7009 - don't reveal revocation failure
		res.WriteHeader(http.StatusOK)
		return
	}

	dcm := &audit.DataChangeMessage{
		EventType: identity.UserLoggedOutServiceEventType,
		UserID:    tokenInfo.BelongsToUser,
		Context:   map[string]any{},
	}
	if err = s.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing user logged out event")
	}

	logger.WithValue(oauthkeys.OAuth2ClientIDKey, clientID).WithValue(identitykeys.UserIDKey, tokenInfo.BelongsToUser).Info("oauth2 token revoked")
	res.WriteHeader(http.StatusOK)
}
