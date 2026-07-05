package fakes

import (
	uploadedmedia "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v3/identifiers"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUploadedMedia builds a fake UploadedMedia.
func BuildFakeUploadedMedia() *uploadedmedia.UploadedMedia {
	return &uploadedmedia.UploadedMedia{
		ID:            identifiers.New(),
		StoragePath:   fake.URL(),
		MimeType:      uploadedmedia.MimeTypeImagePNG,
		CreatedByUser: identifiers.New(),
	}
}

// BuildFakeUploadedMediaCreationRequestInput builds a fake UploadedMediaCreationRequestInput.
func BuildFakeUploadedMediaCreationRequestInput() *uploadedmedia.UploadedMediaCreationRequestInput {
	return &uploadedmedia.UploadedMediaCreationRequestInput{
		StoragePath: fake.URL(),
		MimeType:    uploadedmedia.MimeTypeImagePNG,
	}
}

// BuildFakeUploadedMediaDatabaseCreationInput builds a fake UploadedMediaDatabaseCreationInput.
func BuildFakeUploadedMediaDatabaseCreationInput() *uploadedmedia.UploadedMediaDatabaseCreationInput {
	return &uploadedmedia.UploadedMediaDatabaseCreationInput{
		ID:            identifiers.New(),
		StoragePath:   fake.URL(),
		MimeType:      uploadedmedia.MimeTypeImagePNG,
		CreatedByUser: identifiers.New(),
	}
}

// BuildFakeUploadedMediaUpdateRequestInput builds a fake UploadedMediaUpdateRequestInput.
func BuildFakeUploadedMediaUpdateRequestInput() *uploadedmedia.UploadedMediaUpdateRequestInput {
	storagePath := fake.URL()
	mimeType := uploadedmedia.MimeTypeImageJPEG

	return &uploadedmedia.UploadedMediaUpdateRequestInput{
		StoragePath: &storagePath,
		MimeType:    &mimeType,
	}
}
