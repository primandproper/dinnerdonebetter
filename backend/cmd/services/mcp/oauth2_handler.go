package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v6/authentication/totp"
	"github.com/primandproper/platform-go/v6/routing"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// registerOAuth2Routes adds all OAuth2 authorization server endpoints to the router.
func registerOAuth2Routes(router *routing.Router, ts *tokenStore, baseURL string, identityRepo identity.Repository, authenticator authentication.Authenticator, totpVerifier totp.Verifier) {
	// Protected Resource Metadata (RFC 9728)
	router.Handle(http.MethodGet, "/.well-known/oauth-protected-resource", auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               baseURL,
		AuthorizationServers:   []string{baseURL},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Dinner Done Better MCP Server",
	}))

	// Authorization Server Metadata (RFC 8414)
	router.Handle(http.MethodGet, "/.well-known/oauth-authorization-server", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                baseURL,
			"authorization_endpoint":                baseURL + "/authorize",
			"token_endpoint":                        baseURL + "/token",
			"registration_endpoint":                 baseURL + "/register",
			"response_types_supported":              []string{"code"},
			"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
			"code_challenge_methods_supported":      []string{"S256"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_post", "none"},
		}); err != nil {
			http.Error(w, "failed to encode metadata", http.StatusInternalServerError)
		}
	}))

	// Authorization endpoint — serves login form and processes login
	router.Handle(http.MethodGet, "/authorize", http.HandlerFunc(handleAuthorizeGET))
	router.Handle(http.MethodPost, "/authorize", handleAuthorizePOST(ts, identityRepo, authenticator, totpVerifier))

	// Token endpoint — exchanges codes for tokens and handles refresh
	router.Handle(http.MethodPost, "/token", handleToken(ts))

	// Dynamic Client Registration (RFC 7591)
	router.Handle(http.MethodPost, "/register", handleRegister(ts))
}

// loginFormData is template data for the login form.
type loginFormData struct {
	ClientID            string
	RedirectURI         string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Error               string
}

var loginFormTemplate = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dinner Done Better — Sign In</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .card { background: white; border-radius: 12px; padding: 2rem; width: 100%; max-width: 400px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h1 { font-size: 1.5rem; margin-bottom: 1.5rem; text-align: center; }
        label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 0.25rem; color: #333; }
        input[type="text"], input[type="password"] { width: 100%; padding: 0.625rem; border: 1px solid #ddd; border-radius: 6px; font-size: 1rem; margin-bottom: 1rem; }
        input:focus { outline: none; border-color: #4a90d9; box-shadow: 0 0 0 2px rgba(74,144,217,0.2); }
        button { width: 100%; padding: 0.75rem; background: #4a90d9; color: white; border: none; border-radius: 6px; font-size: 1rem; font-weight: 500; cursor: pointer; }
        button:hover { background: #3a7bc8; }
        .error { background: #fee; color: #c33; padding: 0.75rem; border-radius: 6px; margin-bottom: 1rem; font-size: 0.875rem; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Sign In</h1>
        {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
        <form method="POST" action="/authorize">
            <input type="hidden" name="client_id" value="{{.ClientID}}">
            <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
            <input type="hidden" name="state" value="{{.State}}">
            <input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
            <input type="hidden" name="code_challenge_method" value="{{.CodeChallengeMethod}}">

            <label for="username">Username</label>
            <input type="text" id="username" name="username" required autofocus>

            <label for="password">Password</label>
            <input type="password" id="password" name="password" required>

            <label for="totp_token">TOTP Code</label>
            <input type="text" id="totp_token" name="totp_token" autocomplete="one-time-code" inputmode="numeric" pattern="[0-9]*">

            <button type="submit">Sign In</button>
        </form>
    </div>
</body>
</html>`))

// handleAuthorizeGET serves the login form.
func handleAuthorizeGET(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")

	if responseType != "code" {
		http.Error(w, "unsupported response_type, must be 'code'", http.StatusBadRequest)
		return
	}
	if clientID == "" || redirectURI == "" {
		http.Error(w, "client_id and redirect_uri are required", http.StatusBadRequest)
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		http.Error(w, "code_challenge with method S256 is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := loginFormTemplate.Execute(w, &loginFormData{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}); err != nil {
		log.Printf("error rendering login form: %v", err)
	}
}

// handleAuthorizePOST processes the login form submission by validating credentials directly against the database.
func handleAuthorizePOST(ts *tokenStore, identityRepo identity.Repository, authenticator authentication.Authenticator, totpVerifier totp.Verifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		totpToken := r.FormValue("totp_token")
		clientID := r.FormValue("client_id")
		redirectURI := r.FormValue("redirect_uri")
		state := r.FormValue("state")
		codeChallenge := r.FormValue("code_challenge")
		codeChallengeMethod := r.FormValue("code_challenge_method")

		if codeChallengeMethod != "S256" {
			http.Error(w, "code_challenge_method must be S256", http.StatusBadRequest)
			return
		}

		renderLoginError := func(errMsg string) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			if tmplErr := loginFormTemplate.Execute(w, &loginFormData{
				ClientID:            clientID,
				RedirectURI:         redirectURI,
				State:               state,
				CodeChallenge:       codeChallenge,
				CodeChallengeMethod: codeChallengeMethod,
				Error:               errMsg,
			}); tmplErr != nil {
				log.Printf("error rendering login form: %v", tmplErr)
			}
		}

		// Look up user by username (admin-only).
		user, err := identityRepo.GetAdminUserByUsername(r.Context(), username)
		if err != nil || user == nil {
			renderLoginError("Access denied. Admin credentials required.")
			return
		}

		if user.IsBanned() {
			renderLoginError("Access denied. Account is banned.")
			return
		}

		// Validate the password; TOTP is verified separately below.
		matches, err := authenticator.PasswordMatches(r.Context(), user.HashedPassword, password)
		if err != nil || !matches {
			renderLoginError("Access denied. Admin credentials required.")
			return
		}

		// If TOTP is enabled, verify the code.
		if user.TwoFactorSecretVerifiedAt != nil {
			if verifyErr := totpVerifier.Verify(r.Context(), user.TwoFactorSecret, totpToken); verifyErr != nil {
				if errors.Is(verifyErr, totp.ErrCodeRequired) {
					renderLoginError("TOTP code is required.")
					return
				}
				renderLoginError("Access denied. Admin credentials required.")
				return
			}
		}

		// Get the user's default account.
		accountID, err := identityRepo.GetDefaultAccountIDForUser(r.Context(), user.ID)
		if err != nil {
			log.Printf("error getting default account for user: %v", err)
			http.Error(w, "internal error resolving account", http.StatusInternalServerError)
			return
		}

		// Create authorization code.
		code, err := ts.createAuthCode(user.ID, accountID, codeChallenge, redirectURI, clientID)
		if err != nil {
			http.Error(w, "internal error creating authorization code", http.StatusInternalServerError)
			return
		}

		// Redirect back to the client with the authorization code.
		redirectURL, err := url.Parse(redirectURI)
		if err != nil {
			http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
			return
		}

		q := redirectURL.Query()
		q.Set("code", code)
		if state != "" {
			q.Set("state", state)
		}
		redirectURL.RawQuery = q.Encode()

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	}
}

// tokenResponse is the OAuth2 token endpoint response.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

// tokenErrorResponse is the OAuth2 token endpoint error response.
type tokenErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// handleToken handles the OAuth2 token endpoint.
func handleToken(ts *tokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			writeTokenError(w, "invalid_request", "invalid form data")
			return
		}

		grantType := r.FormValue("grant_type")

		switch grantType {
		case "authorization_code":
			code := r.FormValue("code")
			codeVerifier := r.FormValue("code_verifier")
			clientID := r.FormValue("client_id")
			redirectURI := r.FormValue("redirect_uri")

			if code == "" || codeVerifier == "" || clientID == "" {
				writeTokenError(w, "invalid_request", "code, code_verifier, and client_id are required")
				return
			}

			accessToken, refreshToken, expiresIn, err := ts.exchangeCode(r.Context(), code, codeVerifier, clientID, redirectURI)
			if err != nil {
				writeTokenError(w, "invalid_grant", err.Error())
				return
			}

			writeTokenResponse(w, &tokenResponse{
				AccessToken:  accessToken,
				TokenType:    "Bearer",
				ExpiresIn:    expiresIn,
				RefreshToken: refreshToken,
			})

		case "refresh_token":
			refreshToken := r.FormValue("refresh_token")
			if refreshToken == "" {
				writeTokenError(w, "invalid_request", "refresh_token is required")
				return
			}

			accessToken, newRefreshToken, expiresIn, err := ts.refreshAccessToken(r.Context(), refreshToken)
			if err != nil {
				writeTokenError(w, "invalid_grant", err.Error())
				return
			}

			writeTokenResponse(w, &tokenResponse{
				AccessToken:  accessToken,
				TokenType:    "Bearer",
				ExpiresIn:    expiresIn,
				RefreshToken: newRefreshToken,
			})

		default:
			writeTokenError(w, "unsupported_grant_type", fmt.Sprintf("unsupported grant_type: %s", grantType))
		}
	}
}

// clientRegistrationRequest is the request body for dynamic client registration.
type clientRegistrationRequest struct {
	RedirectURIs []string `json:"redirect_uris"`
	ClientName   string   `json:"client_name,omitempty"`
	GrantTypes   []string `json:"grant_types,omitempty"`
}

// clientRegistrationResponse is the response body for dynamic client registration.
type clientRegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret"`
	ClientName              string   `json:"client_name,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at"`
}

// handleRegister handles dynamic client registration (RFC 7591).
func handleRegister(ts *tokenStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req clientRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_client_metadata","error_description":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}

		if len(req.RedirectURIs) == 0 {
			http.Error(w, `{"error":"invalid_client_metadata","error_description":"redirect_uris is required"}`, http.StatusBadRequest)
			return
		}

		rc, err := ts.registerClient(req.RedirectURIs, req.ClientName)
		if err != nil {
			http.Error(w, `{"error":"server_error","error_description":"failed to register client"}`, http.StatusInternalServerError)
			return
		}

		resp := &clientRegistrationResponse{
			ClientID:                rc.clientID,
			ClientSecret:            rc.clientSecret,
			ClientName:              rc.clientName,
			RedirectURIs:            rc.redirectURIs,
			GrantTypes:              []string{"authorization_code", "refresh_token"},
			TokenEndpointAuthMethod: "client_secret_post",
			ClientIDIssuedAt:        rc.createdAt.Unix(),
			ClientSecretExpiresAt:   0, // never expires
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
			log.Printf("error encoding registration response: %v", encodeErr)
		}
	}
}

func writeTokenResponse(w http.ResponseWriter, resp *tokenResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding token response: %v", err)
	}
}

func writeTokenError(w http.ResponseWriter, errCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(&tokenErrorResponse{
		Error:       errCode,
		Description: description,
	}); err != nil {
		log.Printf("error encoding token error response: %v", err)
	}
}
