package auth

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/converters"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createPasswordResetTokenForTest(t *testing.T, ctx context.Context, examplePasswordResetToken *auth.PasswordResetToken, dbc *repository) *auth.PasswordResetToken {
	t.Helper()

	// create
	if examplePasswordResetToken == nil {
		examplePasswordResetToken = authfakes.BuildFakePasswordResetToken()
	}
	dbInput := authconverters.ConvertPasswordResetTokenToPasswordResetTokenDatabaseCreationInput(examplePasswordResetToken)

	created, err := dbc.CreatePasswordResetToken(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	examplePasswordResetToken.CreatedAt = created.CreatedAt
	assert.Equal(t, examplePasswordResetToken, created)

	passwordResetToken, err := dbc.GetPasswordResetTokenByToken(ctx, created.Token)
	examplePasswordResetToken.CreatedAt = passwordResetToken.CreatedAt
	examplePasswordResetToken.ExpiresAt = passwordResetToken.ExpiresAt

	assert.NoError(t, err)
	assert.Equal(t, passwordResetToken, examplePasswordResetToken)

	return created
}

func TestQuerier_Integration_PasswordResetTokens(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	examplePasswordResetToken := authfakes.BuildFakePasswordResetToken()
	examplePasswordResetToken.BelongsToUser = user.ID
	createdPasswordResetTokens := []*auth.PasswordResetToken{}

	// create
	created := createPasswordResetTokenForTest(t, ctx, examplePasswordResetToken, dbc)
	createdPasswordResetTokens = append(createdPasswordResetTokens, created)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePasswordResetTokens, RelevantID: created.ID},
	})

	// create more
	for range exampleQuantity {
		input := authfakes.BuildFakePasswordResetToken()
		input.BelongsToUser = user.ID
		createdPasswordResetTokens = append(createdPasswordResetTokens, createPasswordResetTokenForTest(t, ctx, input, dbc))
	}

	// redeem (update)
	for _, passwordResetToken := range createdPasswordResetTokens {
		assert.NoError(t, dbc.RedeemPasswordResetToken(ctx, passwordResetToken.ID))
	}
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePasswordResetTokens, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypePasswordResetTokens, RelevantID: created.ID},
	})
}

func TestSQLQuerier_GetPasswordResetTokenByToken(T *testing.T) {
	T.Parallel()

	T.Run("with missing token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetPasswordResetTokenByToken(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestSQLQuerier_CreatePasswordResetToken(T *testing.T) {
	T.Parallel()

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreatePasswordResetToken(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestSQLQuerier_RedeemPasswordResetToken(T *testing.T) {
	T.Parallel()

	T.Run("with missing ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual := c.RedeemPasswordResetToken(ctx, "")
		assert.Error(t, actual)
	})
}
