package manager

import (
	"context"
	"errors"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
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
	tracerProvider tracing.TracerProvider,
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

func (m *webhookManager) CreateWebhook(ctx context.Context, userID, accountID string, input *webhooks.WebhookCreationRequestInput) (*webhooks.Webhook, error) {
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
	for _, ev := range input.Events {
		triggerEventID := ev.ID
		if triggerEventID == "" {
			catalogInput := &webhooks.WebhookTriggerEventDatabaseCreationInput{
				ID:          identifiers.New(),
				Name:        ev.Name,
				Description: ev.Description,
			}
			created, err := m.repo.CreateWebhookTriggerEvent(ctx, catalogInput)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, m.logger, span, "creating catalog trigger event")
			}
			triggerEventID = created.ID
		}
		dbInput.TriggerConfigs = append(dbInput.TriggerConfigs, &webhooks.WebhookTriggerConfigDatabaseCreationInput{
			ID:               identifiers.New(),
			BelongsToWebhook: webhookID,
			TriggerEventID:   triggerEventID,
		})
	}

	created, err := m.repo.CreateWebhook(ctx, dbInput)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, created.ID)

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
		TriggerEventID:   input.TriggerEventID,
	}
	created, err := m.repo.AddWebhookTriggerConfig(ctx, accountID, dbInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "add webhook trigger config")
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, created.ID)

	return created, nil
}

func (m *webhookManager) ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, configID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, configID)

	if err := m.repo.ArchiveWebhookTriggerConfig(ctx, webhookID, configID); err != nil {
		return err
	}

	return nil
}

func (m *webhookManager) CreateWebhookTriggerEvent(ctx context.Context, input *webhooks.WebhookTriggerEventCreationRequestInput) (*webhooks.WebhookTriggerEvent, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, observability.PrepareError(errors.New("nil trigger event creation input"), span, "nil trigger event creation input")
	}
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating trigger event creation input")
	}

	dbInput := &webhooks.WebhookTriggerEventDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        input.Name,
		Description: input.Description,
	}
	return m.repo.CreateWebhookTriggerEvent(ctx, dbInput)
}

func (m *webhookManager) GetWebhookTriggerEvent(ctx context.Context, id string) (*webhooks.WebhookTriggerEvent, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetWebhookTriggerEvent(ctx, id)
}

func (m *webhookManager) GetWebhookTriggerEvents(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.WebhookTriggerEvent], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetWebhookTriggerEvents(ctx, filter)
}

func (m *webhookManager) UpdateWebhookTriggerEvent(ctx context.Context, id string, input *webhooks.WebhookTriggerEventUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	return m.repo.UpdateWebhookTriggerEvent(ctx, id, input)
}

func (m *webhookManager) ArchiveWebhookTriggerEvent(ctx context.Context, id string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.ArchiveWebhookTriggerEvent(ctx, id)
}
