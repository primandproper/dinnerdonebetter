package integration

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServer_Durability covers the two things neither the router test nor the store
// conformance suite can show, and the reason the MCP server was moved onto a durable
// authorization server in the first place: that its records belong to the deployment
// rather than to a process.
//
// Both would have passed against the map-backed server they replaced only by accident of
// there being one replica and no restarts.
func TestMCPServer_Durability(T *testing.T) {
	T.Parallel()

	T.Run("redeems a code at a replica that did not issue it", func(t *testing.T) {
		t.Parallel()

		// Registration and sign-in at one replica, redemption at another. Under the
		// memory store this is the failure a fleet gets in proportion to how well its
		// load balancer spreads traffic: the client is unknown at the second replica and
		// the code it is holding was never issued by anyone.
		issuer := startInstanceForTest(t, fleetBaseURL, nil)
		redeemer := startInstanceForTest(t, fleetBaseURL, nil)
		require.NotEqual(t, issuer.address, redeemer.address)

		registration := issuer.registerClient(t)
		authz := issuer.signIn(t, registration.ClientID, generateTOTPCode(t))

		accessToken := redeemer.redeem(t, authz).AccessToken

		res, err := redeemer.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.NoError(t, err)

		ingredient := &mealplanning.ValidIngredient{}
		requireStructuredContent(t, res, ingredient)
		assert.Equal(t, seededIngredient.ID, ingredient.ID)
	})

	T.Run("keeps a token usable across a restart", func(t *testing.T) {
		t.Parallel()

		// Started without the usual cleanup, because stopping it is the test.
		original, err := startInstance(t.Context(), fleetBaseURL, nil)
		require.NoError(t, err)

		accessToken := original.authenticate(t)

		_, err = original.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.NoError(t, err)

		require.NoError(t, original.stop(context.WithoutCancel(t.Context())))

		// Really gone, so that what follows cannot be the same process answering.
		_, err = original.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.Error(t, err)

		// The replacement binds its own port and advertises the same public address, the
		// way a rescheduled replica does. Nothing was handed over to it: the token is in
		// the database, which is the whole claim.
		successor := startInstanceForTest(t, fleetBaseURL, nil)

		res, err := successor.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.NoError(t, err)

		ingredient := &mealplanning.ValidIngredient{}
		requireStructuredContent(t, res, ingredient)
		assert.Equal(t, seededIngredient.ID, ingredient.ID)
	})

	T.Run("keeps a registered client across a restart", func(t *testing.T) {
		t.Parallel()

		original, err := startInstance(t.Context(), fleetBaseURL, nil)
		require.NoError(t, err)

		// Registration is the expensive half of onboarding an MCP client: it happens
		// once, and a client that has to redo it after every deploy is one whose stored
		// client_id is a lie.
		registration := original.registerClient(t)

		require.NoError(t, original.stop(context.WithoutCancel(t.Context())))

		successor := startInstanceForTest(t, fleetBaseURL, nil)

		authz := successor.signIn(t, registration.ClientID, generateTOTPCode(t))
		assert.NotEmpty(t, successor.redeem(t, authz).AccessToken)
	})
}
