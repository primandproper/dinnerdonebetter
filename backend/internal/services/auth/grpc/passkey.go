package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"

	"github.com/primandproper/platform-go/v9/observability/logging"
)

// passkeyUserStore adapts identityDataManager to webauthn.UserStore.
type passkeyUserStore struct {
	identityDataManager identitymanager.IdentityDataManager
}

func (s *passkeyUserStore) GetUserByID(ctx context.Context, userID string) (*identity.User, error) {
	return s.identityDataManager.GetUser(ctx, userID)
}

func (s *passkeyUserStore) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	return s.identityDataManager.GetUserByUsername(ctx, username)
}

// ProvidePasskeyService creates a WebAuthn passkey service.
func ProvidePasskeyService(
	logger logging.Logger,
	cfg webauthn.Config,
	identityDataManager identitymanager.IdentityDataManager,
	identityRepo identity.Repository,
	sessionStore webauthn.SessionStore,
) (*webauthn.Service, error) {
	userStore := &passkeyUserStore{identityDataManager: identityDataManager}
	return webauthn.NewService(logger, cfg, identityRepo, userStore, sessionStore)
}
