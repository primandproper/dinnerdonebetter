package identity

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	resourceTypeAccounts = "accounts"
)

var (
	_ identity.AccountDataManager = (*repository)(nil)
)

// GetAccount fetches an account from the database.
func (r *repository) GetAccount(ctx context.Context, accountID string) (*identity.Account, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	results, err := r.generatedQuerier.GetAccountByIDWithMemberships(ctx, r.readDB, accountID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing accounts list retrieval query")
	}

	var account *identity.Account
	for _, result := range results {
		if account == nil {
			account = &identity.Account{
				CreatedAt:                  result.CreatedAt,
				SubscriptionPlanID:         database.StringPointerFromNullString(result.SubscriptionPlanID),
				LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
				ContactPhone:               result.ContactPhone,
				BillingStatus:              result.BillingStatus,
				AddressLine1:               result.AddressLine1,
				AddressLine2:               result.AddressLine2,
				City:                       result.City,
				State:                      result.State,
				ZipCode:                    result.ZipCode,
				Country:                    result.Country,
				Latitude:                   database.Float64PointerFromNullString(result.Latitude),
				Longitude:                  database.Float64PointerFromNullString(result.Longitude),
				PaymentProcessorCustomerID: result.PaymentProcessorCustomerID,
				BelongsToUser:              result.BelongsToUser,
				ID:                         result.ID,
				Name:                       result.Name,
				WebhookEncryptionKey:       result.WebhookHmacSecret,
				Members:                    nil,
			}
		}

		account.Members = append(account.Members, &identity.AccountUserMembershipWithUser{
			CreatedAt:     result.MembershipCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.MembershipLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.MembershipArchivedAt),
			ID:            result.MembershipID,
			BelongsToUser: &identity.User{
				CreatedAt:                  result.UserCreatedAt,
				PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
				LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
				LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
				LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
				TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
				Avatar:                     avatarFromRow(result.UserAvatarID, result.UserAvatarStoragePath, result.UserAvatarMimeType, result.UserAvatarCreatedAt, result.UserAvatarLastUpdatedAt, result.UserAvatarArchivedAt, result.UserAvatarCreatedByUser),
				Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
				ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
				AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
				ID:                         result.UserID,
				AccountStatus:              result.UserUserAccountStatus,
				Username:                   result.UserUsername,
				FirstName:                  result.UserFirstName,
				LastName:                   result.UserLastName,
				EmailAddress:               result.UserEmailAddress,
				EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
				RequiresPasswordChange:     result.UserRequiresPasswordChange,
			},
			BelongsToAccount: result.MembershipBelongsToAccount,
			DefaultAccount:   result.MembershipDefaultAccount,
		})
	}

	if account == nil {
		return nil, sql.ErrNoRows
	}

	return account, nil
}

// getAccountsForUser fetches a list of accounts from the database that meet a particular filter.
func (r *repository) getAccountsForUser(ctx context.Context, querier database.SQLQueryExecutor, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	args := &generated.GetAccountsForUserParams{
		BelongsToUser:   userID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	}
	results, err := r.generatedQuerier.GetAccountsForUser(ctx, querier, args)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing accounts list retrieval query")
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	var (
		data                      []*identity.Account
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &identity.Account{
			CreatedAt:                  result.CreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.SubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.ArchivedAt),
			ContactPhone:               result.ContactPhone,
			BillingStatus:              result.BillingStatus,
			AddressLine1:               result.AddressLine1,
			AddressLine2:               result.AddressLine2,
			City:                       result.City,
			State:                      result.State,
			ZipCode:                    result.ZipCode,
			Country:                    result.Country,
			Latitude:                   database.Float64PointerFromNullString(result.Latitude),
			Longitude:                  database.Float64PointerFromNullString(result.Longitude),
			PaymentProcessorCustomerID: result.PaymentProcessorCustomerID,
			BelongsToUser:              result.BelongsToUser,
			ID:                         result.ID,
			Name:                       result.Name,
			Members:                    nil,
		})
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *identity.Account) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (r *repository) GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[identity.Account], err error) {
	return r.getAccountsForUser(ctx, r.readDB, userID, filter)
}

// CreateAccount creates an account in the database.
func (r *repository) CreateAccount(ctx context.Context, input *identity.AccountDatabaseCreationInput) (*identity.Account, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValue(identitykeys.UserIDKey, input.BelongsToUser)

	// begin account creation transaction
	var err error
	var account *identity.Account
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		// create the account.
		if writeErr := r.generatedQuerier.CreateAccount(ctx, tx, &generated.CreateAccountParams{
			City:              input.City,
			Name:              input.Name,
			BillingStatus:     identity.UnpaidAccountBillingStatus,
			ContactPhone:      input.ContactPhone,
			AddressLine1:      input.AddressLine1,
			AddressLine2:      input.AddressLine2,
			ID:                input.ID,
			State:             input.State,
			ZipCode:           input.ZipCode,
			Country:           input.Country,
			BelongsToUser:     input.BelongsToUser,
			WebhookHmacSecret: input.WebhookEncryptionKey,
			Latitude:          database.NullStringFromFloat64Pointer(input.Latitude),
			Longitude:         database.NullStringFromFloat64Pointer(input.Longitude),
		}); writeErr != nil {
			return observability.PrepareError(writeErr, span, "creating account")
		}

		account = &identity.Account{
			ID:            input.ID,
			Name:          input.Name,
			BelongsToUser: input.BelongsToUser,
			BillingStatus: identity.UnpaidAccountBillingStatus,
			ContactPhone:  input.ContactPhone,
			AddressLine1:  input.AddressLine1,
			AddressLine2:  input.AddressLine2,
			City:          input.City,
			State:         input.State,
			ZipCode:       input.ZipCode,
			Country:       input.Country,
			Latitude:      input.Latitude,
			Longitude:     input.Longitude,
			CreatedAt:     r.CurrentTime(),
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &account.ID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccounts,
			RelevantID:       account.ID,
			EventType:        audit.AuditLogEventTypeCreated,
			BelongsToUser:    account.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		accountMembershipID := identifiers.New()
		if err = r.generatedQuerier.AddUserToAccount(ctx, tx, &generated.AddUserToAccountParams{
			ID:               accountMembershipID,
			BelongsToUser:    account.BelongsToUser,
			BelongsToAccount: account.ID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing account membership creation query")
		}

		// Account creators get account_admin role.
		if err = r.generatedQuerier.AssignRoleToUser(ctx, tx, &generated.AssignRoleToUserParams{
			ID:        identifiers.New(),
			UserID:    account.BelongsToUser,
			RoleID:    authorization.AccountAdminRoleID,
			AccountID: sql.NullString{String: account.ID, Valid: true},
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "assigning account role")
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &account.ID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccountUserMemberships,
			RelevantID:       accountMembershipID,
			EventType:        audit.AuditLogEventTypeCreated,
			BelongsToUser:    account.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// account it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.AccountCreatedServiceEventType, account.ID, map[string]any{
			identitykeys.AccountIDKey: account.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing account created event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, identitykeys.AccountIDKey, account.ID)
	logger.Info("account created")

	return account, nil
}

// UpdateAccountBillingFields updates billing-related fields on an account. Used by the payments domain when processing webhook events.
func (r *repository) UpdateAccountBillingFields(ctx context.Context, accountID string, billingStatus, subscriptionPlanID, paymentProcessorCustomerID *string, lastPaymentProviderSyncOccurredAt *time.Time) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if _, err := r.generatedQuerier.UpdateAccountBillingFields(ctx, r.writeDB, &generated.UpdateAccountBillingFieldsParams{
		ID:                                accountID,
		BillingStatus:                     database.NullStringFromStringPointer(billingStatus),
		SubscriptionPlanID:                database.NullStringFromStringPointer(subscriptionPlanID),
		PaymentProcessorCustomerID:        database.NullStringFromStringPointer(paymentProcessorCustomerID),
		LastPaymentProviderSyncOccurredAt: database.NullTimeFromTimePointer(lastPaymentProviderSyncOccurredAt),
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating account billing fields")
	}

	return nil
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the provided input to have a valid ID.
func (r *repository) UpdateAccount(ctx context.Context, updated *identity.Account) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := r.logger.WithValue(identitykeys.AccountIDKey, updated.ID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, updated.ID)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		account, err := r.GetAccount(ctx, updated.ID)
		if err != nil {
			return observability.PrepareError(err, span, "fetching account")
		}

		if _, err = r.generatedQuerier.UpdateAccount(ctx, tx, &generated.UpdateAccountParams{
			Name:          updated.Name,
			ContactPhone:  updated.ContactPhone,
			AddressLine1:  updated.AddressLine1,
			AddressLine2:  updated.AddressLine2,
			City:          updated.City,
			State:         updated.State,
			ZipCode:       updated.ZipCode,
			Country:       updated.Country,
			BelongsToUser: updated.BelongsToUser,
			ID:            updated.ID,
			Latitude:      database.NullStringFromFloat64Pointer(updated.Latitude),
			Longitude:     database.NullStringFromFloat64Pointer(updated.Longitude),
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating account")
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &updated.ID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccounts,
			RelevantID:       updated.ID,
			EventType:        audit.AuditLogEventTypeUpdated,
			BelongsToUser:    account.BelongsToUser,
			Changes:          buildChangesForAccount(account, updated),
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.AccountUpdatedServiceEventType, updated.ID, map[string]any{
			identitykeys.AccountIDKey: updated.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("account updated")

	return nil
}

func buildChangesForAccount(account, updated *identity.Account) map[string]*audit.ChangeLog {
	changes := map[string]*audit.ChangeLog{}

	if account.Name != updated.Name {
		changes["name"] = &audit.ChangeLog{
			OldValue: account.Name,
			NewValue: updated.Name,
		}
	}

	if account.ContactPhone != updated.ContactPhone {
		changes["contact_phone"] = &audit.ChangeLog{
			OldValue: account.ContactPhone,
			NewValue: updated.ContactPhone,
		}
	}

	if account.AddressLine1 != updated.AddressLine1 {
		changes["address_line_1"] = &audit.ChangeLog{
			OldValue: account.AddressLine1,
			NewValue: updated.AddressLine1,
		}
	}

	if account.AddressLine2 != updated.AddressLine2 {
		changes["address_line_2"] = &audit.ChangeLog{
			OldValue: account.AddressLine2,
			NewValue: updated.AddressLine2,
		}
	}

	if account.City != updated.City {
		changes["city"] = &audit.ChangeLog{
			OldValue: account.City,
			NewValue: updated.City,
		}
	}

	if account.State != updated.State {
		changes["state"] = &audit.ChangeLog{
			OldValue: account.State,
			NewValue: updated.State,
		}
	}

	if account.ZipCode != updated.ZipCode {
		changes["zip_code"] = &audit.ChangeLog{
			OldValue: account.ZipCode,
			NewValue: updated.ZipCode,
		}
	}

	if account.Country != updated.Country {
		changes["country"] = &audit.ChangeLog{
			OldValue: account.Country,
			NewValue: updated.Country,
		}
	}

	if account.Latitude != updated.Latitude {
		changes["latitude"] = &audit.ChangeLog{
			OldValue: fmt.Sprintf("%v", account.Latitude),
			NewValue: fmt.Sprintf("%v", updated.Latitude),
		}
	}

	if account.Longitude != updated.Longitude {
		changes["longitude"] = &audit.ChangeLog{
			OldValue: fmt.Sprintf("%v", account.Longitude),
			NewValue: fmt.Sprintf("%v", updated.Longitude),
		}
	}

	return changes
}

// ArchiveAccount archives an account from the database by its ID.
func (r *repository) ArchiveAccount(ctx context.Context, accountID, ownerID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" || ownerID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)
	logger = logger.WithValue(identitykeys.UserIDKey, ownerID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = r.generatedQuerier.ArchiveAccount(ctx, tx, &generated.ArchiveAccountParams{
			BelongsToUser: ownerID,
			ID:            accountID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving account")
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &accountID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccounts,
			RelevantID:       accountID,
			EventType:        audit.AuditLogEventTypeArchived,
			BelongsToUser:    ownerID,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.AccountArchivedServiceEventType, accountID, map[string]any{
			identitykeys.AccountIDKey: accountID,
			identitykeys.UserIDKey:    ownerID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
