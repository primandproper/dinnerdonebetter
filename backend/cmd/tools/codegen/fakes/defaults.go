package main

import (
	"fmt"
	"reflect"
	"strings"
)

// domainAlias is the import alias every generated file uses for its own domain package.
//
// Fixed rather than the package name, because the fakes packages disagreed: some imported
// mealplanning as `types` and some as `mealplanning`, in the same directory. One alias means a
// field override reads the same in every domain, which matters when the override is copied from
// the domain next door.
const domainAlias = "types"

// defaulter turns a struct field into the expression a fake assigns to it.
//
// It holds the entity types the domain declares, because the most useful default for a field
// whose type is another entity is that entity's builder — and knowing which types have builders
// is knowing the declaration.
type defaulter struct {
	// builders maps an entity type to the name of the builder that produces a pointer to it.
	builders map[reflect.Type]string
	// domainPath is the import path of the domain being generated, which is the one package
	// the generated file refers to under a fixed alias.
	domainPath string
}

// expr returns the mechanical default for a field, or ok=false when there is none.
//
// A field with no default is not an error here — it is a field the declaration has to override.
// Reporting it as missing rather than guessing is the whole point: a guessed value for a field
// the type actually constrains produces a fake that fails validation, and a fake that fails
// validation fails it in whatever test happens to use it, a long way from this decision.
func (d *defaulter) expr(field *reflect.StructField) (expression string, ok bool) {
	t := field.Type

	// time.Time is a struct, so it has to be answered before the struct rules below get a
	// chance to look for a builder for it.
	if isTime(t) {
		return "BuildFakeTime()", true
	}

	switch t.Kind() {
	case reflect.String:
		// A named string type is an enumeration in every case this repository has: a
		// status, a kind, a shape. Its valid values are constants the generator cannot
		// enumerate, and a random string is not one of them.
		if isNamed(t) {
			return "", false
		}

		return stringExpr(field.Name), true

	case reflect.Bool:
		return "fake.Bool()", true

	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return d.numberExpr(t), true

	case reflect.Struct:
		if builder, found := d.builders[t]; found {
			return "*" + builder + "()", true
		}

		return "", false

	case reflect.Pointer:
		// An optional field is left unset. Optional fields are optional in the domain's
		// own judgement, so a fake that fills them exercises a shape the API does not
		// require, and every test that wants one set says so itself.
		return "nil", true

	case reflect.Slice, reflect.Map, reflect.Interface:
		// A child collection is almost never independent of its parent — the children
		// carry the parent's ID — so the useful version is a local, not a default.
		return "nil", true

	default:
		return "", false
	}
}

// stringExpr is the default for an unnamed string field.
//
// The name carries the type here in a way the type system does not: `BelongsToAccount` and
// `Description` are both strings, but only one of them is rejected for not being an identifier.
func stringExpr(name string) string {
	// `...User` covers BelongsToUser, CreatedByUser and ByUser, which hold a user's ID and
	// are as much identifiers as anything ending in ID.
	if name == "ID" || strings.HasSuffix(name, "ID") || strings.HasSuffix(name, "User") || strings.HasPrefix(name, "BelongsTo") {
		return "BuildFakeID()"
	}

	return "buildUniqueString()"
}

// numberExpr is the default for a numeric field.
//
// buildFakeNumber returns a float64 above 100 and below 127, which fits every numeric type the
// domains use and stays clear of the zero that so many validators reject.
func (d *defaulter) numberExpr(t reflect.Type) string {
	if t.Kind() == reflect.Float64 && !isNamed(t) {
		return "buildFakeNumber()"
	}

	return fmt.Sprintf("%s(buildFakeNumber())", d.qualify(t))
}

// qualify renders a type as the generated file refers to it.
func (d *defaulter) qualify(t reflect.Type) string {
	if !isNamed(t) {
		return t.String()
	}

	alias := defaultAlias(t.PkgPath())
	if t.PkgPath() == d.domainPath {
		alias = domainAlias
	}

	return fmt.Sprintf("%s.%s", alias, t.Name())
}

// isNamed reports whether t is a defined type declared in some package, as opposed to a
// predeclared type like string or a composite like []byte.
func isNamed(t reflect.Type) bool {
	return t.Name() != "" && t.PkgPath() != ""
}

// isTime reports whether t is time.Time.
func isTime(t reflect.Type) bool {
	return t.PkgPath() == "time" && t.Name() == "Time"
}
