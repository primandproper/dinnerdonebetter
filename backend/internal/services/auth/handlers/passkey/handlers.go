package passkey

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v7/cookies"
	"github.com/primandproper/platform-go/v7/encoding"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

// AuthOptionsRequest is the request body for passkey authentication options.
type AuthOptionsRequest struct {
	Username string `json:"username"`
}

// AuthOptionsResponse is the response for passkey authentication options.
type AuthOptionsResponse struct {
	Challenge                         string `json:"challenge"`
	PublicKeyCredentialRequestOptions []byte `json:"publicKeyCredentialRequestOptions"`
}

// AuthVerifyRequest is the request body for passkey authentication verify.
// AssertionResponse is the raw JSON bytes of the WebAuthn credential (id, rawId, type, response).
type AuthVerifyRequest struct {
	Challenge         string          `json:"challenge"`
	Username          string          `json:"username"`
	AssertionResponse json.RawMessage `json:"assertionResponse"`
}

// RegOptionsResponse is the response for passkey registration options.
type RegOptionsResponse struct {
	Challenge                          string `json:"challenge"`
	PublicKeyCredentialCreationOptions []byte `json:"publicKeyCredentialCreationOptions"`
}

// RegVerifyRequest is the request body for passkey registration verify.
type RegVerifyRequest struct {
	Challenge           string          `json:"challenge"`
	AttestationResponse json.RawMessage `json:"attestationResponse"`
}

// Handlers provides HTTP handlers for passkey authentication and registration.
type Handlers struct {
	tracer                   tracing.Tracer
	logger                   logging.Logger
	encoder                  encoding.ServerEncoderDecoder
	cookieManager            cookies.Manager
	cookieConfig             *cookies.Config
	buildUnauthedClient      func(context.Context) (client.Client, error)
	getClientForRegistration func(*http.Request) (client.Client, error) // optional; nil means registration disabled
}

// NewHandlers creates passkey HTTP handlers. getClientForRegistration is optional (pass nil to disable
// passkey registration); when provided, it is used by RegOptionsHandler and RegVerifyHandler to obtain
// an authenticated gRPC client (e.g. from request context after auth middleware).
func NewHandlers(
	tracer tracing.Tracer,
	logger logging.Logger,
	encoder encoding.ServerEncoderDecoder,
	cookieManager cookies.Manager,
	cookieConfig *cookies.Config,
	buildUnauthedClient func(context.Context) (client.Client, error),
	getClientForRegistration func(*http.Request) (client.Client, error),
) *Handlers {
	return &Handlers{
		tracer:                   tracer,
		logger:                   logger,
		encoder:                  encoder,
		cookieManager:            cookieManager,
		cookieConfig:             cookieConfig,
		buildUnauthedClient:      buildUnauthedClient,
		getClientForRegistration: getClientForRegistration,
	}
}

// AuthOptionsHandler handles POST /auth/passkey/authentication/options.
func (h *Handlers) AuthOptionsHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := h.tracer.StartSpan(req.Context())
	defer span.End()

	logger := h.logger.WithRequest(req)

	var body AuthOptionsRequest
	if err := h.encoder.DecodeRequest(ctx, req, &body); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding passkey options request")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "invalid request"}, http.StatusBadRequest)
		return
	}

	unauthedClient, err := h.buildUnauthedClient(ctx)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "building gRPC client")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "service unavailable"}, http.StatusServiceUnavailable)
		return
	}

	optsRes, err := unauthedClient.BeginPasskeyAuthentication(ctx, &authsvc.BeginPasskeyAuthenticationRequest{
		Username: strings.TrimSpace(body.Username),
	})
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "beginning passkey authentication")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "failed to get passkey options"}, http.StatusInternalServerError)
		return
	}

	h.encoder.EncodeResponseWithStatus(ctx, res, AuthOptionsResponse{
		PublicKeyCredentialRequestOptions: optsRes.PublicKeyCredentialRequestOptions,
		Challenge:                         optsRes.Challenge,
	}, http.StatusOK)
}

// AuthVerifyHandler handles POST /auth/passkey/authentication/verify.
func (h *Handlers) AuthVerifyHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := h.tracer.StartSpan(req.Context())
	defer span.End()

	logger := h.logger.WithRequest(req)

	var body AuthVerifyRequest
	if err := h.encoder.DecodeRequest(ctx, req, &body); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding passkey verify request")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "invalid request"}, http.StatusBadRequest)
		return
	}

	if len([]byte(body.AssertionResponse)) == 0 || body.Challenge == "" {
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "assertion_response and challenge are required"}, http.StatusBadRequest)
		return
	}

	unauthedClient, err := h.buildUnauthedClient(ctx)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "building gRPC client")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "service unavailable"}, http.StatusServiceUnavailable)
		return
	}

	tokenRes, err := unauthedClient.FinishPasskeyAuthentication(ctx, &authsvc.FinishPasskeyAuthenticationRequest{
		AssertionResponse: []byte(body.AssertionResponse),
		Challenge:         body.Challenge,
		Username:          strings.TrimSpace(body.Username),
	})
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "finishing passkey authentication")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "authentication failed"}, http.StatusUnauthorized)
		return
	}

	encodedCookie, err := h.cookieManager.Encode(ctx, h.cookieConfig.CookieName, &identity.TokenResponse{AccessToken: tokenRes.Result.AccessToken})
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding auth cookie")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "failed to complete login"}, http.StatusInternalServerError)
		return
	}

	http.SetCookie(res, BuildCookie(h.cookieConfig, encodedCookie))

	res.Header().Set("HX-Redirect", "/")
	h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"redirect": "/"}, http.StatusOK)
}

// RegOptionsHandler handles POST /auth/passkey/registration/options. Requires authentication.
func (h *Handlers) RegOptionsHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := h.tracer.StartSpan(req.Context())
	defer span.End()

	logger := h.logger.WithRequest(req)

	if h.getClientForRegistration == nil {
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "passkey registration not configured"}, http.StatusNotImplemented)
		return
	}

	authedClient, err := h.getClientForRegistration(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting authenticated client for passkey registration")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "authentication required"}, http.StatusUnauthorized)
		return
	}

	optsRes, err := authedClient.BeginPasskeyRegistration(ctx, &authsvc.BeginPasskeyRegistrationRequest{})
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "beginning passkey registration")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "failed to get passkey options"}, http.StatusInternalServerError)
		return
	}

	h.encoder.EncodeResponseWithStatus(ctx, res, RegOptionsResponse{
		PublicKeyCredentialCreationOptions: optsRes.PublicKeyCredentialCreationOptions,
		Challenge:                          optsRes.Challenge,
	}, http.StatusOK)
}

// RegVerifyHandler handles POST /auth/passkey/registration/verify. Requires authentication.
func (h *Handlers) RegVerifyHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := h.tracer.StartSpan(req.Context())
	defer span.End()

	logger := h.logger.WithRequest(req)

	if h.getClientForRegistration == nil {
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "passkey registration not configured"}, http.StatusNotImplemented)
		return
	}

	var body RegVerifyRequest
	if err := h.encoder.DecodeRequest(ctx, req, &body); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding passkey registration verify request")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "invalid request"}, http.StatusBadRequest)
		return
	}

	if len([]byte(body.AttestationResponse)) == 0 || body.Challenge == "" {
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "attestation_response and challenge are required"}, http.StatusBadRequest)
		return
	}

	authedClient, err := h.getClientForRegistration(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting authenticated client for passkey registration")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "authentication required"}, http.StatusUnauthorized)
		return
	}

	_, err = authedClient.FinishPasskeyRegistration(ctx, &authsvc.FinishPasskeyRegistrationRequest{
		AttestationResponse: []byte(body.AttestationResponse),
		Challenge:           body.Challenge,
	})
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "finishing passkey registration")
		h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"error": "failed to register passkey"}, http.StatusBadRequest)
		return
	}

	h.encoder.EncodeResponseWithStatus(ctx, res, map[string]string{"success": "true"}, http.StatusOK)
}
