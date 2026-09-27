package waitlists

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"

	"github.com/primandproper/platform-go/v14/links"
	waitlistsgrpc "github.com/primandproper/platform-go/v14/waitlists/grpc"
	"github.com/primandproper/primitives-go/v2/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfirmationEmail(T *testing.T) {
	T.Parallel()

	T.Run("carries both links to the address that joined", func(t *testing.T) {
		t.Parallel()

		mail := &waitlistsgrpc.ConfirmationMail{
			Signup:      fakes.BuildFakeWaitlistSignup(),
			Confirm:     &links.Link{URL: "https://example.com/waitlists/confirm?t=" + fake.BuildFakeID()},
			Unsubscribe: &links.Link{URL: "https://example.com/waitlists/unsubscribe?t=" + fake.BuildFakeID()},
		}

		actual, err := buildConfirmationEmail(mail, "https://example.com")
		require.NoError(t, err)

		assert.Equal(t, mail.Signup.Contact, actual.ToAddress)
		assert.Contains(t, actual.HTMLContent, mail.Confirm.URL)
		assert.Contains(t, actual.HTMLContent, mail.Unsubscribe.URL)
	})

	T.Run("with a link missing", func(t *testing.T) {
		t.Parallel()

		mail := &waitlistsgrpc.ConfirmationMail{
			Signup:  fakes.BuildFakeWaitlistSignup(),
			Confirm: &links.Link{URL: "https://example.com/waitlists/confirm?t=" + fake.BuildFakeID()},
		}

		actual, err := buildConfirmationEmail(mail, "https://example.com")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}
