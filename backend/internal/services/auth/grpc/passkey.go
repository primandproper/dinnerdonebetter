package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
)

// passkeyUserStore adapts identityDataManager to webauthn.UserStore.
type passkeyUserStore struct {
	directory platformidentity.Store
	db        database.Client
}

var _ webauthn.UserStore = (*passkeyUserStore)(nil)

func (s *passkeyUserStore) GetUserByID(ctx context.Context, userID string) (*platformidentity.User, error) {
	return s.directory.GetUser(ctx, s.db.Reader(), identity.Scope(), userID)
}

func (s *passkeyUserStore) GetUserByUsername(ctx context.Context, username string) (*platformidentity.User, error) {
	return s.directory.GetUserByUsername(ctx, s.db.Reader(), identity.Scope(), username)
}
