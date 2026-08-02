package settings

import (
	"context"
	"database/sql"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	resourceTypeServiceSettingConfigurations = "service_setting_configurations"
)

var (
	_ types.ServiceSettingConfigurationDataManager = (*Repository)(nil)
)

// ServiceSettingConfigurationExists fetches whether a service setting configuration exists from the database.
func (q *Repository) ServiceSettingConfigurationExists(ctx context.Context, serviceSettingConfigurationID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingConfigurationID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)

	result, err := q.generatedQuerier.CheckServiceSettingConfigurationExistence(ctx, q.readDB, serviceSettingConfigurationID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing service setting configuration existence check")
	}

	return result, nil
}

// GetServiceSettingConfiguration fetches a service setting configuration from the database.
func (q *Repository) GetServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) (*types.ServiceSettingConfiguration, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingConfigurationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)
	logger = logger.WithValue(settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)

	result, err := q.generatedQuerier.GetServiceSettingConfigurationByID(ctx, q.readDB, serviceSettingConfigurationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching service setting configuration")
	}

	usableEnumeration := []string{}
	for x := range strings.SplitSeq(result.ServiceSettingEnumeration, serviceSettingsEnumDelimiter) {
		if strings.TrimSpace(x) != "" {
			usableEnumeration = append(usableEnumeration, x)
		}
	}

	serviceSettingConfiguration := &types.ServiceSettingConfiguration{
		CreatedAt:        result.CreatedAt,
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		ID:               result.ID,
		Value:            result.Value,
		Notes:            result.Notes,
		BelongsToUser:    result.BelongsToUser,
		BelongsToAccount: result.BelongsToAccount,
		ServiceSetting: types.ServiceSetting{
			CreatedAt:     result.ServiceSettingCreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.ServiceSettingDefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.ServiceSettingLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ServiceSettingArchivedAt),
			ID:            result.ServiceSettingID,
			Name:          result.ServiceSettingName,
			Type:          string(result.ServiceSettingType),
			Description:   result.ServiceSettingDescription,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.ServiceSettingAdminsOnly,
		},
	}

	return serviceSettingConfiguration, nil
}

// GetServiceSettingConfigurationForUserByName fetches a service setting configuration from the database.
func (q *Repository) GetServiceSettingConfigurationForUserByName(ctx context.Context, userID, settingName string) (*types.ServiceSettingConfiguration, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	if settingName == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingNameKey, settingName)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingNameKey, settingName)

	result, err := q.generatedQuerier.GetServiceSettingConfigurationForUserBySettingName(ctx, q.readDB, &generated.GetServiceSettingConfigurationForUserBySettingNameParams{
		Name:          settingName,
		BelongsToUser: userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching service setting configuration")
	}

	usableEnumeration := []string{}
	for x := range strings.SplitSeq(result.ServiceSettingEnumeration, serviceSettingsEnumDelimiter) {
		if strings.TrimSpace(x) != "" {
			usableEnumeration = append(usableEnumeration, x)
		}
	}

	serviceSettingConfiguration := &types.ServiceSettingConfiguration{
		CreatedAt:        result.CreatedAt,
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		ID:               result.ID,
		Value:            result.Value,
		Notes:            result.Notes,
		BelongsToUser:    result.BelongsToUser,
		BelongsToAccount: result.BelongsToAccount,
		ServiceSetting: types.ServiceSetting{
			CreatedAt:     result.ServiceSettingCreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.ServiceSettingDefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.ServiceSettingLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ServiceSettingArchivedAt),
			ID:            result.ServiceSettingID,
			Name:          result.ServiceSettingName,
			Type:          string(result.ServiceSettingType),
			Description:   result.ServiceSettingDescription,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.ServiceSettingAdminsOnly,
		},
	}

	return serviceSettingConfiguration, nil
}

// GetServiceSettingConfigurationForAccountByName fetches a service setting configuration from the database.
func (q *Repository) GetServiceSettingConfigurationForAccountByName(ctx context.Context, accountID, settingName string) (*types.ServiceSettingConfiguration, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)

	if settingName == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingNameKey, settingName)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingNameKey, settingName)

	result, err := q.generatedQuerier.GetServiceSettingConfigurationForAccountBySettingName(ctx, q.readDB, &generated.GetServiceSettingConfigurationForAccountBySettingNameParams{
		Name:             settingName,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching service setting configuration")
	}

	usableEnumeration := []string{}
	for x := range strings.SplitSeq(result.ServiceSettingEnumeration, serviceSettingsEnumDelimiter) {
		if strings.TrimSpace(x) != "" {
			usableEnumeration = append(usableEnumeration, x)
		}
	}

	serviceSettingConfiguration := &types.ServiceSettingConfiguration{
		CreatedAt:        result.CreatedAt,
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		ID:               result.ID,
		Value:            result.Value,
		Notes:            result.Notes,
		BelongsToUser:    result.BelongsToUser,
		BelongsToAccount: result.BelongsToAccount,
		ServiceSetting: types.ServiceSetting{
			CreatedAt:     result.ServiceSettingCreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.ServiceSettingDefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.ServiceSettingLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ServiceSettingArchivedAt),
			ID:            result.ServiceSettingID,
			Name:          result.ServiceSettingName,
			Type:          string(result.ServiceSettingType),
			Description:   result.ServiceSettingDescription,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.ServiceSettingAdminsOnly,
		},
	}

	return serviceSettingConfiguration, nil
}

// GetServiceSettingConfigurationsForUser fetches a list of service setting configurations from the database that meet a particular filter.
func (q *Repository) GetServiceSettingConfigurationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSettingConfiguration], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	filter.AttachToLogger(logger)

	results, err := q.generatedQuerier.GetServiceSettingConfigurationsForUser(ctx, q.readDB, &generated.GetServiceSettingConfigurationsForUserParams{
		BelongsToUser:   userID,
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing service setting configurations list retrieval query")
	}

	var (
		data                      = []*types.ServiceSettingConfiguration{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		usableEnumeration := []string{}
		for x := range strings.SplitSeq(result.ServiceSettingEnumeration, serviceSettingsEnumDelimiter) {
			if strings.TrimSpace(x) != "" {
				usableEnumeration = append(usableEnumeration, x)
			}
		}

		serviceSettingConfiguration := &types.ServiceSettingConfiguration{
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			ID:               result.ID,
			Value:            result.Value,
			Notes:            result.Notes,
			BelongsToUser:    result.BelongsToUser,
			BelongsToAccount: result.BelongsToAccount,
			ServiceSetting: types.ServiceSetting{
				CreatedAt:     result.ServiceSettingCreatedAt,
				DefaultValue:  database.StringPointerFromNullString(result.ServiceSettingDefaultValue),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ServiceSettingLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ServiceSettingArchivedAt),
				ID:            result.ServiceSettingID,
				Name:          result.ServiceSettingName,
				Type:          string(result.ServiceSettingType),
				Description:   result.ServiceSettingDescription,
				Enumeration:   usableEnumeration,
				AdminsOnly:    result.ServiceSettingAdminsOnly,
			},
		}

		data = append(data, serviceSettingConfiguration)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.ServiceSettingConfiguration) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetServiceSettingConfigurationsForAccount fetches a list of service setting configurations from the database that meet a particular filter.
func (q *Repository) GetServiceSettingConfigurationsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSettingConfiguration], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	filter.AttachToLogger(logger)

	results, err := q.generatedQuerier.GetServiceSettingConfigurationsForAccount(ctx, q.readDB, &generated.GetServiceSettingConfigurationsForAccountParams{
		BelongsToAccount: accountID,
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing service setting configurations list retrieval query")
	}

	var (
		data                      = []*types.ServiceSettingConfiguration{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		usableEnumeration := []string{}
		for x := range strings.SplitSeq(result.ServiceSettingEnumeration, serviceSettingsEnumDelimiter) {
			if strings.TrimSpace(x) != "" {
				usableEnumeration = append(usableEnumeration, x)
			}
		}

		serviceSettingConfiguration := &types.ServiceSettingConfiguration{
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			ID:               result.ID,
			Value:            result.Value,
			Notes:            result.Notes,
			BelongsToUser:    result.BelongsToUser,
			BelongsToAccount: result.BelongsToAccount,
			ServiceSetting: types.ServiceSetting{
				CreatedAt:     result.ServiceSettingCreatedAt,
				DefaultValue:  database.StringPointerFromNullString(result.ServiceSettingDefaultValue),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ServiceSettingLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ServiceSettingArchivedAt),
				ID:            result.ServiceSettingID,
				Name:          result.ServiceSettingName,
				Type:          string(result.ServiceSettingType),
				Description:   result.ServiceSettingDescription,
				Enumeration:   usableEnumeration,
				AdminsOnly:    result.ServiceSettingAdminsOnly,
			},
		}

		data = append(data, serviceSettingConfiguration)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.ServiceSettingConfiguration) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateServiceSettingConfiguration creates a service setting configuration in the database.
func (q *Repository) CreateServiceSettingConfiguration(ctx context.Context, input *types.ServiceSettingConfigurationDatabaseCreationInput) (*types.ServiceSettingConfiguration, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, input.ID)
	logger := q.logger.WithValue(settingskeys.ServiceSettingConfigurationIDKey, input.ID)

	// begin account creation transaction
	var x *types.ServiceSettingConfiguration
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		// create the service setting configuration.
		if err := q.generatedQuerier.CreateServiceSettingConfiguration(ctx, tx, &generated.CreateServiceSettingConfigurationParams{
			ID:               input.ID,
			Value:            input.Value,
			Notes:            input.Notes,
			ServiceSettingID: input.ServiceSettingID,
			BelongsToUser:    input.BelongsToUser,
			BelongsToAccount: input.BelongsToAccount,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing service setting configuration creation query")
		}

		serviceSetting, err := q.getServiceSetting(ctx, tx, input.ServiceSettingID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching service setting")
		}

		x = &types.ServiceSettingConfiguration{
			ID:               input.ID,
			Value:            input.Value,
			Notes:            input.Notes,
			ServiceSetting:   *serviceSetting,
			BelongsToUser:    input.BelongsToUser,
			BelongsToAccount: input.BelongsToAccount,
			CreatedAt:        q.CurrentTime(),
		}

		if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &input.BelongsToAccount,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeServiceSettingConfigurations,
			RelevantID:       x.ID,
			EventType:        audit.AuditLogEventTypeCreated,
			BelongsToUser:    input.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.ServiceSettingConfigurationCreatedServiceEventType, "", map[string]any{
			settingskeys.ServiceSettingConfigurationIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("service setting configuration created")

	return x, nil
}

// UpdateServiceSettingConfiguration updates a particular service setting configuration.
func (q *Repository) UpdateServiceSettingConfiguration(ctx context.Context, updated *types.ServiceSettingConfiguration) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(settingskeys.ServiceSettingConfigurationIDKey, updated.ID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, updated.ID)

	// begin account creation transaction
	var err error
	if err = q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = q.generatedQuerier.UpdateServiceSettingConfiguration(ctx, tx, &generated.UpdateServiceSettingConfigurationParams{
			Value:            updated.Value,
			Notes:            updated.Notes,
			ServiceSettingID: updated.ServiceSetting.ID,
			BelongsToUser:    updated.BelongsToUser,
			BelongsToAccount: updated.BelongsToAccount,
			ID:               updated.ID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating service setting configuration")
		}

		if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &updated.BelongsToAccount,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeServiceSettingConfigurations,
			RelevantID:       updated.ID,
			EventType:        audit.AuditLogEventTypeUpdated,
			BelongsToUser:    updated.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.ServiceSettingConfigurationUpdatedServiceEventType, "", map[string]any{
			settingskeys.ServiceSettingConfigurationIDKey: updated.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("service setting configuration updated")

	return nil
}

// ArchiveServiceSettingConfiguration archives a service setting configuration from the database by its ID.
func (q *Repository) ArchiveServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingConfigurationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)

	// begin account creation transaction
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, err := q.generatedQuerier.ArchiveServiceSettingConfiguration(ctx, tx, serviceSettingConfigurationID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving service setting configuration")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		// ArchiveServiceSettingConfiguration does not have account ID in signature; create audit entry without it
		if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			ID:           identifiers.New(),
			ResourceType: resourceTypeServiceSettingConfigurations,
			RelevantID:   serviceSettingConfigurationID,
			EventType:    audit.AuditLogEventTypeArchived,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.ServiceSettingConfigurationArchivedServiceEventType, "", map[string]any{
			settingskeys.ServiceSettingConfigurationIDKey: serviceSettingConfigurationID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
