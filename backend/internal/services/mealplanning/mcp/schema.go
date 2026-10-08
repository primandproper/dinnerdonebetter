package mcp

import (
	"fmt"
	"reflect"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/google/jsonschema-go/jsonschema"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// add registers one tool, with its schemas reflected off the handler's types.
//
// The tool is copied before its schemas are filled in, so that two servers registering the same
// surface — every test's — do not write the package-level definition under each other.
func add[In, Out any](server *sdkmcp.Server, tool *sdkmcp.Tool, handler sdkmcp.ToolHandlerFor[In, Out]) {
	registered := *tool
	registered.InputSchema = schemaFor[In](tool.Name)
	registered.OutputSchema = schemaFor[Out](tool.Name)

	sdkmcp.AddTool(server, &registered, handler)
}

// schemaFor is the schema a model is told for T: the SDK's own reflection, with the one type it
// cannot follow described by hand.
//
// A Recipe lists the recipes associated with it, and those are recipes: a cycle the reflection
// refuses rather than unrolls. The nested list is described as a list of objects, which is
// what it is, and the top-level shape stays the type's. It panics as AddTool does for a type it
// cannot describe, because that is a surface that cannot be served rather than a call that
// cannot be answered.
//
// A pointer is described as what it points at. Every handler answers with a pointer, and the
// protocol wants the row's shape rather than a nullable one.
func schemaFor[T any](tool string) *jsonschema.Schema {
	typ := reflect.TypeFor[T]()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	schema, err := jsonschema.ForType(typ, &jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[[]*mealplanning.Recipe](): {
				Type:  "array",
				Items: &jsonschema.Schema{Type: "object"},
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("tool %q: %v", tool, err))
	}

	return schema
}

// readOnly marks a tool that changes nothing, which every tool here is.
func readOnly() *sdkmcp.ToolAnnotations {
	return &sdkmcp.ToolAnnotations{ReadOnlyHint: true}
}
