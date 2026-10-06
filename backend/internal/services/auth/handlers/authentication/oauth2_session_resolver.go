package authentication

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/primandproper/platform-go/v15/authentication/oauth2clients/authserver"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// SignInAuthenticator turns a sign-in token into the caller it names — signingrpc's
// PrincipalExtractor.Authenticate.
type SignInAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*signingrpc.Caller, error)
}

var _ SignInAuthenticator = (*signingrpc.PrincipalExtractor)(nil)

// sessionResolver identifies the resource owner behind an authorization request that already
// carries a sign-in: a first-party client — the web app, the iOS app, pkg/client — presenting the
// token it holds in an Authorization header, which should not be asked for a password it proved a
// moment ago.
//
// It answers through the extractor every API request is resolved by, so a token is spent here on
// exactly the terms it is spent anywhere else: its signature, its login still live, and its user
// still admitted by the directory. That is what the hand-rolled authenticator this replaced did
// not do — it checked a signature and nothing more, so a signed-out session and a banned user both
// got a code, and the latter whenever the token named an account, which every sign-in token does.
//
// It is wrapped in authserver.NewGuardedResolver, which puts oauth2clients.Client.Admits on this
// path to a code as the login form's authenticator puts it on the other.
type sessionResolver struct {
	signIns SignInAuthenticator
}

var _ authserver.ScopedSubjectResolver = (*sessionResolver)(nil)

// NewSessionResolver builds the resolver for a sign-in a request already carries.
func NewSessionResolver(signIns SignInAuthenticator) authserver.ScopedSubjectResolver {
	return &sessionResolver{signIns: signIns}
}

// ResolveScopedSubject implements authserver.ScopedSubjectResolver.
//
// A request with no bearer token is not one of this resolver's, and goes on to the login form. So
// does one whose token the extractor refuses — expired, signed out, naming a user the directory no
// longer admits, or not a sign-in token at all — because the honest answer to "I cannot say who
// this is" is the form, where a banned or signed-out person is refused again by signin itself.
//
// So does an impersonation token. An operator acting as somebody holds a fifteen-minute token with
// no refresh, deliberately; converting it into an OAuth2 grant would hand them a credential that
// outlives the impersonation by weeks, in the subject's name.
//
// A directory or store that could not be read is an error, which ends the attempt: that is not a
// refused credential, and no form fixes it.
func (r *sessionResolver) ResolveScopedSubject(ctx context.Context, req *http.Request) (*oauth2server.Subject, tenancy.Scope, error) {
	bearer := bearerTokenFrom(req)
	if bearer == "" {
		return nil, tenancy.Scope{}, nil
	}

	caller, err := r.signIns.Authenticate(ctx, bearer)
	if err != nil {
		if errors.Is(err, signingrpc.ErrUnauthenticated) {
			return nil, tenancy.Scope{}, nil
		}

		return nil, tenancy.Scope{}, err
	}

	if caller.ActorID() != "" {
		return nil, tenancy.Scope{}, nil
	}

	return &oauth2server.Subject{
		ID:     caller.UserID(),
		Claims: map[string]string{authserver.ClaimAccountID: caller.ActiveAccountID()},
	}, caller.Scope(), nil
}

// bearerTokenFrom reads a bearer credential off a request, or returns empty.
func bearerTokenFrom(req *http.Request) string {
	header := req.Header.Get("Authorization")
	if header == "" {
		return ""
	}

	if token, found := strings.CutPrefix(header, "Bearer "); found {
		return strings.TrimSpace(token)
	}

	return ""
}
