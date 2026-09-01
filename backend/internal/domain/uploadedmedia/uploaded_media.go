/*
Package uploadedmedia is this application's half of platform-go's upload
registry: the namespace its table carries, the tenancy every row is filed under,
the content types this application accepts, and the data change events a write
emits.

The registry itself is platform-go's. It owns the schema, the paging, the
tenancy column, the key uniqueness and the ownership check that decides who may
read an object, because that half is the same in every application. What is not
the same is which content types this application is willing to store, and that
list stays here.

# The row is not the bytes

Nothing in the registry opens, reads or removes an object. uploads.UploadManager
moves bytes and hands back a key; the registry row is what makes that key
answerable — whose object it is, how big it was, and what it hangs off. Archival
is metadata-only on purpose: the row is hidden and the object stays in the
bucket until a retention policy removes it.
*/
package uploadedmedia

import (
	"github.com/primandproper/platform-go/v13/tenancy"
)

// TablePrefix namespaces the platform-go upload registry table, rendering
// ddb_uploads_objects.
//
// The platform's own default is the empty prefix, which renders
// "uploads_objects". Its DDL says CREATE TABLE IF NOT EXISTS, so a collision
// with anything else sharing the database would be a silent no-op followed by a
// store reading columns that are not there.
const TablePrefix = "ddb"

// The data change events an uploaded media write emits. They are declared in the
// webhook event catalog (internal/domain/webhooks/catalog), so a subscriber is
// already able to ask for them.
const (
	// UploadedMediaCreatedServiceEventType indicates uploaded media was created.
	UploadedMediaCreatedServiceEventType = "uploaded_media_created"
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
//
// The registry stores whatever content type it is handed — a registry that
// vetted the vocabulary would be one only this application's consumers could
// use — so the list of what this deployment accepts is enforced at the service
// boundary, before an object is stored. It is also what bounds the cardinality
// of the mime_type dimension the upload meter records.
func IsValidMimeType(mimeType string) bool {
	switch mimeType {
	case MimeTypeImagePNG, MimeTypeImageJPEG, MimeTypeImageGIF, MimeTypeVideoMP4:
		return true
	default:
		return false
	}
}

// Scope is the tenancy every uploaded object in this deployment is filed under.
//
// It is global, and that is a decision rather than a default. An object's
// readability is decided from its owner — the user who uploaded it — and the
// checks the owning service runs before it serves one, not from a tenant
// column: a recipe's step images are readable by everyone who can read the
// recipe, across accounts, so scoping them by account would hide a photograph
// from the household reading the recipe it belongs to.
func Scope() tenancy.Scope { return tenancy.Global() }
