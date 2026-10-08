package mcp

import (
	"fmt"
	"reflect"
	"sync"

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
// A pointer is described as what it points at. Every handler answers with a pointer, and the
// protocol wants the row's shape rather than a nullable one. It panics as AddTool does for a
// type it cannot describe, because that is a surface that cannot be served rather than a call
// that cannot be answered.
func schemaFor[T any](tool string) *jsonschema.Schema {
	typ := reflect.TypeFor[T]()
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	schema, err := jsonschema.ForType(typ, &jsonschema.ForOptions{TypeSchemas: typeSchemas()})
	if err != nil {
		panic(fmt.Sprintf("tool %q: %v", tool, err))
	}

	return schema
}

// typeSchemas is the one override the reflection needs: what a list of recipes looks like.
//
// A Recipe lists the recipes associated with it, and those are recipes: a cycle the reflection
// refuses rather than unrolls. A page of recipes stores its rows as the same list type, so the
// override cannot simply say "a list of objects" without describing every row of GetRecipes and
// SearchForRecipes as an object with no fields. Instead the recipe is reflected once with the
// cycle cut, and that full shape is what every list of recipes is described as: a page's rows
// carry every field, as do the associated recipes of a recipe read on its own, and only the
// list inside one of those is left as bare objects. One level unrolled is what a model reads
// anyway.
var typeSchemas = sync.OnceValue(func() map[reflect.Type]*jsonschema.Schema {
	recipeList := reflect.TypeFor[[]*mealplanning.Recipe]()

	recipe, err := jsonschema.ForType(reflect.TypeFor[mealplanning.Recipe](), &jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			recipeList: {Type: "array", Items: &jsonschema.Schema{Type: "object"}},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("describing a recipe: %v", err))
	}

	return map[reflect.Type]*jsonschema.Schema{
		recipeList: {Type: "array", Items: recipe},
	}
})

// readOnly marks a tool that changes nothing, which every tool here is.
func readOnly() *sdkmcp.ToolAnnotations {
	return &sdkmcp.ToolAnnotations{ReadOnlyHint: true}
}
