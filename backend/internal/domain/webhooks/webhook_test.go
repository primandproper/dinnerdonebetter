package webhooks

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	"github.com/stretchr/testify/assert"
)

// aKnownEventType returns an event type from the real catalog.
//
// Read from the catalog rather than through the fakes package, which imports this one.
func aKnownEventType() string {
	return catalog.Catalog().EventTypes()[0]
}

func TestWebhookCreationInput_Validate(T *testing.T) {
	T.Parallel()

	buildValidWebhookCreationInput := func() *WebhookCreationRequestInput {
		return &WebhookCreationRequestInput{
			Name:        "whatever",
			ContentType: "application/json",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      DeliveryMethod,
			Events:      []string{aKnownEventType()},
		}
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, buildValidWebhookCreationInput().ValidateWithContext(t.Context()))
	})

	T.Run("bad name", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Name = ""

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("bad url", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		// much as we'd like to use testutils.InvalidRawURL here, it causes a cyclical import :'(
		exampleInput.URL = fmt.Sprintf(`%s://verygoodsoftwarenotvirus.ru`, string(byte(2<<6-1)))

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("bad method", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Method = http.MethodPatch

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("bad content type", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.ContentType = "application/xml"

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("event type outside the catalog", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		// A typo'd event type accepted here becomes an endpoint that never fires, and
		// diagnosing that means noticing an absence.
		exampleInput.Events = []string{"reciped_created"}

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("event type published but not deliverable", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		// user_logged_in is emitted, and deliberately not subscribable: an endpoint URL is
		// attacker-supplied, and this would be a live feed of an account's sign-in activity.
		exampleInput.Events = []string{"user_logged_in"}

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("empty events", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Events = []string{}

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})
}

func TestWebhookCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		name := t.Name()
		ctx := t.Context()
		x := &WebhookCreationRequestInput{
			Name:        name,
			ContentType: "application/json",
			URL:         "https://pkg.go.dev",
			Method:      DeliveryMethod,
			Events:      []string{aKnownEventType()},
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestWebhookDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		name := t.Name()
		ctx := t.Context()
		x := &WebhookDatabaseCreationInput{
			ID:               name,
			Name:             name,
			ContentType:      "application/json",
			URL:              "https://pkg.go.dev",
			Method:           DeliveryMethod,
			TriggerConfigs:   []*WebhookTriggerConfigDatabaseCreationInput{{}},
			BelongsToAccount: name,
			CreatedByUser:    name,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
