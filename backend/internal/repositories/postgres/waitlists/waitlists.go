package waitlists

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists/generated"

	"github.com/primandproper/platform-go/v12/database"
	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const (
	resourceTypeWaitlists       = "waitlists"
	resourceTypeWaitlistSignups = "waitlist_signups"
)

var (
	_ types.Repository = (*Repository)(nil)
)

// WaitlistIsNotExpired checks if a waitlist exists and is not expired.
func (r *Repository) WaitlistIsNotExpired(ctx context.Context, waitlistID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(waitlistkeys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlistID)

	exists, err := r.generatedQuerier.CheckWaitlistExistence(ctx, r.readDB, waitlistID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking waitlist existence")
	}

	if !exists {
		return false, sql.ErrNoRows
	}

	result, err := r.generatedQuerier.WaitlistIsNotExpired(ctx, r.readDB, waitlistID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking waitlist expiration status")
	}

	return result, nil
}

// GetWaitlist fetches a waitlist from the database.
func (r *Repository) GetWaitlist(ctx context.Context, waitlistID string) (*types.Waitlist, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(waitlistkeys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlistID)

	result, err := r.generatedQuerier.GetWaitlist(ctx, r.readDB, waitlistID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist")
	}

	waitlist := &types.Waitlist{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Name:          result.Name,
		Description:   result.Description,
		ValidUntil:    result.ValidUntil,
	}

	return waitlist, nil
}

// GetWaitlists fetches waitlists with filtering.
func (r *Repository) GetWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Waitlist], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWaitlists(ctx, r.readDB, &generated.GetWaitlistsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlists from database")
	}

	var (
		data                      []*types.Waitlist
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.Waitlist{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Description:   result.Description,
			ValidUntil:    result.ValidUntil,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Waitlist) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetActiveWaitlists fetches non-expired waitlists with filtering.
func (r *Repository) GetActiveWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Waitlist], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetActiveWaitlists(ctx, r.readDB, &generated.GetActiveWaitlistsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching active waitlists from database")
	}

	var (
		data                      []*types.Waitlist
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.Waitlist{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Description:   result.Description,
			ValidUntil:    result.ValidUntil,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Waitlist) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateWaitlist creates a waitlist in the database.
func (r *Repository) CreateWaitlist(ctx context.Context, input *types.WaitlistDatabaseCreationInput) (*types.Waitlist, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := r.logger.WithValue(waitlistkeys.WaitlistIDKey, input.ID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, input.ID)

	if err := r.withEvent(ctx, logger, types.WaitlistCreatedServiceEventType, "", map[string]any{
		waitlistkeys.WaitlistIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		if err := r.generatedQuerier.CreateWaitlist(ctx, tx, &generated.CreateWaitlistParams{
			ID:          input.ID,
			Name:        input.Name,
			Description: input.Description,
			ValidUntil:  input.ValidUntil,
		}); err != nil {
			return err
		}

		// Inside the transaction, not after it. Record takes the caller's executor so
		// the entry commits with the row it describes — and because the chain head is
		// locked for the length of that transaction, which is what stops two
		// concurrent writers from claiming the same position in this scope's chain.
		return r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeWaitlists,
			RelevantID:   input.ID,
			EventType:    audit.AuditLogEventTypeCreated,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing waitlist creation query")
	}

	x := &types.Waitlist{
		ID:          input.ID,
		CreatedAt:   r.CurrentTime(),
		Name:        input.Name,
		Description: input.Description,
		ValidUntil:  input.ValidUntil,
	}

	logger.Info("waitlist created")
	return x, nil
}

// UpdateWaitlist updates a waitlist.
func (r *Repository) UpdateWaitlist(ctx context.Context, updated *types.Waitlist) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := r.logger.WithValue(waitlistkeys.WaitlistIDKey, updated.ID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, updated.ID)

	if err := r.withEvent(ctx, logger, types.WaitlistUpdatedServiceEventType, "", map[string]any{
		waitlistkeys.WaitlistIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		if _, updateErr := r.generatedQuerier.UpdateWaitlist(ctx, tx, &generated.UpdateWaitlistParams{
			Name:        updated.Name,
			Description: updated.Description,
			ValidUntil:  updated.ValidUntil,
			ID:          updated.ID,
		}); updateErr != nil {
			return updateErr
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeWaitlists,
			RelevantID:   updated.ID,
			EventType:    audit.AuditLogEventTypeUpdated,
		})
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating waitlist")
	}

	return nil
}

// ArchiveWaitlist archives a waitlist.
func (r *Repository) ArchiveWaitlist(ctx context.Context, waitlistID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if waitlistID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(waitlistkeys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlistID)

	return r.withEvent(ctx, logger, types.WaitlistArchivedServiceEventType, "", map[string]any{
		waitlistkeys.WaitlistIDKey: waitlistID,
	}, func(tx database.SQLQueryExecutor) error {
		recordsChanged, err := r.generatedQuerier.ArchiveWaitlist(ctx, tx, waitlistID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving waitlist")
		}

		if recordsChanged == 0 {
			return sql.ErrNoRows
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeWaitlists,
			RelevantID:   waitlistID,
			EventType:    audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		logger.Info("waitlist archived")

		return nil
	})
}

// GetWaitlistSignup fetches a waitlist signup from the database.
func (r *Repository) GetWaitlistSignup(ctx context.Context, waitlistSignupID, waitlistID string) (*types.WaitlistSignup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistSignupID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)

	if waitlistID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(waitlistkeys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlistID)

	result, err := r.generatedQuerier.GetWaitlistSignup(ctx, r.readDB, &generated.GetWaitlistSignupParams{
		ID:                waitlistSignupID,
		BelongsToWaitlist: database.NullStringFromString(waitlistID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signup")
	}

	waitlistSignup := &types.WaitlistSignup{
		CreatedAt:         result.CreatedAt,
		LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
		ID:                result.ID,
		Notes:             result.Notes,
		BelongsToWaitlist: database.StringFromNullString(result.BelongsToWaitlist),
		BelongsToUser:     database.StringFromNullString(result.BelongsToUser),
		BelongsToAccount:  database.StringFromNullString(result.BelongsToAccount),
	}

	return waitlistSignup, nil
}

// GetWaitlistSignupByID fetches a waitlist signup from the database by its ID alone.
func (r *Repository) GetWaitlistSignupByID(ctx context.Context, waitlistSignupID string) (*types.WaitlistSignup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistSignupID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)

	result, err := r.generatedQuerier.GetWaitlistSignupByID(ctx, r.readDB, waitlistSignupID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signup")
	}

	waitlistSignup := &types.WaitlistSignup{
		CreatedAt:         result.CreatedAt,
		LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
		ID:                result.ID,
		Notes:             result.Notes,
		BelongsToWaitlist: database.StringFromNullString(result.BelongsToWaitlist),
		BelongsToUser:     database.StringFromNullString(result.BelongsToUser),
		BelongsToAccount:  database.StringFromNullString(result.BelongsToAccount),
	}

	return waitlistSignup, nil
}

// GetWaitlistSignupsForWaitlist fetches waitlist signups for a waitlist with filtering.
func (r *Repository) GetWaitlistSignupsForWaitlist(ctx context.Context, waitlistID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.WaitlistSignup], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(waitlistkeys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, waitlistID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWaitlistSignupsForWaitlist(ctx, r.readDB, &generated.GetWaitlistSignupsForWaitlistParams{
		BelongsToWaitlist: database.NullStringFromString(waitlistID),
		CreatedAfter:      database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:     database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:     database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:      database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived:   database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:            database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:       database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signups from database")
	}

	var (
		data                      []*types.WaitlistSignup
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.WaitlistSignup{
			CreatedAt:         result.CreatedAt,
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			ID:                result.ID,
			Notes:             result.Notes,
			BelongsToWaitlist: database.StringFromNullString(result.BelongsToWaitlist),
			BelongsToUser:     database.StringFromNullString(result.BelongsToUser),
			BelongsToAccount:  database.StringFromNullString(result.BelongsToAccount),
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.WaitlistSignup) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetWaitlistSignupsForUser fetches waitlist signups for a user with filtering.
func (r *Repository) GetWaitlistSignupsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.WaitlistSignup], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWaitlistSignupsForUser(ctx, r.readDB, &generated.GetWaitlistSignupsForUserParams{
		BelongsToUser:   database.NullStringFromString(userID),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signups for user from database")
	}

	var (
		data                      []*types.WaitlistSignup
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.WaitlistSignup{
			CreatedAt:         result.CreatedAt,
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			ID:                result.ID,
			Notes:             result.Notes,
			BelongsToWaitlist: database.StringFromNullString(result.BelongsToWaitlist),
			BelongsToUser:     database.StringFromNullString(result.BelongsToUser),
			BelongsToAccount:  database.StringFromNullString(result.BelongsToAccount),
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.WaitlistSignup) string {
			return t.ID
		},
		filter,
	), nil
}

// CreateWaitlistSignup creates a waitlist signup in the database.
func (r *Repository) CreateWaitlistSignup(ctx context.Context, input *types.WaitlistSignupDatabaseCreationInput) (*types.WaitlistSignup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := r.logger.WithValue(waitlistkeys.WaitlistSignupIDKey, input.ID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, input.ID)

	if err := r.withEvent(ctx, logger, types.WaitlistSignupCreatedServiceEventType, "", map[string]any{
		waitlistkeys.WaitlistSignupIDKey: input.ID,
		waitlistkeys.WaitlistIDKey:       input.BelongsToWaitlist,
	}, func(tx database.SQLQueryExecutor) error {
		if err := r.generatedQuerier.CreateWaitlistSignup(ctx, tx, &generated.CreateWaitlistSignupParams{
			ID:                input.ID,
			Notes:             input.Notes,
			BelongsToWaitlist: database.NullStringFromString(input.BelongsToWaitlist),
			BelongsToUser:     database.NullStringFromString(input.BelongsToUser),
			BelongsToAccount:  database.NullStringFromString(input.BelongsToAccount),
		}); err != nil {
			return err
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &input.BelongsToAccount,
			// Optional rather than absent: a signup can arrive from someone who is not
			// signed in, and the conversion files an empty actor under
			// UnattributedActorID rather than refusing the write.
			BelongsToUser: input.BelongsToUser,
			ResourceType:  resourceTypeWaitlistSignups,
			RelevantID:    input.ID,
			EventType:     audit.AuditLogEventTypeCreated,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing waitlist signup creation query")
	}

	x := &types.WaitlistSignup{
		ID:                input.ID,
		CreatedAt:         r.CurrentTime(),
		Notes:             input.Notes,
		BelongsToWaitlist: input.BelongsToWaitlist,
		BelongsToUser:     input.BelongsToUser,
		BelongsToAccount:  input.BelongsToAccount,
	}

	logger.Info("waitlist signup created")
	return x, nil
}

// UpdateWaitlistSignup updates a waitlist signup.
func (r *Repository) UpdateWaitlistSignup(ctx context.Context, updated *types.WaitlistSignup) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := r.logger.WithValue(waitlistkeys.WaitlistSignupIDKey, updated.ID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, updated.ID)

	return r.withEvent(ctx, logger, types.WaitlistSignupUpdatedServiceEventType, "", map[string]any{
		waitlistkeys.WaitlistSignupIDKey: updated.ID,
		waitlistkeys.WaitlistIDKey:       updated.BelongsToWaitlist,
	}, func(tx database.SQLQueryExecutor) error {
		if _, err := r.generatedQuerier.UpdateWaitlistSignup(ctx, tx, &generated.UpdateWaitlistSignupParams{
			Notes: updated.Notes,
			ID:    updated.ID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating waitlist signup")
		}

		if err := r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &updated.BelongsToAccount,
			ResourceType:     resourceTypeWaitlistSignups,
			RelevantID:       updated.ID,
			EventType:        audit.AuditLogEventTypeUpdated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	})
}

// ArchiveWaitlistSignup archives a waitlist signup.
func (r *Repository) ArchiveWaitlistSignup(ctx context.Context, waitlistSignupID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if waitlistSignupID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, waitlistSignupID)

	return r.withEvent(ctx, logger, types.WaitlistSignupArchivedServiceEventType, "", map[string]any{
		waitlistkeys.WaitlistSignupIDKey: waitlistSignupID,
	}, func(tx database.SQLQueryExecutor) error {
		recordsChanged, err := r.generatedQuerier.ArchiveWaitlistSignup(ctx, tx, waitlistSignupID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving waitlist signup")
		}

		if recordsChanged == 0 {
			return sql.ErrNoRows
		}

		// ArchiveWaitlistSignup does not have account ID in signature; create audit entry without it
		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeWaitlistSignups,
			RelevantID:   waitlistSignupID,
			EventType:    audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		logger.Info("waitlist signup archived")

		return nil
	})
}
