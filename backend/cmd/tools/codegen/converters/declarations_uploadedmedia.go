package main

// Conversions declared for the uploadedmedia domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("uploadedmedia", []*Conversion{
		{Name: "ConvertUploadedMediaToUploadedMediaCreationRequestInput", From: Param{Name: "x", Type: "UploadedMedia"}, To: "UploadedMediaCreationRequestInput"},
		{Name: "ConvertUploadedMediaToUploadedMediaUpdateRequestInput", From: Param{Name: "x", Type: "UploadedMedia"}, To: "UploadedMediaUpdateRequestInput"},
		{Name: "ConvertUploadedMediaToUploadedMediaDatabaseCreationInput", From: Param{Name: "x", Type: "UploadedMedia"}, To: "UploadedMediaDatabaseCreationInput"},
	})
}
