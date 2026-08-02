package webhooks

import (
	"context"
	"net/http"
	"time"

	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	// WebhookCreatedServiceEventType indicates a webhook was created.
	WebhookCreatedServiceEventType = "webhook_created"
	// WebhookArchivedServiceEventType indicates a webhook was archived.
	WebhookArchivedServiceEventType = "webhook_archived"
	// WebhookTriggerConfigCreatedServiceEventType indicates a webhook trigger config was created.
	WebhookTriggerConfigCreatedServiceEventType = "webhook_trigger_config_created"
	// WebhookTriggerConfigArchivedServiceEventType indicates a webhook trigger config was archived.
	WebhookTriggerConfigArchivedServiceEventType = "webhook_trigger_config_archived"
)

type (
	// Webhook represents a webhook listener, an endpoint to send an HTTP request to upon an event.
	Webhook struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time               `json:"createdAt"`
		ArchivedAt       *time.Time              `json:"archivedAt"`
		LastUpdatedAt    *time.Time              `json:"lastUpdatedAt"`
		Name             string                  `json:"name"`
		URL              string                  `json:"url"`
		Method           string                  `json:"method"`
		ID               string                  `json:"id"`
		BelongsToAccount string                  `json:"belongsToAccount"`
		CreatedByUser    string                  `json:"createdByUser"`
		ContentType      string                  `json:"contentType"`
		TriggerConfigs   []*WebhookTriggerConfig `json:"triggerConfigs"`
	}

	// WebhookTriggerConfig represents a webhook's subscription to a trigger event (join table record).
	WebhookTriggerConfig struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time  `json:"createdAt"`
		ArchivedAt       *time.Time `json:"archivedAt"`
		ID               string     `json:"id"`
		BelongsToWebhook string     `json:"belongsToWebhook"`
		TriggerEventID   string     `json:"triggerEventId"`
	}

	// WebhookTriggerEvent is the catalog entity for available trigger event types.
	WebhookTriggerEvent struct {
		_             struct{}   `json:"-"`
		CreatedAt     time.Time  `json:"createdAt"`
		LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time `json:"archivedAt"`
		ID            string     `json:"id"`
		Name          string     `json:"name"`
		Description   string     `json:"description"`
	}

	// WebhookCreationRequestInput represents what a User could set as input for creating a webhook.
	WebhookCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string                                     `json:"name"`
		ContentType string                                     `json:"contentType"`
		URL         string                                     `json:"url"`
		Method      string                                     `json:"method"`
		Events      []*WebhookTriggerEventCreationRequestInput `json:"events"` // catalog event refs (ID) or new event definitions (Name/Description)
	}

	// WebhookDatabaseCreationInput is used for creating a webhook.
	WebhookDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID               string                                       `json:"-"`
		Name             string                                       `json:"-"`
		ContentType      string                                       `json:"-"`
		URL              string                                       `json:"-"`
		Method           string                                       `json:"-"`
		CreatedByUser    string                                       `json:"-"`
		BelongsToAccount string                                       `json:"-"`
		TriggerConfigs   []*WebhookTriggerConfigDatabaseCreationInput `json:"-"`
	}

	// WebhookTriggerConfigCreationRequestInput represents what a User could set as input for adding a trigger config.
	WebhookTriggerConfigCreationRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToWebhook string `json:"belongsToWebhook"`
		TriggerEventID   string `json:"triggerEventId"`
	}

	// WebhookTriggerConfigDatabaseCreationInput is used for creating a webhook trigger config.
	WebhookTriggerConfigDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID               string `json:"-"`
		BelongsToWebhook string `json:"-"`
		TriggerEventID   string `json:"-"`
	}

	// WebhookTriggerEventCreationRequestInput represents what a User could set as input for creating a catalog trigger event or referencing one by ID.
	// If ID is set, the existing catalog event is used; otherwise Name (and optionally Description) create a new catalog event.
	WebhookTriggerEventCreationRequestInput struct {
		_ struct{} `json:"-"`

		ID          string `json:"id"` // optional: use existing catalog event
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// WebhookTriggerEventDatabaseCreationInput is used for creating a catalog webhook trigger event.
	WebhookTriggerEventDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID          string `json:"-"`
		Name        string `json:"-"`
		Description string `json:"-"`
	}

	// WebhookTriggerEventUpdateRequestInput represents what a User could set as input for updating a catalog trigger event.
	WebhookTriggerEventUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name        string `json:"name"`
		Description string `json:"description"`
	}

	// WebhookExecutionRequest represents a webhook listener, an endpoint to send an HTTP request to upon an event.
	WebhookExecutionRequest struct {
		_ struct{} `json:"-"`

		RequestID    string `json:"id"`
		Payload      any    `json:"payload"`
		WebhookID    string `json:"webhookID"`
		AccountID    string `json:"accountID"`
		TriggerEvent string `json:"triggerEvent"` // catalog event ID
		TestID       string `json:"testID,omitempty"`
	}

	// WebhookDataManager describes a structure capable of storing and retrieving webhooks and trigger events.
	WebhookDataManager interface {
		WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error)
		GetWebhook(ctx context.Context, webhookID, accountID string) (*Webhook, error)
		GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Webhook], error)
		GetWebhooksForAccountAndEvent(ctx context.Context, accountID, triggerEventID string) ([]*Webhook, error)
		CreateWebhook(ctx context.Context, input *WebhookDatabaseCreationInput) (*Webhook, error)
		ArchiveWebhook(ctx context.Context, webhookID, accountID string) error
		AddWebhookTriggerConfig(ctx context.Context, accountID string, input *WebhookTriggerConfigDatabaseCreationInput) (*WebhookTriggerConfig, error)
		ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, configID string) error
		CreateWebhookTriggerEvent(ctx context.Context, input *WebhookTriggerEventDatabaseCreationInput) (*WebhookTriggerEvent, error)
		GetWebhookTriggerEvent(ctx context.Context, id string) (*WebhookTriggerEvent, error)
		GetWebhookTriggerEvents(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[WebhookTriggerEvent], error)
		UpdateWebhookTriggerEvent(ctx context.Context, id string, input *WebhookTriggerEventUpdateRequestInput) error
		ArchiveWebhookTriggerEvent(ctx context.Context, id string) error
	}
)

var _ validation.ValidatableWithContext = (*WebhookCreationRequestInput)(nil)

// ValidateWithContext validates a WebhookCreationRequestInput.
func (w *WebhookCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, is.URL),
		validation.Field(&w.Method, validation.Required, validation.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete)),
		validation.Field(&w.ContentType, validation.Required, validation.In("application/json", "application/xml")),
		validation.Field(&w.Events, validation.Required, validation.Length(1, 100)),
	)
}

var _ validation.ValidatableWithContext = (*WebhookTriggerConfigCreationRequestInput)(nil)

// ValidateWithContext validates a WebhookTriggerConfigCreationRequestInput.
func (w *WebhookTriggerConfigCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.BelongsToWebhook, validation.Required),
		validation.Field(&w.TriggerEventID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*WebhookDatabaseCreationInput)(nil)

// ValidateWithContext validates a WebhookDatabaseCreationInput.
func (w *WebhookDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.ID, validation.Required),
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, is.URL),
		validation.Field(&w.Method, validation.Required, validation.In(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete)),
		validation.Field(&w.ContentType, validation.Required, validation.In(encoding.ContentTypeJSON.String(), encoding.ContentTypeXML.String())),
		validation.Field(&w.TriggerConfigs, validation.Required),
		validation.Field(&w.BelongsToAccount, validation.Required),
		validation.Field(&w.CreatedByUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*WebhookTriggerEventCreationRequestInput)(nil)

// ValidateWithContext validates a WebhookTriggerEventCreationRequestInput.
// Either ID (reference existing catalog event) or Name (create new) must be set.
func (w *WebhookTriggerEventCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.Name, validation.When(w.ID == "", validation.Required)),
		validation.Field(&w.ID, validation.When(w.Name == "", validation.Required)),
	)
}

var _ validation.ValidatableWithContext = (*WebhookTriggerEventDatabaseCreationInput)(nil)

// ValidateWithContext validates a WebhookTriggerEventDatabaseCreationInput.
func (w *WebhookTriggerEventDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.ID, validation.Required),
		validation.Field(&w.Name, validation.Required),
	)
}
