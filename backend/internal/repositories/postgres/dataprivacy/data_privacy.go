package dataprivacy

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/observability"
)

// FetchUserDataCollection retrieves all user-associated data for GDPR/CCPA disclosure.
func (r *repository) FetchUserDataCollection(ctx context.Context, userID string) (*dataprivacy.UserDataCollection, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue("user_id", userID)
	logger.Info("fetching user data collection")

	collection := &dataprivacy.UserDataCollection{
		Identity:      identity.UserDataCollection{},
		Webhooks:      webhooks.UserDataCollection{Data: make(map[string][]webhooks.Webhook)},
		Settings:      settings.UserDataCollection{},
		Notifications: notifications.UserDataCollection{},
	}

	// Fetch user profile
	user, err := r.identityRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user")
	}
	collection.Identity.User = *user

	// Fetch user accounts
	accounts, err := r.identityRepo.GetAccounts(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching accounts")
	}
	for _, account := range accounts.Data {
		collection.Identity.Accounts = append(collection.Identity.Accounts, *account)
	}

	// Fetch account invitations (both sent and received)
	sentInvites, err := r.identityRepo.GetPendingAccountInvitationsFromUser(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching sent invitations")
	}
	receivedInvites, err := r.identityRepo.GetPendingAccountInvitationsForUser(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching received invitations")
	}
	// Combine and deduplicate invitations
	inviteMap := make(map[string]*identity.AccountInvitation)
	for _, invite := range sentInvites.Data {
		inviteMap[invite.ID] = invite
	}
	for _, invite := range receivedInvites.Data {
		inviteMap[invite.ID] = invite
	}
	for _, invite := range inviteMap {
		collection.Identity.AccountInvitations = append(collection.Identity.AccountInvitations, *invite)
	}

	// Fetch audit log entries
	auditLogs, err := r.auditLogRepo.GetAuditLogEntriesForUser(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries")
	}
	for _, entry := range auditLogs.Data {
		collection.AuditLogEntries = append(collection.AuditLogEntries, *entry)
	}

	// Collect domain-specific user-scoped data via registered collectors
	for _, collector := range r.dataCollectors {
		if err = collector.CollectUserData(ctx, collection, userID); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "collecting domain user data")
		}
	}

	// Fetch notifications
	notifs, err := r.notificationsRepo.GetUserNotifications(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching notifications")
	}
	for _, notif := range notifs.Data {
		collection.Notifications.Data = append(collection.Notifications.Data, *notif)
	}

	// Fetch user settings
	userSettings, err := r.settingsRepo.GetServiceSettingConfigurationsForUser(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user settings")
	}
	for _, setting := range userSettings.Data {
		collection.Settings.UserSettings = append(collection.Settings.UserSettings, *setting)
	}

	// Fetch uploaded media
	media, err := r.uploadedMediaRepo.GetUploadedMediaForUser(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media")
	}
	for _, m := range media.Data {
		collection.UploadedMedia = append(collection.UploadedMedia, *m)
	}

	// Fetch waitlist signups
	waitlistSignups, err := r.waitlistsRepo.GetWaitlistSignupsForUser(ctx, userID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signups")
	}
	for _, signup := range waitlistSignups.Data {
		collection.WaitlistSignups = append(collection.WaitlistSignups, *signup)
	}

	// Fetch account-scoped data for each account
	for _, account := range accounts.Data {
		accountLogger := logger.WithValue("account_id", account.ID)

		// Collect domain-specific account-scoped data via registered collectors
		for _, collector := range r.dataCollectors {
			if collectorErr := collector.CollectAccountData(ctx, collection, account.ID); collectorErr != nil {
				return nil, observability.PrepareAndLogError(collectorErr, accountLogger, span, "collecting domain account data")
			}
		}

		// Webhooks
		hooks, webhookErr := r.webhooksRepo.GetWebhooks(ctx, account.ID, nil)
		if webhookErr != nil {
			return nil, observability.PrepareAndLogError(webhookErr, accountLogger, span, "fetching webhooks")
		}
		if len(hooks.Data) > 0 {
			var webhookList []webhooks.Webhook
			for _, hook := range hooks.Data {
				webhookList = append(webhookList, *hook)
			}
			collection.Webhooks.Data[account.ID] = webhookList
		}

		// Account settings
		accountSettings, settingsErr := r.settingsRepo.GetServiceSettingConfigurationsForAccount(ctx, account.ID, nil)
		if settingsErr != nil {
			return nil, observability.PrepareAndLogError(settingsErr, accountLogger, span, "fetching account settings")
		}
		for _, setting := range accountSettings.Data {
			collection.Settings.AccountSettings = append(collection.Settings.AccountSettings, *setting)
		}

		// Issue reports for account
		reports, reportsErr := r.issueReportsRepo.GetIssueReportsForAccount(ctx, account.ID, nil)
		if reportsErr != nil {
			return nil, observability.PrepareAndLogError(reportsErr, accountLogger, span, "fetching issue reports")
		}
		for _, report := range reports.Data {
			collection.IssueReports = append(collection.IssueReports, *report)
		}
	}

	logger.Info("user data collection complete")

	return collection, nil
}

// DeleteUser deletes a user and all associated data via ON DELETE CASCADE.
func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := r.logger.WithValue("user_id", userID)
	logger.Info("deleting user and all associated data")

	// The database schema uses ON DELETE CASCADE on all belongs_to_user foreign keys,
	// so deleting the user record will automatically delete all associated data.
	if err := r.identityRepo.DeleteUser(ctx, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "deleting user")
	}

	logger.Info("user deleted successfully")

	return nil
}
