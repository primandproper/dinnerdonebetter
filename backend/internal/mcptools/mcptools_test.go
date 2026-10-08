package mcptools

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v15/callers"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	oauth2mcp "github.com/primandproper/primitives-go/v2/authentication/oauth2server/mcp"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// examplePermission is a grant the tests ask for; which one does not matter.
const examplePermission = authorization.ReadRecipesPermission

// signedIn is a context carrying a session for userID holding perms service-wide: what the
// authenticator attaches, built by hand.
func signedIn(ctx context.Context, userID string, perms ...authorization.Permission) context.Context {
	return sessions.AttachToContext(ctx, &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:             userID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker(nil, perms),
		},
		ActiveAccountID: fake.BuildFakeID(),
	})
}

// attaching is an authenticator that puts the session for userID on every call.
func attaching(userID string, perms ...authorization.Permission) Authenticator {
	return func(ctx context.Context, _ *mcp.CallToolRequest) (context.Context, error) {
		return signedIn(ctx, userID, perms...), nil
	}
}

// nobody is an authenticator that finds no credential on any call.
func nobody(ctx context.Context, _ *mcp.CallToolRequest) (context.Context, error) { return ctx, nil }

func buildTestGate(t *testing.T, authenticate Authenticator) *Gate {
	t.Helper()

	gate, err := NewGate(authenticate, sessions.PrincipalFromContext, sessions.GrantsFromContext)
	require.NoError(t, err)

	return gate
}

func TestNewGate(T *testing.T) {
	T.Parallel()

	T.Run("refuses a nil authenticator", func(t *testing.T) {
		t.Parallel()

		_, err := NewGate(nil, sessions.PrincipalFromContext, sessions.GrantsFromContext)
		require.ErrorIs(t, err, ErrNilAuthenticator)
	})

	T.Run("refuses a nil principal extractor", func(t *testing.T) {
		t.Parallel()

		_, err := NewGate(nobody, nil, sessions.GrantsFromContext)
		require.ErrorIs(t, err, ErrNilPrincipalExtractor)
	})

	T.Run("refuses a nil grants extractor", func(t *testing.T) {
		t.Parallel()

		_, err := NewGate(nobody, sessions.PrincipalFromContext, nil)
		require.ErrorIs(t, err, ErrNilGrantsExtractor)
	})
}

func TestGate_Begin(T *testing.T) {
	T.Parallel()

	T.Run("admits a caller holding the grant", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		gate := buildTestGate(t, attaching(userID, examplePermission))

		ctx, principal, err := gate.Begin(t.Context(), &mcp.CallToolRequest{}, examplePermission)
		require.NoError(t, err)
		require.NotNil(t, principal)
		assert.Equal(t, userID, principal.UserID())

		// The context handed back is the one carrying the caller, so a tool that reads
		// through it reads as them.
		attached, ok := sessions.PrincipalFromContext(ctx)
		require.True(t, ok)
		assert.Equal(t, userID, attached.UserID())
	})

	T.Run("admits any caller to a tool requiring no grant", func(t *testing.T) {
		t.Parallel()

		gate := buildTestGate(t, attaching(fake.BuildFakeID()))

		_, principal, err := gate.Begin(t.Context(), &mcp.CallToolRequest{})
		require.NoError(t, err)
		assert.NotNil(t, principal)
	})

	T.Run("refuses a call carrying nobody", func(t *testing.T) {
		t.Parallel()

		gate := buildTestGate(t, nobody)

		_, principal, err := gate.Begin(t.Context(), &mcp.CallToolRequest{}, examplePermission)
		require.ErrorIs(t, err, ErrNoPrincipal)
		assert.Nil(t, principal)
	})

	T.Run("refuses a caller without the grant", func(t *testing.T) {
		t.Parallel()

		gate := buildTestGate(t, attaching(fake.BuildFakeID(), authorization.ReadMealsPermission))

		_, principal, err := gate.Begin(t.Context(), &mcp.CallToolRequest{}, examplePermission)
		require.ErrorIs(t, err, ErrPermissionDenied)
		assert.Nil(t, principal)
	})

	T.Run("requires every grant named", func(t *testing.T) {
		t.Parallel()

		gate := buildTestGate(t, attaching(fake.BuildFakeID(), examplePermission))

		_, _, err := gate.Begin(t.Context(), &mcp.CallToolRequest{}, examplePermission, authorization.ReadMealsPermission)
		require.ErrorIs(t, err, ErrPermissionDenied)
	})

	T.Run("reports an authenticator that could not decide", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("directory is down")
		gate := buildTestGate(t, func(ctx context.Context, _ *mcp.CallToolRequest) (context.Context, error) {
			return ctx, expected
		})

		_, _, err := gate.Begin(t.Context(), &mcp.CallToolRequest{}, examplePermission)
		require.ErrorIs(t, err, expected)
		assert.NotErrorIs(t, err, ErrNoPrincipal)
	})
}

// directoryFunc is a PrincipalDirectory answering with one function.
type directoryFunc func(ctx context.Context, scope tenancy.Scope, userID, accountID string) (*platformidentity.Principal, error)

func (f directoryFunc) GetPrincipal(ctx context.Context, _ database.SQLQueryExecutor, scope tenancy.Scope, userID, accountID string) (*platformidentity.Principal, error) {
	return f(ctx, scope, userID, accountID)
}

// rendererFunc is a SessionRenderer answering with one function.
type rendererFunc func(ctx context.Context, principal *platformidentity.Principal) (*sessions.ContextData, error)

func (f rendererFunc) SessionForPrincipal(ctx context.Context, principal *platformidentity.Principal) (*sessions.ContextData, error) {
	return f(ctx, principal)
}

// renderingRoles is a renderer that grants perms to every principal it is handed.
func renderingRoles(perms ...authorization.Permission) SessionRenderer {
	return rendererFunc(func(_ context.Context, principal *platformidentity.Principal) (*sessions.ContextData, error) {
		return &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID:             principal.User.ID,
				ServicePermissions: authorization.NewServiceRolePermissionChecker(principal.ServiceRoles(), perms),
			},
			ActiveAccountID: principal.ActiveAccountID,
		}, nil
	})
}

// mockDBForTest is a client whose reader is nil, because the directory above it is a function
// and never sends a statement.
func mockDBForTest() *mockdatabase.ClientMock {
	return &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
	}
}

// tokenRequest is a tool call carrying a token Protect verified: the access token under the
// key oauth2server/mcp puts it at.
func tokenRequest(userID, accountID string) *mcp.CallToolRequest {
	token := &oauth2server.AccessToken{
		Subject: oauth2server.Subject{ID: userID, Claims: map[string]string{ClaimAccountID: accountID}},
	}

	return &mcp.CallToolRequest{Extra: &mcp.RequestExtra{TokenInfo: &auth.TokenInfo{
		UserID: userID,
		Extra:  map[string]any{oauth2mcp.ExtraAccessToken: token},
	}}}
}

func TestNewAuthenticator(T *testing.T) {
	T.Parallel()

	directory := directoryFunc(func(context.Context, tenancy.Scope, string, string) (*platformidentity.Principal, error) {
		return nil, platformidentity.ErrUserNotFound
	})

	T.Run("refuses a nil directory", func(t *testing.T) {
		t.Parallel()

		_, err := NewAuthenticator(nil, mockDBForTest(), renderingRoles())
		require.ErrorIs(t, err, ErrNilDirectory)
	})

	T.Run("refuses a nil database client", func(t *testing.T) {
		t.Parallel()

		_, err := NewAuthenticator(directory, nil, renderingRoles())
		require.ErrorIs(t, err, ErrNilDatabaseClient)
	})

	T.Run("refuses a nil session renderer", func(t *testing.T) {
		t.Parallel()

		_, err := NewAuthenticator(directory, mockDBForTest(), nil)
		require.ErrorIs(t, err, ErrNilSessionRenderer)
	})
}

func TestAuthenticator(T *testing.T) {
	T.Parallel()

	T.Run("resolves the token's subject in the account it was issued for", func(t *testing.T) {
		t.Parallel()

		exampleUser := identityfakes.BuildFakeUser()
		exampleUser.ServiceRoles = []string{authorization.ServiceAdminRoleName}
		accountID := identityfakes.BuildFakeAccount().ID

		directory := directoryFunc(func(_ context.Context, scope tenancy.Scope, userID, activeAccountID string) (*platformidentity.Principal, error) {
			// The global directory, and the account the claim names rather than the
			// subject's default: a grant made for one account is not a grant for whichever
			// account they hold now.
			assert.Equal(t, tenancy.Global(), scope)
			assert.Equal(t, exampleUser.ID, userID)
			assert.Equal(t, accountID, activeAccountID)

			return &platformidentity.Principal{User: exampleUser, ActiveAccountID: accountID}, nil
		})

		authenticate, err := NewAuthenticator(directory, mockDBForTest(), renderingRoles(examplePermission))
		require.NoError(t, err)

		ctx, err := authenticate(t.Context(), tokenRequest(exampleUser.ID, accountID))
		require.NoError(t, err)

		principal, ok := sessions.PrincipalFromContext(ctx)
		require.True(t, ok)
		assert.Equal(t, exampleUser.ID, principal.UserID())
		assert.Equal(t, accountID, principal.ActiveAccountID())

		// What the roles grant is on the context too, which is what the gate checks and
		// what a platform tool surface reads through sessions.GrantsFromContext.
		grants, ok := sessions.GrantsFromContext(ctx)
		require.True(t, ok)
		assert.True(t, grants.Has(examplePermission))
	})

	T.Run("answers a call carrying no token with nobody", func(t *testing.T) {
		t.Parallel()

		directory := directoryFunc(func(context.Context, tenancy.Scope, string, string) (*platformidentity.Principal, error) {
			t.Fatal("the directory must not be read for a call carrying no token")

			return nil, nil
		})

		authenticate, err := NewAuthenticator(directory, mockDBForTest(), renderingRoles())
		require.NoError(t, err)

		for name, req := range map[string]*mcp.CallToolRequest{
			"nil request":         nil,
			"no extra":            {},
			"no token info":       {Extra: &mcp.RequestExtra{}},
			"token not Protect's": {Extra: &mcp.RequestExtra{TokenInfo: &auth.TokenInfo{UserID: fake.BuildFakeID()}}},
			"subject naming nobody": {Extra: &mcp.RequestExtra{TokenInfo: &auth.TokenInfo{
				Extra: map[string]any{oauth2mcp.ExtraAccessToken: &oauth2server.AccessToken{}},
			}}},
		} {
			ctx, authErr := authenticate(t.Context(), req)
			require.NoError(t, authErr, name)

			_, ok := sessions.PrincipalFromContext(ctx)
			assert.False(t, ok, name)
		}
	})

	T.Run("answers a subject the directory refuses with nobody", func(t *testing.T) {
		t.Parallel()

		for name, refusal := range map[string]error{
			"banned":           platformidentity.ErrSignInNotAdmitted,
			"unknown":          platformidentity.ErrUserNotFound,
			"no longer member": platformidentity.ErrMembershipNotFound,
		} {
			directory := directoryFunc(func(context.Context, tenancy.Scope, string, string) (*platformidentity.Principal, error) {
				return nil, refusal
			})

			authenticate, err := NewAuthenticator(directory, mockDBForTest(), renderingRoles())
			require.NoError(t, err)

			// A refusal is not an outage: the call goes on as nobody and the gate refuses
			// it as unauthenticated, rather than being reported to the model as a tool
			// that broke.
			ctx, err := authenticate(t.Context(), tokenRequest(fake.BuildFakeID(), fake.BuildFakeID()))
			require.NoError(t, err, name)

			_, ok := sessions.PrincipalFromContext(ctx)
			assert.False(t, ok, name)
		}
	})

	T.Run("reports a directory that could not be read", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("connection refused")
		directory := directoryFunc(func(context.Context, tenancy.Scope, string, string) (*platformidentity.Principal, error) {
			return nil, expected
		})

		authenticate, err := NewAuthenticator(directory, mockDBForTest(), renderingRoles())
		require.NoError(t, err)

		_, err = authenticate(t.Context(), tokenRequest(fake.BuildFakeID(), fake.BuildFakeID()))
		require.ErrorIs(t, err, expected)
	})

	T.Run("reports a session that could not be rendered", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("policy unreadable")
		exampleUser := identityfakes.BuildFakeUser()

		directory := directoryFunc(func(context.Context, tenancy.Scope, string, string) (*platformidentity.Principal, error) {
			return &platformidentity.Principal{User: exampleUser}, nil
		})
		renderer := rendererFunc(func(context.Context, *platformidentity.Principal) (*sessions.ContextData, error) {
			return nil, expected
		})

		authenticate, err := NewAuthenticator(directory, mockDBForTest(), renderer)
		require.NoError(t, err)

		_, err = authenticate(t.Context(), tokenRequest(exampleUser.ID, fake.BuildFakeID()))
		require.ErrorIs(t, err, expected)
	})
}

// The gate and the authenticator together: the shape the server composes them in.
func TestGateOverAuthenticator(T *testing.T) {
	T.Parallel()

	T.Run("a verified token reaches a tool as a caller with grants", func(t *testing.T) {
		t.Parallel()

		exampleUser := identityfakes.BuildFakeUser()
		directory := directoryFunc(func(context.Context, tenancy.Scope, string, string) (*platformidentity.Principal, error) {
			return &platformidentity.Principal{User: exampleUser, ActiveAccountID: fake.BuildFakeID()}, nil
		})

		authenticate, err := NewAuthenticator(directory, mockDBForTest(), renderingRoles(examplePermission))
		require.NoError(t, err)

		gate := buildTestGate(t, authenticate)

		var principal callers.Principal
		_, principal, err = gate.Begin(t.Context(), tokenRequest(exampleUser.ID, fake.BuildFakeID()), examplePermission)
		require.NoError(t, err)
		assert.Equal(t, exampleUser.ID, principal.UserID())

		_, _, err = gate.Begin(t.Context(), tokenRequest(exampleUser.ID, fake.BuildFakeID()), authorization.ReadMealsPermission)
		require.ErrorIs(t, err, ErrPermissionDenied)
	})
}
