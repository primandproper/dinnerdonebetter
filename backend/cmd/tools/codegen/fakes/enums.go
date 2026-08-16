package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// enumIndex is the constants declared for each named type, keyed by import path and type name.
//
// Reflection can see that a field's type is `types.ProductKind` and that its kind is string. It
// cannot see that `recurring` and `one_time` are the only two strings that type is ever allowed to
// hold, because constants are erased by the time a program is running. That knowledge only exists
// in the source, so the source is what this reads.
//
// Without it every enumerated field is a field the generator has no default for, and the ninety
// fake builders in this repository between them touch enough of them that "declare an override for
// each" was most of what the declarations had grown into.
// The second and more useful source is the type's own validation. Almost none of the enumerated
// fields in this repository are typed: `MealPlan.Status` is a `string`, and the constants that may
// go in it are untyped constants sharing a name prefix. What does say which values are allowed is
// the rule the type declares about itself:
//
//	validation.Field(&x.Shape, validation.In(VesselShapeCone, VesselShapeSphere, ...))
//
// That list is authoritative in a way a name prefix never is — it is what rejects a bad value at
// runtime — and it is maintained, because a value missing from it fails. Reading it is what lets a
// generated fake be one the domain would accept, which is the property the whole generator is for.
type enumIndex struct {
	// byType maps "import/path.TypeName" to the constant names declared with that type, in
	// source order.
	byType map[string][]string
	// byField maps "import/path.TypeName.FieldName" to the values that type's validation
	// admits for that field, in source order.
	byField map[string][]string
	// root is the backend module's directory, which is what import paths are resolved against.
	root string
	// mu guards the two maps, which are filled lazily. The generator is single-threaded, but
	// its tests are not: this package's convention is that every subtest runs in parallel.
	mu sync.Mutex
}

// newEnumIndex indexes the packages the generated fakes can refer to.
func newEnumIndex(root string) *enumIndex {
	return &enumIndex{byType: map[string][]string{}, byField: map[string][]string{}, root: root}
}

// permitted returns the values a type's validation admits for one of its fields, or nil when it
// says nothing about it.
func (e *enumIndex) permitted(pkgPath, typeName, fieldName string) ([]string, error) {
	if !strings.HasPrefix(pkgPath, backendPrefix) {
		return nil, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.load(pkgPath); err != nil {
		return nil, err
	}

	if values := e.byField[pkgPath+"."+typeName+"."+fieldName]; len(values) > 0 {
		return values, nil
	}

	// An entity usually declares no validation about itself — it is what comes back out of
	// the database, and the database is where it was checked. The rule for its fields is on
	// the input types that put it there, which describe the same field under the same name.
	for _, suffix := range inputSuffixes {
		if values := e.byField[pkgPath+"."+typeName+suffix+"."+fieldName]; len(values) > 0 {
			return values, nil
		}
	}

	return nil, nil
}

// inputSuffixes are the types that carry an entity's validation, in the order they are consulted.
//
// Named rather than matched by prefix: `Recipe` is a prefix of `RecipeStep`, and a rule borrowed
// from the wrong type would be a list of values the field cannot actually hold.
var inputSuffixes = []string{"CreationRequestInput", "DatabaseCreationInput", "UpdateRequestInput"}

// constants returns the constant names declared for a named type, or nil when it has none.
//
// Packages are parsed on first use rather than up front: a domain's fakes refer to a handful of
// packages and there is no reason to read the rest of the tree to find that out.
func (e *enumIndex) constants(pkgPath, typeName string) ([]string, error) {
	if !strings.HasPrefix(pkgPath, backendPrefix) {
		// A named string type from outside this module — an enum in a platform package,
		// say. Its constants are readable in principle, but its source is in the module
		// cache and this generator has no business reading there. Such a field needs an
		// override.
		return nil, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.load(pkgPath); err != nil {
		return nil, err
	}

	return e.byType[pkgPath+"."+typeName], nil
}

// load parses one package's constant declarations, once. Callers hold mu.
func (e *enumIndex) load(pkgPath string) error {
	if _, done := e.byType[pkgPath]; done {
		return nil
	}

	// A sentinel under the package's own path, which is not a legal "path.Type" key, marks
	// the package parsed even when it declares no constants at all.
	e.byType[pkgPath] = nil

	dir := filepath.Join(e.root, strings.TrimPrefix(strings.TrimPrefix(pkgPath, backendPrefix), "/"))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading %s for enumerated constants: %w", pkgPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		file, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, 0)
		if parseErr != nil {
			return fmt.Errorf("parsing %s: %w", entry.Name(), parseErr)
		}

		e.collect(pkgPath, file)
		e.collectValidations(pkgPath, file)
	}

	return nil
}

// collectValidations records the values each type's validation rules admit, per field.
func (e *enumIndex) collectValidations(pkgPath string, file *ast.File) {
	for _, decl := range file.Decls {
		function, isFunction := decl.(*ast.FuncDecl)
		if !isFunction || function.Recv == nil || len(function.Recv.List) != 1 {
			continue
		}

		receiver := receiverTypeName(function.Recv.List[0].Type)
		if receiver == "" {
			continue
		}

		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall || !isSelector(call.Fun, "validation", "Field") || len(call.Args) < 2 {
				return true
			}

			field := validatedFieldName(call.Args[0])
			if field == "" {
				return true
			}

			if values := admittedValues(call.Args[1:]); len(values) > 0 {
				e.byField[pkgPath+"."+receiver+"."+field] = values
			}

			return true
		})
	}
}

// admittedValues pulls the members out of a validation.In rule, wherever it sits in the arguments.
//
// A rule is skipped whole when any member is something other than a name or a literal. The point
// of reading these is to emit a value the domain accepts; half a list would not do that, and an
// expression this cannot render would not compile.
func admittedValues(args []ast.Expr) []string {
	var values []string

	for _, arg := range args {
		ast.Inspect(arg, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall || !isSelector(call.Fun, "validation", "In") || len(call.Args) == 0 {
				return true
			}

			members := make([]string, 0, len(call.Args))

			for _, member := range call.Args {
				switch value := member.(type) {
				case *ast.Ident:
					members = append(members, value.Name)
				case *ast.BasicLit:
					members = append(members, value.Value)
				default:
					return false
				}
			}

			values = members

			return false
		})

		if len(values) > 0 {
			return values
		}
	}

	return values
}

// receiverTypeName is the name of the type a method is declared on.
func receiverTypeName(expr ast.Expr) string {
	if star, isPointer := expr.(*ast.StarExpr); isPointer {
		expr = star.X
	}

	if ident, isIdent := expr.(*ast.Ident); isIdent {
		return ident.Name
	}

	return ""
}

// validatedFieldName is the field a validation.Field rule is about, given its `&x.Field` argument.
func validatedFieldName(expr ast.Expr) string {
	unary, isUnary := expr.(*ast.UnaryExpr)
	if !isUnary || unary.Op != token.AND {
		return ""
	}

	selector, isSelector := unary.X.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}

	return selector.Sel.Name
}

// isSelector reports whether an expression is the call `pkg.name`.
func isSelector(expr ast.Expr, pkg, name string) bool {
	selector, isSelector := expr.(*ast.SelectorExpr)
	if !isSelector || selector.Sel.Name != name {
		return false
	}

	ident, isIdent := selector.X.(*ast.Ident)

	return isIdent && ident.Name == pkg
}

// collect records every typed constant declaration in one file.
func (e *enumIndex) collect(pkgPath string, file *ast.File) {
	for _, decl := range file.Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}

		// Within a const block a spec with neither type nor value repeats the previous
		// spec's, which is how an iota run is written. Carrying the type forward is what
		// makes those members visible here.
		carried := ""

		for _, spec := range general.Specs {
			value, isValue := spec.(*ast.ValueSpec)
			if !isValue {
				continue
			}

			typeName := carried

			if named, isNamed := value.Type.(*ast.Ident); isNamed {
				typeName = named.Name
				carried = typeName
			} else if len(value.Values) > 0 {
				// Typed by its value, not by the block: an untyped constant that
				// happens to sit in the same run. It is not a member.
				typeName = ""
			}

			if typeName == "" {
				continue
			}

			key := pkgPath + "." + typeName

			for _, name := range value.Names {
				if name.IsExported() {
					e.byType[key] = append(e.byType[key], name.Name)
				}
			}
		}
	}
}
