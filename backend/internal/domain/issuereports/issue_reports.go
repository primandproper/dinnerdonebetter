package issuereports

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v12/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// IssueReportCreatedServiceEventType indicates an issue report was created.
	IssueReportCreatedServiceEventType = "issue_report_created"
	// IssueReportUpdatedServiceEventType indicates an issue report was updated.
	IssueReportUpdatedServiceEventType = "issue_report_updated"
	// IssueReportArchivedServiceEventType indicates an issue report was archived.
	IssueReportArchivedServiceEventType = "issue_report_archived"
)

func init() {
	gob.Register(new(IssueReport))
	gob.Register(new(IssueReportCreationRequestInput))
	gob.Register(new(IssueReportDatabaseCreationInput))
	gob.Register(new(IssueReportUpdateRequestInput))
}

type (
	// IssueReport represents a user-submitted issue report.
	IssueReport struct {
		_                struct{}   `json:"-"`
		CreatedAt        time.Time  `json:"createdAt"`
		LastUpdatedAt    *time.Time `json:"lastUpdatedAt"`
		ArchivedAt       *time.Time `json:"archivedAt"`
		ID               string     `json:"id"`
		IssueType        string     `json:"issueType"`
		Details          string     `json:"details"`
		RelevantTable    string     `json:"relevantTable,omitempty"`
		RelevantRecordID string     `json:"relevantRecordID,omitempty"`
		CreatedByUser    string     `json:"createdByUser"`
		BelongsToAccount string     `json:"belongsToAccount"`
	}

	// IssueReportCreationRequestInput represents input for creating an issue report.
	IssueReportCreationRequestInput struct {
		_                struct{} `json:"-"`
		IssueType        string   `json:"issueType"`
		Details          string   `json:"details"`
		RelevantTable    string   `json:"relevantTable,omitempty"`
		RelevantRecordID string   `json:"relevantRecordID,omitempty"`
	}

	// IssueReportDatabaseCreationInput is used for creating an issue report in persistence.
	IssueReportDatabaseCreationInput struct {
		_                struct{} `json:"-"`
		ID               string   `json:"-"`
		IssueType        string   `json:"-"`
		Details          string   `json:"-"`
		RelevantTable    string   `json:"-"`
		RelevantRecordID string   `json:"-"`
		CreatedByUser    string   `json:"-"`
		BelongsToAccount string   `json:"-"`
	}

	// IssueReportUpdateRequestInput represents input for updating an issue report.
	IssueReportUpdateRequestInput struct {
		_                struct{} `json:"-"`
		IssueType        *string  `json:"issueType,omitempty"`
		Details          *string  `json:"details,omitempty"`
		RelevantTable    *string  `json:"relevantTable,omitempty"`
		RelevantRecordID *string  `json:"relevantRecordID,omitempty"`
	}

	// IssueReportDataManager describes a structure capable of storing issue reports.
	IssueReportDataManager interface {
		GetIssueReport(ctx context.Context, issueReportID string) (*IssueReport, error)
		GetIssueReports(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[IssueReport], error)
		GetIssueReportsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[IssueReport], error)
		GetIssueReportsForTable(ctx context.Context, tableName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[IssueReport], error)
		GetIssueReportsForRecord(ctx context.Context, tableName, recordID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[IssueReport], error)
		CreateIssueReport(ctx context.Context, input *IssueReportDatabaseCreationInput) (*IssueReport, error)
		UpdateIssueReport(ctx context.Context, issueReport *IssueReport) error
		ArchiveIssueReport(ctx context.Context, issueReportID string) error
	}
)

// Update merges an IssueReportUpdateRequestInput into an IssueReport.
func (i *IssueReport) Update(input *IssueReportUpdateRequestInput) {
	if input.IssueType != nil && *input.IssueType != i.IssueType {
		i.IssueType = *input.IssueType
	}
	if input.Details != nil && *input.Details != i.Details {
		i.Details = *input.Details
	}
	if input.RelevantTable != nil && *input.RelevantTable != i.RelevantTable {
		i.RelevantTable = *input.RelevantTable
	}
	if input.RelevantRecordID != nil && *input.RelevantRecordID != i.RelevantRecordID {
		i.RelevantRecordID = *input.RelevantRecordID
	}
}

var _ validation.ValidatableWithContext = (*IssueReportCreationRequestInput)(nil)

// ValidateWithContext validates an IssueReportCreationRequestInput.
func (i *IssueReportCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(&i.IssueType, validation.Required),
		validation.Field(&i.Details, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*IssueReportDatabaseCreationInput)(nil)

// ValidateWithContext validates an IssueReportDatabaseCreationInput.
func (i *IssueReportDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(&i.ID, validation.Required),
		validation.Field(&i.IssueType, validation.Required),
		validation.Field(&i.Details, validation.Required),
		validation.Field(&i.CreatedByUser, validation.Required),
		validation.Field(&i.BelongsToAccount, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*IssueReportUpdateRequestInput)(nil)

// ValidateWithContext validates an IssueReportUpdateRequestInput.
func (i *IssueReportUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(&i.IssueType, validation.When(i.IssueType != nil, validation.Required)),
		validation.Field(&i.Details, validation.When(i.Details != nil, validation.Required)),
	)
}
