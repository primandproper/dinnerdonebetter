package identityspike

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/succession"

	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The succession rule, against a real database.
//
// A household whose owner is erased goes to its longest-tenured remaining
// member, and is deleted only if the owner was alone. platform's EraseUser
// leaves the account standing with an owner that resolves to nothing; this is
// what this application does about that.
//
// Every test here runs the rule and EraseUser on one transaction, in the order
// the package documentation requires, because that is the shape the erasure
// will actually have.

// household is one account and the people in it, as the tests set them up.
type household struct {
	account *identity.Account
	owner   *identity.User
}

// registerHousehold creates an owner and their account.
func (f *fixture) registerHousehold(t *testing.T, ctx context.Context) household {
	t.Helper()

	user, account := newRegistration()

	registration, err := f.service.Register(ctx, tenancy.Global(), user, account, []string{"account_admin"})
	require.NoError(t, err)

	return household{account: registration.Account, owner: registration.User}
}

// addMember registers a second user and puts them in the household.
//
// createdAt is stamped onto the membership afterwards, because tenure is what
// the rule sorts on and a test that created three members in one transaction
// would otherwise be sorting rows that share a timestamp.
func (f *fixture) addMember(t *testing.T, ctx context.Context, h household, createdAt time.Time) *identity.User {
	t.Helper()

	member, ownAccount := newRegistration()

	registration, err := f.service.Register(ctx, tenancy.Global(), member, ownAccount, []string{"account_admin"})
	require.NoError(t, err)

	require.NoError(t, f.db.WithTransaction(ctx, func(tx database.Tx) error {
		_, createErr := f.store.CreateMembership(ctx, tx, tenancy.Global(), &identity.Membership{
			ID:               identifiers.New(),
			BelongsToUser:    registration.User.ID,
			BelongsToAccount: h.account.ID,
			Roles:            []string{"account_member"},
		})

		return createErr
	}))

	_, err = f.db.Writer().ExecContext(ctx,
		`UPDATE `+tablePrefix+`_identity_memberships SET created_at = $1 WHERE belongs_to_user = $2 AND belongs_to_account = $3`,
		createdAt, registration.User.ID, h.account.ID)
	require.NoError(t, err)

	return registration.User
}

// addSoloHousehold gives an existing user a second household nobody else is in.
//
// Built through the store rather than by registering, because registering would
// mint a second user and the point is a household whose only member is one this
// test already has.
func (f *fixture) addSoloHousehold(t *testing.T, ctx context.Context, ownerID string) *identity.Account {
	t.Helper()

	account := &identity.Account{ID: identifiers.New(), Name: "solo " + identifiers.New(), OwnerUserID: ownerID}

	require.NoError(t, f.db.WithTransaction(ctx, func(tx database.Tx) error {
		if _, err := f.store.CreateAccount(ctx, tx, tenancy.Global(), account); err != nil {
			return err
		}

		_, err := f.store.CreateMembership(ctx, tx, tenancy.Global(), &identity.Membership{
			ID:               identifiers.New(),
			BelongsToUser:    ownerID,
			BelongsToAccount: account.ID,
			Roles:            []string{"account_admin"},
		})

		return err
	}))

	return account
}

// eraseWithSuccession is the erasure as it will be composed: the rule, then
// platform's EraseUser, on one transaction.
func (f *fixture) eraseWithSuccession(
	t *testing.T,
	ctx context.Context,
	userID string,
	afterErase func(tx database.Tx) error,
) (succession.Outcome, error) {
	t.Helper()

	rule, err := succession.New(f.store, tablePrefix)
	require.NoError(t, err)

	var outcome succession.Outcome

	err = f.db.WithTransaction(ctx, func(tx database.Tx) error {
		var ruleErr error
		if outcome, ruleErr = rule.Apply(ctx, tx, tenancy.Global(), userID); ruleErr != nil {
			return ruleErr
		}

		if _, eraseErr := f.store.EraseUser(ctx, tx, tenancy.Global(), userID); eraseErr != nil {
			return eraseErr
		}

		if afterErase != nil {
			return afterErase(tx)
		}

		return nil
	})

	return outcome, err
}

func (f *fixture) accountOwner(t *testing.T, ctx context.Context, accountID string) (string, bool) {
	t.Helper()

	var owner string

	row := f.db.Reader().QueryRowContext(ctx,
		`SELECT owner_user_id FROM `+tablePrefix+`_identity_accounts WHERE id = $1`, accountID)
	if err := row.Scan(&owner); err != nil {
		return "", false
	}

	return owner, true
}

// TestSuccession_TransfersToTheLongestTenuredMember is the rule's main case.
func TestSuccession_TransfersToTheLongestTenuredMember(T *testing.T) {
	T.Parallel()

	T.Run("the earliest membership inherits the household", func(t *testing.T) {
		t.Parallel()

		f := buildFixture(t)
		ctx := t.Context()

		h := f.registerHousehold(t, ctx)

		// Deliberately added newest-first, so a rule that returned the first row
		// the database handed it would pick the wrong one.
		newest := f.addMember(t, ctx, h, time.Now().Add(-1*time.Hour))
		oldest := f.addMember(t, ctx, h, time.Now().Add(-100*time.Hour))
		middle := f.addMember(t, ctx, h, time.Now().Add(-50*time.Hour))

		outcome, err := f.eraseWithSuccession(t, ctx, h.owner.ID, nil)
		require.NoError(t, err)

		require.Len(t, outcome.Transferred, 1)
		assert.Equal(t, h.account.ID, outcome.Transferred[0].AccountID)
		assert.Equal(t, oldest.ID, outcome.Transferred[0].NewOwnerUserID,
			"the longest-tenured member should have inherited the household")
		assert.Empty(t, outcome.DeletedAccountIDs)

		owner, found := f.accountOwner(t, ctx, h.account.ID)
		require.True(t, found, "the household should still exist")
		assert.Equal(t, oldest.ID, owner)

		// And the others are still in it.
		assert.NotEqual(t, newest.ID, owner)
		assert.NotEqual(t, middle.ID, owner)
		assert.Equal(t, 3, f.count(t, ctx,
			`SELECT COUNT(*) FROM `+tablePrefix+`_identity_memberships WHERE belongs_to_account = $1`, h.account.ID),
			"the remaining members should have kept their memberships")
	})
}

// TestSuccession_DeletesASoloHousehold is the other half of the rule.
func TestSuccession_DeletesASoloHousehold(T *testing.T) {
	T.Parallel()

	T.Run("a household the owner was alone in goes", func(t *testing.T) {
		t.Parallel()

		f := buildFixture(t)
		ctx := t.Context()

		h := f.registerHousehold(t, ctx)

		outcome, err := f.eraseWithSuccession(t, ctx, h.owner.ID, nil)
		require.NoError(t, err)

		assert.Empty(t, outcome.Transferred)
		require.Len(t, outcome.DeletedAccountIDs, 1)
		assert.Equal(t, h.account.ID, outcome.DeletedAccountIDs[0])

		_, found := f.accountOwner(t, ctx, h.account.ID)
		assert.False(t, found, "a solo household should have been deleted, not left standing")
	})
}

// TestSuccession_LeavesNoOwnerlessHousehold is the property the rule exists for,
// stated as the thing that must never be true afterwards.
//
// It is the test that would have caught doing nothing at all: platform's
// EraseUser on its own leaves the account with an owner_user_id naming a user
// that no longer exists, which is the state this package was written to prevent.
func TestSuccession_LeavesNoOwnerlessHousehold(T *testing.T) {
	T.Parallel()

	T.Run("no surviving household names an erased owner", func(t *testing.T) {
		t.Parallel()

		f := buildFixture(t)
		ctx := t.Context()

		shared := f.registerHousehold(t, ctx)
		f.addMember(t, ctx, shared, time.Now().Add(-10*time.Hour))

		// One subject owning two households, one shared and one not, so both paths
		// of the rule run in a single erasure.
		solo := f.addSoloHousehold(t, ctx, shared.owner.ID)

		outcome, err := f.eraseWithSuccession(t, ctx, shared.owner.ID, nil)
		require.NoError(t, err)

		// The invariant first, because it is what this test is for: whatever the
		// rule decided, nothing may be left naming an owner who is gone. It is
		// asserted rather than required so the outcome below is still reported when
		// it fails.
		assert.Zero(t, f.count(t, ctx,
			`SELECT COUNT(*) FROM `+tablePrefix+`_identity_accounts a
			 WHERE NOT EXISTS (SELECT 1 FROM `+tablePrefix+`_identity_users u WHERE u.id = a.owner_user_id)`),
			"every surviving household should name an owner who still exists")

		// And the rule took both paths to get there.
		assert.Len(t, outcome.Transferred, 1)
		if assert.Len(t, outcome.DeletedAccountIDs, 1) {
			assert.Equal(t, solo.ID, outcome.DeletedAccountIDs[0])
		}
	})
}

// TestSuccession_RollsBackWithTheErasure pins that the rule's writes are the
// erasure's writes.
//
// A failure after EraseUser — another domain's eraser, or the record that the
// erasure happened — must take the transfer and the deletion back with it.
// Otherwise a retried erasure would find households already moved and a subject
// still present, which is the half-erased state the shared transaction exists to
// rule out.
func TestSuccession_RollsBackWithTheErasure(T *testing.T) {
	T.Parallel()

	T.Run("a later failure undoes the transfer and the deletion", func(t *testing.T) {
		t.Parallel()

		f := buildFixture(t)
		ctx := t.Context()

		shared := f.registerHousehold(t, ctx)
		successor := f.addMember(t, ctx, shared, time.Now().Add(-10*time.Hour))

		solo := f.addSoloHousehold(t, ctx, shared.owner.ID)

		errLaterEraser := platformerrors.New("a later domain's eraser failed")

		_, err := f.eraseWithSuccession(t, ctx, shared.owner.ID, func(database.Tx) error {
			return errLaterEraser
		})
		require.Error(t, err)

		// The subject is still there.
		assert.Equal(t, 1, f.count(t, ctx,
			`SELECT COUNT(*) FROM `+tablePrefix+`_identity_users WHERE id = $1`, shared.owner.ID))

		// The transfer did not stick.
		owner, found := f.accountOwner(t, ctx, shared.account.ID)
		require.True(t, found)
		assert.Equal(t, shared.owner.ID, owner, "the transfer should have rolled back")
		assert.NotEqual(t, successor.ID, owner)

		// Neither did the deletion.
		_, found = f.accountOwner(t, ctx, solo.ID)
		assert.True(t, found, "the deleted household should have come back with the rollback")
	})
}
