package emails

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v13/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGeneratedPasswordResetTokenEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = new(time.Now())
		token := fake.BuildFakeString()

		actual, err := BuildGeneratedPasswordResetTokenEmail(user, token, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Contains(t, actual.HTMLContent, branding.LogoURL)
		// The link is the only place the secret goes, and it has to actually be in it.
		assert.Contains(t, actual.HTMLContent, token)
	})
}

func TestBuildInviteMemberEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		invitation := fakes.BuildFakeAccountInvitation()

		actual, err := BuildInviteMemberEmail(user, invitation, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestBuildPasswordResetTokenRedeemedEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = new(time.Now())

		actual, err := BuildPasswordResetTokenRedeemedEmail(user, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestBuildUsernameReminderEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = new(time.Now())

		actual, err := BuildUsernameReminderEmail(user, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, actual)
	})
}
