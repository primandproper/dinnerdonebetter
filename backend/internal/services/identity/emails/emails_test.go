package emails

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGeneratedPasswordResetTokenEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = new(time.Now())
		token := authfakes.BuildFakePasswordResetToken()

		actual, err := BuildGeneratedPasswordResetTokenEmail(user, token, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Contains(t, actual.HTMLContent, branding.LogoURL)
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
