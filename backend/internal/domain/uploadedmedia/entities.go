package uploadedmedia

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: UploadedMedia{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "StoragePath", Expr: `fake.URL()`},
					{Name: "MimeType", Expr: `types.MimeTypeImagePNG`},
					{Name: "CreatedAt", Expr: `time.Time{}`},
				},
			},
		},
		{
			Type: UploadedMediaCreationRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "StoragePath", Expr: `fake.URL()`},
					{Name: "MimeType", Expr: `types.MimeTypeImagePNG`},
				},
			},
		},
		{
			Type: UploadedMediaDatabaseCreationInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "StoragePath", Expr: `fake.URL()`},
					{Name: "MimeType", Expr: `types.MimeTypeImagePNG`},
				},
			},
		},
		{
			Type: UploadedMediaUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `storagePath := fake.URL()`},
					{Code: `mimeType := types.MimeTypeImageJPEG`},
				},
				Fields: []entitydecl.Field{
					{Name: "StoragePath", Expr: `&storagePath`},
					{Name: "MimeType", Expr: `&mimeType`},
				},
			},
		},
	},
}
