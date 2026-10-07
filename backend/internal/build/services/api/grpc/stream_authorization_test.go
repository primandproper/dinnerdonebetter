package grpcapi

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/mediaregistry/mediaregistrypb"
	"github.com/primandproper/primitives-go/v2/authentication/tokens/jwt"
	"github.com/primandproper/primitives-go/v2/authorization/static"
	"github.com/primandproper/primitives-go/v2/database"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/random"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// principalDirectory answers GetPrincipal from what it was given.
type principalDirectory struct {
	principals sync.Map
}

func (d *principalDirectory) GetPrincipal(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*platformidentity.Principal, error) {
	principal, ok := d.principals.Load(userID)
	if !ok {
		return nil, platformidentity.ErrUserNotFound
	}

	return principal.(*platformidentity.Principal), nil
}

// buildTestStreamChain is the stream chain the server installs, over the real method table,
// platform's real extractor and a real signer, with only the directory faked.
func buildTestStreamChain(t *testing.T) ([]grpc.StreamServerInterceptor, *jwt.Signer, *principalDirectory) {
	t.Helper()

	key, err := random.GenerateRawBytes(t.Context(), 32)
	require.NoError(t, err)

	signer, err := jwt.NewSigner(identifiers.New(), identifiers.New(), key)
	require.NoError(t, err)

	directory := &principalDirectory{}
	client := &databasemock.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }}

	extractor, err := interceptors.ProvidePrincipalExtractor(signer, client, directory, nil, nil, loggingnoop.NewLogger(), nil)
	require.NoError(t, err)

	policy, err := static.NewResolver(authorization.PlatformPolicy())
	require.NoError(t, err)

	sessionBuilder, err := identitybuild.NewSessionBuilder(policy, nil)
	require.NoError(t, err)

	authInterceptor, err := interceptors.ProvideAuthInterceptor(loggingnoop.NewLogger(), extractor, sessionBuilder, MethodPermissions())
	require.NoError(t, err)

	return BuildStreamServerInterceptors(authInterceptor, buildServerEnforcer(t, authInterceptor)), signer, directory
}

// TestBuildStreamServerInterceptors_enforcesTheMethodPermissionTable drives the stream chain the
// server installs.
//
// UploadObject is the stream it is about: it requires mediaregistry.objects.create, which a person
// holds through service_user and an account membership does not grant.
func TestBuildStreamServerInterceptors_enforcesTheMethodPermissionTable(T *testing.T) {
	T.Parallel()

	chain, signer, directory := buildTestStreamChain(T)
	info := &grpc.StreamServerInfo{FullMethod: mediaregistrypb.MediaRegistryService_UploadObject_FullMethodName, IsClientStream: true}

	// upload opens the stream as somebody holding serviceRoles and a membership of one account,
	// and reports whether the handler ran.
	upload := func(t *testing.T, serviceRoles ...string) (bool, error) {
		t.Helper()

		user := identityfakes.BuildFakeUser()
		user.ServiceRoles = serviceRoles
		user.RequiresPasswordChange = false

		accountID := identifiers.New()
		principal := &platformidentity.Principal{
			User:            user,
			ActiveAccountID: accountID,
			Memberships: []*platformidentity.Membership{{
				ID:               identifiers.New(),
				BelongsToUser:    user.ID,
				BelongsToAccount: accountID,
				Roles:            []string{authorization.AccountMemberRoleName},
			}},
		}
		directory.principals.Store(user.ID, principal)

		claims, err := signin.DefaultClaims(t.Context(), &signin.ClaimsInput{Principal: principal, FamilyID: identifiers.New()})
		require.NoError(t, err)

		token, _, err := signer.IssueToken(t.Context(), user.ID, time.Hour, claims)
		require.NoError(t, err)

		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer "+token))

		called := false
		err = chainStream(chain, info, func(any, grpc.ServerStream) error {
			called = true
			return nil
		})(nil, &fakeServerStream{ctx: ctx})

		return called, err
	}

	T.Run("refuses a signed-in caller without the permission", func(t *testing.T) {
		t.Parallel()

		called, err := upload(t)

		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.False(t, called, "the handler ran for a caller the enforcer should have refused")
	})

	T.Run("admits a caller who holds it", func(t *testing.T) {
		t.Parallel()

		called, err := upload(t, authorization.ServiceUserRoleName)

		require.NoError(t, err)
		assert.True(t, called)
	})
}
