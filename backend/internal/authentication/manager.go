package authentication

import (
	"context"
	"strings"
	"time"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/sessions"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	name = "authentication_manager"
)

type (
	// LoginMetadata holds request metadata for session tracking.
	LoginMetadata struct {
		ClientIP  string
		UserAgent string
	}

	Manager interface {
		ProcessLogin(ctx context.Context, adminOnly bool, loginData *auth.UserLoginInput, meta *LoginMetadata) (*auth.TokenResponse, error)
		ProcessPasskeyLogin(ctx context.Context, userID, desiredAccountID string, meta *LoginMetadata) (*auth.TokenResponse, error)
		ExchangeTokenForUser(ctx context.Context, refreshToken, desiredAccountID string) (*auth.TokenResponse, error)
	}

	manager struct {
		tokenIssuer             tokens.Issuer
		signIn                  *signin.Service
		tracer                  tracing.Tracer
		logger                  logging.Logger
		dataChangesPublisher    messagequeue.Publisher
		directory               platformidentity.SignInReader
		db                      database.Client
		sessionStore            auth.SessionStore
		maxAccessTokenLifetime  time.Duration
		maxRefreshTokenLifetime time.Duration
	}
)

func NewManager(
	ctx context.Context,
	queuesConfig *queuescfg.Config,
	tokenIssuer tokens.Issuer,
	signInService *signin.Service,
	tracingProvider tracing.Provider,
	logger logging.Logger,
	publisherProvider messagequeue.PublisherProvider,
	directory platformidentity.SignInReader,
	db database.Client,
	sessionStore auth.SessionStore,
	cfg *authcfg.TokensConfig,
) (Manager, error) {
	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, queuesConfig.DataChangesTopicName)
	if err != nil {
		return nil, observability.PrepareError(err, nil, "creating data changes publisher")
	}

	m := &manager{
		maxRefreshTokenLifetime: cfg.MaxRefreshTokenLifetime,
		maxAccessTokenLifetime:  cfg.MaxAccessTokenLifetime,
		tracer:                  tracing.NewNamedTracer(tracingProvider, name),
		logger:                  logging.NewNamedLogger(logger, name),
		tokenIssuer:             tokenIssuer,
		signIn:                  signInService,
		dataChangesPublisher:    dataChangesPublisher,
		directory:               directory,
		db:                      db,
		sessionStore:            sessionStore,
	}

	return m, nil
}

func (m *manager) ProcessLogin(ctx context.Context, adminOnly bool, loginData *auth.UserLoginInput, meta *LoginMetadata) (*auth.TokenResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.Clone()

	if err := loginData.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	logger = logger.WithValue(identitykeys.UsernameKey, loginData.Username)

	// The handle is not trimmed and the two secrets are, which is what this replaced did
	// and is worth saying out loud now that one call does all three. The lookup ran on the
	// username as given; the trims happened afterwards, so they only ever reached the
	// password comparison and the TOTP check. Trimming the handle here would make
	// " alice" sign in where it used to fail, and untrimming the code would refuse a
	// pasted one — both are changes, and neither belongs in an adoption.
	credentials := &signin.Credentials{
		Username:        loginData.Username,
		Password:        strings.TrimSpace(loginData.Password),
		TOTPCode:        strings.TrimSpace(loginData.TOTPToken),
		ActiveAccountID: loginData.DesiredAccountID,
	}

	// One call where there were five: the handle read, the password comparison, the
	// status check, the second factor and the principal resolution. platform does them in
	// that order for reasons this application had not thought about and now inherits —
	// the status is checked after the password, so somebody who cannot prove the password
	// cannot learn whether an account exists, is suspended or was terminated; and a
	// handle naming nobody still costs a password hash, so a stopwatch cannot tell the
	// two apart either. This code returned on an unknown handle without hashing anything.
	//
	// The administrative door is the same call through AdminAuthenticate, which adds the
	// service role check and demands a proven second factor whatever the policy says.
	// That is the rule this package enforced by hand, minus the hand.
	authenticate := m.signIn.Authenticate
	if adminOnly {
		authenticate = m.signIn.AdminAuthenticate
	}

	principal, err := authenticate(ctx, ddbidentity.Scope(), credentials)
	if err != nil {
		return nil, observability.PrepareError(err, span, "authenticating")
	}

	user, accountID := principal.User, principal.ActiveAccountID

	logger = logger.WithValue(identitykeys.UserIDKey, user.ID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)
	logger.Debug("login validated")

	response, err := m.issueTokensWithSession(ctx, user, accountID, auth.LoginMethodPassword, meta)
	if err != nil {
		return nil, observability.PrepareError(err, span, "issuing tokens with session")
	}

	dcm := &audit.DataChangeMessage{
		EventType: ddbidentity.UserLoggedInServiceEventType,
		AccountID: accountID,
		UserID:    user.ID,
	}

	if err = m.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		return nil, observability.PrepareError(err, span, "publishing data change")
	}

	return response, nil
}

// ProcessPasskeyLogin issues tokens for a user authenticated via passkey.
func (m *manager) ProcessPasskeyLogin(ctx context.Context, userID, desiredAccountID string, meta *LoginMetadata) (*auth.TokenResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	principal, err := m.signInPrincipal(ctx, userID, desiredAccountID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "resolving the principal signing in")
	}

	user, accountID := principal.User, principal.ActiveAccountID

	response, err := m.issueTokensWithSession(ctx, user, accountID, auth.LoginMethodPasskey, meta)
	if err != nil {
		return nil, observability.PrepareError(err, span, "issuing tokens with session")
	}

	dcm := &audit.DataChangeMessage{
		EventType: ddbidentity.UserLoggedInServiceEventType,
		AccountID: accountID,
		UserID:    user.ID,
	}

	if err = m.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		return nil, observability.PrepareError(err, span, "publishing data change")
	}

	return response, nil
}

func (m *manager) ExchangeTokenForUser(ctx context.Context, refreshToken, desiredAccountID string) (*auth.TokenResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.Clone()

	claims, err := m.tokenIssuer.ParseToken(ctx, refreshToken)
	if err != nil {
		return nil, observability.PrepareError(err, span, "parsing userID from token")
	}
	userID := claims.Subject()

	principal, err := m.signInPrincipal(ctx, userID, desiredAccountID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "resolving the principal exchanging a token")
	}

	user := principal.User

	logger = logger.WithValue(identitykeys.UserIDKey, user.ID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, user.ID)

	// Validate the session the refresh token names, and that this is the refresh token
	// it was last issued with. The second half is the rotation: a refresh token spent
	// once is superseded by the one this call is about to issue, and presenting the old
	// one afterwards has to fail even though the session it names is perfectly live.
	sessionID, _ := claims.GetString("sid")

	session, err := m.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "reading session")
	}
	if session.Data == nil || session.Data.RefreshTokenID != claims.JTI() {
		return nil, observability.PrepareError(ErrSessionSuperseded, span, "validating session")
	}

	accountID := principal.ActiveAccountID

	// Issue new tokens against the same session. The identifier is not rotated: it is
	// not a credential a client ever holds on its own — it rides inside a token this
	// server signs — so there is no identifier an attacker could have planted for the
	// rotation to invalidate. What does rotate is the pair of JTIs below, which is what
	// retires the tokens this call was made with.
	var accessJTI, refreshJTINew string
	response := &auth.TokenResponse{
		UserID:     user.ID,
		AccountID:  accountID,
		ExpiresUTC: time.Now().Add(m.maxAccessTokenLifetime).UTC(),
	}

	extraClaims := tokenClaims(accountID, sessionID)

	response.AccessToken, accessJTI, err = m.tokenIssuer.IssueToken(ctx, user.ID, m.maxAccessTokenLifetime, extraClaims)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating access token")
	}

	response.RefreshToken, refreshJTINew, err = m.tokenIssuer.IssueToken(ctx, user.ID, m.maxRefreshTokenLifetime, extraClaims)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating refresh token")
	}

	// Recorded before the tokens are handed back, and a failure here fails the refresh.
	// The session is what says which pair is current, so a write that did not land is a
	// pair that will not authenticate — better to answer that now than to return two
	// tokens that are already dead.
	if err = m.sessionStore.Save(ctx, sessionID, &auth.SessionPayload{
		SessionTokenID: accessJTI,
		RefreshTokenID: refreshJTINew,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "recording rotated session tokens")
	}

	dcm := &audit.DataChangeMessage{
		EventType: ddbidentity.UserLoggedInServiceEventType,
		AccountID: accountID,
		UserID:    user.ID,
	}

	if err = m.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		return nil, observability.PrepareError(err, span, "publishing data change")
	}

	return response, nil
}

// issueTokensWithSession establishes a session and issues tokens with its identifier
// embedded.
//
// The session comes first because its identifier is minted by the store rather than here,
// and the tokens have to carry it. That order costs a second write — the session is
// established, then saved again once the two JTIs exist — and the alternative would be
// minting the identifier locally, which is how the store would end up not owning the one
// thing it is the authority on.
//
// A session that cannot be established fails the login. The version this replaced logged
// and carried on, which handed the user two tokens naming a session that was not there:
// they authenticated with them exactly zero times, and the failure surfaced as a sign-in
// that appeared to work and then did not.
func (m *manager) issueTokensWithSession(ctx context.Context, user *platformidentity.User, accountID, loginMethod string, meta *LoginMetadata) (*auth.TokenResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	var clientIP, userAgent string
	if meta != nil {
		clientIP = meta.ClientIP
		userAgent = meta.UserAgent
	}

	session, err := m.sessionStore.NewFor(
		ctx,
		auth.SessionHolder(user.ID),
		sessions.Metadata{
			DeviceName:  deriveDeviceName(userAgent),
			IPAddress:   clientIP,
			UserAgent:   userAgent,
			LoginMethod: loginMethod,
		},
		&auth.SessionPayload{},
	)
	if err != nil {
		return nil, observability.PrepareError(err, span, "establishing session")
	}

	response := &auth.TokenResponse{
		UserID:     user.ID,
		AccountID:  accountID,
		ExpiresUTC: time.Now().Add(m.maxAccessTokenLifetime).UTC(),
	}

	var accessJTI, refreshJTI string
	extraClaims := tokenClaims(accountID, session.ID)

	response.AccessToken, accessJTI, err = m.tokenIssuer.IssueToken(ctx, user.ID, m.maxAccessTokenLifetime, extraClaims)
	if err != nil {
		return nil, observability.PrepareError(err, span, "creating access token")
	}

	response.RefreshToken, refreshJTI, err = m.tokenIssuer.IssueToken(ctx, user.ID, m.maxRefreshTokenLifetime, extraClaims)
	if err != nil {
		return nil, observability.PrepareError(err, span, "creating refresh token")
	}

	if err = m.sessionStore.Save(ctx, session.ID, &auth.SessionPayload{
		SessionTokenID: accessJTI,
		RefreshTokenID: refreshJTI,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "recording session tokens")
	}

	return response, nil
}

// tokenClaims builds the extraClaims map passed to tokens.Issuer.IssueToken. Empty
// values are always included so issued tokens have a stable shape; parsers tolerate
// empty string values for these optional claims.
func tokenClaims(accountID, sessionID string) map[string]any {
	return map[string]any{
		"account_id": accountID,
		"sid":        sessionID,
	}
}

// deriveDeviceName produces a simple friendly device name from a User-Agent string.
func deriveDeviceName(userAgent string) string {
	if userAgent == "" {
		return "Unknown Device"
	}

	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "android"):
		return "Android Device"
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		return "Mac"
	case strings.Contains(ua, "windows"):
		return "Windows PC"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Unknown Device"
	}
}

// signInPrincipal resolves who is signing in and which account they land in.
//
// One read where there were three, and it is the read rather than this function that
// makes the three checks: platform's GetPrincipal refuses a user whose account status does
// not admit signing in, refuses a named account the user is not a live member of, and
// falls back to their default when none was named. Each of those used to be a separate
// call here, and the ban check used to be a separate call in three places — so a path that
// forgot one was a path that signed somebody in anyway.
//
// A user who belongs to no account gets a principal with an empty ActiveAccountID rather
// than an error. That is a state and not a failure: somebody has to be able to sign in
// before anybody has put them anywhere.
func (m *manager) signInPrincipal(ctx context.Context, userID, desiredAccountID string) (*platformidentity.Principal, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	principal, err := m.directory.GetPrincipal(ctx, m.db.Reader(), ddbidentity.Scope(), userID, desiredAccountID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "reading principal")
	}

	return principal, nil
}
