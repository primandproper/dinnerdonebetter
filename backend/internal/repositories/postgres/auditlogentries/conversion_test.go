package auditlogentries

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v10/audit"
	"github.com/primandproper/platform-go/v10/identifiers"
	"github.com/primandproper/platform-go/v10/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPlatformEntry(T *testing.T) {
	T.Parallel()

	T.Run("maps an account-scoped entry", func(t *testing.T) {
		t.Parallel()

		accountID, userID, resourceID := identifiers.New(), identifiers.New(), identifiers.New()

		converted := toPlatformEntry(&audit.AuditLogEntry{
			BelongsToAccount: pointer.To(accountID),
			BelongsToUser:    userID,
			ResourceType:     "recipes",
			RelevantID:       resourceID,
			EventType:        audit.AuditLogEventTypeUpdated,
			ActorIP:          "203.0.113.7",
		})

		assert.Equal(t, accountID, converted.Scope)
		assert.Equal(t, userID, converted.Actor.ID)
		assert.Equal(t, platformaudit.ActorUser, converted.Actor.Type)
		assert.Equal(t, "203.0.113.7", converted.Actor.IP)
		assert.Equal(t, resourceID, converted.ResourceID)
		assert.Equal(t, platformaudit.EventUpdated, converted.EventType)
	})

	T.Run("chains a user-scoped entry under the user", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		converted := toPlatformEntry(&audit.AuditLogEntry{
			BelongsToUser: userID,
			ResourceType:  "users",
			EventType:     audit.AuditLogEventTypeCreated,
		})

		assert.Equal(t, userID, converted.Scope)
		assert.Equal(t, userID, converted.Actor.ID)
	})

	// The platform refuses an entry with no actor, and plenty of this application's
	// repository methods take an ID and nothing else. Naming the absence is what
	// keeps those writes working without pretending somebody was responsible.
	T.Run("names the actor when the call site has none", func(t *testing.T) {
		t.Parallel()

		converted := toPlatformEntry(&audit.AuditLogEntry{
			ResourceType: "service_settings",
			EventType:    audit.AuditLogEventTypeArchived,
		})

		assert.Equal(t, audit.UnattributedActorID, converted.Actor.ID)
		assert.Equal(t, platformaudit.ActorSystem, converted.Actor.Type)

		// The three things the platform validates before it will record anything. An
		// empty actor here is the failure this branch exists to prevent.
		require.NotEmpty(t, converted.Actor.ID)
		require.NotEmpty(t, converted.ResourceType)
		require.NotEmpty(t, converted.EventType)
	})
}

func TestFromPlatformEntry(T *testing.T) {
	T.Parallel()

	T.Run("recovers the account from the scope", func(t *testing.T) {
		t.Parallel()

		accountID, userID := identifiers.New(), identifiers.New()
		recordedAt := time.Now().UTC().Truncate(time.Microsecond)

		converted := fromPlatformEntry(&platformaudit.Entry{
			Scope:        accountID,
			Actor:        platformaudit.Actor{ID: userID, Type: platformaudit.ActorUser},
			RecordedAt:   recordedAt,
			ResourceType: "recipes",
			EventType:    platformaudit.EventUpdated,
			Seq:          7,
			Hash:         "abc",
			PrevHash:     "def",
		})

		require.NotNil(t, converted.BelongsToAccount)
		assert.Equal(t, accountID, *converted.BelongsToAccount)
		assert.Equal(t, userID, converted.BelongsToUser)
		assert.Equal(t, recordedAt, converted.CreatedAt)
		assert.Equal(t, int64(7), converted.Seq)
		assert.Equal(t, "abc", converted.Hash)
		assert.Equal(t, "def", converted.PrevHash)
	})

	// A scope equal to the actor's own ID is a user chain, not an account. Getting
	// this backwards would report a user ID as an account ID to every reader.
	T.Run("reports no account when the scope is the actor", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()

		converted := fromPlatformEntry(&platformaudit.Entry{
			Scope:        userID,
			Actor:        platformaudit.Actor{ID: userID},
			ResourceType: "users",
			EventType:    platformaudit.EventCreated,
		})

		assert.Nil(t, converted.BelongsToAccount)
		assert.Equal(t, userID, converted.BelongsToUser)
	})

	T.Run("reports no account for a platform-scoped entry", func(t *testing.T) {
		t.Parallel()

		converted := fromPlatformEntry(&platformaudit.Entry{
			Actor:        platformaudit.Actor{ID: audit.UnattributedActorID},
			ResourceType: "service_settings",
			EventType:    platformaudit.EventArchived,
		})

		assert.Nil(t, converted.BelongsToAccount)
	})
}

// The round trip is what the read path depends on: an entry written through
// Record and read back through the Reader has to describe the same event.
func TestConversionRoundTrip(T *testing.T) {
	T.Parallel()

	T.Run("preserves an account-scoped entry", func(t *testing.T) {
		t.Parallel()

		original := &audit.AuditLogEntry{
			BelongsToAccount: pointer.To(identifiers.New()),
			BelongsToUser:    identifiers.New(),
			ID:               identifiers.New(),
			ResourceType:     "recipes",
			RelevantID:       identifiers.New(),
			EventType:        audit.AuditLogEventTypeUpdated,
			CreatedAt:        time.Now().UTC().Truncate(time.Microsecond),
			Changes:          map[string]audit.Change{"name": {Old: "before", New: "after"}},
		}

		result := fromPlatformEntry(toPlatformEntry(original))

		assert.Equal(t, *original.BelongsToAccount, *result.BelongsToAccount)
		assert.Equal(t, original.BelongsToUser, result.BelongsToUser)
		assert.Equal(t, original.ID, result.ID)
		assert.Equal(t, original.ResourceType, result.ResourceType)
		assert.Equal(t, original.RelevantID, result.RelevantID)
		assert.Equal(t, original.EventType, result.EventType)
		assert.Equal(t, original.CreatedAt, result.CreatedAt)
		assert.Equal(t, original.Changes, result.Changes)
	})

	T.Run("preserves a user-scoped entry", func(t *testing.T) {
		t.Parallel()

		original := &audit.AuditLogEntry{
			BelongsToUser: identifiers.New(),
			ResourceType:  "users",
			RelevantID:    identifiers.New(),
			EventType:     audit.AuditLogEventTypeCreated,
		}

		result := fromPlatformEntry(toPlatformEntry(original))

		assert.Nil(t, result.BelongsToAccount)
		assert.Equal(t, original.BelongsToUser, result.BelongsToUser)
	})
}
