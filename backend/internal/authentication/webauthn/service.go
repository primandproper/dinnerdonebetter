package webauthn

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability/logging"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	serviceO11yName = "webauthn_service"

	defaultSessionTTL = 5 * time.Minute
)

// Config holds WebAuthn configuration.
type Config struct {
	RPID          string   // Relying Party ID (e.g. "dinnerdonebetter.com" or "localhost")
	RPDisplayName string   // Display name for the RP
	RPOrigins     []string // Allowed origins (e.g. "https://dinnerdonebetter.com", "https://localhost:8080")
}

// Service provides passkey registration and authentication.
type Service struct {
	logger       logging.Logger
	webauthn     *webauthn.WebAuthn
	credStore    identity.WebAuthnCredentialDataManager
	userStore    UserStore
	sessionStore SessionStore
}

// UserStore provides user lookup for WebAuthn.
type UserStore interface {
	GetUserByID(ctx context.Context, userID string) (*identity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*identity.User, error)
}

// NewService creates a new WebAuthn service.
func NewService(logger logging.Logger, cfg Config, credStore identity.WebAuthnCredentialDataManager, userStore UserStore, sessionStore SessionStore) (*Service, error) {
	w, err := webauthn.New(&webauthn.Config{
		RPID:          cfg.RPID,
		RPDisplayName: cfg.RPDisplayName,
		RPOrigins:     cfg.RPOrigins,
		Timeouts: webauthn.TimeoutsConfig{
			Login:        webauthn.TimeoutConfig{Timeout: 60000},
			Registration: webauthn.TimeoutConfig{Timeout: 60000},
		},
	})
	if err != nil {
		return nil, err
	}
	return &Service{
		logger:       logging.NewNamedLogger(logger, serviceO11yName),
		webauthn:     w,
		credStore:    credStore,
		userStore:    userStore,
		sessionStore: sessionStore,
	}, nil
}

// BeginRegistrationOptions returns PublicKeyCredentialCreationOptions for the given user.
func (s *Service) BeginRegistrationOptions(ctx context.Context, userID string) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, nil, err
	}
	creds, err := s.credStore.GetWebAuthnCredentialsForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	waUser := &WebAuthnUser{
		User:        user,
		Credentials: creds,
		UserID:      []byte(user.ID),
	}
	creation, session, err := s.webauthn.BeginRegistration(waUser)
	if err != nil {
		return nil, nil, err
	}
	ttl := defaultSessionTTL
	if session.Expires.After(time.Now()) {
		ttl = time.Until(session.Expires)
	}
	if saveErr := s.sessionStore.SaveSession(ctx, session.Challenge, session, ttl); saveErr != nil {
		return nil, nil, saveErr
	}
	return creation, session, nil
}

// FinishRegistration validates the attestation and stores the credential.
func (s *Service) FinishRegistration(ctx context.Context, userID string, req *http.Request) error {
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return err
	}
	parsed, err := protocol.ParseCredentialCreationResponse(req)
	if err != nil {
		return err
	}
	session, err := s.sessionStore.GetSession(ctx, parsed.Response.CollectedClientData.Challenge)
	if err != nil || session == nil {
		return err
	}
	creds, err := s.credStore.GetWebAuthnCredentialsForUser(ctx, userID)
	if err != nil {
		return err
	}
	waUser := &WebAuthnUser{
		User:        user,
		Credentials: creds,
		UserID:      []byte(user.ID),
	}
	credential, err := s.webauthn.CreateCredential(waUser, *session, parsed)
	if err != nil {
		return err
	}
	transportsJSON := ""
	if len(credential.Transport) > 0 {
		t := make([]string, len(credential.Transport))
		for i, tr := range credential.Transport {
			t[i] = string(tr)
		}
		b, marshalErr := json.Marshal(t)
		if marshalErr != nil {
			return marshalErr
		}
		transportsJSON = string(b)
	}
	_, err = s.credStore.CreateWebAuthnCredential(ctx, &identity.WebAuthnCredentialCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: userID,
		CredentialID:  credential.ID,
		PublicKey:     credential.PublicKey,
		SignCount:     credential.Authenticator.SignCount,
		Transports:    transportsJSON,
		FriendlyName:  "",
	})
	return err
}

// BeginAuthenticationOptions returns PublicKeyCredentialRequestOptions for the given username.
func (s *Service) BeginAuthenticationOptions(ctx context.Context, username string) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	var waUser *WebAuthnUser
	if username != "" {
		user, err := s.userStore.GetUserByUsername(ctx, username)
		if err != nil || user == nil {
			return nil, nil, err
		}
		creds, err := s.credStore.GetWebAuthnCredentialsForUser(ctx, user.ID)
		if err != nil {
			return nil, nil, err
		}
		waUser = &WebAuthnUser{
			User:        user,
			Credentials: creds,
			UserID:      []byte(user.ID),
		}
	}
	var assertion *protocol.CredentialAssertion
	var session *webauthn.SessionData
	var err error
	if waUser != nil {
		assertion, session, err = s.webauthn.BeginLogin(waUser)
	} else {
		assertion, session, err = s.webauthn.BeginDiscoverableLogin()
	}
	if err != nil {
		return nil, nil, err
	}
	ttl := defaultSessionTTL
	if session.Expires.After(time.Now()) {
		ttl = time.Until(session.Expires)
	}
	if saveErr := s.sessionStore.SaveSession(ctx, session.Challenge, session, ttl); saveErr != nil {
		return nil, nil, saveErr
	}
	return assertion, session, nil
}

// sessionDeleter is optionally implemented by SessionStore backends that support explicit deletion.
// It lets FinishAuthentication invalidate a challenge immediately after a successful assertion, so the
// one-time challenge can't be replayed within its TTL (sign-count clone detection doesn't help when an
// authenticator reports a counter of 0).
type sessionDeleter interface {
	DeleteSession(ctx context.Context, challenge string) error
}

// deleteSessionByChallenge best-effort deletes the ceremony session for a challenge. A failed
// delete is logged but not surfaced: the session expires via TTL regardless, and a successful
// authentication shouldn't fail over cleanup.
func (s *Service) deleteSessionByChallenge(ctx context.Context, challenge string) {
	if deleter, ok := s.sessionStore.(sessionDeleter); ok {
		if err := deleter.DeleteSession(ctx, challenge); err != nil {
			s.logger.Error("deleting webauthn ceremony session", err)
		}
	}
}

// FinishAuthenticationResult holds the result of a successful passkey authentication.
type FinishAuthenticationResult struct {
	UserID       string
	CredentialID string
	SignCount    uint32
}

// FinishAuthentication validates the assertion and returns the authenticated user.
func (s *Service) FinishAuthentication(ctx context.Context, username string, req *http.Request) (*FinishAuthenticationResult, error) {
	parsed, err := protocol.ParseCredentialRequestResponse(req)
	if err != nil {
		return nil, err
	}
	session, err := s.sessionStore.GetSession(ctx, parsed.Response.CollectedClientData.Challenge)
	if err != nil || session == nil {
		return nil, err
	}
	var credential *webauthn.Credential
	var user *identity.User
	if username != "" {
		u, userErr := s.userStore.GetUserByUsername(ctx, username)
		if userErr != nil || u == nil {
			return nil, userErr
		}

		creds, credErr := s.credStore.GetWebAuthnCredentialsForUser(ctx, u.ID)
		if credErr != nil {
			return nil, credErr
		}
		waUser := &WebAuthnUser{
			User:            u,
			Credentials:     creds,
			UserID:          []byte(u.ID),
			AssertionCredID: parsed.RawID,
			AssertionFlags:  parsed.Response.AuthenticatorData.Flags,
		}
		credential, err = s.webauthn.ValidateLogin(waUser, *session, parsed)
		if err != nil {
			return nil, err
		}

		user = u
	} else {
		handler := s.discoverableUserHandlerWithParsed(ctx, parsed)
		var waUser webauthn.User
		waUser, credential, err = s.webauthn.ValidatePasskeyLogin(handler, *session, parsed)
		if err != nil {
			return nil, err
		}
		if waUser != nil {
			if wu, ok := waUser.(*WebAuthnUser); ok {
				user = wu.User
			}
		}
	}
	if user == nil || credential == nil {
		if err != nil {
			return nil, err
		}
		return nil, protocol.ErrBadRequest.WithDetails("authentication failed")
	}

	// The assertion validated: delete the one-time challenge/session now so it can't be replayed
	// within its TTL, even if the sign-count bookkeeping below fails.
	s.deleteSessionByChallenge(ctx, parsed.Response.CollectedClientData.Challenge)

	stored, err := s.credStore.GetWebAuthnCredentialByCredentialID(ctx, credential.ID)
	if err != nil || stored == nil {
		return nil, err
	}
	if err = s.credStore.UpdateWebAuthnCredentialSignCount(ctx, stored.ID, credential.Authenticator.SignCount); err != nil {
		return nil, err
	}
	return &FinishAuthenticationResult{
		UserID:       user.ID,
		CredentialID: stored.ID,
		SignCount:    credential.Authenticator.SignCount,
	}, nil
}

// FinishRegistrationFromBytes validates attestation bytes and stores the credential. For gRPC use.
func (s *Service) FinishRegistrationFromBytes(ctx context.Context, userID string, attestationResponse []byte, challenge string) error {
	session, err := s.sessionStore.GetSession(ctx, challenge)
	if err != nil || session == nil {
		return err
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(attestationResponse)
	if err != nil {
		return err
	}
	if parsed.Response.CollectedClientData.Challenge != challenge {
		return protocol.ErrChallengeMismatch
	}
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return err
	}
	creds, err := s.credStore.GetWebAuthnCredentialsForUser(ctx, userID)
	if err != nil {
		return err
	}
	waUser := &WebAuthnUser{User: user, Credentials: creds, UserID: []byte(user.ID)}
	credential, err := s.webauthn.CreateCredential(waUser, *session, parsed)
	if err != nil {
		return err
	}
	transportsJSON := ""
	if len(credential.Transport) > 0 {
		t := make([]string, len(credential.Transport))
		for i, tr := range credential.Transport {
			t[i] = string(tr)
		}
		b, marshalErr := json.Marshal(t)
		if marshalErr != nil {
			return marshalErr
		}
		transportsJSON = string(b)
	}
	_, err = s.credStore.CreateWebAuthnCredential(ctx, &identity.WebAuthnCredentialCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: userID,
		CredentialID:  credential.ID,
		PublicKey:     credential.PublicKey,
		SignCount:     credential.Authenticator.SignCount,
		Transports:    transportsJSON,
		FriendlyName:  "",
	})
	return err
}

// FinishAuthenticationFromBytes validates assertion bytes and returns the authenticated user. For gRPC use.
func (s *Service) FinishAuthenticationFromBytes(ctx context.Context, username string, assertionResponse []byte, challenge string) (*FinishAuthenticationResult, error) {
	session, err := s.sessionStore.GetSession(ctx, challenge)
	if err != nil || session == nil {
		return nil, err
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(assertionResponse)
	if err != nil {
		return nil, err
	}
	if parsed.Response.CollectedClientData.Challenge != challenge {
		return nil, protocol.ErrChallengeMismatch
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", io.NopCloser(bytes.NewReader(assertionResponse)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return s.FinishAuthentication(ctx, username, req)
}

// GetCredentialsForUser returns all active passkey credentials for the given user.
func (s *Service) GetCredentialsForUser(ctx context.Context, userID string) ([]*identity.WebAuthnCredential, error) {
	return s.credStore.GetWebAuthnCredentialsForUser(ctx, userID)
}

// ArchiveCredentialForUser archives a passkey credential only if it belongs to the given user.
func (s *Service) ArchiveCredentialForUser(ctx context.Context, credentialID, userID string) error {
	return s.credStore.ArchiveWebAuthnCredentialForUser(ctx, credentialID, userID)
}

// DiscoverableUserHandler creates a webauthn.DiscoverableUserHandler for passwordless login.
func (s *Service) DiscoverableUserHandler(ctx context.Context) webauthn.DiscoverableUserHandler {
	return s.discoverableUserHandlerWithParsed(ctx, nil)
}

// discoverableUserHandlerWithParsed creates a handler that returns WebAuthnUser with assertion flags
// when parsed is non-nil, satisfying go-webauthn's BackupEligible consistency check.
func (s *Service) discoverableUserHandlerWithParsed(ctx context.Context, parsed *protocol.ParsedCredentialAssertionData) webauthn.DiscoverableUserHandler {
	return func(rawID, userHandle []byte) (webauthn.User, error) {
		stored, err := s.credStore.GetWebAuthnCredentialByCredentialID(ctx, rawID)
		if err != nil || stored == nil {
			return nil, err
		}
		user, err := s.userStore.GetUserByID(ctx, stored.BelongsToUser)
		if err != nil || user == nil {
			return nil, err
		}
		creds, err := s.credStore.GetWebAuthnCredentialsForUser(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		waUser := &WebAuthnUser{
			User:        user,
			Credentials: creds,
			UserID:      []byte(user.ID),
		}
		if parsed != nil {
			waUser.AssertionCredID = parsed.RawID
			waUser.AssertionFlags = parsed.Response.AuthenticatorData.Flags
		}
		return waUser, nil
	}
}
