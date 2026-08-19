package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcpserver"

	"github.com/primandproper/platform-go/v11/authentication/oauth2server"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// redirectURI is where an authorization response is sent. Nothing listens there: the
// code is read off the Location header of the 302 rather than followed, and a loopback
// redirect on a port the client picked is what a native MCP client registers anyway.
const redirectURI = "http://127.0.0.1:33418/callback"

// instance is one MCP server's worth of state: a Service, and the listener its handler
// is served on.
//
// baseURL and address are separate on purpose. baseURL is the fleet's public address —
// what the discovery documents advertise and what a token's audience names — and every
// replica in a run shares it. address is where this particular process is listening,
// which is where the suite sends its requests. In a deployment those differ by a load
// balancer; here they differ by a port.
type instance struct {
	_ struct{} `json:"-"`

	service *mcpserver.Service
	server  *http.Server
	baseURL string
	address string
}

// startInstance binds a port, builds a service against the suite's database, and serves
// it. An empty baseURL means the replica is its own fleet, advertising the address it
// just bound.
//
// configure runs against this replica's own copy of the config, which is how a test asks
// for a deployment that differs from the rendered one — a configured issuer, say —
// without disturbing anyone else's.
func startInstance(ctx context.Context, baseURL string, configure func(*config.MCPServiceConfig)) (*instance, error) {
	// Bound before the service is built rather than after: a port picked and then
	// released is a port something else in this process — a container's host mapping,
	// another replica — can take in between.
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("binding listener: %w", err)
	}

	address := "http://" + listener.Addr().String()
	if baseURL == "" {
		baseURL = address
	}

	cfg := *mcpServiceConfig
	if configure != nil {
		configure(&cfg)
	}

	service, err := mcpserver.NewService(ctx, &cfg, baseURL)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("building MCP service: %w", err), listener.Close())
	}

	handler, err := service.Handler(ctx, mcpserver.TransportHTTP)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("building MCP handler: %w", err), listener.Close(), service.Shutdown(ctx))
	}

	server := &http.Server{
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		// ErrServerClosed is what stop() produces, and it is the only error this
		// goroutine can report to nobody.
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			panic(serveErr)
		}
	}()

	return &instance{service: service, server: server, baseURL: baseURL, address: address}, nil
}

// startInstanceForTest starts a replica and stops it when the test ends. Pass
// fleetBaseURL for another process serving the same deployment, or an empty string for a
// server that is a resource of its own.
func startInstanceForTest(t *testing.T, baseURL string, configure func(*config.MCPServiceConfig)) *instance {
	t.Helper()

	i, err := startInstance(t.Context(), baseURL, configure)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, i.stop(context.WithoutCancel(t.Context()))) })

	return i
}

// stop drains the HTTP server and then releases the container's database pool, in that
// order: a pool closed under an in-flight request is a request that fails rather than
// one that finishes.
func (i *instance) stop(ctx context.Context) error {
	return errors.Join(i.server.Shutdown(ctx), i.service.Shutdown(ctx))
}

// authorizationServerMetadata fetches the RFC 8414 document a replica publishes.
func (i *instance) authorizationServerMetadata(t *testing.T) *oauth2server.AuthorizationServerMetadata {
	t.Helper()

	doc := &oauth2server.AuthorizationServerMetadata{}
	getJSON(t, i.address+oauth2server.PathAuthorizationServerMetadata, doc)

	return doc
}

// protectedResourceMetadata fetches the RFC 9728 document a replica publishes.
func (i *instance) protectedResourceMetadata(t *testing.T) *oauth2server.ProtectedResourceMetadata {
	t.Helper()

	doc := &oauth2server.ProtectedResourceMetadata{}
	getJSON(t, i.address+oauth2server.PathProtectedResourceMetadata, doc)

	return doc
}

// registerClient performs RFC 7591 dynamic registration against a replica.
//
// token_endpoint_auth_method is none, because an MCP client running on a user's machine
// is a public client: it can keep no secret, which is what makes PKCE the whole of the
// protection here.
func (i *instance) registerClient(t *testing.T) *oauth2server.RegistrationResponse {
	t.Helper()

	res := i.post(t, oauth2server.PathRegister, "application/json",
		fmt.Sprintf(`{"redirect_uris":[%q],"client_name":"Integration Test Client","token_endpoint_auth_method":"none"}`, redirectURI))
	defer closeBody(t, res)

	body := readBody(t, res)
	require.Equal(t, http.StatusCreated, res.StatusCode, body)

	registration := &oauth2server.RegistrationResponse{}
	require.NoError(t, json.Unmarshal([]byte(body), registration))
	require.NotEmpty(t, registration.ClientID)

	return registration
}

// authorization is a code in hand, and the verifier that will redeem it.
type authorization struct {
	_ struct{} `json:"-"`

	code     string
	verifier string
	clientID string
}

// signIn drives GET /authorize for the login form and then POST /authorize with the
// admin's credentials, returning the code off the redirect.
//
// The resource parameter is the fleet's base URL rather than this replica's address: RFC
// 8707 names the protected resource a token is for, and a fleet is one resource however
// many processes serve it.
func (i *instance) signIn(t *testing.T, clientID, totpToken string) *authorization {
	t.Helper()

	// A fresh verifier per sign-in, from x/oauth2 rather than a constant: the challenge
	// binding is what PKCE is, and a suite that reused one would still pass if the
	// server stopped checking it.
	verifier := oauth2.GenerateVerifier()

	query := url.Values{
		"response_type":         {oauth2server.ResponseTypeCode},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {oauth2server.S256Challenge(verifier)},
		"code_challenge_method": {oauth2server.CodeChallengeMethodS256},
		"resource":              {fleetBaseURL},
		"state":                 {"opaque-state"},
	}.Encode()

	// The form first. A client that cannot render this cannot sign a human in, and it
	// is served by this application rather than by the platform.
	form := i.get(t, i.address+oauth2server.PathAuthorize+"?"+query)
	defer closeBody(t, form)
	require.Equal(t, http.StatusOK, form.StatusCode)
	require.Contains(t, readBody(t, form), `name="totp_token"`)

	res := i.post(t, oauth2server.PathAuthorize+"?"+query, "application/x-www-form-urlencoded", url.Values{
		"username":   {adminUser.Username},
		"password":   {adminUserPassword},
		"totp_token": {totpToken},
	}.Encode())
	defer closeBody(t, res)

	require.Equal(t, http.StatusFound, res.StatusCode, readBody(t, res))

	location, err := url.Parse(res.Header.Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "opaque-state", location.Query().Get("state"))

	code := location.Query().Get("code")
	require.NotEmpty(t, code, location.String())

	return &authorization{code: code, verifier: verifier, clientID: clientID}
}

// redeem exchanges an authorization code for a token at a replica — not necessarily the
// one that issued it.
func (i *instance) redeem(t *testing.T, authz *authorization) *oauth2server.TokenResponse {
	t.Helper()

	res := i.post(t, oauth2server.PathToken, "application/x-www-form-urlencoded", tokenRequestBody(authz))
	defer closeBody(t, res)

	body := readBody(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	token := &oauth2server.TokenResponse{}
	require.NoError(t, json.Unmarshal([]byte(body), token))
	require.NotEmpty(t, token.AccessToken)

	return token
}

// tokenRequestBody renders the form a code is redeemed with.
func tokenRequestBody(authz *authorization) string {
	return url.Values{
		"grant_type":    {oauth2server.GrantTypeAuthorizationCode},
		"code":          {authz.code},
		"code_verifier": {authz.verifier},
		"client_id":     {authz.clientID},
		"redirect_uri":  {redirectURI},
	}.Encode()
}

// authenticate runs the whole flow against a replica and hands back the access token:
// register, sign in, redeem.
func (i *instance) authenticate(t *testing.T) string {
	t.Helper()

	registration := i.registerClient(t)

	return i.redeem(t, i.signIn(t, registration.ClientID, generateTOTPCode(t))).AccessToken
}

// connect opens an MCP session against a replica with the given token.
//
// The session is initialized as part of connecting, so a token this server will not
// accept fails here rather than at whatever was going to be called with it.
func (i *instance) connect(t *testing.T, accessToken string) (*mcp.ClientSession, error) {
	t.Helper()

	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-integration-tests", Version: "v1.0.0"}, nil)

	return client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint:   i.address + "/mcp",
		HTTPClient: &http.Client{Transport: bearerTransport(accessToken), Timeout: 30 * time.Second},
		// The handler is stateless and answers with JSON, so there is no
		// server-initiated stream to hold open — and a client that opens one anyway
		// leaves a request in flight when the replica is asked to stop.
		DisableStandaloneSSE: true,
		// No retries: an unreachable replica is the assertion in the durability tests,
		// and retrying it would turn a failed dial into a slow one.
		MaxRetries: -1,
	}, nil)
}

// getValidIngredient opens a session and reads one row back through the MCP tool that
// serves it.
//
// One tool stands in for all of them here. What varies between them is the repository
// method behind the handler, which is covered elsewhere; what this suite is asking is
// whether a call arrives authenticated, with an account on the token, at a repository
// pointed at a real database — and any of the sixty tools would ask that the same way.
func (i *instance) getValidIngredient(t *testing.T, accessToken, ingredientID string) (*mcp.CallToolResult, error) {
	t.Helper()

	session, err := i.connect(t, accessToken)
	if err != nil {
		return nil, err
	}
	defer func() { require.NoError(t, session.Close()) }()

	return session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "GetValidIngredient",
		Arguments: map[string]any{"ValidIngredientID": ingredientID},
	})
}

// listTools opens a session and asks the replica what it serves.
func (i *instance) listTools(t *testing.T, accessToken string) (*mcp.ListToolsResult, error) {
	t.Helper()

	session, err := i.connect(t, accessToken)
	if err != nil {
		return nil, err
	}
	defer func() { require.NoError(t, session.Close()) }()

	return session.ListTools(t.Context(), &mcp.ListToolsParams{})
}

// get issues an unauthenticated GET to an absolute address.
func (i *instance) get(t *testing.T, address string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
	require.NoError(t, err)

	return do(t, req)
}

// post issues an unauthenticated POST to a path on this replica.
func (i *instance) post(t *testing.T, path, contentType, body string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, i.address+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)

	return do(t, req)
}

// do sends a request without following redirects, which is the whole of what this suite
// needs from an HTTP client: an authorization code arrives on the Location header of a
// 302 whose target is a port nothing is listening on.
func do(t *testing.T, req *http.Request) *http.Response {
	t.Helper()

	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	res, err := client.Do(req) //nolint:gosec // G704: every address here is a listener this suite bound on loopback.
	require.NoError(t, err)

	return res
}

// bearerTransport presents a token on every request the MCP SDK makes.
type bearerTransport string

func (b bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Cloned: RoundTrip must not modify the request it is given.
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+string(b))

	return http.DefaultTransport.RoundTrip(req)
}

func generateTOTPCode(t *testing.T) string {
	t.Helper()

	code, err := totp.GenerateCode(strings.ToUpper(adminUser.TwoFactorSecret), time.Now().UTC())
	require.NoError(t, err)

	return code
}

func getJSON(t *testing.T, address string, into any) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
	require.NoError(t, err)

	res := do(t, req)
	defer closeBody(t, res)

	body := readBody(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, json.Unmarshal([]byte(body), into))
}

func readBody(t *testing.T, res *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return string(body)
}

func closeBody(t *testing.T, res *http.Response) {
	t.Helper()

	require.NoError(t, res.Body.Close())
}

// requireStructuredContent decodes a tool result's structured output into a domain type.
//
// Through the JSON the client actually received rather than by asserting on the map it
// decoded to: the field names on the wire are the ones a model sees, and a schema that
// disagrees with them is the failure this is worth catching.
func requireStructuredContent(t *testing.T, res *mcp.CallToolResult, into any) {
	t.Helper()

	require.NotNil(t, res)
	require.False(t, res.IsError, res.Content)
	require.NotNil(t, res.StructuredContent)

	body, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, into))
}
