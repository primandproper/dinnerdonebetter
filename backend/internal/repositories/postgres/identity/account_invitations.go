package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeAccountInvitations = "account_invitations"
)

var (
	_ identity.AccountInvitationDataManager = (*repository)(nil)
)

// AccountInvitationExists fetches whether an account invitation exists from the database.
func (r *repository) AccountInvitationExists(ctx context.Context, accountInvitationID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountInvitationID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, identitykeys.AccountInvitationIDKey, accountInvitationID)

	result, err := r.generatedQuerier.CheckAccountInvitationExistence(ctx, r.readDB, accountInvitationID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing account invitation existence check")
	}

	return result, nil
}

// GetAccountInvitationByAccountAndID fetches an invitation from the database.
func (r *repository) GetAccountInvitationByAccountAndID(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if accountInvitationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, identitykeys.AccountInvitationIDKey, accountInvitationID)

	result, err := r.generatedQuerier.GetAccountInvitationByAccountAndID(ctx, r.readDB, &generated.GetAccountInvitationByAccountAndIDParams{
		DestinationAccount: accountID,
		ID:                 accountInvitationID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	accountInvitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			Avatar:                     r.avatarFor(ctx, result.UserAvatarID),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return accountInvitation, nil
}

// GetAccountInvitationByTokenAndID fetches an invitation from the database.
func (r *repository) GetAccountInvitationByTokenAndID(ctx context.Context, token, invitationID string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if token == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	if invitationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountInvitationIDKey, invitationID)
	tracing.AttachToSpan(span, identitykeys.AccountInvitationIDKey, invitationID)

	logger.Debug("fetching account invitation")

	result, err := r.generatedQuerier.GetAccountInvitationByTokenAndID(ctx, r.readDB, &generated.GetAccountInvitationByTokenAndIDParams{
		Token: token,
		ID:    invitationID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	accountInvitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			Avatar:                     r.avatarFor(ctx, result.UserAvatarID),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return accountInvitation, nil
}

// GetAccountInvitationByToken fetches an invitation from the database.
func (r *repository) GetAccountInvitationByToken(ctx context.Context, token string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if token == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	logger.Debug("fetching account invitation")

	result, err := r.generatedQuerier.GetAccountInvitationByToken(ctx, r.readDB, token)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	accountInvitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			Avatar:                     r.avatarFor(ctx, result.UserAvatarID),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return accountInvitation, nil
}

// GetAccountInvitationByEmailAndToken fetches an invitation from the database.
func (r *repository) GetAccountInvitationByEmailAndToken(ctx context.Context, emailAddress, token string) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if emailAddress == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserEmailAddressKey, emailAddress)
	tracing.AttachToSpan(span, identitykeys.UserEmailAddressKey, emailAddress)

	if token == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	result, err := r.generatedQuerier.GetAccountInvitationByEmailAndToken(ctx, r.readDB, &generated.GetAccountInvitationByEmailAndTokenParams{
		ToEmail: emailAddress,
		Token:   token,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
	}

	invitation := &identity.AccountInvitation{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ToUser:        database.StringPointerFromNullString(result.ToUser),
		Status:        string(result.Status),
		ToEmail:       result.ToEmail,
		StatusNote:    result.StatusNote,
		Token:         result.Token,
		ID:            result.ID,
		Note:          result.Note,
		ToName:        result.ToName,
		ExpiresAt:     result.ExpiresAt,
		DestinationAccount: identity.Account{
			CreatedAt:                  result.AccountCreatedAt,
			SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
			ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
			ContactPhone:               result.AccountContactPhone,
			BillingStatus:              result.AccountBillingStatus,
			AddressLine1:               result.AccountAddressLine1,
			AddressLine2:               result.AccountAddressLine2,
			City:                       result.AccountCity,
			State:                      result.AccountState,
			ZipCode:                    result.AccountZipCode,
			Country:                    result.AccountCountry,
			Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
			Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
			PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
			BelongsToUser:              result.AccountBelongsToUser,
			ID:                         result.AccountID,
			Name:                       result.AccountName,
			Members:                    nil,
		},
		FromUser: identity.User{
			CreatedAt:                  result.UserCreatedAt,
			PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
			LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
			LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
			LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
			TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
			Avatar:                     r.avatarFor(ctx, result.UserAvatarID),
			Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
			ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
			AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
			TwoFactorSecret:            result.UserTwoFactorSecret,
			HashedPassword:             result.UserHashedPassword,
			ID:                         result.UserID,
			AccountStatus:              result.UserUserAccountStatus,
			Username:                   result.UserUsername,
			FirstName:                  result.UserFirstName,
			LastName:                   result.UserLastName,
			EmailAddress:               result.UserEmailAddress,
			EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
			RequiresPasswordChange:     result.UserRequiresPasswordChange,
		},
	}

	return invitation, nil
}

// CreateAccountInvitation creates an invitation in a database.
func (r *repository) CreateAccountInvitation(ctx context.Context, input *identity.AccountInvitationDatabaseCreationInput) (*identity.AccountInvitation, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := r.logger.WithValue(identitykeys.AccountInvitationIDKey, input.ID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.DestinationAccountID)

	if input.ToUser == nil && input.ToEmail != "" {
		if invitee, err := r.GetUserByEmail(ctx, input.ToEmail); err == nil {
			input.ToUser = &invitee.ID
		}
	}

	if err := r.withEvent(ctx, logger, identity.AccountInvitationCreatedServiceEventType, input.DestinationAccountID, map[string]any{
		identitykeys.AccountInvitationIDKey:  input.ID,
		identitykeys.UserIDKey:               input.FromUser,
		identitykeys.DestinationAccountIDKey: input.DestinationAccountID,
	}, func(tx database.Tx) error {
		if err := r.generatedQuerier.CreateAccountInvitation(ctx, tx, &generated.CreateAccountInvitationParams{
			ExpiresAt:          input.ExpiresAt,
			ID:                 input.ID,
			FromUser:           input.FromUser,
			ToName:             input.ToName,
			Note:               input.Note,
			ToEmail:            input.ToEmail,
			Token:              input.Token,
			DestinationAccount: input.DestinationAccountID,
			ToUser:             database.NullStringFromStringPointer(input.ToUser),
		}); err != nil {
			return err
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &input.DestinationAccountID,
			ResourceType:     resourceTypeAccountInvitations,
			RelevantID:       input.ID,
			EventType:        audit.AuditLogEventTypeCreated,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account invitation creation query")
	}

	x := &identity.AccountInvitation{
		ID:                 input.ID,
		FromUser:           identity.User{ID: input.FromUser},
		ToUser:             input.ToUser,
		Note:               input.Note,
		ToName:             input.ToName,
		ToEmail:            input.ToEmail,
		Token:              input.Token,
		StatusNote:         "",
		Status:             string(identity.PendingAccountInvitationStatus),
		DestinationAccount: identity.Account{ID: input.DestinationAccountID},
		ExpiresAt:          input.ExpiresAt,
		CreatedAt:          r.CurrentTime(),
	}

	tracing.AttachToSpan(span, identitykeys.AccountInvitationIDKey, x.ID)
	logger = logger.WithValue(identitykeys.AccountInvitationIDKey, x.ID)

	logger.Info("account invitation created")

	return x, nil
}

// GetPendingAccountInvitationsFromUser fetches pending account invitations sent from a given user.
func (r *repository) GetPendingAccountInvitationsFromUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)
	filter.AttachToLogger(logger)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := r.generatedQuerier.GetPendingInvitesFromUser(ctx, r.readDB, &generated.GetPendingInvitesFromUserParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		Status:          generated.InvitationState(identity.PendingAccountInvitationStatus),
		FromUser:        userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account invitation query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.GetPendingInvitesFromUserRow) *identity.AccountInvitation {
			return &identity.AccountInvitation{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				ToUser:        database.StringPointerFromNullString(result.ToUser),
				Status:        string(result.Status),
				ToEmail:       result.ToEmail,
				StatusNote:    result.StatusNote,
				Token:         result.Token,
				ID:            result.ID,
				Note:          result.Note,
				ToName:        result.ToName,
				ExpiresAt:     result.ExpiresAt,
				DestinationAccount: identity.Account{
					CreatedAt:                  result.AccountCreatedAt,
					SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
					LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
					ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
					ContactPhone:               result.AccountContactPhone,
					BillingStatus:              result.AccountBillingStatus,
					AddressLine1:               result.AccountAddressLine1,
					AddressLine2:               result.AccountAddressLine2,
					City:                       result.AccountCity,
					State:                      result.AccountState,
					ZipCode:                    result.AccountZipCode,
					Country:                    result.AccountCountry,
					Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
					Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
					PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
					BelongsToUser:              result.AccountBelongsToUser,
					ID:                         result.AccountID,
					Name:                       result.AccountName,
					Members:                    nil,
				},
				FromUser: identity.User{
					CreatedAt:                  result.UserCreatedAt,
					PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
					LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
					LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
					LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
					TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
					Avatar:                     r.avatarFor(ctx, result.UserAvatarID),
					Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
					ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
					AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
					TwoFactorSecret:            result.UserTwoFactorSecret,
					HashedPassword:             result.UserHashedPassword,
					ID:                         result.UserID,
					AccountStatus:              result.UserUserAccountStatus,
					Username:                   result.UserUsername,
					FirstName:                  result.UserFirstName,
					LastName:                   result.UserLastName,
					EmailAddress:               result.UserEmailAddress,
					EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
					RequiresPasswordChange:     result.UserRequiresPasswordChange,
				},
			}
		},
		func(result *generated.GetPendingInvitesFromUserRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(t *identity.AccountInvitation) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetPendingAccountInvitationsForUser fetches pending account invitations sent to a given user.
func (r *repository) GetPendingAccountInvitationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)
	filter.AttachToLogger(logger)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := r.generatedQuerier.GetPendingInvitesForUser(ctx, r.readDB, &generated.GetPendingInvitesForUserParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		Status:          generated.InvitationState(identity.PendingAccountInvitationStatus),
		ToUser:          database.NullStringFromString(userID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing account invitation query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.GetPendingInvitesForUserRow) *identity.AccountInvitation {
			return &identity.AccountInvitation{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				ToUser:        database.StringPointerFromNullString(result.ToUser),
				Status:        string(result.Status),
				ToEmail:       result.ToEmail,
				StatusNote:    result.StatusNote,
				Token:         result.Token,
				ID:            result.ID,
				Note:          result.Note,
				ToName:        result.ToName,
				ExpiresAt:     result.ExpiresAt,
				DestinationAccount: identity.Account{
					CreatedAt:                  result.AccountCreatedAt,
					SubscriptionPlanID:         database.StringPointerFromNullString(result.AccountSubscriptionPlanID),
					LastUpdatedAt:              database.TimePointerFromNullTime(result.AccountLastUpdatedAt),
					ArchivedAt:                 database.TimePointerFromNullTime(result.AccountArchivedAt),
					ContactPhone:               result.AccountContactPhone,
					BillingStatus:              result.AccountBillingStatus,
					AddressLine1:               result.AccountAddressLine1,
					AddressLine2:               result.AccountAddressLine2,
					City:                       result.AccountCity,
					State:                      result.AccountState,
					ZipCode:                    result.AccountZipCode,
					Country:                    result.AccountCountry,
					Latitude:                   database.Float64PointerFromNullString(result.AccountLatitude),
					Longitude:                  database.Float64PointerFromNullString(result.AccountLongitude),
					PaymentProcessorCustomerID: result.AccountPaymentProcessorCustomerID,
					BelongsToUser:              result.AccountBelongsToUser,
					ID:                         result.AccountID,
					Name:                       result.AccountName,
					Members:                    nil,
				},
				FromUser: identity.User{
					CreatedAt:                  result.UserCreatedAt,
					PasswordLastChangedAt:      database.TimePointerFromNullTime(result.UserPasswordLastChangedAt),
					LastUpdatedAt:              database.TimePointerFromNullTime(result.UserLastUpdatedAt),
					LastAcceptedTermsOfService: database.TimePointerFromNullTime(result.UserLastAcceptedTermsOfService),
					LastAcceptedPrivacyPolicy:  database.TimePointerFromNullTime(result.UserLastAcceptedPrivacyPolicy),
					TwoFactorSecretVerifiedAt:  database.TimePointerFromNullTime(result.UserTwoFactorSecretVerifiedAt),
					Avatar:                     r.avatarFor(ctx, result.UserAvatarID),
					Birthday:                   database.TimePointerFromNullTime(result.UserBirthday),
					ArchivedAt:                 database.TimePointerFromNullTime(result.UserArchivedAt),
					AccountStatusExplanation:   result.UserUserAccountStatusExplanation,
					TwoFactorSecret:            result.UserTwoFactorSecret,
					HashedPassword:             result.UserHashedPassword,
					ID:                         result.UserID,
					AccountStatus:              result.UserUserAccountStatus,
					Username:                   result.UserUsername,
					FirstName:                  result.UserFirstName,
					LastName:                   result.UserLastName,
					EmailAddress:               result.UserEmailAddress,
					EmailAddressVerifiedAt:     database.TimePointerFromNullTime(result.UserEmailAddressVerifiedAt),
					RequiresPasswordChange:     result.UserRequiresPasswordChange,
				},
			}
		},
		func(result *generated.GetPendingInvitesForUserRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(t *identity.AccountInvitation) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

func (r *repository) setInvitationStatus(ctx context.Context, querier database.Tx, accountInvitationID, note, status string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithValue("new_status", status)

	if accountInvitationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, identitykeys.AccountInvitationIDKey, accountInvitationID)

	if err := r.generatedQuerier.SetAccountInvitationStatus(ctx, querier, &generated.SetAccountInvitationStatusParams{
		Status:     generated.InvitationState(status),
		StatusNote: note,
		ID:         accountInvitationID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "changing account invitation status")
	}

	logger.Debug("account invitation updated")

	return nil
}

// CancelAccountInvitation cancels an account invitation by its ID with a note.
func (r *repository) CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return r.withEvent(ctx, r.logger, identity.AccountInvitationCanceledServiceEventType, accountID, map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}, func(tx database.Tx) error {
		return r.setInvitationStatus(ctx, tx, accountInvitationID, note, string(identity.CancelledAccountInvitationStatus))
	})
}

// AcceptAccountInvitation accepts an account invitation by its ID with a note.
func (r *repository) AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID, token, note string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountInvitationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountInvitationIDKey, accountInvitationID)
	tracing.AttachToSpan(span, identitykeys.AccountInvitationIDKey, accountInvitationID)

	if token == "" {
		return platformerrors.ErrNilInputParameter
	}

	if err := r.WithTransaction(ctx, func(tx database.Tx) error {
		invitation, err := r.GetAccountInvitationByTokenAndID(ctx, token, accountInvitationID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching account invitation")
		}

		if err = r.setInvitationStatus(ctx, tx, accountInvitationID, note, string(identity.AcceptedAccountInvitationStatus)); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "accepting account invitation")
		}

		addUserInput := &identity.AccountUserMembershipDatabaseCreationInput{
			ID:        identifiers.New(),
			Reason:    fmt.Sprintf("accepted account invitation %s", accountInvitationID),
			AccountID: invitation.DestinationAccount.ID,
		}
		if invitation.ToUser != nil {
			addUserInput.UserID = *invitation.ToUser
			if err = r.addUserToAccount(ctx, tx, addUserInput); err != nil {
				return observability.PrepareAndLogError(err, logger, span, "adding user to account")
			}
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.AccountInvitationAcceptedServiceEventType, accountID, map[string]any{
			identitykeys.AccountInvitationIDKey:  accountInvitationID,
			identitykeys.DestinationAccountIDKey: invitation.DestinationAccount.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// RejectAccountInvitation rejects an account invitation by its ID with a note.
func (r *repository) RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return r.withEvent(ctx, r.logger, identity.AccountInvitationRejectedServiceEventType, accountID, map[string]any{
		identitykeys.AccountInvitationIDKey: accountInvitationID,
	}, func(tx database.Tx) error {
		return r.setInvitationStatus(ctx, tx, accountInvitationID, note, string(identity.RejectedAccountInvitationStatus))
	})
}

func (r *repository) attachInvitationsToUser(ctx context.Context, querier database.Tx, userEmail, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger

	if userEmail == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserEmailAddressKey, userEmail)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, userEmail)

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	rowCount, err := r.generatedQuerier.AttachAccountInvitationsToUserID(ctx, querier, &generated.AttachAccountInvitationsToUserIDParams{
		ToEmail: userEmail,
		ToUser:  database.NullStringFromString(userID),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareAndLogError(err, logger, span, "attaching invitations to user")
	}

	logger.WithValue("rows_affected", rowCount).Info("invitations associated with user")

	return nil
}

func (r *repository) acceptInvitationForUser(ctx context.Context, querier database.Tx, input *identity.UserDatabaseCreationInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithValue(identitykeys.UsernameKey, input.Username).WithValue(identitykeys.UserEmailAddressKey, input.EmailAddress)

	invitation, tokenCheckErr := r.GetAccountInvitationByToken(ctx, input.InvitationToken)
	if tokenCheckErr != nil {
		return observability.PrepareError(tokenCheckErr, span, "fetching account invitation")
	}

	if err := r.generatedQuerier.CreateAccountUserMembershipForNewUser(ctx, querier, &generated.CreateAccountUserMembershipForNewUserParams{
		ID:               identifiers.New(),
		BelongsToUser:    input.ID,
		BelongsToAccount: input.DestinationAccountID,
		DefaultAccount:   true,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "writing destination account membership")
	}

	// Assign account_member role for the invited account.
	if err := r.generatedQuerier.AssignRoleToUser(ctx, querier, &generated.AssignRoleToUserParams{
		ID:        identifiers.New(),
		UserID:    input.ID,
		RoleID:    authorization.AccountMemberRoleID,
		AccountID: sql.NullString{String: input.DestinationAccountID, Valid: true},
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "assigning account role for invitation")
	}

	logger.Debug("created membership via invitation")

	if err := r.setInvitationStatus(ctx, querier, invitation.ID, "", string(identity.AcceptedAccountInvitationStatus)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "accepting account invitation")
	}

	logger.Debug("marked invitation as accepted")

	return nil
}
