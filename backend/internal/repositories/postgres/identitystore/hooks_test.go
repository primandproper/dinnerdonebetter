package identitystore

import (
	"testing"
	"time"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangedFields(T *testing.T) {
	T.Parallel()

	// Every save stamps lastUpdatedAt, so naming it would tell a subscriber nothing
	// about what the account holder edited.
	T.Run("names the edited fields and not the save stamp", func(t *testing.T) {
		t.Parallel()

		earlier, later := time.Now().Add(-time.Hour), time.Now()
		before := &platformidentity.Account{Name: identifiers.New(), LastUpdatedAt: &earlier}
		after := &platformidentity.Account{Name: identifiers.New(), LastUpdatedAt: &later}

		changes, err := platformaudit.Diff(before, after)
		require.NoError(t, err)
		require.Contains(t, changes, "lastUpdatedAt", "the audit diff keeps the stamp")

		assert.Equal(t, []string{"name"}, changedFields(changes))
	})
}
