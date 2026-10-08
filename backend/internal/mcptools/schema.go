package mcptools

const (
	jsonSchemaVersion = "https://json-schema.org/draft/2020-12/schema"

	objType    = "object"
	arrType    = "array"
	strType    = "string"
	boolType   = "boolean"
	intType    = "integer"
	numberType = "number"

	dtFmt = "date-time"

	// The JSON Schema keywords the helpers below emit by hand.
	KeyDescription = "description"
	KeyType        = "type"
)

// The property names the hand-written tool schemas repeat across domains. They are the wire
// names a model reads and writes, so they are spelled once here rather than once per tool. A
// domain's own names live beside its tools.
const (
	FieldArchivedAt    = "ArchivedAt"
	FieldCreatedAt     = "CreatedAt"
	FieldDescription   = "Description"
	FieldFilter        = "Filter"
	FieldLastUpdatedAt = "LastUpdatedAt"
	FieldName          = "Name"
	FieldQuery         = "Query"
	FieldResults       = "Results"
)

// The schema for a QueryFilter is not here. filtering.QueryFilterSchema reflects it off the
// struct, and the hand-written mirror that used to live here is why: it described SortBy as the
// field to sort by rather than the direction to sort in and carried no enum, declared
// MaxResponseSize as an unbounded integer, and — because the invocation structs hand the
// decoded object straight to encoding/json — keyed every property on the Go field name against
// camelCase json tags, so a filter a model supplied was dropped in full and the list came back
// unfiltered and plausible.
//
// The rest of this file has the same exposure. Every domain type's schema is written out by
// hand beside a struct that can move without it, and the drift would look the same. platform's
// mcptool package replaces these with schemas reflected off the type; see the package doc.

// SchemaObject is a top-level object schema over properties.
func SchemaObject(properties map[string]any) map[string]any {
	return map[string]any{
		"$schema":    jsonSchemaVersion,
		KeyType:      objType,
		"properties": properties,
	}
}

// ObjectType is a nested object schema over fieldSchema, requiring requiredFields.
func ObjectType(fieldSchema map[string]any, requiredFields ...string) map[string]any {
	x := map[string]any{
		KeyType:      objType,
		"properties": fieldSchema,
	}

	if len(requiredFields) > 0 {
		x["required"] = requiredFields
	}

	return x
}

// ArrayType is an array schema whose items are fieldSchema.
func ArrayType(fieldSchema map[string]any) map[string]any {
	return map[string]any{
		KeyType: arrType,
		"items": fieldSchema,
	}
}

// FloatField is a number property.
func FloatField(description string) map[string]any {
	return map[string]any{
		KeyType:        numberType,
		KeyDescription: description,
	}
}

// UintField is a non-negative integer property.
func UintField(description string) map[string]any {
	return map[string]any{
		KeyType:        intType,
		KeyDescription: description,
		"minimum":      0,
	}
}

// BoolField is a boolean property.
func BoolField(description string) map[string]any {
	return map[string]any{
		KeyType:        boolType,
		KeyDescription: description,
	}
}

// StringField is a string property.
func StringField(description string) map[string]any {
	return map[string]any{
		KeyType:        strType,
		KeyDescription: description,
	}
}

// TimestampField is a string property in RFC 3339 date-time format.
func TimestampField(description string) map[string]any {
	return StringFieldWithFormat(description, dtFmt)
}

// StringFieldWithFormat is a string property carrying a JSON Schema format.
func StringFieldWithFormat(description, format string) map[string]any {
	return map[string]any{
		KeyType:        strType,
		KeyDescription: description,
		"format":       format,
	}
}
