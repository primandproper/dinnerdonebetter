package settings

import (
	"context"
	"database/sql"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	platformkeys "github.com/primandproper/platform-go/v13/observability/keys"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeServiceSettings  = "service_settings"
	serviceSettingsEnumDelimiter = "|"
)

var (
	_ types.ServiceSettingDataManager = (*Repository)(nil)
)

// ServiceSettingExists fetches whether a service setting exists from the database.
func (q *Repository) ServiceSettingExists(ctx context.Context, serviceSettingID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingIDKey, serviceSettingID)

	result, err := q.generatedQuerier.CheckServiceSettingExistence(ctx, q.readDB, serviceSettingID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing service setting existence check")
	}

	return result, nil
}

// GetServiceSetting fetches a service setting from the database.
func (q *Repository) GetServiceSetting(ctx context.Context, serviceSettingID string) (*types.ServiceSetting, error) {
	return q.getServiceSetting(ctx, q.readDB, serviceSettingID)
}

// getServiceSetting fetches a service setting from the database.
func (q *Repository) getServiceSetting(ctx context.Context, db database.SQLQueryExecutor, serviceSettingID string) (*types.ServiceSetting, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingIDKey, serviceSettingID)

	result, err := q.generatedQuerier.GetServiceSetting(ctx, db, serviceSettingID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing service setting fetch")
	}

	usableEnumeration := []string{}
	for x := range strings.SplitSeq(result.Enumeration, serviceSettingsEnumDelimiter) {
		if strings.TrimSpace(x) != "" {
			usableEnumeration = append(usableEnumeration, x)
		}
	}

	serviceSetting := &types.ServiceSetting{
		CreatedAt:     result.CreatedAt,
		DefaultValue:  database.StringPointerFromNullString(result.DefaultValue),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Name:          result.Name,
		Type:          string(result.Type),
		Description:   result.Description,
		Enumeration:   usableEnumeration,
		AdminsOnly:    result.AdminsOnly,
	}

	return serviceSetting, nil
}

// SearchForServiceSettings fetches a service setting from the database.
func (q *Repository) SearchForServiceSettings(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSetting], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForServiceSettings(ctx, q.readDB, &generated.SearchForServiceSettingsParams{
		NameQuery:       query,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing service settings list retrieval query")
	}

	var (
		data                      = []*types.ServiceSetting{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		rawEnumeration := strings.Split(result.Enumeration, serviceSettingsEnumDelimiter)
		usableEnumeration := []string{}
		for _, y := range rawEnumeration {
			if strings.TrimSpace(y) != "" {
				usableEnumeration = append(usableEnumeration, y)
			}
		}

		serviceSetting := &types.ServiceSetting{
			CreatedAt:     result.CreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.DefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Type:          string(result.Type),
			Description:   result.Description,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.AdminsOnly,
		}

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
		data = append(data, serviceSetting)
	}

	result := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.ServiceSetting) string {
			return t.ID
		},
		filter,
	)

	return result, nil
}

// GetServiceSettings fetches a list of service settings from the database that meet a particular filter.
func (q *Repository) GetServiceSettings(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSetting], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetServiceSettings(ctx, q.readDB, &generated.GetServiceSettingsParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing service settings list retrieval query")
	}

	var (
		data                      = []*types.ServiceSetting{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		rawEnumeration := strings.Split(result.Enumeration, serviceSettingsEnumDelimiter)
		usableEnumeration := []string{}
		for _, y := range rawEnumeration {
			if strings.TrimSpace(y) != "" {
				usableEnumeration = append(usableEnumeration, y)
			}
		}

		data = append(data, &types.ServiceSetting{
			CreatedAt:     result.CreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.DefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Type:          string(result.Type),
			Description:   result.Description,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.AdminsOnly,
		})
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.ServiceSetting) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateServiceSetting creates a service setting in the database.
func (q *Repository) CreateServiceSetting(ctx context.Context, input *types.ServiceSettingDatabaseCreationInput) (*types.ServiceSetting, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, settingskeys.ServiceSettingIDKey, input.ID)
	logger := q.logger.WithValue(settingskeys.ServiceSettingIDKey, input.ID)

	var err error
	var x *types.ServiceSetting
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		// create the service setting.
		if err = q.generatedQuerier.CreateServiceSetting(ctx, tx, &generated.CreateServiceSettingParams{
			ID:           input.ID,
			Name:         input.Name,
			Type:         generated.SettingType(input.Type),
			Description:  input.Description,
			Enumeration:  strings.Join(input.Enumeration, serviceSettingsEnumDelimiter),
			DefaultValue: database.NullStringFromStringPointer(input.DefaultValue),
			AdminsOnly:   input.AdminsOnly,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing service setting creation query")
		}

		x = &types.ServiceSetting{
			ID:           input.ID,
			Name:         input.Name,
			Type:         input.Type,
			Description:  input.Description,
			DefaultValue: input.DefaultValue,
			AdminsOnly:   input.AdminsOnly,
			Enumeration:  input.Enumeration,
			CreatedAt:    q.CurrentTime(),
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeServiceSettings,
			RelevantID:   x.ID,
			EventType:    audit.AuditLogEventTypeCreated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.ServiceSettingCreatedServiceEventType, "", map[string]any{
			settingskeys.ServiceSettingIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("service setting created")

	return x, nil
}

// ArchiveServiceSetting archives a service setting from the database by its ID.
func (q *Repository) ArchiveServiceSetting(ctx context.Context, serviceSettingID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingIDKey, serviceSettingID)

	if err := q.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, err := q.generatedQuerier.ArchiveServiceSetting(ctx, tx, serviceSettingID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating service setting")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType: resourceTypeServiceSettings,
			RelevantID:   serviceSettingID,
			EventType:    audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.ServiceSettingArchivedServiceEventType, "", map[string]any{
			settingskeys.ServiceSettingIDKey: serviceSettingID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
