package grpc

import (
	"context"
	"testing"
	"time"

	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	qrcodesnoop "github.com/primandproper/primitives-go/v2/qrcodes/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeSignIns is a SignIns that answers from its fields and records what it was asked.
type fakeSignIns struct {
	registerErr  error
	exchangeErr  error
	issueErr     error
	registered   *signin.Registered
	exchanged    *signin.SignIn
	issued       *signin.SignIn
	registration *signin.Registration
	issuedFor    string
	issuedIn     string
	signedOut    []string
	issueOptions int
}

func (f *fakeSignIns) Register(_ context.Context, _ tenancy.Scope, registration *signin.Registration) (*signin.Registered, error) {
	f.registration = registration
	return f.registered, f.registerErr
}

func (f *fakeSignIns) ExchangeRefreshToken(context.Context, tenancy.Scope, string) (*signin.SignIn, error) {
	return f.exchanged, f.exchangeErr
}

func (f *fakeSignIns) IssueForPrincipal(_ context.Context, _ tenancy.Scope, userID, activeAccountID string, opts ...signin.IssueOption) (*signin.SignIn, error) {
	f.issuedFor, f.issuedIn, f.issueOptions = userID, activeAccountID, len(opts)
	return f.issued, f.issueErr
}

func (f *fakeSignIns) SignOut(_ context.Context, _ tenancy.Scope, refreshToken string) error {
	f.signedOut = append(f.signedOut, refreshToken)
	return nil
}

func buildTestService(signIns SignIns) *serviceImpl {
	return NewAuthService(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), signIns, qrcodesnoop.NewBuilder()).(*serviceImpl)
}

func signInFor(userID, accountID string) *signin.SignIn {
	return &signin.SignIn{
		Principal:    &platformidentity.Principal{User: &platformidentity.User{ID: userID}, ActiveAccountID: accountID},
		Token:        identifiers.New(),
		RefreshToken: identifiers.New(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}
}

func validRegistrationInput() *authsvc.UserRegistrationInput {
	return &authsvc.UserRegistrationInput{
		Username:              "  someone  ",
		EmailAddress:          " Someone@Example.com ",
		Password:              "a password with plenty of entropy",
		AccountName:           "household",
		AcceptedTos:           true,
		AcceptedPrivacyPolicy: true,
	}
}

func TestServiceImpl_RegisterUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID, accountID := identifiers.New(), identifiers.New()
		signIns := &fakeSignIns{registered: &signin.Registered{
			User:           &platformidentity.User{ID: userID, Username: "someone", AccountStatus: platformidentity.StatusGood},
			Membership:     &platformidentity.Membership{BelongsToAccount: accountID},
			TOTPEnrollment: &totp.Enrollment{Secret: "SECRET"},
		}}

		res, err := buildTestService(signIns).RegisterUser(t.Context(), &authsvc.RegisterUserRequest{Input: validRegistrationInput()})
		require.NoError(t, err)

		assert.Equal(t, userID, res.GetCreated().GetCreatedUserId())
		assert.Equal(t, accountID, res.GetCreated().GetCreatedAccountId())
		assert.Equal(t, "SECRET", res.GetCreated().GetTwoFactorSecret())

		// The handle and address are trimmed, the address folded, before signin sees them.
		assert.Equal(t, "someone", signIns.registration.User.Username)
		assert.Equal(t, "someone@example.com", signIns.registration.User.EmailAddress)
		assert.ElementsMatch(t, []platformidentity.Agreement{platformidentity.TermsOfService, platformidentity.PrivacyPolicy}, signIns.registration.Agreements)
	})

	T.Run("refuses an input that does not validate before asking signin", func(t *testing.T) {
		t.Parallel()

		signIns := &fakeSignIns{}
		input := validRegistrationInput()
		input.EmailAddress = "not an address"

		_, err := buildTestService(signIns).RegisterUser(t.Context(), &authsvc.RegisterUserRequest{Input: input})
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, signIns.registration)
	})

	T.Run("refuses a missing input", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(&fakeSignIns{}).RegisterUser(t.Context(), &authsvc.RegisterUserRequest{})
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("refuses a registration missing an agreement before asking signin", func(t *testing.T) {
		t.Parallel()

		signIns := &fakeSignIns{}
		input := validRegistrationInput()
		input.AcceptedPrivacyPolicy = false

		_, err := buildTestService(signIns).RegisterUser(t.Context(), &authsvc.RegisterUserRequest{Input: input})
		require.Error(t, err)
		assert.Nil(t, signIns.registration)
	})

	T.Run("a registration with no second factor is a wiring failure", func(t *testing.T) {
		t.Parallel()

		signIns := &fakeSignIns{registered: &signin.Registered{User: &platformidentity.User{ID: identifiers.New()}}}

		_, err := buildTestService(signIns).RegisterUser(t.Context(), &authsvc.RegisterUserRequest{Input: validRegistrationInput()})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestServiceImpl_ExchangeToken(T *testing.T) {
	T.Parallel()

	T.Run("with no account named, is a refresh", func(t *testing.T) {
		t.Parallel()

		exchanged := signInFor(identifiers.New(), identifiers.New())
		signIns := &fakeSignIns{exchanged: exchanged}

		res, err := buildTestService(signIns).ExchangeToken(t.Context(), &authsvc.ExchangeTokenRequest{RefreshToken: identifiers.New()})
		require.NoError(t, err)

		assert.Equal(t, exchanged.Token, res.GetAccessToken())
		assert.Equal(t, exchanged.RefreshToken, res.GetRefreshToken())
		assert.Equal(t, exchanged.Principal.ActiveAccountID, res.GetAccountId())
		assert.Empty(t, signIns.issuedFor)
		assert.Empty(t, signIns.signedOut)
	})

	T.Run("naming the login's own account is a refresh", func(t *testing.T) {
		t.Parallel()

		exchanged := signInFor(identifiers.New(), identifiers.New())
		signIns := &fakeSignIns{exchanged: exchanged}

		res, err := buildTestService(signIns).ExchangeToken(t.Context(), &authsvc.ExchangeTokenRequest{
			RefreshToken:     identifiers.New(),
			DesiredAccountId: exchanged.Principal.ActiveAccountID,
		})
		require.NoError(t, err)

		assert.Equal(t, exchanged.Token, res.GetAccessToken())
		assert.Empty(t, signIns.issuedFor)
	})

	T.Run("switches to another account, ending the login it replaced", func(t *testing.T) {
		t.Parallel()

		userID, otherAccount := identifiers.New(), identifiers.New()
		exchanged := signInFor(userID, identifiers.New())
		switched := signInFor(userID, otherAccount)
		signIns := &fakeSignIns{exchanged: exchanged, issued: switched}

		res, err := buildTestService(signIns).ExchangeToken(t.Context(), &authsvc.ExchangeTokenRequest{
			RefreshToken:     identifiers.New(),
			DesiredAccountId: otherAccount,
		})
		require.NoError(t, err)

		assert.Equal(t, switched.Token, res.GetAccessToken())
		assert.Equal(t, switched.RefreshToken, res.GetRefreshToken())
		assert.Equal(t, otherAccount, res.GetAccountId())
		assert.Equal(t, userID, signIns.issuedFor)
		assert.Equal(t, otherAccount, signIns.issuedIn)
		assert.Equal(t, []string{exchanged.RefreshToken}, signIns.signedOut)
	})

	T.Run("an administrative login stays administrative", func(t *testing.T) {
		t.Parallel()

		exchanged := signInFor(identifiers.New(), identifiers.New())
		exchanged.Administrative = true
		signIns := &fakeSignIns{exchanged: exchanged, issued: signInFor(identifiers.New(), identifiers.New())}

		_, err := buildTestService(signIns).ExchangeToken(t.Context(), &authsvc.ExchangeTokenRequest{
			RefreshToken:     identifiers.New(),
			DesiredAccountId: identifiers.New(),
		})
		require.NoError(t, err)

		// The credential kind, the multi-factor waiver, and the administrative door.
		assert.Equal(t, 3, signIns.issueOptions)
	})

	T.Run("a refused switch ends the login it would have replaced", func(t *testing.T) {
		t.Parallel()

		exchanged := signInFor(identifiers.New(), identifiers.New())
		signIns := &fakeSignIns{exchanged: exchanged, issueErr: platformidentity.ErrMembershipNotFound}

		_, err := buildTestService(signIns).ExchangeToken(t.Context(), &authsvc.ExchangeTokenRequest{
			RefreshToken:     identifiers.New(),
			DesiredAccountId: identifiers.New(),
		})
		require.Error(t, err)
		assert.Equal(t, []string{exchanged.RefreshToken}, signIns.signedOut)
	})

	T.Run("a refused exchange switches nothing", func(t *testing.T) {
		t.Parallel()

		signIns := &fakeSignIns{exchangeErr: signin.ErrInvalidCredentials}

		_, err := buildTestService(signIns).ExchangeToken(t.Context(), &authsvc.ExchangeTokenRequest{
			RefreshToken:     identifiers.New(),
			DesiredAccountId: identifiers.New(),
		})
		require.Error(t, err)
		assert.Empty(t, signIns.issuedFor)
		assert.Empty(t, signIns.signedOut)
	})
}
