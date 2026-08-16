package main

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// derivation is what an entity's own field types imply beyond a one-line-per-field literal.
//
// A collection of child entities cannot be written as a field expression, because building one
// takes a loop and because the children have to agree with their parent about the parent's ID.
// That was the whole reason the declarations grew a Locals section, and it meant the same fifteen
// lines were written out once per parent — nineteen times in mealplanning alone, differing only in
// the type names, which are exactly what the generator can already see.
type derivation struct {
	// exprs are the field expressions the locals exist to supply.
	exprs map[string]string
	// locals are statements emitted above the composite literal, in order.
	locals []string
}

// derive works out an entity's child collections from its field types.
//
// overridden are the fields the declaration speaks for; derivation leaves those alone, which is
// the escape hatch for a parent whose children are not the mechanical `exampleQuantity` of them.
func derive(p *entityPlan, defaults *defaulter, children *childPlanner, overridden map[string]struct{}) (*derivation, error) {
	d := &derivation{exprs: map[string]string{}}

	parentIDVar := lowerFirst(p.typ.Name()) + "ID"
	linksToParent := false

	var loops []string

	for field := range p.typ.Fields() {
		if !field.IsExported() || field.Type.Kind() != reflect.Slice {
			continue
		}

		if _, declared := overridden[field.Name]; declared {
			continue
		}

		if !children.derives(p.typ, field.Name) {
			continue
		}

		element := field.Type.Elem()

		pointerElement := element.Kind() == reflect.Pointer
		if pointerElement {
			element = element.Elem()
		}

		builder, isEntity := defaults.builders[element]
		if !isEntity {
			continue
		}

		link := parentLinkField(element, p.typ)
		if link != "" {
			linksToParent = true
		}

		loops = append(loops, childLoop(&field, element, builder, pointerElement, link, parentIDVar))
		d.exprs[field.Name] = lowerFirst(field.Name)
	}

	if linksToParent {
		if _, declared := overridden["ID"]; declared {
			return nil, fmt.Errorf(
				"%s: its children are given %s.ID, so the generator has to name it, but the declaration overrides ID. "+
					"Drop the ID override, or override the child collection instead and build the children yourself",
				p.builderName, p.typ.Name(),
			)
		}

		d.locals = append(d.locals, fmt.Sprintf("%s := BuildFakeID()", parentIDVar))
		d.exprs["ID"] = parentIDVar
	}

	d.locals = append(d.locals, loops...)

	return d, nil
}

// childLoop renders the statement that fills one child collection.
func childLoop(field *reflect.StructField, element reflect.Type, builder string, pointerElement bool, link, parentIDVar string) string {
	slice := lowerFirst(field.Name)
	qualified := domainAlias + "." + element.Name()

	if pointerElement {
		qualified = "*" + qualified
	}

	built := builder + "()"
	if !pointerElement {
		built = "*" + built
	}

	if link == "" {
		return fmt.Sprintf(`var %s []%s
for range exampleQuantity {
	%s = append(%s, %s)
}`, slice, qualified, slice, slice, built)
	}

	child := lowerFirst(element.Name())
	if child == slice {
		child += "Child"
	}

	return fmt.Sprintf(`var %s []%s
for range exampleQuantity {
	%s := %s
	%s.%s = %s
	%s = append(%s, %s)
}`, slice, qualified, child, built, child, link, parentIDVar, slice, slice, child)
}

// parentLinkField is the child's field holding its parent's ID, or empty when it has none.
//
// A child that carries `BelongsToMealPlan` and a parent named `MealPlan` are talking about each
// other, and a fake that leaves them disagreeing is a fake that a repository test will reject for
// reasons that have nothing to do with what it was testing.
func parentLinkField(child, parent reflect.Type) string {
	name := "BelongsTo" + parent.Name()

	field, found := child.FieldByName(name)
	if !found || field.Type.Kind() != reflect.String {
		return ""
	}

	if _, hasID := parent.FieldByName("ID"); !hasID {
		return ""
	}

	return name
}

// childPlanner decides which child collections a domain can fill without recursing forever.
//
// A filled collection makes a fake call other fakes several levels deep, and a cycle in that graph
// is a stack overflow in every test that touches either end of it. Recipe.AssociatedRecipes is one:
// a recipe holds recipes, and a fake that filled it would still be building recipes now.
//
// A collection that closes a cycle is left nil rather than rejected. Nil is a shape the field
// genuinely has — a recipe with no associated recipes is a recipe — whereas refusing to generate
// would make every self-referential type the author's problem for no gain. Cycles that do not run
// through a collection are a different matter: a required field cannot be left unset, so those are
// still an error.
type childPlanner struct {
	// edges is the graph of which entity's fake builds which, after the decisions below.
	edges map[reflect.Type]map[reflect.Type]bool
	// filled is, per entity, the collection fields that are safe to fill.
	filled map[reflect.Type]map[string]bool
}

// derives reports whether a collection field should be filled with children.
func (c *childPlanner) derives(parent reflect.Type, field string) bool {
	return c.filled[parent][field]
}

// planChildren works out the whole domain's child graph before any of it is rendered.
func planChildren(plans []entityPlan, defaults *defaulter) (*childPlanner, error) {
	c := &childPlanner{
		edges:  map[reflect.Type]map[reflect.Type]bool{},
		filled: map[reflect.Type]map[string]bool{},
	}

	type candidate struct {
		parent reflect.Type
		child  reflect.Type
		field  string
	}

	var collections []candidate

	// A field that is not a collection is required, so its edge is not negotiable and goes
	// in first. Collections are then offered the edges that are left.
	for _, p := range plans {
		if p.fake.Bespoke {
			// A hand-written builder's calls are its own business; the generator
			// cannot see them and does not get to rule on them.
			continue
		}

		overridden := map[string]struct{}{}
		for _, f := range p.fake.Fields {
			overridden[f.Name] = struct{}{}
		}

		for field := range p.typ.Fields() {
			if !field.IsExported() {
				continue
			}

			if _, declared := overridden[field.Name]; declared {
				continue
			}

			referenced := elementType(field.Type)
			if _, isEntity := defaults.builders[referenced]; !isEntity {
				continue
			}

			if field.Type.Kind() == reflect.Slice {
				collections = append(collections, candidate{parent: p.typ, child: referenced, field: field.Name})

				continue
			}

			c.connect(p.typ, referenced)
		}
	}

	if cycle := c.findCycle(plans); cycle != "" {
		return nil, fmt.Errorf("fakes would build each other forever: %s. Override one of the fields to break the cycle", cycle)
	}

	for _, collection := range collections {
		if c.reaches(collection.child, collection.parent) {
			continue
		}

		c.connect(collection.parent, collection.child)

		if c.filled[collection.parent] == nil {
			c.filled[collection.parent] = map[string]bool{}
		}

		c.filled[collection.parent][collection.field] = true
	}

	return c, nil
}

// connect records that one entity's fake builds another's.
func (c *childPlanner) connect(from, to reflect.Type) {
	if c.edges[from] == nil {
		c.edges[from] = map[reflect.Type]bool{}
	}

	c.edges[from][to] = true
}

// reaches reports whether from's fake would, directly or eventually, build to.
func (c *childPlanner) reaches(from, to reflect.Type) bool {
	if from == to {
		return true
	}

	seen := map[reflect.Type]bool{}

	var walk func(t reflect.Type) bool

	walk = func(t reflect.Type) bool {
		if seen[t] {
			return false
		}

		seen[t] = true

		for next := range c.edges[t] {
			if next == to || walk(next) {
				return true
			}
		}

		return false
	}

	return walk(from)
}

// findCycle returns a readable path through a cycle, or empty when the graph is acyclic.
func (c *childPlanner) findCycle(plans []entityPlan) string {
	visiting := map[reflect.Type]bool{}
	done := map[reflect.Type]bool{}

	var walk func(t reflect.Type, path []string) string

	walk = func(t reflect.Type, path []string) string {
		if visiting[t] {
			return strings.Join(append(path, t.Name()), " -> ")
		}

		if done[t] {
			return ""
		}

		visiting[t] = true

		for next := range c.edges[t] {
			if found := walk(next, append(path, t.Name())); found != "" {
				return found
			}
		}

		visiting[t] = false
		done[t] = true

		return ""
	}

	for _, p := range plans {
		if found := walk(p.typ, nil); found != "" {
			return found
		}
	}

	return ""
}

// elementType unwraps a pointer or a slice to the type a builder would be looked up by.
func elementType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	return t
}

// lowerFirst renders a type or field name as a local variable.
func lowerFirst(name string) string {
	if name == "" {
		return name
	}

	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}
