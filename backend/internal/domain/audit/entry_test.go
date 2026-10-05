package audit

import (
	"testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
)

func TestScopeFor(T *testing.T) {
	T.Parallel()

	T.Run("prefers the account", func(t *testing.T) {
		t.Parallel()

		accountID, userID := identifiers.New(), identifiers.New()

		assert.Equal(t, tenancy.Of(accountID), ScopeFor(accountID, userID))
	})

	// Signup, login and password reset all happen outside an account. Filing them
	// under the empty scope would put every login in the application behind one row
	// lock; per-user chains keep them independent.
	T.Run("falls back to the user when there is no account", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		assert.Equal(t, tenancy.Of(userID), ScopeFor("", userID))
	})

	// Global, not the zero Scope: under platform-go v14 the zero value means nobody
	// said, and every read refuses it. An event that genuinely belongs to the
	// platform has to say so.
	T.Run("is global for events belonging to neither", func(t *testing.T) {
		t.Parallel()

		scope := ScopeFor("", "")
		assert.True(t, scope.IsGlobal())
		assert.Equal(t, tenancy.Global(), scope)
	})
}

func TestNewEntry(T *testing.T) {
	T.Parallel()

	T.Run("attributes the entry to the user and files it under the account", func(t *testing.T) {
		t.Parallel()

		accountID, userID, resourceID := identifiers.New(), identifiers.New(), identifiers.New()

		entry := NewEntry(userID, accountID, "recipes", resourceID, platformaudit.EventCreated)

		assert.Equal(t, platformaudit.Actor{ID: userID, Type: platformaudit.ActorUser}, entry.Actor)
		assert.Equal(t, tenancy.Of(accountID), entry.Scope)
		assert.Equal(t, "recipes", entry.ResourceType)
		assert.Equal(t, resourceID, entry.ResourceID)
		assert.Equal(t, platformaudit.EventCreated, entry.EventType)
	})

	// Recorded under a name rather than blank, so the writers still owed a requester can be
	// found: see UnattributedActorID.
	T.Run("names the system as actor when there is no requester", func(t *testing.T) {
		t.Parallel()

		entry := NewEntry("", "", "meal_plan_events", identifiers.New(), platformaudit.EventUpdated)

		assert.Equal(t, platformaudit.Actor{ID: UnattributedActorID, Type: platformaudit.ActorSystem}, entry.Actor)
		assert.Equal(t, tenancy.Global(), entry.Scope)
	})
}
