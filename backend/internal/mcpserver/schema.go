package mcpserver

const (
	jsonSchemaVersion = "https://json-schema.org/draft/2020-12/schema"

	objType    = "object"
	arrType    = "array"
	strType    = "string"
	boolType   = "boolean"
	intType    = "integer"
	numberType = "number"

	dtFmt = "date-time"
)

// The schema for a QueryFilter is not here. filtering.QueryFilterSchema reflects it off the
// struct, and the hand-written mirror that used to live here is why: it described SortBy as the
// field to sort by rather than the direction to sort in and carried no enum, declared
// MaxResponseSize as an unbounded integer, and — because the invocation structs below hand the
// decoded object straight to encoding/json — keyed every property on the Go field name against
// camelCase json tags, so a filter a model supplied was dropped in full and the list came back
// unfiltered and plausible.
//
// The rest of this file has the same exposure. Every domain type's schema below is written out
// by hand beside a struct that can move without it, and the drift would look the same. Replacing
// those is a larger job than this one — the types are ours rather than platform's, so it means
// deciding where their descriptions live — and it is not done here.

func schemaObject(properties map[string]any) map[string]any {
	return map[string]any{
		"$schema":    jsonSchemaVersion,
		"type":       objType,
		"properties": properties,
	}
}

func objectType(fieldSchema map[string]any, requiredFields ...string) map[string]any {
	x := map[string]any{
		"type":       objType,
		"properties": fieldSchema,
	}

	if len(requiredFields) > 0 {
		x["required"] = requiredFields
	}

	return x
}

func arrayType(fieldSchema map[string]any) map[string]any {
	return map[string]any{
		"type":  arrType,
		"items": fieldSchema,
	}
}

func floatField(description string) map[string]any {
	return map[string]any{
		"type":        numberType,
		"description": description,
	}
}

func uintField(description string) map[string]any {
	return map[string]any{
		"type":        intType,
		"description": description,
		"minimum":     0,
	}
}

func boolField(description string) map[string]any {
	return map[string]any{
		"type":        boolType,
		"description": description,
	}
}

func stringField(description string) map[string]any {
	x := map[string]any{
		"type":        strType,
		"description": description,
	}

	return x
}

func timestampField(description string) map[string]any {
	return stringFieldWithFormat(description, dtFmt)
}

func stringFieldWithFormat(description, format string) map[string]any {
	x := map[string]any{
		"type":        strType,
		"description": description,
		"format":      format,
	}

	return x
}
