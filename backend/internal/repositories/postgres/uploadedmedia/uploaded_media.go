package uploadedmedia

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediakeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/keys"
	generated "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia/generated"

	"github.com/primandproper/platform-go/v7/database"
	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/identifiers"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

const (
	resourceTypeUploadedMedia = "uploaded_media"
)

var (
	_ types.UploadedMediaDataManager = (*repository)(nil)
)

// GetUploadedMedia fetches uploaded media from the database.
func (r *repository) GetUploadedMedia(ctx context.Context, uploadedMediaID string) (*types.UploadedMedia, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if uploadedMediaID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(uploadedmediakeys.UploadedMediaIDKey, uploadedMediaID)
	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, uploadedMediaID)

	result, err := r.generatedQuerier.GetUploadedMedia(ctx, r.readDB, uploadedMediaID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media")
	}

	uploadedMedia := &types.UploadedMedia{
		ID:            result.ID,
		StoragePath:   result.StoragePath,
		MimeType:      string(result.MimeType),
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		CreatedByUser: result.CreatedByUser,
	}

	return uploadedMedia, nil
}

// GetUploadedMediaWithIDs fetches a list of uploaded media from the database by their IDs.
func (r *repository) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*types.UploadedMedia, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if len(ids) == 0 {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue("ids", ids)
	tracing.AttachToSpan(span, "id_count", len(ids))

	results, err := r.generatedQuerier.GetUploadedMediaWithIDs(ctx, r.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media with IDs")
	}

	var uploadedMediaList []*types.UploadedMedia
	for _, result := range results {
		uploadedMediaList = append(uploadedMediaList, &types.UploadedMedia{
			ID:            result.ID,
			StoragePath:   result.StoragePath,
			MimeType:      string(result.MimeType),
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser: result.CreatedByUser,
		})
	}

	return uploadedMediaList, nil
}

// GetUploadedMediaForUser fetches a list of uploaded media for a specific user from the database that meet a particular filter.
func (r *repository) GetUploadedMediaForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.UploadedMedia], error) {
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

	results, err := r.generatedQuerier.GetUploadedMediaForUser(ctx, r.readDB, &generated.GetUploadedMediaForUserParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		CreatedByUser:   userID,
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media from database")
	}

	var (
		data                      []*types.UploadedMedia
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.UploadedMedia{
			ID:            result.ID,
			StoragePath:   result.StoragePath,
			MimeType:      string(result.MimeType),
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser: result.CreatedByUser,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.UploadedMedia) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateUploadedMedia creates uploaded media in the database.
func (r *repository) CreateUploadedMedia(ctx context.Context, input *types.UploadedMediaDatabaseCreationInput) (*types.UploadedMedia, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.CreatedByUser)
	logger = logger.WithValue(identitykeys.UserIDKey, input.CreatedByUser)

	logger.Debug("CreateUploadedMedia invoked")

	var err error
	var x *types.UploadedMedia
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.CreateUploadedMedia(ctx, tx, &generated.CreateUploadedMediaParams{
			ID:            input.ID,
			StoragePath:   input.StoragePath,
			MimeType:      generated.UploadedMediaMimeType(input.MimeType),
			CreatedByUser: input.CreatedByUser,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing uploaded media creation query")
		}

		x = &types.UploadedMedia{
			ID:            input.ID,
			StoragePath:   input.StoragePath,
			MimeType:      input.MimeType,
			CreatedByUser: input.CreatedByUser,
			CreatedAt:     r.CurrentTime(),
		}

		userID := x.CreatedByUser
		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToUser: userID,
			ID:            identifiers.New(),
			ResourceType:  resourceTypeUploadedMedia,
			RelevantID:    x.ID,
			EventType:     audit.AuditLogEventTypeCreated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, x.ID)

	return x, nil
}

// UpdateUploadedMedia updates uploaded media in the database.
func (r *repository) UpdateUploadedMedia(ctx context.Context, uploadedMedia *types.UploadedMedia) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if uploadedMedia == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(uploadedmediakeys.UploadedMediaIDKey, uploadedMedia.ID)
	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, uploadedMedia.ID)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, err := r.generatedQuerier.UpdateUploadedMedia(ctx, tx, &generated.UpdateUploadedMediaParams{
			StoragePath: uploadedMedia.StoragePath,
			MimeType:    generated.UploadedMediaMimeType(uploadedMedia.MimeType),
			ID:          uploadedMedia.ID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating uploaded media")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		userID := uploadedMedia.CreatedByUser
		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToUser: userID,
			ID:            identifiers.New(),
			ResourceType:  resourceTypeUploadedMedia,
			RelevantID:    uploadedMedia.ID,
			EventType:     audit.AuditLogEventTypeUpdated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("uploaded media updated")

	return nil
}

// ArchiveUploadedMedia archives uploaded media in the database.
func (r *repository) ArchiveUploadedMedia(ctx context.Context, uploadedMediaID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if uploadedMediaID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(uploadedmediakeys.UploadedMediaIDKey, uploadedMediaID)
	tracing.AttachToSpan(span, uploadedmediakeys.UploadedMediaIDKey, uploadedMediaID)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, err := r.generatedQuerier.ArchiveUploadedMedia(ctx, tx, uploadedMediaID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving uploaded media")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			ID:           identifiers.New(),
			ResourceType: resourceTypeUploadedMedia,
			RelevantID:   uploadedMediaID,
			EventType:    audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("uploaded media archived")

	return nil
}
