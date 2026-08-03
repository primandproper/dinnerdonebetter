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

// queryFilterSchema returns the JSON schema for a QueryFilter object.
func queryFilterSchema() map[string]any {
	return objectType(map[string]any{
		"SortBy":          stringField("Field to sort by"),
		"CreatedAfter":    timestampField("Filter results created after this timestamp (ISO 8601)"),
		"CreatedBefore":   timestampField("Filter results created before this timestamp (ISO 8601)"),
		"UpdatedAfter":    timestampField("Filter results updated after this timestamp (ISO 8601)"),
		"UpdatedBefore":   timestampField("Filter results updated before this timestamp (ISO 8601)"),
		"MaxResponseSize": intField("Maximum number of results to return"),
		"IncludeArchived": boolField("Whether to include archived items"),
		"Cursor":          stringField("Pagination cursor for fetching next page"),
	})
}

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

func intField(description string) map[string]any {
	return map[string]any{
		"type":        intType,
		"description": description,
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
