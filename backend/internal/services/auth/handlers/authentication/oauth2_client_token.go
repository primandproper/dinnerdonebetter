package authentication

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/go-oauth2/oauth2/v4"
)

type tokenImpl struct {
	Token types.OAuth2ClientToken
}

func (t *tokenImpl) New() oauth2.TokenInfo {
	return &tokenImpl{}
}

func (t *tokenImpl) GetClientID() string {
	return t.Token.ClientID
}

func (t *tokenImpl) SetClientID(s string) {
	t.Token.ClientID = s
}

func (t *tokenImpl) GetUserID() string {
	return t.Token.BelongsToUser
}

func (t *tokenImpl) SetUserID(s string) {
	t.Token.BelongsToUser = s
}

func (t *tokenImpl) GetRedirectURI() string {
	return t.Token.RedirectURI
}

func (t *tokenImpl) SetRedirectURI(s string) {
	t.Token.RedirectURI = s
}

func (t *tokenImpl) GetScope() string {
	return "n/a"
}

func (t *tokenImpl) SetScope(_ string) {}

func (t *tokenImpl) GetCode() string {
	return t.Token.Code
}

func (t *tokenImpl) SetCode(s string) {
	t.Token.Code = s
}

func (t *tokenImpl) GetCodeCreateAt() time.Time {
	return t.Token.CodeCreatedAt
}

func (t *tokenImpl) SetCodeCreateAt(x time.Time) {
	t.Token.CodeCreatedAt = x
}

func (t *tokenImpl) GetCodeExpiresIn() time.Duration {
	return t.Token.CodeExpiresAt
}

func (t *tokenImpl) SetCodeExpiresIn(duration time.Duration) {
	t.Token.CodeExpiresAt = duration
}

func (t *tokenImpl) GetCodeChallenge() string {
	return t.Token.CodeChallenge
}

func (t *tokenImpl) SetCodeChallenge(s string) {
	t.Token.CodeChallenge = s
}

func (t *tokenImpl) GetCodeChallengeMethod() oauth2.CodeChallengeMethod {
	return oauth2.CodeChallengeMethod(t.Token.CodeChallengeMethod)
}

func (t *tokenImpl) SetCodeChallengeMethod(method oauth2.CodeChallengeMethod) {
	t.Token.CodeChallengeMethod = method.String()
}

func (t *tokenImpl) GetAccess() string {
	return t.Token.Access
}

func (t *tokenImpl) SetAccess(s string) {
	t.Token.Access = s
}

func (t *tokenImpl) GetAccessCreateAt() time.Time {
	return t.Token.AccessCreatedAt
}

func (t *tokenImpl) SetAccessCreateAt(x time.Time) {
	t.Token.AccessCreatedAt = x
}

func (t *tokenImpl) GetAccessExpiresIn() time.Duration {
	return t.Token.AccessExpiresAt
}

func (t *tokenImpl) SetAccessExpiresIn(duration time.Duration) {
	t.Token.AccessExpiresAt = duration
}

func (t *tokenImpl) GetRefresh() string {
	return t.Token.Refresh
}

func (t *tokenImpl) SetRefresh(s string) {
	t.Token.Refresh = s
}

func (t *tokenImpl) GetRefreshCreateAt() time.Time {
	return t.Token.RefreshCreatedAt
}

func (t *tokenImpl) SetRefreshCreateAt(x time.Time) {
	t.Token.RefreshCreatedAt = x
}

func (t *tokenImpl) GetRefreshExpiresIn() time.Duration {
	return t.Token.RefreshExpiresAt
}

func (t *tokenImpl) SetRefreshExpiresIn(duration time.Duration) {
	t.Token.RefreshExpiresAt = duration
}

func convertTokenToImpl(t *types.OAuth2ClientToken) oauth2.TokenInfo {
	return &tokenImpl{Token: *t}
}
