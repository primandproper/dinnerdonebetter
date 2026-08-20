package fakes

import (
	uploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUploadedMedia builds a fake UploadedMedia.
func BuildFakeUploadedMedia() *uploadedmedia.UploadedMedia {
	media := fake.BuildFakeRecord[uploadedmedia.UploadedMedia]()

	// A MIME type is a value the uploads path matches against a list of the ones it
	// accepts, and a storage path is a URL something later fetches.
	media.MimeType = uploadedmedia.MimeTypeImagePNG
	media.StoragePath = gofakeit.URL()

	return media
}

// BuildFakeUploadedMediaCreationRequestInput builds a fake UploadedMediaCreationRequestInput.
func BuildFakeUploadedMediaCreationRequestInput() *uploadedmedia.UploadedMediaCreationRequestInput {
	input := fake.BuildFakeRecord[uploadedmedia.UploadedMediaCreationRequestInput]()
	input.MimeType = uploadedmedia.MimeTypeImagePNG
	input.StoragePath = gofakeit.URL()

	return input
}

// BuildFakeUploadedMediaDatabaseCreationInput builds a fake UploadedMediaDatabaseCreationInput.
func BuildFakeUploadedMediaDatabaseCreationInput() *uploadedmedia.UploadedMediaDatabaseCreationInput {
	input := fake.BuildFakeRecord[uploadedmedia.UploadedMediaDatabaseCreationInput]()
	input.MimeType = uploadedmedia.MimeTypeImagePNG
	input.StoragePath = gofakeit.URL()

	return input
}

// BuildFakeUploadedMediaUpdateRequestInput builds a fake UploadedMediaUpdateRequestInput.
//
// Both fields are optional, and an update input whose fields are all absent updates
// nothing, so they are filled here rather than left to BuildFakeRecord. The JPEG is
// deliberately not the PNG the builders above use: an update that changes nothing is
// an update whose effect no assertion can see.
func BuildFakeUploadedMediaUpdateRequestInput() *uploadedmedia.UploadedMediaUpdateRequestInput {
	return &uploadedmedia.UploadedMediaUpdateRequestInput{
		StoragePath: pointer.To(gofakeit.URL()),
		MimeType:    pointer.To(uploadedmedia.MimeTypeImageJPEG),
	}
}
