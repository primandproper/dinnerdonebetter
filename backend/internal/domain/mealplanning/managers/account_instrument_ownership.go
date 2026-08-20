package managers

import (
	"context"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

func (m *mealPlanningManager) ListAccountInstrumentOwnerships(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.AccountInstrumentOwnership], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValue(identitykeys.AccountIDKey, ownerID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, ownerID)

	results, err := m.db.GetAccountInstrumentOwnerships(ctx, ownerID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching instrument ownerships")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidInstrumentsNotOwnedByAccount(ctx context.Context, accountID, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	results, err := m.db.SearchForValidInstrumentsNotOwnedByAccount(ctx, accountID, query, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid instruments not owned by account")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateAccountInstrumentOwnership(ctx context.Context, ownerID string, input *types.AccountInstrumentOwnershipCreationRequestInput) (*types.AccountInstrumentOwnership, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	convertedInput := converters.ConvertAccountInstrumentOwnershipCreationRequestInputToAccountInstrumentOwnershipDatabaseCreationInput(input)
	convertedInput.BelongsToAccount = ownerID

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.AccountInstrumentOwnershipIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, convertedInput.ID)

	created, err := m.db.CreateAccountInstrumentOwnership(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating instrument ownership")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadAccountInstrumentOwnership(ctx context.Context, ownerID, instrumentOwnershipID string) (*types.AccountInstrumentOwnership, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		identitykeys.AccountIDKey:                        ownerID,
		mealplanningkeys.AccountInstrumentOwnershipIDKey: instrumentOwnershipID,
	})
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, ownerID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, instrumentOwnershipID)

	result, err := m.db.GetAccountInstrumentOwnership(ctx, instrumentOwnershipID, ownerID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching instrument ownership")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateAccountInstrumentOwnership(ctx context.Context, instrumentOwnershipID, ownerID string, input *types.AccountInstrumentOwnershipUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		identitykeys.AccountIDKey:                        ownerID,
		mealplanningkeys.AccountInstrumentOwnershipIDKey: instrumentOwnershipID,
	})
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, ownerID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, instrumentOwnershipID)

	existingAccountInstrumentOwnership, err := m.db.GetAccountInstrumentOwnership(ctx, instrumentOwnershipID, ownerID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching instrument ownership to update")
	}

	existingAccountInstrumentOwnership.Update(input)
	if err = m.db.UpdateAccountInstrumentOwnership(ctx, existingAccountInstrumentOwnership); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating instrument ownership")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveAccountInstrumentOwnership(ctx context.Context, ownerID, instrumentOwnershipID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		identitykeys.AccountIDKey:                        ownerID,
		mealplanningkeys.AccountInstrumentOwnershipIDKey: instrumentOwnershipID,
	})
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, ownerID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, instrumentOwnershipID)

	if err := m.db.ArchiveAccountInstrumentOwnership(ctx, instrumentOwnershipID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving instrument ownership")
	}

	return nil
}
