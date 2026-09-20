package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"

	analyticsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"
	authgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	dataprivacygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	internalopsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	auditgrpc "github.com/primandproper/platform-go/v14/audit/auditpb"
	oauth2clientsgrpc "github.com/primandproper/platform-go/v14/authentication/oauth2clients/oauth2clientspb"
	paymentsgrpc "github.com/primandproper/platform-go/v14/billing/billingpb"
	commentsgrpc "github.com/primandproper/platform-go/v14/comments/commentspb"
	identitygrpc "github.com/primandproper/platform-go/v14/identity/identitypb"
	issuereportsgrpc "github.com/primandproper/platform-go/v14/issuereports/issuereportspb"
	notificationsgrpc "github.com/primandproper/platform-go/v14/notifications/notificationspb"
	settingsgrpc "github.com/primandproper/platform-go/v14/settings/settingspb"
	waitlistsgrpc "github.com/primandproper/platform-go/v14/waitlists/waitlistspb"
	webhooksgrpc "github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/httpclient"
	"github.com/primandproper/primitives-go/v2/random"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
)

const (
	zuckModeUserHeader    = "X-Zuck-Mode-User"
	zuckModeAccountHeader = "X-Zuck-Mode-Account"
)

type Client interface {
	analyticsgrpc.AnalyticsServiceClient
	authgrpc.AuthServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	mealplanninggrpc.MealPlanningServiceClient
	notificationsgrpc.NotificationsServiceClient
	oauth2clientsgrpc.OAuth2ClientsServiceClient
	paymentsgrpc.BillingServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient

	// IdentityService returns the directory client: users, accounts, memberships and
	// invitations.
	//
	// It is an accessor rather than an embed because identity and waitlists both name an
	// RPC Invite — an offer of membership in an account on one, an offer of a place off a
	// waiting list on the other — and a type embedding both has an ambiguous selector and
	// does not compile. Identity is the one moved because its RPCs are platform's and
	// reaching for them deliberately is no worse than reaching for the comments client
	// below, which is here for the same reason.
	IdentityService() identitygrpc.IdentityServiceClient

	// CommentsService returns the standalone CommentsService client. Use this to call
	// CommentsService RPCs directly instead of via MealPlanningService.
	CommentsService() commentsgrpc.CommentsServiceClient

	// WebhooksService returns the webhooks client.
	//
	// It is an accessor rather than an embed because billing and webhooks both
	// name three RPCs Subscription — a paid plan on one, an endpoint's interest in
	// an event type on the other — and a type embedding both has three ambiguous
	// selectors and does not compile.
	WebhooksService() webhooksgrpc.WebhooksServiceClient

	// Close releases the underlying gRPC connection. Callers that build clients repeatedly
	// must call this to avoid leaking connections.
	Close() error
}

type client struct {
	analyticsgrpc.AnalyticsServiceClient
	authgrpc.AuthServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	mealplanninggrpc.MealPlanningServiceClient
	notificationsgrpc.NotificationsServiceClient
	oauth2clientsgrpc.OAuth2ClientsServiceClient
	paymentsgrpc.BillingServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient

	identityClient identitygrpc.IdentityServiceClient
	commentsClient commentsgrpc.CommentsServiceClient
	webhooksClient webhooksgrpc.WebhooksServiceClient
	conn           *grpc.ClientConn
}

// BuildClient builds a new Client.
func BuildClient(grpcServerAddress string, opts ...grpc.DialOption) (Client, error) {
	conn, err := grpc.NewClient(grpcServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("building grpc client: %w", err)
	}

	c := &client{
		AnalyticsServiceClient:     analyticsgrpc.NewAnalyticsServiceClient(conn),
		AuthServiceClient:          authgrpc.NewAuthServiceClient(conn),
		AuditServiceClient:         auditgrpc.NewAuditServiceClient(conn),
		DataPrivacyServiceClient:   dataprivacygrpc.NewDataPrivacyServiceClient(conn),
		InternalOperationsClient:   internalopsgrpc.NewInternalOperationsClient(conn),
		IssueReportsServiceClient:  issuereportsgrpc.NewIssueReportsServiceClient(conn),
		MealPlanningServiceClient:  mealplanninggrpc.NewMealPlanningServiceClient(conn),
		NotificationsServiceClient: notificationsgrpc.NewNotificationsServiceClient(conn),
		OAuth2ClientsServiceClient: oauth2clientsgrpc.NewOAuth2ClientsServiceClient(conn),
		BillingServiceClient:       paymentsgrpc.NewBillingServiceClient(conn),
		SettingsServiceClient:      settingsgrpc.NewSettingsServiceClient(conn),
		UploadedMediaServiceClient: uploadedmediagrpc.NewUploadedMediaServiceClient(conn),
		WaitlistsServiceClient:     waitlistsgrpc.NewWaitlistsServiceClient(conn),
		identityClient:             identitygrpc.NewIdentityServiceClient(conn),
		commentsClient:             commentsgrpc.NewCommentsServiceClient(conn),
		webhooksClient:             webhooksgrpc.NewWebhooksServiceClient(conn),
		conn:                       conn,
	}

	return c, nil
}

func (c *client) IdentityService() identitygrpc.IdentityServiceClient {
	return c.identityClient
}

func (c *client) CommentsService() commentsgrpc.CommentsServiceClient {
	return c.commentsClient
}

// WebhooksService returns the webhooks client. See the interface for why it is
// not embedded.
func (c *client) WebhooksService() webhooksgrpc.WebhooksServiceClient {
	return c.webhooksClient
}

// Close releases the underlying gRPC connection.
func (c *client) Close() error {
	if c.conn == nil {
		return nil
	}

	return c.conn.Close()
}

// BuildUnauthenticatedGRPCClient connects without TLS or auth tokens.
// Use only for plaintext backends (e.g. kubectl port-forward).
func BuildUnauthenticatedGRPCClient(grpcServerAddr string, opts ...grpc.DialOption) (Client, error) {
	return BuildClient(grpcServerAddr, append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, opts...)...)
}

// BuildTLSGRPCClient connects with TLS but no auth tokens.
// Suitable for reaching a TLS-enabled gRPC server (e.g. api.dinnerdonebetter.com:443)
// without supplying OAuth2 credentials.
func BuildTLSGRPCClient(grpcServerAddr string, opts ...grpc.DialOption) (Client, error) {
	return BuildClient(grpcServerAddr, append([]grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))}, opts...)...)
}

func WithOAuth2Credentials(
	ctx context.Context,
	authServerAddress,
	clientID,
	clientSecret,
	authToken string,
) (grpc.DialOption, error) {
	state, err := random.GenerateBase64EncodedString(ctx, 32)
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  authServerAddress,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   authServerAddress + oauth2server.PathAuthorize,
			TokenURL:  authServerAddress + oauth2server.PathToken,
		},
	}

	// PKCE (RFC 7636) with the S256 challenge method.
	pkceVerifier := oauth2.GenerateVerifier()

	authCodeURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(pkceVerifier),
	)

	// POST rather than GET: the authorization server renders a login form on GET — the answer
	// for a browser arriving without a session — and runs the authenticator that reads this
	// bearer token only on POST. The authorization parameters stay in the query string on both,
	// so the request that issues the code is the one that was validated.
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		authCodeURL,
		http.NoBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build oauth2 code retrieval request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c, err := httpclient.NewHTTPClient(httpclient.WithTracing(true))
	if err != nil {
		return nil, fmt.Errorf("failed to build oauth2 code retrieval client: %w", err)
	}

	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 code: %w", err)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			log.Println("failed to close oauth2 response body", err)
		}
	}()

	const (
		codeKey = "code"
	)

	rl, err := res.Location()
	if err != nil {
		return nil, err
	}

	code := rl.Query().Get(codeKey)
	if code == "" {
		return nil, errors.New("code not returned from oauth2 redirect")
	}

	oauth2Token, err := oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(pkceVerifier))
	if err != nil {
		return nil, err
	}

	ts := oauth2.ReuseTokenSource(oauth2Token, oauth2Config.TokenSource(ctx, oauth2Token))

	return grpc.WithPerRPCCredentials(oauth.TokenSource{
		TokenSource: ts,
	}), nil
}

func ImpersonateUseAndAccountContext(ctx context.Context, userID, accountID string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		zuckModeUserHeader:    userID,
		zuckModeAccountHeader: accountID,
	}))
}

// bearerTokenCredential adds a static Bearer token to each RPC.
type bearerTokenCredential struct {
	token string
}

func (b bearerTokenCredential) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + b.token,
	}, nil
}

func (bearerTokenCredential) RequireTransportSecurity() bool {
	return false
}

// WithBearerTokenCredentials returns a DialOption that attaches the given token as a Bearer token to every RPC.
// Use this when the token is a JWT (e.g. from LoginForToken) that should be sent directly without OAuth2 exchange.
func WithBearerTokenCredentials(token string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(bearerTokenCredential{token: token})
}

// BuildUnauthenticatedGRPCClientWithBearerToken connects with a Bearer token (e.g. JWT from LoginForToken).
// Use this when the token includes an account_id claim and you want the session to use that account.
func BuildUnauthenticatedGRPCClientWithBearerToken(grpcServerAddr, token string) (Client, error) {
	return BuildClient(grpcServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		WithBearerTokenCredentials(token),
	)
}
