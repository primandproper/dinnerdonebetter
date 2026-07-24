package manager

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/keys"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
)

const (
	o11yName = "waitlist_data_manager"
)

// waitlistRepository avoids wire cycles: manager takes this interface and produces waitlists.Repository.
type waitlistRepository interface {
	waitlists.Repository
}

var (
	_ waitlists.Repository = (*waitlistManager)(nil)
	_ WaitlistsDataManager = (*waitlistManager)(nil)
)

type waitlistManager struct {
	tracer               tracing.Tracer
	logger               logging.Logger
	repo                 waitlistRepository
	dataChangesPublisher messagequeue.Publisher
}

// NewWaitlistDataManager returns a new manager that wraps the repository and emits data change events.
func NewWaitlistDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo waitlistRepository,
	cfg *msgconfig.QueuesConfig,
	publisherProvider messagequeue.PublisherProvider,
) (WaitlistsDataManager, error) {
	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide publisher for data changes topic: %w", err)
	}

	return &waitlistManager{
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:               logging.NewNamedLogger(logger, o11yName),
		repo:                 repo,
		dataChangesPublisher: dataChangesPublisher,
	}, nil
}

func (m *waitlistManager) WaitlistIsNotExpired(ctx context.Context, waitlistID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.WaitlistIsNotExpired(ctx, waitlistID)
}

func (m *waitlistManager) GetWaitlist(ctx context.Context, waitlistID string) (*waitlists.Waitlist, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetWaitlist(ctx, waitlistID)
}

func (m *waitlistManager) GetWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetWaitlists(ctx, filter)
}

func (m *waitlistManager) GetActiveWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetActiveWaitlists(ctx, filter)
}

func (m *waitlistManager) CreateWaitlist(ctx context.Context, input *waitlists.WaitlistDatabaseCreationInput) (*waitlists.Waitlist, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	created, err := m.repo.CreateWaitlist(ctx, input)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, created.ID)
	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, waitlists.WaitlistCreatedServiceEventType, map[string]any{
		waitlistkeys.WaitlistIDKey: created.ID,
	}))

	return created, nil
}

func (m *waitlistManager) UpdateWaitlist(ctx context.Context, waitlist *waitlists.Waitlist) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if waitlist == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, waitlist.ID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlist.ID)

	if err := m.repo.UpdateWaitlist(ctx, waitlist); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, waitlists.WaitlistUpdatedServiceEventType, map[string]any{
		waitlistkeys.WaitlistIDKey: waitlist.ID,
	}))

	return nil
}

func (m *waitlistManager) ArchiveWaitlist(ctx context.Context, waitlistID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlistID)

	if err := m.repo.ArchiveWaitlist(ctx, waitlistID); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, waitlists.WaitlistArchivedServiceEventType, map[string]any{
		waitlistkeys.WaitlistIDKey: waitlistID,
	}))

	return nil
}

func (m *waitlistManager) GetWaitlistSignup(ctx context.Context, waitlistSignupID, waitlistID string) (*waitlists.WaitlistSignup, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetWaitlistSignup(ctx, waitlistSignupID, waitlistID)
}

func (m *waitlistManager) GetWaitlistSignupByID(ctx context.Context, waitlistSignupID string) (*waitlists.WaitlistSignup, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetWaitlistSignupByID(ctx, waitlistSignupID)
}

func (m *waitlistManager) GetWaitlistSignupsForWaitlist(ctx context.Context, waitlistID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetWaitlistSignupsForWaitlist(ctx, waitlistID, filter)
}

func (m *waitlistManager) GetWaitlistSignupsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetWaitlistSignupsForUser(ctx, userID, filter)
}

func (m *waitlistManager) CreateWaitlistSignup(ctx context.Context, input *waitlists.WaitlistSignupDatabaseCreationInput) (*waitlists.WaitlistSignup, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	created, err := m.repo.CreateWaitlistSignup(ctx, input)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, created.ID)
	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, waitlists.WaitlistSignupCreatedServiceEventType, map[string]any{
		waitlistkeys.WaitlistSignupIDKey: created.ID,
		waitlistkeys.WaitlistIDKey:       created.BelongsToWaitlist,
	}))

	return created, nil
}

func (m *waitlistManager) UpdateWaitlistSignup(ctx context.Context, signup *waitlists.WaitlistSignup) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if signup == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, signup.ID).WithValue(waitlistkeys.WaitlistIDKey, signup.BelongsToWaitlist)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, signup.ID)

	if err := m.repo.UpdateWaitlistSignup(ctx, signup); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, waitlists.WaitlistSignupUpdatedServiceEventType, map[string]any{
		waitlistkeys.WaitlistSignupIDKey: signup.ID,
		waitlistkeys.WaitlistIDKey:       signup.BelongsToWaitlist,
	}))

	return nil
}

func (m *waitlistManager) ArchiveWaitlistSignup(ctx context.Context, waitlistSignupID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)

	if err := m.repo.ArchiveWaitlistSignup(ctx, waitlistSignupID); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, waitlists.WaitlistSignupArchivedServiceEventType, map[string]any{
		waitlistkeys.WaitlistSignupIDKey: waitlistSignupID,
	}))

	return nil
}
