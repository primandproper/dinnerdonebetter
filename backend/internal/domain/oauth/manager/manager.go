package manager

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/converters"
	oauthkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/keys"

	perrors "github.com/primandproper/platform-go/v5/errors"
	errorsgrpc "github.com/primandproper/platform-go/v5/errors/grpc"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/identifiers"
	"github.com/primandproper/platform-go/v5/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/random"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "oauth_manager"

	clientIDSize     = 16
	clientSecretSize = 16
)

type OAuth2Manager interface {
	CreateOAuth2Client(ctx context.Context, input *oauth.OAuth2ClientCreationRequestInput) (*oauth.OAuth2Client, error)
	GetOAuth2Client(ctx context.Context, oauth2ClientID string) (*oauth.OAuth2Client, error)
	GetOAuth2Clients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[oauth.OAuth2Client], error)
	ArchiveOAuth2Client(ctx context.Context, oauth2ClientID string) error
}

type manager struct {
	tracer                    tracing.Tracer
	logger                    logging.Logger
	sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
	secretGenerator           random.Generator
	dataChangesPublisher      messagequeue.Publisher
	oauthRepository           oauth.Repository
}

func NewOAuth2Manager(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	secretGenerator random.Generator,
	sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error),
	publisherProvider messagequeue.PublisherProvider,
	oauthRepository oauth.Repository,
	queuesConfig *msgconfig.QueuesConfig,
) (OAuth2Manager, error) {
	if queuesConfig == nil {
		return nil, perrors.ErrNilInputProvided
	}

	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, queuesConfig.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide publisher for data changes topic: %w", err)
	}

	return &manager{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		sessionContextDataFetcher: sessionContextDataFetcher,
		secretGenerator:           secretGenerator,
		oauthRepository:           oauthRepository,
		dataChangesPublisher:      dataChangesPublisher,
	}, nil
}

func (m *manager) CreateOAuth2Client(ctx context.Context, input *oauth.OAuth2ClientCreationRequestInput) (*oauth.OAuth2Client, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	sessionContextData, err := m.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching session context data")
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid oauth2 client creation request")
	}

	dbInput := converters.ConvertOAuth2ClientCreationRequestInputToOAuth2ClientDatabaseCreationInput(input)
	dbInput.ID = identifiers.New()

	if dbInput.ClientID, err = m.secretGenerator.GenerateHexEncodedString(ctx, clientIDSize); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "generating client id")
	}

	if dbInput.ClientSecret, err = m.secretGenerator.GenerateHexEncodedString(ctx, clientSecretSize); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "generating client secret")
	}

	created, err := m.oauthRepository.CreateOAuth2Client(ctx, dbInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating oauth2 client")
	}

	m.dataChangesPublisher.PublishAsync(ctx, &audit.DataChangeMessage{
		EventType: oauth.OAuth2ClientCreatedServiceEventType,
		Context: map[string]any{
			oauthkeys.OAuth2ClientIDKey: created.ID,
		},
		UserID:    sessionContextData.GetUserID(),
		AccountID: sessionContextData.GetActiveAccountID(),
	})

	return created, nil
}

func (m *manager) GetOAuth2Client(ctx context.Context, oauth2ClientID string) (*oauth.OAuth2Client, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		oauthkeys.OAuth2ClientIDKey: oauth2ClientID,
	}, span, m.logger)

	oauth2Client, err := m.oauthRepository.GetOAuth2ClientByDatabaseID(ctx, oauth2ClientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client by database ID")
	}

	return oauth2Client, nil
}

func (m *manager) GetOAuth2Clients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[oauth.OAuth2Client], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(m.logger.WithSpan(span))

	oauth2Clients, err := m.oauthRepository.GetOAuth2Clients(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "getting oauth2 clients")
	}

	return oauth2Clients, nil
}

func (m *manager) ArchiveOAuth2Client(ctx context.Context, oauth2ClientID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := m.sessionContextDataFetcher(ctx)
	if err != nil {
		return observability.PrepareAndLogError(err, m.logger, span, "fetching session context data")
	}

	logger := observability.ObserveValues(map[string]any{
		oauthkeys.OAuth2ClientIDKey: oauth2ClientID,
	}, span, m.logger)

	if err = m.oauthRepository.ArchiveOAuth2Client(ctx, oauth2ClientID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving oauth2 client")
	}

	m.dataChangesPublisher.PublishAsync(ctx, &audit.DataChangeMessage{
		EventType: oauth.OAuth2ClientArchivedServiceEventType,
		Context: map[string]any{
			oauthkeys.OAuth2ClientIDKey: oauth2ClientID,
		},
		UserID:    sessionContextData.GetUserID(),
		AccountID: sessionContextData.GetActiveAccountID(),
	})

	return nil
}
