package authentication

import (
	"context"
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v4/identifiers"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"

	"github.com/go-oauth2/oauth2/v4"
)

var _ oauth2.TokenStore = (*oauth2TokenStoreImpl)(nil)

type oauth2TokenStoreImpl struct {
	tracer      tracing.Tracer
	logger      logging.Logger
	dataManager types.OAuth2ClientTokenDataManager
}

// Create implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) Create(ctx context.Context, info oauth2.TokenInfo) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	info.SetAccessExpiresIn(24 * time.Hour)
	info.SetRefreshExpiresIn(72 * time.Hour)

	logger := s.logger.WithValue("info", info)

	input := &types.OAuth2ClientTokenDatabaseCreationInput{
		RefreshCreatedAt:    info.GetRefreshCreateAt(),
		AccessCreatedAt:     info.GetAccessCreateAt(),
		CodeCreatedAt:       info.GetCodeCreateAt(),
		RedirectURI:         info.GetRedirectURI(),
		Code:                info.GetCode(),
		CodeChallenge:       info.GetCodeChallenge(),
		CodeChallengeMethod: info.GetCodeChallengeMethod().String(),
		BelongsToUser:       info.GetUserID(),
		Access:              info.GetAccess(),
		ClientID:            info.GetClientID(),
		Refresh:             info.GetRefresh(),
		ID:                  identifiers.New(),
		CodeExpiresIn:       info.GetCodeExpiresIn(),
		AccessExpiresIn:     info.GetAccessExpiresIn(),
		RefreshExpiresIn:    info.GetRefreshExpiresIn(),
	}

	if _, err := s.dataManager.CreateOAuth2ClientToken(ctx, input); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating oauth2 client token")
	}

	return nil
}

// RemoveByCode implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) RemoveByCode(ctx context.Context, code string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("code", code)
	logger.Debug("RemoveByCode invoked")

	if err := s.dataManager.DeleteOAuth2ClientTokenByCode(ctx, code); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing oauth2 client token by code")
	}

	return nil
}

// RemoveByAccess implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) RemoveByAccess(ctx context.Context, access string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("access", access)
	logger.Debug("RemoveByAccess invoked")

	if err := s.dataManager.DeleteOAuth2ClientTokenByAccess(ctx, access); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing oauth2 client token by access")
	}

	return nil
}

// RemoveByRefresh implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) RemoveByRefresh(ctx context.Context, refresh string) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("refresh", refresh)
	logger.Debug("RemoveByRefresh invoked")

	if err := s.dataManager.DeleteOAuth2ClientTokenByRefresh(ctx, refresh); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing oauth2 client token by refresh")
	}

	return nil
}

// GetByCode implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("code", code)
	logger.Debug("GetByCode invoked")

	token, err := s.dataManager.GetOAuth2ClientTokenByCode(ctx, code)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client token by code")
	}

	return convertTokenToImpl(token), nil
}

// GetByAccess implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("access", access)
	logger.Debug("GetByAccess invoked")

	token, err := s.dataManager.GetOAuth2ClientTokenByAccess(ctx, access)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client token by access")
	}

	return convertTokenToImpl(token), nil
}

// GetByRefresh implements the requisite oauth2 interface.
func (s *oauth2TokenStoreImpl) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("refresh", refresh)
	logger.Debug("GetByRefresh invoked")

	token, err := s.dataManager.GetOAuth2ClientTokenByRefresh(ctx, refresh)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client token by refresh")
	}

	return convertTokenToImpl(token), nil
}
