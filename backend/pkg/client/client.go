package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"

	analyticsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"
	auditgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	authgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	commentsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	dataprivacygrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	identitygrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	internalopsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	issuereportsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	mealplanninggrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	notificationsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	oauthgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
	paymentsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	settingsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	uploadedmediagrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	waitlistsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	webhooksgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	"github.com/primandproper/platform-go/v5/httpclient"
	"github.com/primandproper/platform-go/v5/random"

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
	identitygrpc.IdentityServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	mealplanninggrpc.MealPlanningServiceClient
	notificationsgrpc.UserNotificationsServiceClient
	oauthgrpc.OAuthServiceClient
	paymentsgrpc.PaymentsServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient
	webhooksgrpc.WebhooksServiceClient

	// CommentsService returns the standalone CommentsService client. Use this to call
	// CommentsService RPCs directly instead of via MealPlanningService.
	CommentsService() commentsgrpc.CommentsServiceClient

	// Close releases the underlying gRPC connection. Callers that build clients repeatedly
	// must call this to avoid leaking connections.
	Close() error
}

type client struct {
	analyticsgrpc.AnalyticsServiceClient
	authgrpc.AuthServiceClient
	identitygrpc.IdentityServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	mealplanninggrpc.MealPlanningServiceClient
	notificationsgrpc.UserNotificationsServiceClient
	oauthgrpc.OAuthServiceClient
	paymentsgrpc.PaymentsServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient
	webhooksgrpc.WebhooksServiceClient

	commentsClient commentsgrpc.CommentsServiceClient
	conn           *grpc.ClientConn
}

// BuildClient builds a new Client.
func BuildClient(grpcServerAddress string, opts ...grpc.DialOption) (Client, error) {
	conn, err := grpc.NewClient(grpcServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("building grpc client: %w", err)
	}

	c := &client{
		AnalyticsServiceClient:         analyticsgrpc.NewAnalyticsServiceClient(conn),
		AuthServiceClient:              authgrpc.NewAuthServiceClient(conn),
		IdentityServiceClient:          identitygrpc.NewIdentityServiceClient(conn),
		AuditServiceClient:             auditgrpc.NewAuditServiceClient(conn),
		DataPrivacyServiceClient:       dataprivacygrpc.NewDataPrivacyServiceClient(conn),
		InternalOperationsClient:       internalopsgrpc.NewInternalOperationsClient(conn),
		IssueReportsServiceClient:      issuereportsgrpc.NewIssueReportsServiceClient(conn),
		MealPlanningServiceClient:      mealplanninggrpc.NewMealPlanningServiceClient(conn),
		UserNotificationsServiceClient: notificationsgrpc.NewUserNotificationsServiceClient(conn),
		OAuthServiceClient:             oauthgrpc.NewOAuthServiceClient(conn),
		PaymentsServiceClient:          paymentsgrpc.NewPaymentsServiceClient(conn),
		SettingsServiceClient:          settingsgrpc.NewSettingsServiceClient(conn),
		UploadedMediaServiceClient:     uploadedmediagrpc.NewUploadedMediaServiceClient(conn),
		WaitlistsServiceClient:         waitlistsgrpc.NewWaitlistsServiceClient(conn),
		WebhooksServiceClient:          webhooksgrpc.NewWebhooksServiceClient(conn),
		commentsClient:                 commentsgrpc.NewCommentsServiceClient(conn),
		conn:                           conn,
	}

	return c, nil
}

func (c *client) CommentsService() commentsgrpc.CommentsServiceClient {
	return c.commentsClient
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
			AuthURL:   authServerAddress + "/oauth2/authorize",
			TokenURL:  authServerAddress + "/oauth2/token",
		},
	}

	authCodeURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "plain"),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		authCodeURL,
		http.NoBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build oauth2 code retrieval request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	c := httpclient.NewHTTPClient(&httpclient.Config{EnableTracing: true})
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := c.Do(req) //nolint:gosec // G704: authCodeURL from OAuth config (authServerAddress), not user-controlled
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

	oauth2Token, err := oauth2Config.Exchange(ctx, code)
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
