package mcpserver

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

const (
	authCodeLifetime     = 5 * time.Minute
	accessTokenLifetime  = 24 * time.Hour
	refreshTokenLifetime = 7 * 24 * time.Hour
	cleanupInterval      = 10 * time.Minute
)

// authCodeEntry stores a pending authorization code.
type authCodeEntry struct {
	expiresAt     time.Time
	userID        string
	accountID     string
	codeChallenge string
	redirectURI   string
	clientID      string
}

// accessTokenEntry maps an MCP access token to a user.
type accessTokenEntry struct {
	expiresAt time.Time
	userID    string
	accountID string
}

// refreshTokenEntry stores data needed to issue new access tokens.
type refreshTokenEntry struct {
	expiresAt time.Time
	userID    string
	accountID string
}

// registeredClient stores a dynamically registered OAuth2 client.
type registeredClient struct {
	createdAt    time.Time
	clientID     string
	clientSecret string
	clientName   string
	redirectURIs []string
}

// tokenStore manages all auth state for the MCP server's OAuth2 authorization server.
type tokenStore struct {
	authCodes     map[string]*authCodeEntry
	accessTokens  map[string]*accessTokenEntry
	refreshTokens map[string]*refreshTokenEntry
	clients       map[string]*registeredClient
	mu            sync.RWMutex
}

func newTokenStore() *tokenStore {
	return &tokenStore{
		authCodes:     make(map[string]*authCodeEntry),
		accessTokens:  make(map[string]*accessTokenEntry),
		refreshTokens: make(map[string]*refreshTokenEntry),
		clients:       make(map[string]*registeredClient),
	}
}

// startCleanupLoop periodically evicts expired entries.
func (ts *tokenStore) startCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ts.evictExpired()
			}
		}
	}()
}

func (ts *tokenStore) evictExpired() {
	now := time.Now()
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for k, v := range ts.authCodes {
		if now.After(v.expiresAt) {
			delete(ts.authCodes, k)
		}
	}
	for k, v := range ts.accessTokens {
		if now.After(v.expiresAt) {
			delete(ts.accessTokens, k)
		}
	}
	for k, v := range ts.refreshTokens {
		if now.After(v.expiresAt) {
			delete(ts.refreshTokens, k)
		}
	}
}

// verifyToken implements auth.TokenVerifier for the MCP SDK's RequireBearerToken middleware.
func (ts *tokenStore) verifyToken(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
	ts.mu.RLock()
	entry, ok := ts.accessTokens[token]
	ts.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: unknown token", auth.ErrInvalidToken)
	}
	if time.Now().After(entry.expiresAt) {
		return nil, fmt.Errorf("%w: token expired", auth.ErrInvalidToken)
	}

	return &auth.TokenInfo{
		UserID:     entry.userID,
		Expiration: entry.expiresAt,
		Extra:      map[string]any{"raw_token": token},
	}, nil
}

// userContextForToken returns the user ID and account ID associated with an MCP access token.
func (ts *tokenStore) userContextForToken(token string) (userID, accountID string, err error) {
	ts.mu.RLock()
	entry, ok := ts.accessTokens[token]
	ts.mu.RUnlock()

	if !ok {
		return "", "", fmt.Errorf("no entry for token")
	}
	if time.Now().After(entry.expiresAt) {
		return "", "", fmt.Errorf("token expired")
	}

	return entry.userID, entry.accountID, nil
}

// createAuthCode stores an authorization code that can be exchanged for tokens.
func (ts *tokenStore) createAuthCode(userID, accountID, codeChallenge, redirectURI, clientID string) (string, error) {
	code, err := generateOpaqueToken(32)
	if err != nil {
		return "", fmt.Errorf("generating auth code: %w", err)
	}

	ts.mu.Lock()
	ts.authCodes[code] = &authCodeEntry{
		userID:        userID,
		accountID:     accountID,
		codeChallenge: codeChallenge,
		redirectURI:   redirectURI,
		clientID:      clientID,
		expiresAt:     time.Now().Add(authCodeLifetime),
	}
	ts.mu.Unlock()

	return code, nil
}

// exchangeCode validates an authorization code and PKCE verifier, and returns MCP access and refresh tokens.
func (ts *tokenStore) exchangeCode(_ context.Context, code, codeVerifier, clientID, redirectURI string) (accessToken, refreshToken string, expiresIn int, err error) {
	ts.mu.Lock()
	entry, ok := ts.authCodes[code]
	if ok {
		delete(ts.authCodes, code) // one-time use
	}
	ts.mu.Unlock()

	if !ok {
		return "", "", 0, fmt.Errorf("invalid authorization code")
	}
	if time.Now().After(entry.expiresAt) {
		return "", "", 0, fmt.Errorf("authorization code expired")
	}
	if entry.clientID != clientID {
		return "", "", 0, fmt.Errorf("client_id mismatch")
	}
	if entry.redirectURI != redirectURI {
		return "", "", 0, fmt.Errorf("redirect_uri mismatch")
	}

	// Validate PKCE S256: SHA256(code_verifier) == code_challenge
	if !validatePKCES256(codeVerifier, entry.codeChallenge) {
		return "", "", 0, fmt.Errorf("PKCE verification failed")
	}

	// Generate opaque tokens.
	accessToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", 0, fmt.Errorf("generating access token: %w", err)
	}
	refreshToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", 0, fmt.Errorf("generating refresh token: %w", err)
	}

	now := time.Now()
	ts.mu.Lock()
	ts.accessTokens[accessToken] = &accessTokenEntry{
		userID:    entry.userID,
		accountID: entry.accountID,
		expiresAt: now.Add(accessTokenLifetime),
	}
	ts.refreshTokens[refreshToken] = &refreshTokenEntry{
		userID:    entry.userID,
		accountID: entry.accountID,
		expiresAt: now.Add(refreshTokenLifetime),
	}
	ts.mu.Unlock()

	return accessToken, refreshToken, int(accessTokenLifetime.Seconds()), nil
}

// refreshAccessToken uses a refresh token to issue new tokens.
func (ts *tokenStore) refreshAccessToken(_ context.Context, oldRefreshToken string) (accessToken, refreshToken string, expiresIn int, err error) {
	ts.mu.Lock()
	entry, ok := ts.refreshTokens[oldRefreshToken]
	if ok {
		delete(ts.refreshTokens, oldRefreshToken) // one-time use
	}
	ts.mu.Unlock()

	if !ok {
		return "", "", 0, fmt.Errorf("invalid refresh token")
	}
	if time.Now().After(entry.expiresAt) {
		return "", "", 0, fmt.Errorf("refresh token expired")
	}

	accessToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", 0, fmt.Errorf("generating access token: %w", err)
	}
	refreshToken, err = generateOpaqueToken(32)
	if err != nil {
		return "", "", 0, fmt.Errorf("generating refresh token: %w", err)
	}

	now := time.Now()
	ts.mu.Lock()
	ts.accessTokens[accessToken] = &accessTokenEntry{
		userID:    entry.userID,
		accountID: entry.accountID,
		expiresAt: now.Add(accessTokenLifetime),
	}
	ts.refreshTokens[refreshToken] = &refreshTokenEntry{
		userID:    entry.userID,
		accountID: entry.accountID,
		expiresAt: now.Add(refreshTokenLifetime),
	}
	ts.mu.Unlock()

	return accessToken, refreshToken, int(accessTokenLifetime.Seconds()), nil
}

// registerClient performs dynamic client registration (RFC 7591).
func (ts *tokenStore) registerClient(redirectURIs []string, clientName string) (*registeredClient, error) {
	id, err := generateOpaqueToken(16)
	if err != nil {
		return nil, fmt.Errorf("generating client_id: %w", err)
	}
	secret, err := generateOpaqueToken(32)
	if err != nil {
		return nil, fmt.Errorf("generating client_secret: %w", err)
	}

	rc := &registeredClient{
		clientID:     id,
		clientSecret: secret,
		redirectURIs: redirectURIs,
		clientName:   clientName,
		createdAt:    time.Now(),
	}

	ts.mu.Lock()
	ts.clients[id] = rc
	ts.mu.Unlock()

	return rc, nil
}

// generateOpaqueToken generates a cryptographically random hex-encoded token.
func generateOpaqueToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// validatePKCES256 checks that SHA256(base64url(code_verifier)) == code_challenge.
func validatePKCES256(codeVerifier, codeChallenge string) bool {
	h := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return computed == codeChallenge
}
