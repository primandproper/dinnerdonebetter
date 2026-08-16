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

const (
	// exprUnset is the expression for a field a fake deliberately leaves empty.
	exprUnset = "nil"
	// exprID is the expression for a field holding an identifier.
	exprID = "BuildFakeID()"
	// packageTime is the standard library package whose two types need answering by name.
	packageTime = "time"
)

// defaulter turns a struct field into the expression a fake assigns to it.
//
// It holds the entity types the domain declares, because the most useful default for a field
// whose type is another entity is that entity's builder — and knowing which types have builders
// is knowing the declaration.
type defaulter struct {
	// builders maps an entity type to the name of the builder that produces a pointer to it.
	builders map[reflect.Type]string
	// enums is what each type says its own fields may hold, which is how an enumerated field
	// gets a value the domain accepts.
	enums *enumIndex
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
func (d *defaulter) expr(owner reflect.Type, field *reflect.StructField) (expression string, ok bool, err error) {
	// What the owning type's own validation admits beats anything the field's type implies:
	// a `string` field the domain will only accept six values for is an enumeration whatever
	// Go thinks it is, and five of those values are what a test needs to see.
	permitted, err := d.enums.permitted(owner.PkgPath(), owner.Name(), field.Name)
	if err != nil {
		return "", false, err
	}

	if len(permitted) > 0 {
		return d.chooseFrom(owner.PkgPath(), permitted, field.Type), true, nil
	}

	return d.exprFor(field.Name, field.Type)
}

// chooseFrom renders a choice between the values a field is allowed to hold.
func (d *defaulter) chooseFrom(pkgPath string, values []string, t reflect.Type) string {
	alias := d.qualifierFor(pkgPath)

	qualified := make([]string, 0, len(values))

	for _, value := range values {
		if strings.HasPrefix(value, `"`) {
			qualified = append(qualified, value)

			continue
		}

		qualified = append(qualified, alias+"."+value)
	}

	choice := qualified[0]
	if len(qualified) > 1 {
		choice = fmt.Sprintf("pickOne(%s)", strings.Join(qualified, ", "))
	}

	// The rule is declared about the value, so an optional field needs it wrapped back up.
	if t.Kind() == reflect.Pointer {
		return fmt.Sprintf("pointer.To(%s)", choice)
	}

	return choice
}

// exprFor is expr, split from it so that a pointer's element can be asked the same question its
// field would have been asked had it not been optional.
func (d *defaulter) exprFor(name string, t reflect.Type) (expression string, ok bool, err error) {
	// time.Time is a struct, so it has to be answered before the struct rules below get a
	// chance to look for a builder for it.
	if isTime(t) {
		return "BuildFakeTime()", true, nil
	}

	// A duration is an int64 of nanoseconds, so the numeric default would produce about a
	// tenth of a microsecond — a value that is technically in range and that no field
	// measuring a timeout, an interval or a lifetime could mean.
	if isDuration(t) {
		return "time.Duration(buildFakeNumber()) * time.Minute", true, nil
	}

	switch t.Kind() {
	case reflect.String:
		// A named string type is an enumeration in every case this repository has: a
		// status, a kind, a shape, and a random string is not one of its members. The
		// members are constants, so they are read out of the source; see enums.go.
		if isNamed(t) {
			return d.enumExpr(t)
		}

		return stringExpr(name), true, nil

	case reflect.Bool:
		return "fake.Bool()", true, nil

	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return d.numberExpr(t), true, nil

	case reflect.Struct:
		if builder, found := d.builders[t]; found {
			return "*" + builder + "()", true, nil
		}

		return "", false, nil

	case reflect.Pointer:
		return d.pointerExpr(name, t.Elem())

	case reflect.Slice, reflect.Map, reflect.Interface:
		// A slice of entities is filled by the child-list derivation in render.go, which
		// gets there first. What reaches here is a collection of something the domain
		// does not declare, and there is no faithful shape to invent for one.
		return exprUnset, true, nil

	default:
		return "", false, nil
	}
}

// pointerExpr is the default for an optional field.
//
// Optional fields are filled rather than left nil. Leaving them nil makes every assertion about
// one vacuous — a test that checks an optional field survives a round trip passes whether or not
// the code copies it — and vacuous assertions are the failure this package exists to prevent. A
// test that specifically wants the unset shape sets the field to nil itself, which is one line and
// is visible where it matters.
func (d *defaulter) pointerExpr(name string, elem reflect.Type) (expression string, ok bool, err error) {
	// Except where nil is the field's meaning rather than the absence of one. ArchivedAt is a
	// tombstone: a row with one set is deleted, and every method that walks a collection skips
	// it. A fake that filled it would be an example of a deleted entity, which is not what any
	// caller asking for a fake wants — Webhook.EventTypes() returns nothing for one, and the
	// tests that noticed were the lucky ones.
	if strings.HasSuffix(name, "ArchivedAt") {
		return exprUnset, true, nil
	}

	// A builder already returns a pointer, so wrapping it would produce a **Entity.
	if builder, found := d.builders[elem]; found && !isTime(elem) {
		return builder + "()", true, nil
	}

	// pointer.To needs an addressable value, and a nil slice or map behind a pointer is a
	// shape no domain here uses anyway.
	switch elem.Kind() {
	case reflect.Slice, reflect.Map, reflect.Interface, reflect.Pointer:
		return exprUnset, true, nil
	default:
	}

	inner, ok, err := d.exprFor(name, elem)
	if err != nil {
		return "", false, err
	}

	if !ok {
		// An entity from another domain, whose builder lives in a fakes package this one
		// does not import. Optional is the one place a field with no default is not an
		// error, because nil is a value the type genuinely allows.
		return exprUnset, true, nil
	}

	return fmt.Sprintf("pointer.To(%s)", inner), true, nil
}

// enumExpr is the default for a named string type.
//
// Every member is offered rather than the first, because a fake that always picks the same member
// is a fake that lets a switch missing every other member pass its tests.
func (d *defaulter) enumExpr(t reflect.Type) (expression string, ok bool, err error) {
	constants, err := d.enums.constants(t.PkgPath(), t.Name())
	if err != nil {
		return "", false, err
	}

	if len(constants) == 0 {
		return "", false, nil
	}

	alias := d.qualifierFor(t.PkgPath())

	qualified := make([]string, 0, len(constants))
	for _, name := range constants {
		qualified = append(qualified, alias+"."+name)
	}

	if len(qualified) == 1 {
		return qualified[0], true, nil
	}

	return fmt.Sprintf("pickOne(%s)", strings.Join(qualified, ", ")), true, nil
}

// stringExpr is the default for an unnamed string field.
//
// The name carries the type here in a way the type system does not: `BelongsToAccount` and
// `Description` are both strings, but only one of them is rejected for not being an identifier.
func stringExpr(name string) string {
	// `...User` covers BelongsToUser, CreatedByUser and ByUser, which hold a user's ID and
	// are as much identifiers as anything ending in ID.
	if name == "ID" || strings.HasSuffix(name, "ID") || strings.HasSuffix(name, "User") || strings.HasPrefix(name, "BelongsTo") {
		return exprID
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

	return fmt.Sprintf("%s.%s", d.qualifierFor(t.PkgPath()), t.Name())
}

// qualifierFor is the identifier the generated file refers to a package by.
func (d *defaulter) qualifierFor(pkgPath string) string {
	if pkgPath == d.domainPath {
		return domainAlias
	}

	return defaultAlias(pkgPath)
}

// isNamed reports whether t is a defined type declared in some package, as opposed to a
// predeclared type like string or a composite like []byte.
func isNamed(t reflect.Type) bool {
	return t.Name() != "" && t.PkgPath() != ""
}

// isTime reports whether t is time.Time.
func isTime(t reflect.Type) bool {
	return t.PkgPath() == packageTime && t.Name() == "Time"
}

// isDuration reports whether t is time.Duration.
func isDuration(t reflect.Type) bool {
	return t.PkgPath() == packageTime && t.Name() == "Duration"
}
