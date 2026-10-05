package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildEmitterForTest(t *testing.T) *events.Emitter {
	t.Helper()

	writer, err := outbox.NewWriter(dialect.Postgres)
	require.NoError(t, err)

	return events.NewEmitter(writer, t.Name(), nil, nil)
}

func TestPasswordResetDirectory_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("writes the password and queues the reset mail on the same transaction", func(t *testing.T) {
		t.Parallel()

		harness := buildSignInHooksHarness(t, nil)
		userID := identifiers.New()

		store := &identitymock.StoreMock{
			UpdateUserPasswordFunc: func(_ context.Context, tx database.Tx, _ tenancy.Scope, id, _ string) error {
				assert.Equal(t, harness.tx, tx)
				assert.Equal(t, userID, id)
				return nil
			},
		}

		directory := NewPasswordResetDirectory(loggingnoop.NewLogger(), store, buildEmitterForTest(t))

		require.NoError(t, directory.UpdateUserPassword(t.Context(), harness.tx, tenancy.Global(), userID, "hashed"))

		enqueued := harness.enqueued(t)
		require.Len(t, enqueued, 1)
		assert.Equal(t, ddbidentity.PasswordResetTokenRedeemedEventType, enqueued[0].EventType)
		assert.Equal(t, userID, enqueued[0].UserID)
	})

	T.Run("queues nothing when the write fails", func(t *testing.T) {
		t.Parallel()

		harness := buildSignInHooksHarness(t, nil)
		writeErr := errors.New("blah")

		directory := NewPasswordResetDirectory(loggingnoop.NewLogger(), &identitymock.StoreMock{
			UpdateUserPasswordFunc: func(context.Context, database.Tx, tenancy.Scope, string, string) error { return writeErr },
		}, buildEmitterForTest(t))

		require.ErrorIs(t, directory.UpdateUserPassword(t.Context(), harness.tx, tenancy.Global(), identifiers.New(), "hashed"), writeErr)
		assert.Empty(t, harness.enqueued(t))
	})
}

func TestSignInMailers_SendPasswordReset(T *testing.T) {
	T.Parallel()

	T.Run("queues the reset mail carrying the link's secret", func(t *testing.T) {
		t.Parallel()

		harness := buildSignInHooksHarness(t, nil)
		db := &mockdatabase.ClientMock{
			WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error { return fn(harness.tx) },
		}

		mailers := NewSignInMailers(loggingnoop.NewLogger(), db, buildEmitterForTest(t))

		user := &identity.User{ID: identifiers.New()}
		issuance := &passwordreset.Issuance{
			Token:  &passwordreset.Token{ID: identifiers.New(), ExpiresAt: time.Now().Add(time.Hour)},
			Secret: identifiers.New(),
		}

		require.NoError(t, mailers.SendPasswordReset(t.Context(), &passwordreset.Mail{User: user, Issuance: issuance}))

		enqueued := harness.enqueued(t)
		require.Len(t, enqueued, 1)
		assert.Equal(t, ddbidentity.PasswordResetTokenCreatedEventType, enqueued[0].EventType)
		assert.Equal(t, user.ID, enqueued[0].UserID)
		assert.Equal(t, issuance.Secret, enqueued[0].Context[authkeys.PasswordResetTokenSecretKey])
		assert.Equal(t, issuance.Token.ID, enqueued[0].Context[authkeys.PasswordResetTokenIDKey])
	})

	T.Run("refuses a mail that names nobody", func(t *testing.T) {
		t.Parallel()

		mailers := NewSignInMailers(loggingnoop.NewLogger(), &mockdatabase.ClientMock{}, buildEmitterForTest(t))

		assert.Error(t, mailers.SendPasswordReset(t.Context(), &passwordreset.Mail{}))
	})
}
