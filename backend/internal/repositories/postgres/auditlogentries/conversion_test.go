package auditlogentries

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/pointer"
	"github.com/primandproper/primitives-go/v2/tenancy"

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

		assert.Equal(t, tenancy.Of(accountID), converted.Scope)
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

		assert.Equal(t, tenancy.Of(userID), converted.Scope)
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
