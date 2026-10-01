package grpc

import (
	"context"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/converters"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/errors"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"

	"google.golang.org/grpc/codes"
)

// CredentialKindAccountSwitch is how a login moved into another account by ExchangeToken is
// recorded: the refresh token it spent is what proved the person, and nothing else was asked.
const CredentialKindAccountSwitch signin.CredentialKind = "account_switch"

// errNoSecondFactorMinted is a registration that came back without the second factor
// authentication.RegistrationPolicy asks for — a sign-in service built without that policy.
// It is a wiring failure, and refusing to answer is better than answering a registrant with
// an empty secret they would be asked for at their first sign-in.
var errNoSecondFactorMinted = platformerrors.New("registration minted no second factor")

// RegisterUser signs somebody up: a user, the credentials they will sign in with, and either
// the account they now own or the one the invitation they answered admits them to.
//
// It is signin.Service.Register with no registrar, which platform's SignInService does not
// offer yet (platform-go#1068). Who may sign up, and as what — standing, roles, the second
// factor, the agreements — is authentication.RegistrationPolicy's, which the sign-in service
// applies to every registration it writes; the password is authentication.PasswordPolicy's.
// What this adds is the wire: the agreements as two booleans, and the QR code rendered from
// the second factor.
func (s *serviceImpl) RegisterUser(ctx context.Context, request *authsvc.RegisterUserRequest) (*authsvc.RegisterUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertGRPCUserRegistrationInputToUserRegistrationInput(request.GetInput())
	if input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.ErrNilInputParameter, s.logger, span, codes.InvalidArgument, "registering user")
	}

	// Trimmed before validation rather than after, so the address the email rule parses is
	// the address the column receives. The handle is not folded here — the store folds it on
	// write and on every lookup, and a second fold in front of it is a second place for the
	// two to disagree.
	input.Username = strings.TrimSpace(input.Username)
	input.EmailAddress = strings.TrimSpace(strings.ToLower(input.EmailAddress))
	input.Password = strings.TrimSpace(input.Password)

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UsernameKey:            input.Username,
		identitykeys.UserEmailAddressKey:    input.EmailAddress,
		identitykeys.AccountInvitationIDKey: input.InvitationID,
	}, span, s.logger)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid registration input provided")
	}

	// The agreements the registrant accepted, which RegistrationPolicy refuses the
	// registration without.
	var agreements []platformidentity.Agreement
	if input.AcceptedTOS {
		agreements = append(agreements, platformidentity.TermsOfService)
	}
	if input.AcceptedPrivacyPolicy {
		agreements = append(agreements, platformidentity.PrivacyPolicy)
	}

	registered, err := s.signIns.Register(ctx, ddbidentity.Scope(), &signin.Registration{
		User: &platformidentity.User{
			ID:           identifiers.New(),
			Username:     input.Username,
			EmailAddress: input.EmailAddress,
			FirstName:    input.FirstName,
			LastName:     input.LastName,
		},
		Account:         &platformidentity.Account{Name: input.AccountName},
		Credential:      signin.Password(input.Password),
		InvitationID:    input.InvitationID,
		InvitationToken: input.InvitationToken,
		Agreements:      agreements,
	})
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "registering user")
	}

	if registered.TOTPEnrollment == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNoSecondFactorMinted, logger, span, codes.Internal, "registering user")
	}

	user := registered.User
	twoFactorSecret := registered.TOTPEnrollment.Secret

	// The QR code is a convenience rendering of the secret, which is returned beside it;
	// don't fail a signup over it.
	twoFactorQRCode, qrCodeErr := s.qrCodes.BuildQRCode(ctx, user.Username, twoFactorSecret)
	if qrCodeErr != nil {
		observability.AcknowledgeError(qrCodeErr, logger, span, "building two factor QR code")
	}

	var accountID string
	if registered.Membership != nil {
		accountID = registered.Membership.BelongsToAccount
	}

	// No signup event is published here: the registration's hook writes user_signed_up into
	// the outbox on the transaction that made the user. See
	// internal/repositories/postgres/identitystore/hooks.go.

	return &authsvc.RegisterUserResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertUserCreationResponseToGRPCUserCreationResponse(&auth.UserCreationResponse{
			CreatedAt:        user.CreatedAt,
			Username:         user.Username,
			EmailAddress:     user.EmailAddress,
			TwoFactorQRCode:  twoFactorQRCode,
			CreatedUserID:    user.ID,
			CreatedAccountID: accountID,
			AccountStatus:    string(user.AccountStatus),
			TwoFactorSecret:  twoFactorSecret,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
		}),
	}, nil
}

// ExchangeToken spends a refresh token SignInService issued and answers with a pair for the
// account the caller asked for, which is how somebody in two households moves between them.
//
// It is here until SignInService can do it itself (platform-go#1069), and it is built from the
// doors that exist, in an order that fails safe:
//
//  1. The refresh token is exchanged, exactly as SignInService.ExchangeRefreshToken would.
//     That is the proof — of who is asking, that their login is live, and that the token was
//     not spent before — and it rotates the login, so the presented token is retired whatever
//     happens next.
//  2. If the account asked for is the login's own, or none was named, the successor is the
//     answer. A refresh is a refresh.
//  3. Otherwise a new login is issued on that account through IssueForPrincipal, which reads
//     the membership and refuses an account the user is not a live member of. The refresh
//     token was the request's whole authority and proved the person as completely as their
//     sign-in did, so the new login asks for no second factor; an administrative login stays
//     administrative.
//  4. The login the exchange rotated is ended with its successor, so the switch leaves one
//     login rather than two. A switch that was refused ends it too: its successor was never
//     handed back, and a live login nobody holds is worse than one more sign-in.
//
// What platform's version will do better is keep the login one login across the switch —
// the family, its listing in ListSignIns, its reuse detection. Here a switch is a new login.
func (s *serviceImpl) ExchangeToken(ctx context.Context, request *authsvc.ExchangeTokenRequest) (*authsvc.ExchangeTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	exchanged, err := s.signIns.ExchangeRefreshToken(ctx, ddbidentity.Scope(), request.GetRefreshToken())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "exchanging refresh token")
	}

	desired := strings.TrimSpace(request.GetDesiredAccountId())
	if desired == "" || desired == exchanged.Principal.ActiveAccountID {
		return exchangeTokenResponse(span.SpanContext().TraceID().String(), exchanged), nil
	}

	userID := exchanged.Principal.User.ID
	logger = logger.WithValue(identitykeys.UserIDKey, userID).WithValue(identitykeys.AccountIDKey, desired)

	opts := []signin.IssueOption{signin.WithCredentialKind(CredentialKindAccountSwitch), signin.MultiFactor()}
	if exchanged.Administrative {
		opts = append(opts, signin.Administrative())
	}

	switched, switchErr := s.signIns.IssueForPrincipal(ctx, ddbidentity.Scope(), userID, desired, opts...)

	if err = s.signIns.SignOut(ctx, ddbidentity.Scope(), exchanged.RefreshToken); err != nil {
		observability.AcknowledgeError(err, logger, span, "ending the login an account switch replaced")
	}

	if switchErr != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(switchErr, logger, span, codes.Internal, "switching accounts")
	}

	return exchangeTokenResponse(span.SpanContext().TraceID().String(), switched), nil
}

func exchangeTokenResponse(traceID string, signedIn *signin.SignIn) *authsvc.ExchangeTokenResponse {
	x := &authsvc.ExchangeTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: traceID,
		},
		AccessToken:  signedIn.Token,
		RefreshToken: signedIn.RefreshToken,
		ExpiresUtc:   grpcconverters.ConvertTimeToPBTimestamp(signedIn.ExpiresAt),
	}

	if signedIn.Principal != nil {
		x.AccountId = signedIn.Principal.ActiveAccountID
		if signedIn.Principal.User != nil {
			x.UserId = signedIn.Principal.User.ID
		}
	}

	return x
}
