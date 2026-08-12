package manager

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"

	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/identifiers"
	"github.com/primandproper/platform-go/v10/observability"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

const (
	o11yName = "webhook_data_manager"
)

var _ WebhookDataManager = (*webhookManager)(nil)

type webhookManager struct {
	tracer tracing.Tracer
	logger logging.Logger
	repo   webhooks.Repository
}

// NewWebhookDataManager returns a new WebhookDataManager that delegates to the webhooks repository.
//
// Data change events are enqueued into the outbox by the repository, inside the same
// transaction as the write they describe; see internal/repositories/postgres/events.
func NewWebhookDataManager(
	ctx context.Context,
	tracerProvider tracing.Provider,
	logger logging.Logger,
	repo webhooks.Repository,
) (WebhookDataManager, error) {
	return &webhookManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
	}, nil
}

func (m *webhookManager) WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.WebhookExists(ctx, webhookID, accountID)
}

// CreateWebhook registers a webhook and returns it together with its signing secret.
//
// The secret is returned here and nowhere else. There is no read path that can produce it, which
// is what makes "stored to sign with" and "handed to the account once" the only two things that
// ever happen to it.
func (m *webhookManager) CreateWebhook(ctx context.Context, userID, accountID string, input *webhooks.WebhookCreationRequestInput) (*webhooks.WebhookCreationResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span)
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating webhook creation input")
	}

	webhookID := identifiers.New()
	dbInput := &webhooks.WebhookDatabaseCreationInput{
		ID:               webhookID,
		Name:             input.Name,
		ContentType:      input.ContentType,
		URL:              input.URL,
		Method:           input.Method,
		CreatedByUser:    userID,
		BelongsToAccount: accountID,
		TriggerConfigs:   nil,
	}

	// Validation has already rejected any event type outside the catalog, so there is nothing
	// to resolve here — an event type is its own identity now, rather than a foreign key into a
	// table of randomly-identified rows that the fan-out could never match.
	for _, eventType := range input.Events {
		dbInput.TriggerConfigs = append(dbInput.TriggerConfigs, &webhooks.WebhookTriggerConfigDatabaseCreationInput{
			ID:               identifiers.New(),
			BelongsToWebhook: webhookID,
			EventType:        eventType,
		})
	}

	created, err := m.repo.CreateWebhook(ctx, dbInput)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, created.Webhook.ID)

	return created, nil
}

func (m *webhookManager) GetWebhook(ctx context.Context, webhookID, accountID string) (*webhooks.Webhook, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetWebhook(ctx, webhookID, accountID)
}

func (m *webhookManager) GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetWebhooks(ctx, accountID, filter)
}

func (m *webhookManager) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(webhookkeys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)

	if err := m.repo.ArchiveWebhook(ctx, webhookID, accountID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archive webhook")
	}

	return nil
}

func (m *webhookManager) AddWebhookTriggerConfig(ctx context.Context, accountID string, input *webhooks.WebhookTriggerConfigCreationRequestInput) (*webhooks.WebhookTriggerConfig, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, observability.PrepareError(errors.New("nil trigger config creation input"), span, "nil trigger config creation input")
	}
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating trigger config creation input")
	}

	dbInput := &webhooks.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: input.BelongsToWebhook,
		EventType:        input.EventType,
	}
	created, err := m.repo.AddWebhookTriggerConfig(ctx, accountID, dbInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "add webhook trigger config")
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, created.ID)

	return created, nil
}

func (m *webhookManager) ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, accountID, configID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, configID)

	if err := m.repo.ArchiveWebhookTriggerConfig(ctx, webhookID, accountID, configID); err != nil {
		return err
	}

	return nil
}

// RotateWebhookSecret mints a new signing secret and returns it, once.
//
// Deliveries are signed under both the new key and the outgoing one until this is called again,
// so a subscriber can accept either while it switches over. That window is the whole reason the
// secret is per-endpoint rather than per-account: a single account-wide key cannot be rolled
// without breaking every subscriber for that account at the same instant, which in practice
// means it never gets rolled at all.
func (m *webhookManager) RotateWebhookSecret(ctx context.Context, webhookID, accountID string) (string, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)

	secret, err := m.repo.RotateWebhookSecret(ctx, webhookID, accountID)
	if err != nil {
		return "", observability.PrepareAndLogError(err, m.logger.WithSpan(span), span, "rotating webhook signing secret")
	}

	return secret, nil
}

func (m *webhookManager) GetWebhookEventTypes(context.Context) []*webhooks.WebhookEventType {
	return webhooks.EventTypeCatalog()
}
