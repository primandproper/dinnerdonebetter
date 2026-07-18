package dataprivacy

import (
	"context"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v5/filtering"
)

// UserDataDisclosureStatus represents the status of a user data disclosure request.
type UserDataDisclosureStatus string

const (
	// UserDataDisclosureStatusPending indicates the disclosure request is pending.
	UserDataDisclosureStatusPending UserDataDisclosureStatus = "pending"
	// UserDataDisclosureStatusProcessing indicates the disclosure request is being processed.
	UserDataDisclosureStatusProcessing UserDataDisclosureStatus = "processing"
	// UserDataDisclosureStatusCompleted indicates the disclosure request is complete.
	UserDataDisclosureStatusCompleted UserDataDisclosureStatus = "completed"
	// UserDataDisclosureStatusFailed indicates the disclosure request failed.
	UserDataDisclosureStatusFailed UserDataDisclosureStatus = "failed"
	// UserDataDisclosureStatusExpired indicates the disclosure request has expired.
	UserDataDisclosureStatusExpired UserDataDisclosureStatus = "expired"
)

type (
	// DataDeletionResponse is returned when a user requests their data be deleted.
	DataDeletionResponse struct {
		Successful bool `json:"Successful"`
	}

	// UserDataAggregationRequest represents a message queue event meant to aggregate data for a user.
	UserDataAggregationRequest struct {
		_ struct{} `json:"-"`

		RequestID string `json:"id"`
		ReportID  string `json:"reportID"`
		UserID    string `json:"userID"`
		TestID    string `json:"testID,omitempty"`
	}

	// UserDataCollectionResponse represents the response to a UserDataAggregationRequest.
	UserDataCollectionResponse struct {
		_ struct{} `json:"-"`

		ReportID string `json:"reportID"`
	}

	// UserDataCollection contains all user-associated data for GDPR/CCPA disclosure.
	// This type intentionally aggregates all domain types as a cross-cutting concern.
	// When adding or removing a domain from this template, update the fields here accordingly.
	UserDataCollection struct {
		Identity        identity.UserDataCollection      `json:"identity"`
		MealPlanning    mealplanning.UserDataCollection  `json:"meal_planning"`
		Webhooks        webhooks.UserDataCollection      `json:"webhooks"`
		Settings        settings.UserDataCollection      `json:"settings"`
		Notifications   notifications.UserDataCollection `json:"notifications"`
		AuditLogEntries []audit.AuditLogEntry            `json:"audit_log_entries,omitempty"`
		IssueReports    []issuereports.IssueReport       `json:"issue_reports,omitempty"`
		UploadedMedia   []uploadedmedia.UploadedMedia    `json:"uploaded_media,omitempty"`
		WaitlistSignups []waitlists.WaitlistSignup       `json:"waitlist_signups,omitempty"`
	}

	// UserDataDisclosure represents a user data disclosure request for GDPR/CCPA compliance.
	UserDataDisclosure struct {
		ExpiresAt     time.Time                `json:"expiresAt"`
		CreatedAt     time.Time                `json:"createdAt"`
		LastUpdatedAt *time.Time               `json:"lastUpdatedAt,omitempty"`
		CompletedAt   *time.Time               `json:"completedAt,omitempty"`
		ArchivedAt    *time.Time               `json:"archivedAt,omitempty"`
		ID            string                   `json:"id"`
		BelongsToUser string                   `json:"belongsToUser"`
		Status        UserDataDisclosureStatus `json:"status"`
		ReportID      string                   `json:"reportId,omitempty"`
	}

	// UserDataDisclosureCreationInput represents the input for creating a user data disclosure.
	UserDataDisclosureCreationInput struct {
		ExpiresAt     time.Time `json:"expiresAt"`
		ID            string    `json:"-"`
		BelongsToUser string    `json:"-"`
	}

	// UserDataCollector collects domain-specific user data for GDPR/CCPA disclosure.
	// Each domain implements this interface to contribute its data to the aggregate collection.
	// When adding or removing a domain from this template, implement this interface and register the collector.
	UserDataCollector interface {
		CollectUserData(ctx context.Context, collection *UserDataCollection, userID string) error
		CollectAccountData(ctx context.Context, collection *UserDataCollection, accountID string) error
	}

	// DataPrivacyDataManager contains data privacy management functions.
	DataPrivacyDataManager interface {
		FetchUserDataCollection(ctx context.Context, userID string) (*UserDataCollection, error)
		DeleteUser(ctx context.Context, userID string) error
	}

	// UserDataDisclosureDataManager contains user data disclosure management functions.
	UserDataDisclosureDataManager interface {
		CreateUserDataDisclosure(ctx context.Context, input *UserDataDisclosureCreationInput) (*UserDataDisclosure, error)
		GetUserDataDisclosure(ctx context.Context, disclosureID string) (*UserDataDisclosure, error)
		GetUserDataDisclosuresForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[UserDataDisclosure], error)
		MarkUserDataDisclosureCompleted(ctx context.Context, disclosureID, reportID string) error
		MarkUserDataDisclosureFailed(ctx context.Context, disclosureID string) error
		ArchiveUserDataDisclosure(ctx context.Context, disclosureID string) error
	}
)
