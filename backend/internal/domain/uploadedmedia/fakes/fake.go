package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

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
func BuildFakeUploadedMedia() *mediaregistry.Object {
	object := fake.BuildFakeRecord[mediaregistry.Object]()

	object.Scope = uploadedmedia.Scope()
	object.ContentType = uploadedmedia.MimeTypeImagePNG
	object.Key = gofakeit.URL()
	object.BelongsTo = mediaregistry.Subject{}

	return object
}

// BuildFakeUploadedMediaFor builds a fake registry Object attached to subject.
func BuildFakeUploadedMediaFor(subject mediaregistry.Subject) *mediaregistry.Object {
	object := BuildFakeUploadedMedia()
	object.BelongsTo = subject

	return object
}

// BuildFakeUploadedMediaList builds a faked page of registry Objects.
func BuildFakeUploadedMediaList() *filtering.QueryFilteredResult[mediaregistry.Object] {
	return fake.BuildFakePage(BuildFakeUploadedMedia)
}

// BuildFakeUploadedMediaInput builds what a caller hands mediaregistry.Store.RecordObject.
//
// It is derived from a fake Object rather than faked independently, so the two cannot drift:
// what a test registers and what it then asserts on are the same values. The id is carried
// across rather than left for the store to mint, because this application builds a storage key
// from the id — the bytes have to know where they are going before anything writes them.
func BuildFakeUploadedMediaInput() mediaregistry.ObjectInput {
	return InputFromUploadedMedia(BuildFakeUploadedMedia())
}

// InputFromUploadedMedia is the RecordObject input that would store object as it stands.
func InputFromUploadedMedia(object *mediaregistry.Object) mediaregistry.ObjectInput {
	return mediaregistry.ObjectInput{
		BelongsTo:   object.BelongsTo,
		ID:          object.ID,
		Key:         object.Key,
		ContentType: object.ContentType,
		OwnerID:     object.OwnerID,
		Size:        object.Size,
	}
}
