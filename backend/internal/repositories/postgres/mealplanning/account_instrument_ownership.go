package mealplanning

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v8/database"
	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	resourceTypeAccountInstrumentOwnerships = "account_instrument_ownerships"
)

var (
	_ types.AccountInstrumentOwnershipDataManager = (*repository)(nil)
)

// AccountInstrumentOwnershipExists fetches whether an account instrument ownership exists from the database.
func (q *repository) AccountInstrumentOwnershipExists(ctx context.Context, accountInstrumentOwnershipID, accountID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if accountInstrumentOwnershipID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.AccountInstrumentOwnershipIDKey, accountInstrumentOwnershipID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, accountInstrumentOwnershipID)

	if accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := q.generatedQuerier.CheckAccountInstrumentOwnershipExistence(ctx, q.readDB, &generated.CheckAccountInstrumentOwnershipExistenceParams{
		ID:               accountInstrumentOwnershipID,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing account instrument ownership existence check")
	}

	return result, nil
}

// GetAccountInstrumentOwnership fetches an account instrument ownership from the database.
func (q *repository) GetAccountInstrumentOwnership(ctx context.Context, accountInstrumentOwnershipID, accountID string) (*types.AccountInstrumentOwnership, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if accountInstrumentOwnershipID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.AccountInstrumentOwnershipIDKey, accountInstrumentOwnershipID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, accountInstrumentOwnershipID)

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := q.generatedQuerier.GetAccountInstrumentOwnership(ctx, q.readDB, &generated.GetAccountInstrumentOwnershipParams{
		ID:               accountInstrumentOwnershipID,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account instrument ownership")
	}

	accountInstrumentOwnership := &types.AccountInstrumentOwnership{
		CreatedAt:        result.CreatedAt,
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:               result.ID,
		Notes:            result.Notes,
		BelongsToAccount: result.BelongsToAccount,
		Quantity:         uint16(result.Quantity),
		Instrument: types.ValidInstrument{
			CreatedAt:                      result.ValidInstrumentCreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
			IconPath:                       result.ValidInstrumentIconPath,
			ID:                             result.ValidInstrumentID,
			Name:                           result.ValidInstrumentName,
			PluralName:                     result.ValidInstrumentPluralName,
			Description:                    result.ValidInstrumentDescription,
			Slug:                           result.ValidInstrumentSlug,
			DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
			IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
			UsableForStorage:               result.ValidInstrumentUsableForStorage,
		},
	}

	return accountInstrumentOwnership, nil
}

// GetAccountInstrumentOwnerships fetches a list of account instrument ownerships from the database that meet a particular filter.
func (q *repository) GetAccountInstrumentOwnerships(ctx context.Context, accountID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.AccountInstrumentOwnership], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.AccountInstrumentOwnership
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetAccountInstrumentOwnerships(ctx, q.readDB, &generated.GetAccountInstrumentOwnershipsParams{
		AccountID:       accountID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing account instrument ownerships list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.AccountInstrumentOwnership{
			CreatedAt:        result.CreatedAt,
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:               result.ID,
			Notes:            result.Notes,
			BelongsToAccount: result.BelongsToAccount,
			Quantity:         uint16(result.Quantity),
			Instrument: types.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath,
				ID:                             result.ValidInstrumentID,
				Name:                           result.ValidInstrumentName,
				PluralName:                     result.ValidInstrumentPluralName,
				Description:                    result.ValidInstrumentDescription,
				Slug:                           result.ValidInstrumentSlug,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
				UsableForStorage:               result.ValidInstrumentUsableForStorage,
			},
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(aio *types.AccountInstrumentOwnership) string { return aio.ID },
		filter,
	)

	return x, nil
}

// CreateAccountInstrumentOwnership creates an account instrument ownership in the database.
func (q *repository) CreateAccountInstrumentOwnership(ctx context.Context, input *types.AccountInstrumentOwnershipDatabaseCreationInput) (*types.AccountInstrumentOwnership, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.AccountInstrumentOwnershipIDKey, input.ID)

	// create the account instrument ownership.
	if err := q.generatedQuerier.CreateAccountInstrumentOwnership(ctx, q.writeDB, &generated.CreateAccountInstrumentOwnershipParams{
		ID:                input.ID,
		Notes:             input.Notes,
		ValidInstrumentID: input.ValidInstrumentID,
		BelongsToAccount:  input.BelongsToAccount,
		Quantity:          int32(input.Quantity),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account instrument ownership creation query")
	}

	if _, err := q.auditLogEntryRepo.CreateAuditLogEntry(ctx, q.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &input.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountInstrumentOwnerships,
		RelevantID:       input.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	x := &types.AccountInstrumentOwnership{
		ID:               input.ID,
		Notes:            input.Notes,
		Quantity:         input.Quantity,
		Instrument:       types.ValidInstrument{ID: input.ValidInstrumentID},
		BelongsToAccount: input.BelongsToAccount,
		CreatedAt:        q.CurrentTime(),
	}

	logger.Info("account instrument ownership created")

	return x, nil
}

// UpdateAccountInstrumentOwnership updates a particular account instrument ownership.
func (q *repository) UpdateAccountInstrumentOwnership(ctx context.Context, updated *types.AccountInstrumentOwnership) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.AccountInstrumentOwnershipIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, updated.ID)

	if _, err := q.generatedQuerier.UpdateAccountInstrumentOwnership(ctx, q.writeDB, &generated.UpdateAccountInstrumentOwnershipParams{
		Notes:             updated.Notes,
		ValidInstrumentID: updated.Instrument.ID,
		ID:                updated.ID,
		BelongsToAccount:  updated.BelongsToAccount,
		Quantity:          int32(updated.Quantity),
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating account instrument ownership")
	}

	if _, err := q.auditLogEntryRepo.CreateAuditLogEntry(ctx, q.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &updated.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountInstrumentOwnerships,
		RelevantID:       updated.ID,
		EventType:        audit.AuditLogEventTypeUpdated,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	logger.Info("account instrument ownership updated")

	return nil
}

// ArchiveAccountInstrumentOwnership archives an account instrument ownership from the database by its ID.
func (q *repository) ArchiveAccountInstrumentOwnership(ctx context.Context, accountInstrumentOwnershipID, accountID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if accountInstrumentOwnershipID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.AccountInstrumentOwnershipIDKey, accountInstrumentOwnershipID)
	tracing.AttachToSpan(span, mealplanningkeys.AccountInstrumentOwnershipIDKey, accountInstrumentOwnershipID)

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	rowsAffected, err := q.generatedQuerier.ArchiveAccountInstrumentOwnership(ctx, q.writeDB, &generated.ArchiveAccountInstrumentOwnershipParams{
		ID:               accountInstrumentOwnershipID,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving account instrument ownership")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, q.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountInstrumentOwnerships,
		RelevantID:       accountInstrumentOwnershipID,
		EventType:        audit.AuditLogEventTypeArchived,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	return nil
}
