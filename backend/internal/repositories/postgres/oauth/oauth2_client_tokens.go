package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	oauthkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/oauth/generated"

	platformerrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

var _ types.OAuth2ClientTokenDataManager = (*repository)(nil)

// hashToken computes a deterministic blind index for an OAuth2 token value. The code/access/refresh
// columns are stored with randomized (nonce-based) encryption, so they can't be queried directly;
// lookups happen against this HMAC instead.
func (q *repository) hashToken(value string) string {
	mac := hmac.New(sha256.New, q.oauth2ClientTokenHashKey)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetOAuth2ClientTokenByCode fetches an OAuth2 client token from the database.
func (q *repository) GetOAuth2ClientTokenByCode(ctx context.Context, code string) (*types.OAuth2ClientToken, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if code == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientTokenCodeKey, code)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientTokenCodeKey, code)

	result, err := q.generatedQuerier.GetOAuth2ClientTokenByCode(ctx, q.readDB, q.hashToken(code))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client token by code")
	}

	oauth2ClientToken := &types.OAuth2ClientToken{
		RefreshCreatedAt:    result.RefreshCreatedAt,
		AccessCreatedAt:     result.AccessCreatedAt,
		CodeCreatedAt:       result.CodeCreatedAt,
		RedirectURI:         result.RedirectUri,
		Code:                result.Code,
		CodeChallenge:       result.CodeChallenge,
		CodeChallengeMethod: result.CodeChallengeMethod,
		BelongsToUser:       result.BelongsToUser,
		Access:              result.Access,
		ClientID:            result.ClientID,
		Refresh:             result.Refresh,
		ID:                  result.ID,
		CodeExpiresAt:       0,
		AccessExpiresAt:     0,
		RefreshExpiresAt:    0,
	}

	oauth2ClientToken.CodeExpiresAt = result.CodeExpiresAt.Sub(result.CodeCreatedAt)
	oauth2ClientToken.AccessExpiresAt = result.AccessExpiresAt.Sub(result.AccessCreatedAt)
	oauth2ClientToken.RefreshExpiresAt = result.RefreshExpiresAt.Sub(result.RefreshCreatedAt)

	decryptedCode, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Code)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token code")
	}
	oauth2ClientToken.Code = decryptedCode

	decryptedAccess, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Access)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token access")
	}
	oauth2ClientToken.Access = decryptedAccess

	decryptedRefresh, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Refresh)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token refresh")
	}
	oauth2ClientToken.Refresh = decryptedRefresh

	return oauth2ClientToken, nil
}

// GetOAuth2ClientTokenByAccess fetches an OAuth2 client token from the database.
func (q *repository) GetOAuth2ClientTokenByAccess(ctx context.Context, access string) (*types.OAuth2ClientToken, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if access == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientTokenAccessKey, access)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientTokenAccessKey, access)

	result, err := q.generatedQuerier.GetOAuth2ClientTokenByAccess(ctx, q.readDB, q.hashToken(access))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client token by access")
	}

	oauth2ClientToken := &types.OAuth2ClientToken{
		RefreshCreatedAt:    result.RefreshCreatedAt,
		AccessCreatedAt:     result.AccessCreatedAt,
		CodeCreatedAt:       result.CodeCreatedAt,
		RedirectURI:         result.RedirectUri,
		Code:                result.Code,
		CodeChallenge:       result.CodeChallenge,
		CodeChallengeMethod: result.CodeChallengeMethod,
		BelongsToUser:       result.BelongsToUser,
		Access:              result.Access,
		ClientID:            result.ClientID,
		Refresh:             result.Refresh,
		ID:                  result.ID,
		CodeExpiresAt:       0,
		AccessExpiresAt:     0,
		RefreshExpiresAt:    0,
	}

	oauth2ClientToken.CodeExpiresAt = result.CodeExpiresAt.Sub(result.CodeCreatedAt)
	oauth2ClientToken.AccessExpiresAt = result.AccessExpiresAt.Sub(result.AccessCreatedAt)
	oauth2ClientToken.RefreshExpiresAt = result.RefreshExpiresAt.Sub(result.RefreshCreatedAt)

	decryptedCode, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Code)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token code")
	}
	oauth2ClientToken.Code = decryptedCode

	decryptedAccess, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Access)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token access")
	}
	oauth2ClientToken.Access = decryptedAccess

	decryptedRefresh, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Refresh)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token refresh")
	}
	oauth2ClientToken.Refresh = decryptedRefresh

	return oauth2ClientToken, nil
}

// GetOAuth2ClientTokenByRefresh fetches an OAuth2 client token from the database.
func (q *repository) GetOAuth2ClientTokenByRefresh(ctx context.Context, refresh string) (*types.OAuth2ClientToken, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if refresh == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientTokenRefreshKey, refresh)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientTokenRefreshKey, refresh)

	result, err := q.generatedQuerier.GetOAuth2ClientTokenByRefresh(ctx, q.readDB, q.hashToken(refresh))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting oauth2 client token by refresh")
	}

	oauth2ClientToken := &types.OAuth2ClientToken{
		RefreshCreatedAt:    result.RefreshCreatedAt,
		AccessCreatedAt:     result.AccessCreatedAt,
		CodeCreatedAt:       result.CodeCreatedAt,
		RedirectURI:         result.RedirectUri,
		Code:                result.Code,
		CodeChallenge:       result.CodeChallenge,
		CodeChallengeMethod: result.CodeChallengeMethod,
		BelongsToUser:       result.BelongsToUser,
		Access:              result.Access,
		ClientID:            result.ClientID,
		Refresh:             result.Refresh,
		ID:                  result.ID,
		CodeExpiresAt:       0,
		AccessExpiresAt:     0,
		RefreshExpiresAt:    0,
	}

	oauth2ClientToken.CodeExpiresAt = result.CodeExpiresAt.Sub(result.CodeCreatedAt)
	oauth2ClientToken.AccessExpiresAt = result.AccessExpiresAt.Sub(result.AccessCreatedAt)
	oauth2ClientToken.RefreshExpiresAt = result.RefreshExpiresAt.Sub(result.RefreshCreatedAt)

	decryptedCode, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Code)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token code")
	}
	oauth2ClientToken.Code = decryptedCode

	decryptedAccess, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Access)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token access")
	}
	oauth2ClientToken.Access = decryptedAccess

	decryptedRefresh, err := q.oauth2ClientTokenEncDec.Decrypt(ctx, result.Refresh)
	if err != nil {
		return nil, observability.PrepareError(err, span, "decrypting oauth2 token refresh")
	}
	oauth2ClientToken.Refresh = decryptedRefresh

	return oauth2ClientToken, nil
}

// CreateOAuth2ClientToken creates an OAuth2 client token in the database.
func (q *repository) CreateOAuth2ClientToken(ctx context.Context, input *types.OAuth2ClientTokenDatabaseCreationInput) (*types.OAuth2ClientToken, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValue(oauthkeys.OAuth2ClientTokenIDKey, input.ID)
	now := q.CurrentTime()

	encryptedCode, err := q.oauth2ClientTokenEncDec.Encrypt(ctx, input.Code)
	if err != nil {
		return nil, observability.PrepareError(err, span, "encrypting oauth2 token code")
	}

	encryptedAccess, err := q.oauth2ClientTokenEncDec.Encrypt(ctx, input.Access)
	if err != nil {
		return nil, observability.PrepareError(err, span, "encrypting oauth2 token access")
	}

	encryptedRefresh, err := q.oauth2ClientTokenEncDec.Encrypt(ctx, input.Refresh)
	if err != nil {
		return nil, observability.PrepareError(err, span, "encrypting oauth2 token refresh")
	}

	// create the oauth2 client token.
	if err = q.generatedQuerier.CreateOAuth2ClientToken(ctx, q.writeDB, &generated.CreateOAuth2ClientTokenParams{
		AccessExpiresAt:     now.Add(input.AccessExpiresIn),
		CodeExpiresAt:       now.Add(input.CodeExpiresIn),
		RefreshExpiresAt:    now.Add(input.RefreshExpiresIn),
		RefreshCreatedAt:    now,
		CodeCreatedAt:       now,
		AccessCreatedAt:     now,
		CodeChallenge:       input.CodeChallenge,
		CodeChallengeMethod: input.CodeChallengeMethod,
		ClientID:            input.ClientID,
		Access:              encryptedAccess,
		Code:                encryptedCode,
		ID:                  input.ID,
		Refresh:             encryptedRefresh,
		RedirectUri:         input.RedirectURI,
		BelongsToUser:       input.BelongsToUser,
		CodeHash:            q.hashToken(input.Code),
		AccessHash:          q.hashToken(input.Access),
		RefreshHash:         q.hashToken(input.Refresh),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing oauth2 client token creation query")
	}

	logger.Info("oauth2 client token created")

	oauth2ClientToken := &types.OAuth2ClientToken{
		RefreshCreatedAt:    input.RefreshCreatedAt,
		AccessCreatedAt:     input.AccessCreatedAt,
		CodeCreatedAt:       input.CodeCreatedAt,
		RedirectURI:         input.RedirectURI,
		Code:                input.Code,
		CodeChallenge:       input.CodeChallenge,
		CodeChallengeMethod: input.CodeChallengeMethod,
		BelongsToUser:       input.BelongsToUser,
		Access:              input.Access,
		ClientID:            input.ClientID,
		Refresh:             input.Refresh,
		ID:                  input.ID,
		CodeExpiresAt:       input.CodeExpiresIn,
		AccessExpiresAt:     input.AccessExpiresIn,
		RefreshExpiresAt:    input.RefreshExpiresIn,
	}

	return oauth2ClientToken, nil
}

// DeleteOAuth2ClientTokenByAccess deletes an OAuth2 client token from the database by its access token.
func (q *repository) DeleteOAuth2ClientTokenByAccess(ctx context.Context, access string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if access == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientTokenAccessKey, access)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientTokenAccessKey, access)

	if _, err := q.generatedQuerier.DeleteOAuth2ClientTokenByAccess(ctx, q.writeDB, q.hashToken(access)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "deleting oauth2 client token by access")
	}

	return nil
}

// DeleteOAuth2ClientTokenByCode deletes an OAuth2 client token from the database by its authorization code.
func (q *repository) DeleteOAuth2ClientTokenByCode(ctx context.Context, code string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if code == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientTokenCodeKey, code)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientTokenCodeKey, code)

	if _, err := q.generatedQuerier.DeleteOAuth2ClientTokenByCode(ctx, q.writeDB, q.hashToken(code)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "deleting oauth2 client token by code")
	}

	return nil
}

// DeleteOAuth2ClientTokenByRefresh deletes an OAuth2 client token from the database by its refresh token.
func (q *repository) DeleteOAuth2ClientTokenByRefresh(ctx context.Context, refresh string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if refresh == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(oauthkeys.OAuth2ClientTokenRefreshKey, refresh)
	tracing.AttachToSpan(span, oauthkeys.OAuth2ClientTokenRefreshKey, refresh)

	if _, err := q.generatedQuerier.DeleteOAuth2ClientTokenByRefresh(ctx, q.writeDB, q.hashToken(refresh)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "deleting oauth2 client token by refresh")
	}

	return nil
}
