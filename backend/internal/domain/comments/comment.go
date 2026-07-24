package comments

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v5/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// CommentCreatedServiceEventType indicates a comment was created.
	CommentCreatedServiceEventType = "comment_created"
	// CommentUpdatedServiceEventType indicates a comment was updated.
	CommentUpdatedServiceEventType = "comment_updated"
	// CommentArchivedServiceEventType indicates a comment was archived.
	CommentArchivedServiceEventType = "comment_archived"
)

func init() {
	gob.Register(new(Comment))
	gob.Register(new(CommentCreationRequestInput))
	gob.Register(new(CommentUpdateRequestInput))
}

type (
	// Comment represents a comment on a reference (recipe, meal, meal_plan).
	Comment struct {
		_ struct{} `json:"-"`

		CreatedAt       time.Time  `json:"createdAt"`
		ParentCommentID *string    `json:"parentCommentId,omitempty"`
		LastUpdatedAt   *time.Time `json:"lastUpdatedAt,omitempty"`
		ArchivedAt      *time.Time `json:"archivedAt,omitempty"`
		ID              string     `json:"id"`
		Content         string     `json:"content"`
		TargetType      string     `json:"targetType"`
		ReferencedID    string     `json:"referencedId"`
		BelongsToUser   string     `json:"belongsToUser"`
	}

	// CommentCreationRequestInput represents what a user could set as input for creating comments.
	CommentCreationRequestInput struct {
		_ struct{} `json:"-"`

		Content         string  `json:"content"`
		TargetType      string  `json:"targetType"`
		ReferencedID    string  `json:"referencedId"`
		ParentCommentID *string `json:"parentCommentId,omitempty"`
		BelongsToUser   string  `json:"belongsToUser"`
	}

	// CommentDatabaseCreationInput represents what is stored when creating a comment.
	CommentDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID              string  `json:"-"`
		Content         string  `json:"-"`
		TargetType      string  `json:"-"`
		ReferencedID    string  `json:"-"`
		ParentCommentID *string `json:"-"`
		BelongsToUser   string  `json:"-"`
	}

	// CommentUpdateRequestInput represents what a user could set as input for updating comments.
	CommentUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Content string `json:"content"`
	}

	// CommentDatabaseUpdateInput represents what is stored when updating a comment.
	CommentDatabaseUpdateInput struct {
		_ struct{} `json:"-"`

		ID            string `json:"-"`
		BelongsToUser string `json:"-"`
		Content       string `json:"-"`
	}

	// CommentDataManager describes a structure capable of storing comments permanently.
	CommentDataManager interface {
		CreateComment(ctx context.Context, input *CommentDatabaseCreationInput) (*Comment, error)
		GetComment(ctx context.Context, id string) (*Comment, error)
		GetCommentsForReference(ctx context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Comment], error)
		GetCommentsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Comment], error)
		UpdateComment(ctx context.Context, id, belongsToUser, content string) error
		ArchiveComment(ctx context.Context, id string) error
		ArchiveCommentsForReference(ctx context.Context, targetType, referencedID string) error
	}
)

var _ validation.ValidatableWithContext = (*CommentCreationRequestInput)(nil)

// ValidateWithContext validates a CommentCreationRequestInput.
func (x *CommentCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Content, validation.Required, validation.Length(1, 10000)),
		validation.Field(&x.TargetType, validation.Required),
		validation.Field(&x.ReferencedID, validation.Required),
		validation.Field(&x.BelongsToUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*CommentDatabaseCreationInput)(nil)

// ValidateWithContext validates a CommentDatabaseCreationInput.
func (x *CommentDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Content, validation.Required, validation.Length(1, 10000)),
		validation.Field(&x.TargetType, validation.Required),
		validation.Field(&x.ReferencedID, validation.Required),
		validation.Field(&x.BelongsToUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*CommentUpdateRequestInput)(nil)

// ValidateWithContext validates a CommentUpdateRequestInput.
func (x *CommentUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Content, validation.Required, validation.Length(1, 10000)),
	)
}

var _ validation.ValidatableWithContext = (*CommentDatabaseUpdateInput)(nil)

// ValidateWithContext validates a CommentDatabaseUpdateInput.
func (x *CommentDatabaseUpdateInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.BelongsToUser, validation.Required),
		validation.Field(&x.Content, validation.Required, validation.Length(1, 10000)),
	)
}
