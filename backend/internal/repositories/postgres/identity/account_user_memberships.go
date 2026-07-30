package identity

import (
	"context"
	"database/sql"
	"errors"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	"github.com/primandproper/platform-go/v8/database"
	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	resourceTypeAccountUserMemberships = "account_user_memberships"
)

var (
	_ identity.AccountUserMembershipDataManager = (*repository)(nil)
)

// BuildSessionContextDataForUser queries the database for the memberships of a user and constructs a ContextData struct from the results.
// When activeAccountID is non-empty, it is used as the active account if the user is a member; otherwise the default account is used.
func (r *repository) BuildSessionContextDataForUser(ctx context.Context, userID, activeAccountID string) (*sessions.ContextData, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	// Get account memberships for default_account logic.
	memberships, err := r.generatedQuerier.GetAccountUserMembershipsForUser(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user's memberships from database")
	}

	defaultAccountID := ""
	for _, result := range memberships {
		if result.DefaultAccount {
			defaultAccountID = result.BelongsToAccount
		}
	}

	effectiveAccountID := defaultAccountID
	if activeAccountID != "" {
		isMember, memberErr := r.UserIsMemberOfAccount(ctx, userID, activeAccountID)
		if memberErr != nil {
			return nil, observability.PrepareAndLogError(memberErr, logger, span, "checking account membership")
		}
		if !isMember {
			return nil, observability.PrepareAndLogError(errors.New("user is not a member of the specified account"), logger, span, "user is not a member of the specified account")
		}
		effectiveAccountID = activeAccountID
	}

	// Load service-level permissions and role names from DB.
	servicePermNames, err := r.generatedQuerier.GetServicePermissionsForUser(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching service permissions for user")
	}

	servicePerms := make([]authorization.Permission, len(servicePermNames))
	for i, name := range servicePermNames {
		servicePerms[i] = authorization.Permission(name)
	}

	serviceRoleNames, err := r.generatedQuerier.GetServiceRoleNamesForUser(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching service role names for user")
	}

	// Load account-level permissions from DB (all accounts).
	accountPermRows, err := r.generatedQuerier.GetAccountPermissionsForUser(ctx, r.readDB, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account permissions for user")
	}

	accountPermsMap := map[string][]authorization.Permission{}
	for _, row := range accountPermRows {
		if row.AccountID.Valid {
			accountPermsMap[row.AccountID.String] = append(accountPermsMap[row.AccountID.String], authorization.Permission(row.PermissionName))
		}
	}

	accountRolesMap := map[string]authorization.AccountRolePermissionsChecker{}
	for accountID, perms := range accountPermsMap {
		accountRolesMap[accountID] = authorization.NewAccountRolePermissionChecker(perms)
	}

	user, err := r.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user from database")
	}

	sessionCtxData := &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   user.ID,
			Username:                 user.Username,
			EmailAddress:             user.EmailAddress,
			AccountStatus:            user.AccountStatus,
			AccountStatusExplanation: user.AccountStatusExplanation,
			ServicePermissions:       authorization.NewServiceRolePermissionChecker(serviceRoleNames, servicePerms),
		},
		AccountPermissions: accountRolesMap,
		ActiveAccountID:    effectiveAccountID,
	}

	logger.Debug("fetched session context data for user")

	return sessionCtxData, nil
}

// GetDefaultAccountIDForUser retrieves the default account ID for a user.
func (r *repository) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if userID == "" {
		return "", platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	id, err := r.generatedQuerier.GetDefaultAccountIDForUser(ctx, r.readDB, userID)
	if err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "fetching default account ID for user")
	}

	return id, nil
}

// markAccountAsUserDefault marks a given account as the user's default.
func (r *repository) markAccountAsUserDefault(ctx context.Context, querier database.SQLQueryExecutor, userID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValues(map[string]any{
		identitykeys.UserIDKey:    userID,
		identitykeys.AccountIDKey: accountID,
	})

	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.MarkAccountUserMembershipAsUserDefault(ctx, querier, &generated.MarkAccountUserMembershipAsUserDefaultParams{
			BelongsToUser:    userID,
			BelongsToAccount: accountID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "assigning user default account")
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &accountID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccountUserMemberships,
			EventType:        audit.AuditLogEventTypeUpdated,
			BelongsToUser:    userID,
			Changes: map[string]*audit.ChangeLog{
				"default_account": {
					OldValue: "false",
					NewValue: "true",
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Debug("account marked as default")

	return nil
}

// MarkAccountAsUserDefault does a thing.
func (r *repository) MarkAccountAsUserDefault(ctx context.Context, userID, accountID string) error {
	return r.markAccountAsUserDefault(ctx, r.writeDB, userID, accountID)
}

// UserIsMemberOfAccount does a thing.
func (r *repository) UserIsMemberOfAccount(ctx context.Context, userID, accountID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := r.generatedQuerier.UserIsAccountMember(ctx, r.readDB, &generated.UserIsAccountMemberParams{
		BelongsToAccount: accountID,
		BelongsToUser:    userID,
	})
	if err != nil {
		return false, observability.PrepareError(err, span, "performing user membership check query")
	}

	return result, nil
}

// ModifyUserPermissions does a thing.
func (r *repository) ModifyUserPermissions(ctx context.Context, accountID, userID string, input *identity.ModifyUserPermissionsInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
		identitykeys.UserIDKey:    userID,
		"new_roles":               input.NewRole,
	})

	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	// Look up the new role by name.
	newRole, roleErr := r.generatedQuerier.GetUserRoleByName(ctx, r.readDB, input.NewRole)
	if roleErr != nil {
		return observability.PrepareAndLogError(roleErr, logger, span, "fetching role by name")
	}

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		// Update the user's account-level role assignment.
		if err = r.generatedQuerier.UpdateAccountRoleAssignment(ctx, tx, &generated.UpdateAccountRoleAssignmentParams{
			NewRoleID: newRole.ID,
			UserID:    userID,
			AccountID: sql.NullString{String: accountID, Valid: true},
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating account role assignment")
		}

		if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &accountID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccountUserMemberships,
			EventType:        audit.AuditLogEventTypeUpdated,
			BelongsToUser:    userID,
			Changes: map[string]*audit.ChangeLog{
				"role": {
					NewValue: input.NewRole,
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user permissions modified")

	return nil
}

// TransferAccountOwnership does a thing.
func (r *repository) TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
		"current_owner":           input.CurrentOwner,
		"new_owner":               input.NewOwner,
	})

	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.NewOwner)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		// create the membership.
		if err := r.generatedQuerier.TransferAccountOwnership(ctx, tx, &generated.TransferAccountOwnershipParams{
			NewOwner:  input.NewOwner,
			OldOwner:  input.CurrentOwner,
			AccountID: accountID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "transferring account to new owner")
		}

		if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			BelongsToAccount: &accountID,
			ID:               identifiers.New(),
			ResourceType:     resourceTypeAccountUserMemberships,
			EventType:        audit.AuditLogEventTypeUpdated,
			BelongsToUser:    input.NewOwner,
			Changes: map[string]*audit.ChangeLog{
				"belongs_to_user": {
					OldValue: input.CurrentOwner,
					NewValue: input.NewOwner,
				},
			},
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		isMember, err := r.UserIsMemberOfAccount(ctx, input.NewOwner, accountID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "checking if user is member of account")
		}

		if !isMember {
			if err = r.addUserToAccount(ctx, tx, &identity.AccountUserMembershipDatabaseCreationInput{
				ID:        identifiers.New(),
				Reason:    "transferred ownership",
				UserID:    input.NewOwner,
				AccountID: accountID,
			}); err != nil {
				return observability.PrepareAndLogError(err, logger, span, "adding user to account")
			}
			// audit log created above
		}

		if err = r.removeUserFromAccount(ctx, tx, input.CurrentOwner, accountID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "removing user from account")
		}

		// audit log created above

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, identity.AccountOwnershipTransferredServiceEventType, accountID, map[string]any{
			identitykeys.AccountIDKey: accountID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("account transferred to new owner")

	return nil
}

// addUserToAccount does a thing.
func (r *repository) addUserToAccount(ctx context.Context, querier database.SQLQueryExecutor, input *identity.AccountUserMembershipDatabaseCreationInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := r.logger.WithValues(map[string]any{
		identitykeys.UserIDKey:    input.UserID,
		identitykeys.AccountIDKey: input.AccountID,
	})

	tracing.AttachToSpan(span, identitykeys.UserIDKey, input.UserID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.AccountID)

	// create the membership.
	if err := r.generatedQuerier.AddUserToAccount(ctx, querier, &generated.AddUserToAccountParams{
		ID:               input.ID,
		BelongsToUser:    input.UserID,
		BelongsToAccount: input.AccountID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "performing user account membership creation query")
	}

	// Assign account_member role for this account.
	if err := r.generatedQuerier.AssignRoleToUser(ctx, querier, &generated.AssignRoleToUserParams{
		ID:        identifiers.New(),
		UserID:    input.UserID,
		RoleID:    authorization.AccountMemberRoleID,
		AccountID: sql.NullString{String: input.AccountID, Valid: true},
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "assigning account role to user")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, querier, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &input.AccountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeCreated,
		BelongsToUser:    input.UserID,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	logger.Info("user added to account")

	return nil
}

// removeUserFromAccount removes a user's membership to an account.
func (r *repository) removeUserFromAccount(ctx context.Context, querier database.SQLQueryExecutor, userID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		identitykeys.UserIDKey:    userID,
		identitykeys.AccountIDKey: accountID,
	})

	// archive role assignments for this account.
	if err := r.generatedQuerier.ArchiveRoleAssignmentsForUserAndAccount(ctx, querier, &generated.ArchiveRoleAssignmentsForUserAndAccountParams{
		UserID:    userID,
		AccountID: sql.NullString{String: accountID, Valid: true},
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving role assignments")
	}

	// remove the membership.
	if err := r.generatedQuerier.RemoveUserFromAccount(ctx, querier, &generated.RemoveUserFromAccountParams{
		BelongsToAccount: accountID,
		BelongsToUser:    userID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing user from account")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, querier, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeAccountUserMemberships,
		EventType:        audit.AuditLogEventTypeArchived,
		BelongsToUser:    userID,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	remainingAccounts, err := r.getAccountsForUser(ctx, querier, userID, nil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, span, "fetching remaining accounts")
	}

	if remainingAccounts == nil || len(remainingAccounts.Data) < 1 {
		logger.Debug("user has no remaining accounts, creating a new one")
		if _, err = r.createAccountForUser(ctx, querier, false, "", userID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating account for new user")
		}
		return nil
	}

	account := remainingAccounts.Data[0]
	logger.WithValue("new_default_account", account.ID).Info("setting new default account")
	if err = r.markAccountAsUserDefault(ctx, querier, userID, account.ID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking account as default")
	}

	logger.Info("marked account as default")

	return nil
}

// RemoveUserFromAccount removes a user's membership to an account.
func (r *repository) RemoveUserFromAccount(ctx context.Context, userID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		identitykeys.UserIDKey:    userID,
		identitykeys.AccountIDKey: accountID,
	})

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.removeUserFromAccount(ctx, tx, userID, accountID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "removing user from account")
		}

		// audit log entry created above

		return nil
	}); err != nil {
		return err
	}

	return nil
}
