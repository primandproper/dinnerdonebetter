package authentication

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/keys"

	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/go-oauth2/oauth2/v4"
)

var _ oauth2.ClientStore = (*oauth2ClientStoreImpl)(nil)

type oauth2ClientStoreImpl struct {
	tracer      tracing.Tracer
	logger      logging.Logger
	dataManager types.OAuth2ClientDataManager
	domain      string
}

func newOAuth2ClientStore(domain string, logger logging.Logger, tracer tracing.Tracer, dataManager types.OAuth2ClientDataManager) oauth2.ClientStore {
	return &oauth2ClientStoreImpl{
		domain:      domain,
		tracer:      tracer,
		logger:      logging.EnsureLogger(logger),
		dataManager: dataManager,
	}
}

// GetByID implements the oauth2.ClientStore interface.
func (s *oauth2ClientStoreImpl) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(oauthkeys.OAuth2ClientIDKey, id)
	logger.Debug("getting oauth2 client by ID")

	client, err := s.dataManager.GetOAuth2ClientByClientID(ctx, id)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client by ID")
	}

	c := &oauth2ClientInfoImpl{
		domain: s.domain,
		client: client,
	}

	return c, nil
}
