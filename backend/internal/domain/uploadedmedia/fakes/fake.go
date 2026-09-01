package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUploadedMedia builds a fake registry Object.
//
// Three fields are fixed rather than randomized, because each is a value
// something downstream matches rather than merely carries: the scope is the one
// this deployment files every object under and a read in another scope finds
// nothing, the content type is checked against the list of what this
// application accepts, and the key is a path something later fetches from a
// bucket. The subject is left unattached, which is the ordinary state of a
// standalone upload — BuildFakeUploadedMediaFor attaches one.
func BuildFakeUploadedMedia() *registry.Object {
	object := fake.BuildFakeRecord[registry.Object]()

	object.Scope = uploadedmedia.Scope()
	object.ContentType = uploadedmedia.MimeTypeImagePNG
	object.Key = gofakeit.URL()
	object.BelongsTo = registry.Subject{}

	return object
}

// BuildFakeUploadedMediaFor builds a fake registry Object attached to subject.
func BuildFakeUploadedMediaFor(subject registry.Subject) *registry.Object {
	object := BuildFakeUploadedMedia()
	object.BelongsTo = subject

	return object
}

// BuildFakeUploadedMediaList builds a faked page of registry Objects.
func BuildFakeUploadedMediaList() *filtering.QueryFilteredResult[registry.Object] {
	return fake.BuildFakePage(BuildFakeUploadedMedia)
}
