package interceptors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/authentication/tokens/jwt"
	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
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

const (
	// operatorMethod needs a grant only an operator holds.
	operatorMethod = "/test.Service/OperatorWork"
	// personalMethod needs a grant every person holds about themselves, through service_user.
	personalMethod = "/test.Service/PersonalWork"
	// accountMethod needs a grant held through a membership.
	accountMethod = "/test.Service/AccountWork"
)

// fakeDirectory answers GetPrincipal from what it was given, refusing anybody else as the real
// directory would.
type fakeDirectory struct {
	principals map[string]*platformidentity.Principal
}

func (d *fakeDirectory) GetPrincipal(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*platformidentity.Principal, error) {
	principal, ok := d.principals[userID]
	if !ok {
		return nil, platformidentity.ErrUserNotFound
	}

	if !principal.User.AccountStatus.AdmitsSignIn() {
		return nil, platformidentity.ErrSignInNotAdmitted
	}

	return principal, nil
}

// interceptorHarness is an AuthInterceptor over platform's real extractor and a real JWT signer,
// with only the directory faked, followed by an authorization Enforcer over the same table, as
// the server chains them.
type interceptorHarness struct {
	signer      *jwt.Signer
	directory   *fakeDirectory
	interceptor *AuthInterceptor
	enforcer    *authzgrpc.Enforcer
}

func newInterceptorHarness(t *testing.T) *interceptorHarness {
	t.Helper()

	key, err := random.GenerateRawBytes(t.Context(), 32)
	require.NoError(t, err)

	signer, err := jwt.NewSigner(identifiers.New(), identifiers.New(), key)
	require.NoError(t, err)

	directory := &fakeDirectory{principals: map[string]*platformidentity.Principal{}}
	client := &databasemock.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }}

	extractor, err := ProvidePrincipalExtractor(signer, client, directory, nil, nil, loggingnoop.NewLogger(), nil)
	require.NoError(t, err)

	policy, err := static.NewResolver(authorization.PlatformPolicy())
	require.NoError(t, err)

	builder, err := identitybuild.NewSessionBuilder(policy, nil)
	require.NoError(t, err)

	table := MethodPermissionsMap{
		operatorMethod: {authorization.ImpersonateUserPermission},
		personalMethod: {authorization.ReadMediaObjectsPermission},
		accountMethod:  {authorization.ReadAuditLogEntriesPermission},
		signinpb.SignInService_UpdatePassword_FullMethodName: {},
	}

	interceptor, err := ProvideAuthInterceptor(loggingnoop.NewLogger(), extractor, builder, table)
	require.NoError(t, err)

	requirements := authzgrpc.NewRequirements()
	for method, perms := range table {
		if len(perms) == 0 {
			requirements.Public(method)
			continue
		}
		requirements.Require(method, authorization.ToPlatformPermissions(perms)...)
	}
	for _, method := range interceptor.UnauthenticatedRoutes() {
		if _, declared := table[method]; !declared {
			requirements.Public(method)
		}
	}

	built, err := requirements.Build()
	require.NoError(t, err)

	enforcer, err := authzgrpc.NewEnforcer(built, sessions.GrantsFromContext)
	require.NoError(t, err)

	return &interceptorHarness{signer: signer, directory: directory, interceptor: interceptor, enforcer: enforcer}
}

// addUser puts somebody in the directory, a member of one account, holding the service roles
// named.
func (h *interceptorHarness) addUser(serviceRoles ...string) *platformidentity.Principal {
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

	h.directory.principals[user.ID] = principal

	return principal
}

// tokenFor mints the token signin would for this principal, with the claims signin's
// DefaultClaims gives one.
func (h *interceptorHarness) tokenFor(t *testing.T, principal *platformidentity.Principal, administrative bool, actorID string) string {
	t.Helper()

	input := &signin.ClaimsInput{
		Principal:      principal,
		FamilyID:       identifiers.New(),
		Administrative: administrative,
		ActorID:        actorID,
	}
	if actorID != "" {
		input.ActorScope = tenancy.Global()
	}

	claims, err := signin.DefaultClaims(t.Context(), input)
	require.NoError(t, err)

	token, _, err := h.signer.IssueToken(t.Context(), principal.User.ID, time.Hour, claims)
	require.NoError(t, err)

	return token
}

// call makes one unary call to method with token through the interceptor and then the enforcer,
// and returns the session the handler saw.
func (h *interceptorHarness) call(t *testing.T, method, token string) (*sessions.ContextData, error) {
	t.Helper()

	ctx := t.Context()
	if token != "" {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
	}

	info := &grpc.UnaryServerInfo{FullMethod: method}
	enforce := h.enforcer.UnaryServerInterceptor()

	var seen *sessions.ContextData
	_, err := h.interceptor.UnaryServerInterceptor()(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return enforce(ctx, req, info, func(ctx context.Context, _ any) (any, error) {
			seen = sessions.FromContext(ctx)
			return nil, nil
		})
	})

	return seen, err
}

func TestAuthInterceptor_UnaryServerInterceptor(T *testing.T) {
	T.Parallel()

	T.Run("an administrative token carries the operator's grants", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		operator := h.addUser(authorization.ServiceAdminRoleName)

		session, err := h.call(t, operatorMethod, h.tokenFor(t, operator, true, ""))
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, operator.User.ID, session.GetUserID())
		assert.True(t, session.ServiceRolePermissionChecker().CanImpersonateUsers())
	})

	T.Run("an ordinary token carries no operator permission", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		operator := h.addUser(authorization.ServiceAdminRoleName)
		token := h.tokenFor(t, operator, false, "")

		_, err := h.call(t, operatorMethod, token)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// Still the person, with what a person holds about themselves.
		session, err := h.call(t, personalMethod, token)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.False(t, session.ServiceRolePermissionChecker().IsServiceAdmin())
		assert.False(t, session.ServiceRolePermissionChecker().CanImpersonateUsers())
	})

	T.Run("a data administrator's ordinary token carries none of their grants either", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		dataAdmin := h.addUser(authorization.ServiceDataAdminRoleName)

		session, err := h.call(t, personalMethod, h.tokenFor(t, dataAdmin, false, ""))
		require.NoError(t, err)
		require.NotNil(t, session)

		for _, permission := range authorization.ServiceDataAdminPermissions {
			assert.False(t, session.ServiceRolePermissionChecker().HasPermission(permission), permission)
		}
	})

	T.Run("leaves the permission to the enforcer", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		operator := h.addUser(authorization.ServiceAdminRoleName)
		token := h.tokenFor(t, operator, false, "")
		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer "+token))

		// The interceptor alone admits a caller the enforcer refuses: who is calling is its
		// question, and what they may do is the enforcer's.
		called := false
		_, err := h.interceptor.UnaryServerInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: operatorMethod}, func(context.Context, any) (any, error) {
			called = true
			return nil, nil
		})
		require.NoError(t, err)
		assert.True(t, called)

		_, err = h.call(t, operatorMethod, token)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("an impersonation token carries only the subject's permissions", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		operator := h.addUser(authorization.ServiceAdminRoleName)
		subject := h.addUser(authorization.ServiceUserRoleName)

		token := h.tokenFor(t, subject, false, operator.User.ID)

		session, err := h.call(t, accountMethod, token)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, subject.User.ID, session.GetUserID())
		assert.Equal(t, operator.User.ID, session.ImpersonatorID)
		assert.Equal(t, subject.ActiveAccountID, session.GetActiveAccountID())
		assert.False(t, session.ServiceRolePermissionChecker().IsServiceAdmin())

		_, err = h.call(t, operatorMethod, token)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("an impersonation whose operator no longer stands is refused", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		operator := h.addUser(authorization.ServiceAdminRoleName)
		subject := h.addUser(authorization.ServiceUserRoleName)
		token := h.tokenFor(t, subject, false, operator.User.ID)

		operator.User.AccountStatus = platformidentity.StatusBanned

		// Suspending an operator mid-impersonation stops them on their next request, not when
		// the token lapses.
		_, err := h.call(t, accountMethod, token)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("a banned user is refused as the directory refuses them", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		user := h.addUser(authorization.ServiceUserRoleName)
		token := h.tokenFor(t, user, false, "")

		user.User.AccountStatus = platformidentity.StatusBanned

		_, err := h.call(t, personalMethod, token)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("names the login a sign-in token belongs to", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		user := h.addUser(authorization.ServiceUserRoleName)

		session, err := h.call(t, personalMethod, h.tokenFor(t, user, false, ""))
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.NotEmpty(t, session.SignInFamilyID)
	})

	T.Run("holds a caller who owes a password change at the form", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		user := h.addUser(authorization.ServiceUserRoleName)
		user.User.RequiresPasswordChange = true
		token := h.tokenFor(t, user, false, "")

		_, err := h.call(t, personalMethod, token)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))

		session, err := h.call(t, signinpb.SignInService_UpdatePassword_FullMethodName, token)
		require.NoError(t, err)
		assert.Equal(t, user.User.ID, session.GetUserID())
	})

	T.Run("refuses a required method's anonymous caller", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)

		_, err := h.call(t, personalMethod, "")
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("refuses a token nobody signed", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)

		_, err := h.call(t, personalMethod, identifiers.New())
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("serves a public method with no credential and no extractor", func(t *testing.T) {
		t.Parallel()

		interceptor, err := ProvideAuthInterceptor(loggingnoop.NewLogger(), nil, nil, MethodPermissionsMap{})
		require.NoError(t, err)

		called := false
		_, err = interceptor.UnaryServerInterceptor()(t.Context(), nil,
			&grpc.UnaryServerInfo{FullMethod: signinpb.SignInService_LoginForToken_FullMethodName},
			func(context.Context, any) (any, error) {
				called = true
				return nil, nil
			})
		require.NoError(t, err)
		assert.True(t, called)
	})

	T.Run("refuses an optional method's caller whose token does not work", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)

		// A dead token is not a visitor: GetAuthStatus has to say it expired, not that nobody
		// is signed in.
		_, err := h.call(t, signinpb.SignInService_GetAuthStatus_FullMethodName, identifiers.New())
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("resolves an optional method's signed-in caller", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		user := h.addUser(authorization.ServiceUserRoleName)

		session, err := h.call(t, signinpb.SignInService_GetAuthStatus_FullMethodName, h.tokenFor(t, user, false, ""))
		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, user.User.ID, session.GetUserID())
	})

	T.Run("serves an optional method's visitor as nobody", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)

		session, err := h.call(t, signinpb.SignInService_GetAuthStatus_FullMethodName, "")
		require.NoError(t, err)
		assert.Nil(t, session)
	})
}

func TestAuthInterceptor_HTTPMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("attaches the caller's session", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		user := h.addUser(authorization.ServiceUserRoleName)

		var seen *sessions.ContextData
		handler := h.interceptor.HTTPMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			seen = sessions.FromContext(req.Context())
		}))

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+h.tokenFor(t, user, false, ""))
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		require.NotNil(t, seen)
		assert.Equal(t, user.User.ID, seen.GetUserID())
	})

	T.Run("refuses a caller who owes a password change", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)
		user := h.addUser(authorization.ServiceUserRoleName)
		user.User.RequiresPasswordChange = true

		called := false
		handler := h.interceptor.HTTPMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			called = true
		}))

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+h.tokenFor(t, user, false, ""))
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("lets a request with no credential through as nobody", func(t *testing.T) {
		t.Parallel()

		h := newInterceptorHarness(t)

		called := false
		handler := h.interceptor.HTTPMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			called = true
			assert.Nil(t, sessions.FromContext(req.Context()))
		}))

		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))

		assert.True(t, called)
	})
}

var _ signingrpc.PrincipalDirectory = (*fakeDirectory)(nil)
