package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	o11yName = "data_privacy_manager"
)

var (
	_ dataprivacy.Repository = (*dataPrivacyManager)(nil)
	_ DataPrivacyManager     = (*dataPrivacyManager)(nil)
)

type dataPrivacyManager struct {
	tracer tracing.Tracer
	logger logging.Logger
	repo   dataprivacy.Repository
}

// NewDataPrivacyManager returns a new DataPrivacyManager that wraps the data privacy repository.
func NewDataPrivacyManager(
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo dataprivacy.Repository,
) DataPrivacyManager {
	return &dataPrivacyManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
	}
}

func (m *dataPrivacyManager) FetchUserDataCollection(ctx context.Context, userID string) (*dataprivacy.UserDataCollection, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.FetchUserDataCollection(ctx, userID)
}

func (m *dataPrivacyManager) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.DeleteUser(ctx, userID)
}

func (m *dataPrivacyManager) CreateUserDataDisclosure(ctx context.Context, input *dataprivacy.UserDataDisclosureCreationInput) (*dataprivacy.UserDataDisclosure, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, input.BelongsToUser)

	created, err := m.repo.CreateUserDataDisclosure(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating user data disclosure")
	}

	return created, nil
}

func (m *dataPrivacyManager) GetUserDataDisclosure(ctx context.Context, disclosureID string) (*dataprivacy.UserDataDisclosure, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetUserDataDisclosure(ctx, disclosureID)
}

func (m *dataPrivacyManager) GetUserDataDisclosuresForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetUserDataDisclosuresForUser(ctx, userID, filter)
}

func (m *dataPrivacyManager) MarkUserDataDisclosureCompleted(ctx context.Context, disclosureID, reportID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.MarkUserDataDisclosureCompleted(ctx, disclosureID, reportID)
}

func (m *dataPrivacyManager) MarkUserDataDisclosureFailed(ctx context.Context, disclosureID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.MarkUserDataDisclosureFailed(ctx, disclosureID)
}

func (m *dataPrivacyManager) ArchiveUserDataDisclosure(ctx context.Context, disclosureID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.ArchiveUserDataDisclosure(ctx, disclosureID)
}
