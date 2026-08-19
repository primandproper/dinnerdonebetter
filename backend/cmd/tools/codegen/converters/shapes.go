package main

import (
	"fmt"
	"slices"
	"strings"
)

// The domain packages name an entity's shapes by suffixing the entity: a Webhook is created from a
// WebhookCreationRequestInput, which becomes a WebhookDatabaseCreationInput on the way to the
// repository, and is amended by a WebhookUpdateRequestInput. That convention is not a coincidence
// to be transcribed once per entity — it is the whole specification, and everything below reads it
// rather than restating it.

// entityShape is one suffix an entity's types are named with. The empty suffix is the entity.
type entityShape string

const (
	entity                = entityShape("")
	creationRequestInput  = entityShape("CreationRequestInput")
	databaseCreationInput = entityShape("DatabaseCreationInput")
	updateRequestInput    = entityShape("UpdateRequestInput")
)

// conversionShapes are the conversions that exist for every entity that has the types for them.
//
// Four rules, in the order the generated file lists them. They cover 180 of the 198 conversions
// this replaced; the rest are exceptions, and exceptions.go is where they live. Adding a fifth
// rule here adds it for every entity at once, which is the property the old hand-written corpus
// did not have — six entities had an UpdateRequestInput converter and thirty-eight did not, for no
// reason anybody recorded.
var conversionShapes = []struct {
	From entityShape
	To   entityShape
}{
	{From: entity, To: creationRequestInput},
	{From: entity, To: databaseCreationInput},
	{From: entity, To: updateRequestInput},
	{From: creationRequestInput, To: databaseCreationInput},
}

// allShapes is every suffix the convention knows, longest first so that a type is matched against
// the most specific one it ends with.
var allShapes = []entityShape{creationRequestInput, databaseCreationInput, updateRequestInput}

// typeName is the type an entity takes in a given shape.
func (s entityShape) typeName(entityName string) string {
	return entityName + string(s)
}

// entityOf returns the entity a type name belongs to, and whether the name is an entity itself
// rather than one of its shapes.
//
// A type that ends in no known suffix is an entity. That is what makes the enumeration total: the
// set of entities is not a list somebody maintains, it is every struct that is not the shape of
// another one.
func entityOf(typeName string) (string, bool) {
	for _, shape := range allShapes {
		if base, found := strings.CutSuffix(typeName, string(shape)); found && base != "" {
			return base, false
		}
	}

	return typeName, true
}

// conversionName is what a conversion between two types is called.
//
// Derived rather than declared, because 179 of the 188 conversions this replaced already spelled
// it this way and the nine that did not were abbreviations of their own argument's type —
// ConvertAccountCreationRequestInputToAccountDatabaseCreationInput takes an AccountCreationRequestInput.
// A name that is computed cannot drift from the thing it names.
func conversionName(from, to string) string {
	return fmt.Sprintf("Convert%sTo%s", from, to)
}

// sourceParam is what every generated conversion calls its argument.
//
// One name, always. The corpus this replaced used x, input, webhook, mealPlan, cfg, client and
// recipeStepProduct, and no reader ever needed to know which: the parameter's type is in the
// signature directly above it, and its name carries nothing the type does not.
const sourceParam = "x"

// converterFor names the conversion an element of a nested collection goes through.
//
// It does not check that the conversion exists, because the enumeration that emits this one emits
// that one too: every entity's shapes are generated together, so a slice of Child in a
// ChildInput-shaped destination has a ConvertChildToChildInput by construction. Where that is not
// true the generated file does not compile, which is the right place to find out.
func converterFor(fromElement, toElement string) string {
	return conversionName(fromElement, toElement)
}

// sortConversions puts a domain's conversions in a stable, readable order: by entity, then by the
// order the shape rules are declared in.
func sortConversions(conversions []*Conversion) {
	slices.SortFunc(conversions, func(a, b *Conversion) int {
		if a.Entity != b.Entity {
			return strings.Compare(a.Entity, b.Entity)
		}

		return a.shapeIndex - b.shapeIndex
	})
}
