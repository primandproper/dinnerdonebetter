package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// buildSignInClientForTest dials platform's SignInService on the server under test.
//
// It is the generated client over a plain connection, not platform's own client package,
// on purpose. That package decodes refusals into Go sentinels, and what these tests pin is
// the wire: the codes a client in another language — @primandproper/platform-client — sees
// and branches on.
func buildSignInClientForTest(t *testing.T) signinpb.SignInServiceClient {
	t.Helper()

	conn, err := grpc.NewClient(
		fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	return signinpb.NewSignInServiceClient(conn)
}

// withBearerToken is ctx carrying token the way a client sends it.
func withBearerToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

// signInForTest signs a freshly registered user in through SignInService and returns them
// with the token it issued.
func signInForTest(t *testing.T, signIn signinpb.SignInServiceClient) (token *signinpb.IssuedToken, userID string) {
	t.Helper()

	user := createServiceUserForTest(t, true, buildUserRegistrationInputForTest(t))

	res, err := signIn.LoginForToken(t.Context(), &signinpb.LoginForTokenRequest{
		Credentials: &signinpb.Credentials{
			Username: user.Username,
			Password: user.HashedPassword,
			TotpCode: generateTOTPCodeForUserForTest(t, user),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res.GetToken())

	return res.GetToken(), user.ID
}

func TestSignIn_LoginForToken(T *testing.T) {
	T.Parallel()

	T.Run("issues a rotating pair whose access token this application's services accept", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		token, userID := signInForTest(t, buildSignInClientForTest(t))

		assert.NotEmpty(t, token.GetToken())
		assert.NotEmpty(t, token.GetRefreshToken(), "a sign-in with a refresh token store names issues a refresh token")
		assert.NotEmpty(t, token.GetFamilyId())
		assert.False(t, token.GetAdministrative())

		// The token is platform's, and AuthService is this application's: the interceptor
		// has to recognize a token it did not mint for the two doors to be one sign-in.
		ddbClient, err := buildAuthedGRPCClientWithBearerToken(token.GetToken())
		require.NoError(t, err)

		self, err := ddbClient.GetSelf(ctx, &authsvc.GetSelfRequest{})
		require.NoError(t, err)
		assert.Equal(t, userID, self.GetResult().GetId())
	})

	T.Run("refuses a wrong password as Unauthenticated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user := createServiceUserForTest(t, true, buildUserRegistrationInputForTest(t))

		_, err := buildSignInClientForTest(t).LoginForToken(ctx, &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{
				Username: user.Username,
				Password: user.HashedPassword + user.HashedPassword,
				TotpCode: generateTOTPCodeForUserForTest(t, user),
			},
		})

		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestSignIn_AdminLoginForToken(T *testing.T) {
	T.Parallel()

	T.Run("issues an administrative token to a service administrator", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		code, err := totp.GenerateCode(strings.ToUpper(premadeAdminUser.TwoFactorSecret), time.Now().UTC())
		require.NoError(t, err)

		res, err := buildSignInClientForTest(t).AdminLoginForToken(ctx, &signinpb.AdminLoginForTokenRequest{
			Credentials: &signinpb.Credentials{
				Username: premadeAdminUser.Username,
				Password: adminUserPassword,
				TotpCode: code,
			},
		})
		require.NoError(t, err)

		assert.True(t, res.GetToken().GetAdministrative())
		assert.NotEmpty(t, res.GetToken().GetRefreshToken())
	})

	T.Run("refuses somebody who is not one", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user := createServiceUserForTest(t, true, buildUserRegistrationInputForTest(t))

		_, err := buildSignInClientForTest(t).AdminLoginForToken(ctx, &signinpb.AdminLoginForTokenRequest{
			Credentials: &signinpb.Credentials{
				Username: user.Username,
				Password: user.HashedPassword,
				TotpCode: generateTOTPCodeForUserForTest(t, user),
			},
		})

		assert.Error(t, err)
	})
}

func TestSignIn_ExchangeRefreshToken(T *testing.T) {
	T.Parallel()

	T.Run("rotates the refresh token and keeps the login", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		first, _ := signInForTest(t, signIn)

		res, err := signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: first.GetRefreshToken()})
		require.NoError(t, err)

		successor := res.GetToken()
		assert.NotEqual(t, first.GetRefreshToken(), successor.GetRefreshToken())
		assert.Equal(t, first.GetFamilyId(), successor.GetFamilyId(), "an exchange continues the login rather than starting one")
	})

	T.Run("a spent refresh token presented again ends the login", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		first, _ := signInForTest(t, signIn)

		res, err := signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: first.GetRefreshToken()})
		require.NoError(t, err)
		successor := res.GetToken()

		// The replay: somebody presents the token the rightful holder has already spent.
		_, err = signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: first.GetRefreshToken()})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))

		// And the successor, which is still unspent, went with the family.
		_, err = signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: successor.GetRefreshToken()})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestSignIn_SignOut(T *testing.T) {
	T.Parallel()

	T.Run("ends the login's refresh tokens", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		token, _ := signInForTest(t, signIn)

		_, err := signIn.SignOut(ctx, &signinpb.SignOutRequest{RefreshToken: token.GetRefreshToken()})
		require.NoError(t, err)

		_, err = signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: token.GetRefreshToken()})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("leaves the access token in hand working until it expires", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// This pins the model rather than a defect. A platform access token is checked by
		// its signature and nothing else, so a sign-out stops it being replaced and does
		// not stop it being used; its lifetime is how long a sign-out takes to take effect.
		// A token from this application's AuthService behaves differently, and a change
		// that made these two agree should have to change this test to do it.
		signIn := buildSignInClientForTest(t)
		token, userID := signInForTest(t, signIn)

		_, err := signIn.SignOut(ctx, &signinpb.SignOutRequest{RefreshToken: token.GetRefreshToken()})
		require.NoError(t, err)

		self, err := signIn.GetSelf(withBearerToken(ctx, token.GetToken()), &signinpb.GetSelfRequest{})
		require.NoError(t, err)
		assert.Equal(t, userID, self.GetUser().GetId())
	})
}

func TestSignIn_SignOutEverywhere(T *testing.T) {
	T.Parallel()

	T.Run("ends every login the caller holds", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		user := createServiceUserForTest(t, true, buildUserRegistrationInputForTest(t))

		login := func() *signinpb.IssuedToken {
			res, err := signIn.LoginForToken(ctx, &signinpb.LoginForTokenRequest{
				Credentials: &signinpb.Credentials{
					Username: user.Username,
					Password: user.HashedPassword,
					TotpCode: generateTOTPCodeForUserForTest(t, user),
				},
			})
			require.NoError(t, err)

			return res.GetToken()
		}

		here, elsewhere := login(), login()
		require.NotEqual(t, here.GetFamilyId(), elsewhere.GetFamilyId())

		_, err := signIn.SignOutEverywhere(withBearerToken(ctx, here.GetToken()), &signinpb.SignOutEverywhereRequest{})
		require.NoError(t, err)

		for _, token := range []*signinpb.IssuedToken{here, elsewhere} {
			_, err = signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: token.GetRefreshToken()})
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		}
	})

	T.Run("requires a signed-in caller", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := buildSignInClientForTest(t).SignOutEverywhere(ctx, &signinpb.SignOutEverywhereRequest{})

		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestSignIn_GetAuthStatus(T *testing.T) {
	T.Parallel()

	T.Run("answers an anonymous caller", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		res, err := buildSignInClientForTest(t).GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
		require.NoError(t, err)

		assert.False(t, res.GetAuthenticated())
	})

	T.Run("reads the token a signed-in caller sends", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		token, _ := signInForTest(t, signIn)

		res, err := signIn.GetAuthStatus(withBearerToken(ctx, token.GetToken()), &signinpb.GetAuthStatusRequest{})
		require.NoError(t, err)

		assert.True(t, res.GetAuthenticated())
	})

	T.Run("refuses a token that does not work rather than answering as though none was sent", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Told "not signed in", a client with an expired token would sign its user out.
		// Told Unauthenticated, it refreshes — which is what it should do.
		signIn := buildSignInClientForTest(t)
		token, _ := signInForTest(t, signIn)

		_, err := signIn.GetAuthStatus(withBearerToken(ctx, token.GetToken()+token.GetToken()), &signinpb.GetAuthStatusRequest{})

		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

// TestSignIn_WithheldMethods pins the methods internal/build/signin leaves out. They write a
// password, and platform's server applies no password policy, so reaching them here would let
// through a password AuthService refuses.
func TestSignIn_WithheldMethods(T *testing.T) {
	T.Parallel()

	T.Run("Register is not reachable anonymously", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := buildSignInClientForTest(t).Register(ctx, &signinpb.RegisterRequest{})

		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("UpdatePassword is refused to a signed-in caller", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		token, _ := signInForTest(t, signIn)

		_, err := signIn.UpdatePassword(withBearerToken(ctx, token.GetToken()), &signinpb.UpdatePasswordRequest{})

		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}
