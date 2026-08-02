package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
)

// ConvertUploadedMediaToUploadedMediaCreationRequestInput creates a UploadedMediaCreationRequestInput from a UploadedMedia.
func ConvertUploadedMediaToUploadedMediaCreationRequestInput(x *uploadedmedia.UploadedMedia) *uploadedmedia.UploadedMediaCreationRequestInput {
	return &uploadedmedia.UploadedMediaCreationRequestInput{
		StoragePath: x.StoragePath,
		MimeType:    x.MimeType,
	}
}

// ConvertUploadedMediaToUploadedMediaUpdateRequestInput creates a UploadedMediaUpdateRequestInput from a UploadedMedia.
func ConvertUploadedMediaToUploadedMediaUpdateRequestInput(x *uploadedmedia.UploadedMedia) *uploadedmedia.UploadedMediaUpdateRequestInput {
	return &uploadedmedia.UploadedMediaUpdateRequestInput{
		StoragePath: &x.StoragePath,
		MimeType:    &x.MimeType,
	}
}

// ConvertUploadedMediaToUploadedMediaDatabaseCreationInput creates a UploadedMediaDatabaseCreationInput from a UploadedMedia.
func ConvertUploadedMediaToUploadedMediaDatabaseCreationInput(x *uploadedmedia.UploadedMedia) *uploadedmedia.UploadedMediaDatabaseCreationInput {
	return &uploadedmedia.UploadedMediaDatabaseCreationInput{
		ID:            x.ID,
		StoragePath:   x.StoragePath,
		MimeType:      x.MimeType,
		CreatedByUser: x.CreatedByUser,
	}
}
