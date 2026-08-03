package payments

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests ---

func TestCreateSubscription(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateSubscription(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestGetSubscription(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetSubscription(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestGetSubscriptionsForAccount(T *testing.T) {
	T.Parallel()

	T.Run("with empty account id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetSubscriptionsForAccount(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestUpdateSubscription(T *testing.T) {
	T.Parallel()

	T.Run("with nil subscription", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateSubscription(ctx, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestUpdateSubscriptionStatus(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateSubscriptionStatus(ctx, "", payments.SubscriptionStatusActive)
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestArchiveSubscription(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveSubscription(ctx, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests ---

func TestQuerier_Integration_Subscriptions(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)
	product := createProductForTest(t, ctx, nil, dbc)

	now := time.Now().UTC()
	periodEnd := now.AddDate(0, 1, 0)
	input := &payments.SubscriptionDatabaseCreationInput{
		ID:                     identifiers.New(),
		BelongsToAccount:       account.ID,
		ProductID:              product.ID,
		ExternalSubscriptionID: "ext_sub_" + identifiers.New(),
		Status:                 payments.SubscriptionStatusActive,
		CurrentPeriodStart:     now,
		CurrentPeriodEnd:       periodEnd,
	}

	created, err := dbc.CreateSubscription(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, input.ID, created.ID)
	assert.Equal(t, account.ID, created.BelongsToAccount)
	assert.Equal(t, product.ID, created.ProductID)

	// Get by ID
	fetched, err := dbc.GetSubscription(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)

	// Get by external ID
	byExt, err := dbc.GetSubscriptionByExternalID(ctx, input.ExternalSubscriptionID)
	require.NoError(t, err)
	require.NotNil(t, byExt)
	assert.Equal(t, created.ID, byExt.ID)

	// List for account
	subs, err := dbc.GetSubscriptionsForAccount(ctx, account.ID, nil)
	require.NoError(t, err)
	assert.Len(t, subs.Data, 1)

	// Update status
	err = dbc.UpdateSubscriptionStatus(ctx, created.ID, payments.SubscriptionStatusCancelled)
	require.NoError(t, err)

	updated, err := dbc.GetSubscription(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, payments.SubscriptionStatusCancelled, updated.Status)

	// Archive
	err = dbc.ArchiveSubscription(ctx, created.ID)
	require.NoError(t, err)

	_, err = dbc.GetSubscription(ctx, created.ID)
	assert.Error(t, err)
}
