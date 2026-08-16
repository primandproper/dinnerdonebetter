package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
)

// passkeyUserStore adapts identityDataManager to webauthn.UserStore.
type passkeyUserStore struct {
	identityDataManager identitymanager.IdentityDataManager
}

var _ webauthn.UserStore = (*passkeyUserStore)(nil)

func (s *passkeyUserStore) GetUserByID(ctx context.Context, userID string) (*identity.User, error) {
	return s.identityDataManager.GetUser(ctx, userID)
}

func (s *passkeyUserStore) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	return s.identityDataManager.GetUserByUsername(ctx, username)
}
