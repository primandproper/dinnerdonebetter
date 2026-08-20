package uploadedmedia

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v12/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// UploadedMediaCreatedServiceEventType indicates uploaded media was created.
	UploadedMediaCreatedServiceEventType = "uploaded_media_created"
	// UploadedMediaUpdatedServiceEventType indicates uploaded media was updated.
	UploadedMediaUpdatedServiceEventType = "uploaded_media_updated"
	// UploadedMediaArchivedServiceEventType indicates uploaded media was archived.
	UploadedMediaArchivedServiceEventType = "uploaded_media_archived"
)

// Supported MIME types for uploaded media.
const (
	MimeTypeImagePNG  = "image/png"
	MimeTypeImageJPEG = "image/jpeg"
	MimeTypeImageGIF  = "image/gif"
	MimeTypeVideoMP4  = "video/mp4"
)

// IsValidMimeType checks if a MIME type is supported.
func IsValidMimeType(mimeType string) bool {
	switch mimeType {
	case MimeTypeImagePNG, MimeTypeImageJPEG, MimeTypeImageGIF, MimeTypeVideoMP4:
		return true
	default:
		return false
	}
}

func init() {
	gob.Register(new(UploadedMedia))
	gob.Register(new(UploadedMediaCreationRequestInput))
	gob.Register(new(UploadedMediaDatabaseCreationInput))
	gob.Register(new(UploadedMediaUpdateRequestInput))
}

type (
	// UploadedMedia represents a media file uploaded by a user.
	//
	// It is not a duplicate of platform-go's uploads types. Those — ObjectInfo, Attributes —
	// are blob-store descriptors: what the bucket says about a path, produced by the store and
	// owned by it. This is the application's record that a user put something there, and it
	// holds what the bucket cannot know: who uploaded it, when this application accepted it,
	// and whether it has since been archived. StoragePath is the join between the two, and the
	// direction of that join is the point — the row references the object, and the object has
	// never heard of the row. The platform's uploader is consumed directly wherever a file is
	// actually written; see internal/services/mealplanning/grpc/image_uploads.go.
	UploadedMedia struct {
		_             struct{}   `json:"-"`
		CreatedAt     time.Time  `json:"createdAt"`
		LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time `json:"archivedAt"`
		ID            string     `json:"id"`
		StoragePath   string     `json:"storagePath"`
		MimeType      string     `json:"mimeType"`
		CreatedByUser string     `json:"createdByUser"`
	}

	// UploadedMediaCreationRequestInput represents input for creating uploaded media.
	UploadedMediaCreationRequestInput struct {
		_           struct{} `json:"-"`
		StoragePath string   `json:"storagePath"`
		MimeType    string   `json:"mimeType"`
	}

	// UploadedMediaDatabaseCreationInput is used for creating uploaded media in persistence.
	UploadedMediaDatabaseCreationInput struct {
		_             struct{} `json:"-"`
		ID            string   `json:"-"`
		StoragePath   string   `json:"-"`
		MimeType      string   `json:"-"`
		CreatedByUser string   `json:"-"`
	}

	// UploadedMediaUpdateRequestInput represents input for updating uploaded media.
	UploadedMediaUpdateRequestInput struct {
		_           struct{} `json:"-"`
		StoragePath *string  `json:"storagePath,omitempty"`
		MimeType    *string  `json:"mimeType,omitempty"`
	}

	// UploadedMediaDataManager describes a structure capable of storing uploaded media.
	UploadedMediaDataManager interface {
		GetUploadedMedia(ctx context.Context, uploadedMediaID string) (*UploadedMedia, error)
		GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*UploadedMedia, error)
		GetUploadedMediaForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[UploadedMedia], error)
		CreateUploadedMedia(ctx context.Context, input *UploadedMediaDatabaseCreationInput) (*UploadedMedia, error)
		UpdateUploadedMedia(ctx context.Context, uploadedMedia *UploadedMedia) error
		ArchiveUploadedMedia(ctx context.Context, uploadedMediaID string) error
	}
)

// Update merges an UploadedMediaUpdateRequestInput into an UploadedMedia.
func (u *UploadedMedia) Update(input *UploadedMediaUpdateRequestInput) {
	if input.StoragePath != nil && *input.StoragePath != u.StoragePath {
		u.StoragePath = *input.StoragePath
	}
	if input.MimeType != nil && *input.MimeType != u.MimeType {
		u.MimeType = *input.MimeType
	}
}

var _ validation.ValidatableWithContext = (*UploadedMediaCreationRequestInput)(nil)

// ValidateWithContext validates an UploadedMediaCreationRequestInput.
func (u *UploadedMediaCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		u,
		validation.Field(&u.StoragePath, validation.Required),
		validation.Field(&u.MimeType, validation.Required, validation.In(
			MimeTypeImagePNG,
			MimeTypeImageJPEG,
			MimeTypeImageGIF,
			MimeTypeVideoMP4,
		)),
	)
}

var _ validation.ValidatableWithContext = (*UploadedMediaDatabaseCreationInput)(nil)

// ValidateWithContext validates an UploadedMediaDatabaseCreationInput.
func (u *UploadedMediaDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		u,
		validation.Field(&u.ID, validation.Required),
		validation.Field(&u.StoragePath, validation.Required),
		validation.Field(&u.MimeType, validation.Required, validation.In(
			MimeTypeImagePNG,
			MimeTypeImageJPEG,
			MimeTypeImageGIF,
			MimeTypeVideoMP4,
		)),
		validation.Field(&u.CreatedByUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*UploadedMediaUpdateRequestInput)(nil)

// ValidateWithContext validates an UploadedMediaUpdateRequestInput.
func (u *UploadedMediaUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		u,
		validation.Field(&u.StoragePath, validation.When(u.StoragePath != nil, validation.Required)),
		validation.Field(&u.MimeType, validation.When(u.MimeType != nil, validation.Required, validation.In(
			MimeTypeImagePNG,
			MimeTypeImageJPEG,
			MimeTypeImageGIF,
			MimeTypeVideoMP4,
		))),
	)
}
