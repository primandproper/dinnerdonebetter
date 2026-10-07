package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/devices"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	"github.com/primandproper/primitives-go/v2/identifiers"

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
		fmt.Sprintf(":%d", apiServiceConfig.Service.GRPCServer.Port),
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

	user := createServiceUserForTest(t, buildUserRegistrationInputForTest(t))

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

// TestSignIn_ThisApplicationsRules pins what this application adds to platform's doors: its
// password floor on a password written through them, and its registration policy on Register,
// which is open to anybody. The conformance suites assert the doors themselves; these are the
// rules no suite can know.
func TestSignIn_ThisApplicationsRules(T *testing.T) {
	T.Parallel()

	T.Run("Register is open to somebody with no session", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := buildSignInClientForTest(t).Register(ctx, registrationForTest(true))
		require.NoError(t, err)
	})

	T.Run("Register reads a signed-in caller's token, and refuses one that no longer works", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		token, _ := signInForTest(t, signIn)

		_, err := signIn.Register(withBearerToken(ctx, token.GetToken()), registrationForTest(true))
		require.NoError(t, err)

		_, err = signIn.SignOut(ctx, &signinpb.SignOutRequest{RefreshToken: token.GetRefreshToken()})
		require.NoError(t, err)

		_, err = signIn.Register(withBearerToken(ctx, token.GetToken()), registrationForTest(true))
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("a registration that accepts no agreements is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := buildSignInClientForTest(t).Register(ctx, registrationForTest(false))

		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("a registration that chooses no password is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		request := registrationForTest(true)
		request.Credential = &signinpb.RegisterRequest_NoPassword{NoPassword: &signinpb.NoPassword{}}

		_, err := buildSignInClientForTest(t).Register(ctx, request)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("a registration is shaped by this application's policy", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		res, err := buildSignInClientForTest(t).Register(ctx, registrationForTest(true))
		require.NoError(t, err)

		registered := res.GetRegistration()
		assert.Equal(t, []string{authorization.ServiceUserRoleName}, registered.GetUser().GetServiceRoles())
		assert.Equal(t, []string{authorization.AccountAdminRoleName}, registered.GetMembership().GetRoles(),
			"a registrant owns their account with this application's owner role")
		assert.NotEmpty(t, registered.GetTotpEnrollment().GetSecret(), "every registrant is issued a second factor")
		assert.NotNil(t, registered.GetUser().GetLastAcceptedTermsOfService())
		assert.NotNil(t, registered.GetUser().GetLastAcceptedPrivacyPolicy())
	})

	T.Run("a password change through platform's door answers to this application's floor", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		user := createServiceUserForTest(t, buildUserRegistrationInputForTest(t))
		token, err := signIn.LoginForToken(ctx, &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{
				Username: user.Username,
				Password: user.HashedPassword,
				TotpCode: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)

		_, err = signIn.UpdatePassword(withBearerToken(ctx, token.GetToken().GetToken()), &signinpb.UpdatePasswordRequest{
			CurrentPassword: user.HashedPassword,
			NewPassword:     "password",
			TotpCode:        generateTOTPCodeForUserForTest(t, user),
		})

		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

// TestSignIn_AccountPasswordPolicy pins the half of the password rule that needs the account.
func TestSignIn_AccountPasswordPolicy(T *testing.T) {
	T.Parallel()

	T.Run("a password change may not keep the password it is changing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		user := createServiceUserForTest(t, buildUserRegistrationInputForTest(t))
		token, err := signIn.LoginForToken(ctx, &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{
				Username: user.Username,
				Password: user.HashedPassword,
				TotpCode: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)

		_, err = signIn.UpdatePassword(withBearerToken(ctx, token.GetToken().GetToken()), &signinpb.UpdatePasswordRequest{
			CurrentPassword: user.HashedPassword,
			NewPassword:     user.HashedPassword,
			TotpCode:        generateTOTPCodeForUserForTest(t, user),
		})

		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, status.Convert(err).Message(), "current password")
	})
}

// TestSignIn_DevicesAreListedBesideTheirLogins pins "where you're signed in": the device a token
// was issued to is recorded on the token's own transaction and answered as the login's attributes.
func TestSignIn_DevicesAreListedBesideTheirLogins(T *testing.T) {
	T.Parallel()

	T.Run("a login lists the device it was signed in from", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		user := createServiceUserForTest(t, buildUserRegistrationInputForTest(t))

		address, userAgent, deviceName := "203.0.113.42", "integration-test/"+identifiers.New(), "test device "+identifiers.New()

		// As a web app forwards the browser it is serving, and the iOS app names itself.
		signingIn := metadata.AppendToOutgoingContext(ctx,
			devices.ClientAddressMetadataKey, address,
			devices.ClientUserAgentMetadataKey, userAgent,
			devices.DeviceNameMetadataKey, deviceName,
		)

		token, err := signIn.LoginForToken(signingIn, &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{
				Username: user.Username,
				Password: user.HashedPassword,
				TotpCode: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)

		listed, err := signIn.ListSignIns(withBearerToken(ctx, token.GetToken().GetToken()), &signinpb.ListSignInsRequest{})
		require.NoError(t, err)

		var current *signinpb.ActiveSignIn
		for _, candidate := range listed.GetSignIns() {
			if candidate.GetCurrent() {
				current = candidate
			}
		}
		require.NotNil(t, current, "the login the listing was made through is listed")

		assert.Equal(t, map[string]string{
			devices.AttributeIPAddress:  address,
			devices.AttributeUserAgent:  userAgent,
			devices.AttributeDeviceName: deviceName,
		}, current.GetAttributes())
	})
}

// registrationForTest is a registration for somebody nobody has registered.
func registrationForTest(agreeing bool) *signinpb.RegisterRequest {
	username := "reg_" + identifiers.New()

	request := &signinpb.RegisterRequest{
		User: &identitypb.UserRegistrationInput{
			Username:     username,
			EmailAddress: username + "@example.invalid",
		},
		Account:    &identitypb.AccountCreationInput{},
		Credential: &signinpb.RegisterRequest_Password{Password: identifiers.New() + identifiers.New()},
	}

	if agreeing {
		request.Agreements = []identitypb.Agreement{
			identitypb.Agreement_AGREEMENT_TERMS_OF_SERVICE,
			identitypb.Agreement_AGREEMENT_PRIVACY_POLICY,
		}
	}

	return request
}
