package webauthn

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformwebauthn "github.com/primandproper/platform-go/v13/authentication/webauthn"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/go-webauthn/webauthn/protocol"
)

const o11yName = "webauthn_service"

var (
	// ErrNilRelyingParty indicates NewService was called without a relying party.
	ErrNilRelyingParty = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil webauthn relying party")

	// ErrUserNotFound indicates a ceremony named a user this service has no record of.
	ErrUserNotFound = platformerrors.New("webauthn user not found")

	// ErrCredentialNotFound indicates an assertion verified against a credential that is
	// no longer stored. The sign count has nowhere to be written back to, and a login that
	// cannot record the count it was answered with is a login clone detection cannot see.
	ErrCredentialNotFound = platformerrors.New("webauthn credential not found")
)

type (
	// UserStore provides user lookup for WebAuthn.
	UserStore interface {
		GetUserByID(ctx context.Context, userID string) (*identity.User, error)
		GetUserByUsername(ctx context.Context, username string) (*identity.User, error)
	}

	// Service is passkey registration and login for this application's users.
	//
	// The ceremony belongs to platform's RelyingParty: the challenge, the durable one-shot
	// store it lives in, and the single deadline that bounds it. What is here is the half
	// that package deliberately does not name — which user a ceremony is for, where the
	// credential it produces is stored, and the sign count written back afterwards.
	Service struct {
		_ struct{} `json:"-"`

		logger       logging.Logger
		tracer       tracing.Tracer
		relyingParty *platformwebauthn.RelyingParty
		credStore    identity.WebAuthnCredentialDataManager
		userStore    UserStore
	}

	// FinishAuthenticationResult holds the result of a successful passkey authentication.
	FinishAuthenticationResult struct {
		_ struct{} `json:"-"`

		UserID       string
		CredentialID string
		SignCount    uint32
	}
)

// NewService creates a new WebAuthn service over an already-built relying party.
func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	relyingParty *platformwebauthn.RelyingParty,
	credStore identity.WebAuthnCredentialDataManager,
	userStore UserStore,
) (*Service, error) {
	if relyingParty == nil {
		return nil, ErrNilRelyingParty
	}

	return &Service{
		logger:       logging.NewNamedLogger(logger, o11yName),
		tracer:       tracing.NewNamedTracer(tracerProvider, o11yName),
		relyingParty: relyingParty,
		credStore:    credStore,
		userStore:    userStore,
	}, nil
}

// BeginRegistrationOptions returns PublicKeyCredentialCreationOptions for the given user,
// alongside the challenge the client echoes back to FinishRegistration.
//
// The challenge is read off the options rather than out of the ceremony state, which is not
// returned and is not the caller's to hold: the relying party stored it, and FinishRegistration
// reads it back from the challenge the client answers with.
func (s *Service) BeginRegistrationOptions(ctx context.Context, userID string) (*protocol.CredentialCreation, string, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, userID)

	user, err := s.webAuthnUserByID(ctx, userID)
	if err != nil {
		return nil, "", observability.PrepareAndLogError(err, logger, span, "assembling webauthn user")
	}

	creation, err := s.relyingParty.BeginRegistration(ctx, user)
	if err != nil {
		return nil, "", observability.PrepareAndLogError(err, logger, span, "beginning passkey registration")
	}

	return creation, creation.Response.Challenge.String(), nil
}

// FinishRegistration validates an attestation response and stores the credential it produced.
func (s *Service) FinishRegistration(ctx context.Context, userID string, attestationResponse []byte, challenge string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, userID)

	// Parsed here only to check the challenge the caller passed against the one the client
	// actually answered. The ceremony state is looked up by the latter inside the relying
	// party, so a mismatch is a client bug rather than an attack — but the request field is
	// required by the API, and honoring a value we asked for is cheaper than explaining why
	// we ignore it.
	parsed, err := protocol.ParseCredentialCreationResponseBytes(attestationResponse)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "parsing passkey attestation response")
	}

	if parsed.Response.CollectedClientData.Challenge != challenge {
		return observability.PrepareAndLogError(protocol.ErrChallengeMismatch, logger, span, "checking passkey attestation challenge")
	}

	user, err := s.webAuthnUserByID(ctx, userID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "assembling webauthn user")
	}

	credential, err := s.relyingParty.FinishRegistrationBody(ctx, user, bytes.NewReader(attestationResponse))
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "finishing passkey registration")
	}

	transports, err := encodeTransports(credential.Transport)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "encoding passkey transports")
	}

	if _, err = s.credStore.CreateWebAuthnCredential(ctx, &identity.WebAuthnCredentialCreationInput{
		ID:            identifiers.New(),
		BelongsToUser: userID,
		CredentialID:  credential.ID,
		PublicKey:     credential.PublicKey,
		SignCount:     credential.Authenticator.SignCount,
		Transports:    transports,
		FriendlyName:  "",
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "storing passkey credential")
	}

	return nil
}

// BeginAuthenticationOptions returns PublicKeyCredentialRequestOptions for the given username,
// alongside the challenge the client echoes back to FinishAuthentication. An empty username
// begins a discoverable login, where the passkey names the user.
func (s *Service) BeginAuthenticationOptions(ctx context.Context, username string) (*protocol.CredentialAssertion, string, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(identitykeys.UsernameKey, username)

	var (
		assertion *protocol.CredentialAssertion
		err       error
	)

	if username == "" {
		assertion, err = s.relyingParty.BeginDiscoverableLogin(ctx)
	} else {
		var user *WebAuthnUser
		if user, err = s.webAuthnUserByUsername(ctx, username); err != nil {
			return nil, "", observability.PrepareAndLogError(err, logger, span, "assembling webauthn user")
		}

		assertion, err = s.relyingParty.BeginLogin(ctx, user)
	}

	if err != nil {
		return nil, "", observability.PrepareAndLogError(err, logger, span, "beginning passkey authentication")
	}

	return assertion, assertion.Response.Challenge.String(), nil
}

// FinishAuthentication validates an assertion response and returns the authenticated user.
//
// The ceremony state is consumed inside the relying party, in one operation, so an assertion
// replayed inside its TTL finds nothing the second time. Nothing here needs to remember to
// delete it, and there is no window between the read and the removal for a second caller to
// be handed the same challenge.
func (s *Service) FinishAuthentication(ctx context.Context, username string, assertionResponse []byte, challenge string) (*FinishAuthenticationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(identitykeys.UsernameKey, username)

	// Parsed up front for two reasons: the challenge check below, and the authenticator flags
	// the user adapter needs to satisfy go-webauthn's BackupEligible consistency rule. The
	// relying party parses the body again for the verification itself.
	parsed, err := protocol.ParseCredentialRequestResponseBytes(assertionResponse)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "parsing passkey assertion response")
	}

	if parsed.Response.CollectedClientData.Challenge != challenge {
		return nil, observability.PrepareAndLogError(protocol.ErrChallengeMismatch, logger, span, "checking passkey assertion challenge")
	}

	var (
		user       *identity.User
		credential *platformwebauthn.Credential
	)

	if username == "" {
		var resolved platformwebauthn.User
		if resolved, credential, err = s.relyingParty.FinishDiscoverableLoginBody(
			ctx,
			s.discoverableUserHandler(ctx, parsed),
			bytes.NewReader(assertionResponse),
		); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "finishing discoverable passkey authentication")
		}

		if resolvedUser, ok := resolved.(*WebAuthnUser); ok {
			user = resolvedUser.User
		}
	} else {
		var waUser *WebAuthnUser
		if waUser, err = s.webAuthnUserByUsername(ctx, username); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "assembling webauthn user")
		}

		waUser.AssertionCredID = parsed.RawID
		waUser.AssertionFlags = parsed.Response.AuthenticatorData.Flags

		if credential, err = s.relyingParty.FinishLoginBody(ctx, waUser, bytes.NewReader(assertionResponse)); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "finishing passkey authentication")
		}

		user = waUser.User
	}

	if user == nil || credential == nil {
		return nil, observability.PrepareAndLogError(ErrUserNotFound, logger, span, "resolving passkey owner")
	}

	stored, err := s.credStore.GetWebAuthnCredentialByCredentialID(ctx, credential.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching passkey credential")
	}

	if stored == nil {
		return nil, observability.PrepareAndLogError(ErrCredentialNotFound, logger, span, "fetching passkey credential")
	}

	// Surfaced rather than swallowed. A sign count that is not written back is a count the
	// next login compares against a stale value, which is clone detection that reports
	// nothing — so a login whose bookkeeping failed is a login that did not happen.
	if err = s.credStore.UpdateWebAuthnCredentialSignCount(ctx, stored.ID, credential.Authenticator.SignCount); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "recording passkey sign count")
	}

	return &FinishAuthenticationResult{
		UserID:       user.ID,
		CredentialID: stored.ID,
		SignCount:    credential.Authenticator.SignCount,
	}, nil
}

// GetCredentialsForUser returns all active passkey credentials for the given user.
func (s *Service) GetCredentialsForUser(ctx context.Context, userID string) ([]*identity.WebAuthnCredential, error) {
	return s.credStore.GetWebAuthnCredentialsForUser(ctx, userID)
}

// ArchiveCredentialForUser archives a passkey credential only if it belongs to the given user.
func (s *Service) ArchiveCredentialForUser(ctx context.Context, credentialID, userID string) error {
	return s.credStore.ArchiveWebAuthnCredentialForUser(ctx, credentialID, userID)
}

// discoverableUserHandler resolves the user behind a credential during a discoverable login.
// The parsed assertion supplies the authenticator flags the adapter needs; it is never nil on
// the path this is called from.
func (s *Service) discoverableUserHandler(ctx context.Context, parsed *protocol.ParsedCredentialAssertionData) platformwebauthn.DiscoverableUserHandler {
	return func(rawID, _ []byte) (platformwebauthn.User, error) {
		stored, err := s.credStore.GetWebAuthnCredentialByCredentialID(ctx, rawID)
		if err != nil {
			return nil, err
		}

		if stored == nil {
			return nil, ErrCredentialNotFound
		}

		user, err := s.webAuthnUserByID(ctx, stored.BelongsToUser)
		if err != nil {
			return nil, err
		}

		user.AssertionCredID = parsed.RawID
		user.AssertionFlags = parsed.Response.AuthenticatorData.Flags

		return user, nil
	}
}

// webAuthnUserByID assembles the go-webauthn user for a user ID.
func (s *Service) webAuthnUserByID(ctx context.Context, userID string) (*WebAuthnUser, error) {
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.webAuthnUser(ctx, user)
}

// webAuthnUserByUsername assembles the go-webauthn user for a username.
func (s *Service) webAuthnUserByUsername(ctx context.Context, username string) (*WebAuthnUser, error) {
	user, err := s.userStore.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return s.webAuthnUser(ctx, user)
}

// webAuthnUser hangs a user's registered credentials off the adapter.
//
// A missing user is an error rather than a nil user handed onward. The relying party would
// reject the nil, but it would reject it as "nil webauthn user", which says nothing about
// the lookup that came up empty.
func (s *Service) webAuthnUser(ctx context.Context, user *identity.User) (*WebAuthnUser, error) {
	if user == nil {
		return nil, ErrUserNotFound
	}

	credentials, err := s.credStore.GetWebAuthnCredentialsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &WebAuthnUser{
		User:        user,
		Credentials: credentials,
		UserID:      []byte(user.ID),
	}, nil
}

// encodeTransports renders an authenticator's transports as the JSON array the credential
// table stores. No transports is an empty column rather than a "null" literal.
func encodeTransports(transports []protocol.AuthenticatorTransport) (string, error) {
	if len(transports) == 0 {
		return "", nil
	}

	names := make([]string, len(transports))
	for i, transport := range transports {
		names[i] = string(transport)
	}

	encoded, err := json.Marshal(names)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}
