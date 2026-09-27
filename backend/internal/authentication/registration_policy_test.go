package authentication

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func agreedRegistration(t *testing.T) *signin.Registration {
	t.Helper()

	user := identityfakes.BuildFakeUser()

	return &signin.Registration{
		User: &platformidentity.User{
			Username:     user.Username,
			EmailAddress: user.EmailAddress,
			// What a request may say, and the policy replaces.
			AccountStatus: platformidentity.StatusUnverified,
			ServiceRoles:  []string{authorization.ServiceAdminRoleName},
		},
		Account:    &platformidentity.Account{Name: identityfakes.BuildFakeAccount().Name},
		OwnerRoles: []string{identityfakes.BuildFakeUser().ID},
		Agreements: []platformidentity.Agreement{platformidentity.TermsOfService, platformidentity.PrivacyPolicy},
	}
}

func TestRegistrationPolicy(T *testing.T) {
	T.Parallel()

	T.Run("shapes a registrant into this application's", func(t *testing.T) {
		t.Parallel()

		registration := agreedRegistration(t)
		accountName := registration.Account.Name

		require.NoError(t, RegistrationPolicy(t.Context(), registration))

		assert.Equal(t, platformidentity.StatusGood, registration.User.AccountStatus)
		assert.Equal(t, []string{authorization.ServiceUserRoleName}, registration.User.ServiceRoles,
			"a registrant chose their own service role")
		assert.Equal(t, []string{authorization.AccountAdminRoleName}, registration.OwnerRoles,
			"a registrant chose their own account role")
		assert.True(t, registration.EnrollTOTP)
		assert.Equal(t, accountName, registration.Account.Name)
	})

	T.Run("names an account nobody named", func(t *testing.T) {
		t.Parallel()

		registration := agreedRegistration(t)
		registration.Account = nil

		require.NoError(t, RegistrationPolicy(t.Context(), registration))
		require.NotNil(t, registration.Account)
		assert.Equal(t, registration.User.Username+"'s account", registration.Account.Name)
	})

	T.Run("refuses a registration missing an agreement", func(t *testing.T) {
		t.Parallel()

		for _, agreements := range [][]platformidentity.Agreement{
			nil,
			{platformidentity.TermsOfService},
			{platformidentity.PrivacyPolicy},
		} {
			registration := agreedRegistration(t)
			registration.Agreements = agreements

			require.ErrorIs(t, RegistrationPolicy(t.Context(), registration), ErrAgreementsRequired)
		}
	})
}
