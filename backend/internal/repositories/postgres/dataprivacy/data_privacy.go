package dataprivacy

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
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
	accounts, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
		return r.identityRepo.GetAccounts(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching accounts")
	}
	for _, account := range accounts {
		collection.Identity.Accounts = append(collection.Identity.Accounts, *account)
	}

	// Fetch account invitations (both sent and received)
	sentInvites, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
		return r.identityRepo.GetPendingAccountInvitationsFromUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching sent invitations")
	}
	receivedInvites, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
		return r.identityRepo.GetPendingAccountInvitationsForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching received invitations")
	}
	// Combine and deduplicate invitations
	inviteMap := make(map[string]*identity.AccountInvitation)
	for _, invite := range sentInvites {
		inviteMap[invite.ID] = invite
	}
	for _, invite := range receivedInvites {
		inviteMap[invite.ID] = invite
	}
	for _, invite := range inviteMap {
		collection.Identity.AccountInvitations = append(collection.Identity.AccountInvitations, *invite)
	}

	// Fetch audit log entries
	auditLogs, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
		return r.auditLogRepo.GetAuditLogEntriesForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching audit log entries")
	}
	for _, entry := range auditLogs {
		collection.AuditLogEntries = append(collection.AuditLogEntries, *entry)
	}

	// Fetch comments authored by the user
	userComments, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
		return r.commentsRepo.GetCommentsForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching comments")
	}
	for _, comment := range userComments {
		collection.Comments = append(collection.Comments, *comment)
	}

	// Collect domain-specific user-scoped data via registered collectors
	for _, collector := range r.dataCollectors {
		if err = collector.CollectUserData(ctx, collection, userID); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "collecting domain user data")
		}
	}

	// Fetch notifications
	notifs, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
		return r.notificationsRepo.GetUserNotifications(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching notifications")
	}
	for _, notif := range notifs {
		collection.Notifications.Data = append(collection.Notifications.Data, *notif)
	}

	// Fetch user settings
	userSettings, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
		return r.settingsRepo.GetServiceSettingConfigurationsForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user settings")
	}
	for _, setting := range userSettings {
		collection.Settings.UserSettings = append(collection.Settings.UserSettings, *setting)
	}

	// Fetch uploaded media
	media, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
		return r.uploadedMediaRepo.GetUploadedMediaForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching uploaded media")
	}
	for _, m := range media {
		collection.UploadedMedia = append(collection.UploadedMedia, *m)
	}

	// Fetch waitlist signups
	waitlistSignups, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
		return r.waitlistsRepo.GetWaitlistSignupsForUser(ctx, userID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signups")
	}
	for _, signup := range waitlistSignups {
		collection.WaitlistSignups = append(collection.WaitlistSignups, *signup)
	}

	// Fetch account-scoped data for each account
	for _, account := range accounts {
		accountLogger := logger.WithValue("account_id", account.ID)

		// Collect domain-specific account-scoped data via registered collectors
		for _, collector := range r.dataCollectors {
			if collectorErr := collector.CollectAccountData(ctx, collection, account.ID); collectorErr != nil {
				return nil, observability.PrepareAndLogError(collectorErr, accountLogger, span, "collecting domain account data")
			}
		}

		// Webhooks
		hooks, webhookErr := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
			return r.webhooksRepo.GetWebhooks(ctx, account.ID, filter)
		})
		if webhookErr != nil {
			return nil, observability.PrepareAndLogError(webhookErr, accountLogger, span, "fetching webhooks")
		}
		if len(hooks) > 0 {
			var webhookList []webhooks.Webhook
			for _, hook := range hooks {
				webhookList = append(webhookList, *hook)
			}
			collection.Webhooks.Data[account.ID] = webhookList
		}

		// Account settings
		accountSettings, settingsErr := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
			return r.settingsRepo.GetServiceSettingConfigurationsForAccount(ctx, account.ID, filter)
		})
		if settingsErr != nil {
			return nil, observability.PrepareAndLogError(settingsErr, accountLogger, span, "fetching account settings")
		}
		for _, setting := range accountSettings {
			collection.Settings.AccountSettings = append(collection.Settings.AccountSettings, *setting)
		}

		// Issue reports for account
		reports, reportsErr := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
			return r.issueReportsRepo.GetIssueReportsForAccount(ctx, account.ID, filter)
		})
		if reportsErr != nil {
			return nil, observability.PrepareAndLogError(reportsErr, accountLogger, span, "fetching issue reports")
		}
		for _, report := range reports {
			collection.IssueReports = append(collection.IssueReports, *report)
		}

		// Payments: subscriptions, purchases, and payment transactions
		subscriptions, subsErr := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error) {
			return r.paymentsRepo.GetSubscriptionsForAccount(ctx, account.ID, filter)
		})
		if subsErr != nil {
			return nil, observability.PrepareAndLogError(subsErr, accountLogger, span, "fetching subscriptions")
		}
		for _, subscription := range subscriptions {
			collection.Payments.Subscriptions = append(collection.Payments.Subscriptions, *subscription)
		}

		purchases, purchasesErr := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Purchase], error) {
			return r.paymentsRepo.GetPurchasesForAccount(ctx, account.ID, filter)
		})
		if purchasesErr != nil {
			return nil, observability.PrepareAndLogError(purchasesErr, accountLogger, span, "fetching purchases")
		}
		for _, purchase := range purchases {
			collection.Payments.Purchases = append(collection.Payments.Purchases, *purchase)
		}

		transactions, transactionsErr := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.PaymentTransaction], error) {
			return r.paymentsRepo.GetPaymentTransactionsForAccount(ctx, account.ID, filter)
		})
		if transactionsErr != nil {
			return nil, observability.PrepareAndLogError(transactionsErr, accountLogger, span, "fetching payment transactions")
		}
		for _, transaction := range transactions {
			collection.Payments.PaymentTransactions = append(collection.Payments.PaymentTransactions, *transaction)
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
