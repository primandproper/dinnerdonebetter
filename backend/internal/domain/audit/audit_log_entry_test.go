package audit

import (
	"testing"

	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopeFor(T *testing.T) {
	T.Parallel()

	T.Run("prefers the account", func(t *testing.T) {
		t.Parallel()

		accountID, userID := identifiers.New(), identifiers.New()

		assert.Equal(t, accountID, ScopeFor(pointer.To(accountID), userID))
	})

	// Signup, login and password reset all happen outside an account. Filing them
	// under the empty scope would put every login in the application behind one row
	// lock; per-user chains keep them independent.
	T.Run("falls back to the user when there is no account", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		assert.Equal(t, userID, ScopeFor(nil, userID))
	})

	// An account ID that is present but empty is not an account. Treating it as one
	// would put the entry in the platform scope while the read path looked for it in
	// the user's, which is the failure this function exists to make impossible.
	T.Run("treats an empty account ID as no account", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		assert.Equal(t, userID, ScopeFor(pointer.To(""), userID))
	})

	T.Run("is empty for events belonging to neither", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ScopeFor(nil, ""))
	})
}

func TestDiff(T *testing.T) {
	T.Parallel()

	type resource struct {
		Name   string `json:"name"`
		Secret string `json:"secret" audit:"-"`
		Hidden string `json:"-"`
		Count  int    `json:"count"`
	}

	T.Run("reports changed fields under their encoded names", func(t *testing.T) {
		t.Parallel()

		before := &resource{Name: "old", Count: 1}
		after := &resource{Name: "new", Count: 1}

		changes, err := Diff(before, after)
		require.NoError(t, err)

		assert.Equal(t, map[string]Change{"name": {Old: "old", New: "new"}}, changes)
	})

	// Values stay typed through the log rather than being rendered to strings at the
	// write site, so a numeric field reads back as a number.
	T.Run("keeps values typed", func(t *testing.T) {
		t.Parallel()

		changes, err := Diff(&resource{Count: 1}, &resource{Count: 2})
		require.NoError(t, err)

		require.Contains(t, changes, "count")
		assert.Equal(t, 1, changes["count"].Old)
		assert.Equal(t, 2, changes["count"].New)
	})

	// The static counterpart to Redactions: a field tagged audit:"-" is kept out of
	// every diff, wherever the struct is diffed.
	T.Run("skips fields tagged out", func(t *testing.T) {
		t.Parallel()

		changes, err := Diff(
			&resource{Secret: "before", Hidden: "before"},
			&resource{Secret: "after", Hidden: "after"},
		)
		require.NoError(t, err)

		assert.Empty(t, changes)
	})

	T.Run("returns an empty map when nothing changed", func(t *testing.T) {
		t.Parallel()

		changes, err := Diff(&resource{Name: "same"}, &resource{Name: "same"})
		require.NoError(t, err)

		assert.NotNil(t, changes)
		assert.Empty(t, changes)
	})
}

// The redaction policy is the one thing here that is load-bearing for secrecy
// rather than for correctness, so it is worth asserting rather than assuming.
func TestRedactions(T *testing.T) {
	T.Parallel()

	T.Run("drops credential-shaped fields under the catch-all", func(t *testing.T) {
		t.Parallel()

		catchAll, ok := Redactions[""]
		require.True(t, ok, "the catch-all is what makes the policy about the word rather than about one table")

		for _, field := range []string{"password", "hashed_password", "two_factor_secret", "clientSecret"} {
			assert.Contains(t, catchAll.Drop, field)
		}
	})

	// Hashed rather than dropped: rotating a credential is a real event worth
	// recording, and the digest still answers whether the token presented was the
	// token issued.
	T.Run("hashes tokens rather than dropping them", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, Redactions[""].Hash, "token")
		assert.Contains(t, Redactions["password_reset_tokens"].Hash, "token")
	})
}
