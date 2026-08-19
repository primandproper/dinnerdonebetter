package main

// Almost every field of almost every conversion is a copy, and the shapes that are not copies are
// themselves regular: an identifier minted for a row that does not have one, a nested entity
// reduced to its ID, a slice mapped through the converter for its element. All of that is derived
// from the two structs in plan.go and declared nowhere.
//
// What is left is genuinely per-field knowledge — a field the destination cannot be given a value
// for, an expression that is particular to one entity. Those are the exceptions in exceptions.go,
// and every one of them carries the reason it exists.

// Conversion is one generated function, as the enumeration produces it.
//
// Nothing here is written by hand: Entity comes from the type names, From and To from the shape
// rule applied to it, Name from the two of those. The struct exists so that planning and rendering
// have something to hold, not so that anybody fills one in.
type Conversion struct {
	// Fields carries the exceptions declared for this conversion, if any.
	Fields map[string]Rule
	// Entity is the base type both shapes belong to.
	Entity string
	// From and To are the source and destination type names, unqualified.
	From string
	To   string
	// Name is the generated function's name.
	Name string

	// shapeIndex is which of the shape rules produced this conversion, and orders the
	// generated file so that an entity's conversions read in a consistent order.
	shapeIndex int
}

// ruleKind is what happens to one destination field.
//
// Most of these are never declared: they are what the planner derives, and they are named here so
// that the planner and the renderer have one vocabulary rather than two.
type ruleKind uint8

const (
	// ruleCopy reads the source field of the same name.
	ruleCopy ruleKind = iota
	// ruleRef takes its address, for a destination that holds a pointer to what the source
	// holds by value.
	ruleRef
	// ruleDeref reads through a source pointer. Never derived: a nil source field makes it a
	// panic, so it is a decision somebody makes and signs.
	ruleDeref
	// ruleNestedID reduces a nested entity to its identifier.
	ruleNestedID
	// ruleNewID mints an identifier for a row that does not have one yet.
	ruleNewID
	// ruleMapSlice converts a slice element by element.
	ruleMapSlice
	// ruleExpr is an arbitrary expression, the escape hatch that keeps the rest small.
	ruleExpr
	// ruleSkip leaves a destination field at its zero value, deliberately.
	ruleSkip
)

// Rule is what happens to one destination field.
type Rule struct {
	// sourceField names the field on the source this rule reads, where that is not the
	// destination field's own name.
	sourceField string

	// converter and elementType describe a derived ruleMapSlice.
	converter   string
	elementType string

	// expr is the rendered expression for ruleExpr.
	expr string

	// why is the reason this field is not a copy. It is rendered into the generated source
	// above the field, so the reasoning survives where a reader of the output meets it.
	why string

	kind ruleKind

	// guarded marks a derived ruleNestedID whose destination can hold the absence, and so is
	// read behind a nil check rather than straight through.
	guarded bool
}

// Expr assigns an arbitrary expression, written in terms of the source parameter x.
func Expr(expr, why string) Rule {
	return Rule{kind: ruleExpr, expr: expr, why: why}
}

// Deref reads through a source pointer, which panics if it is nil.
func Deref(why string) Rule {
	return Rule{kind: ruleDeref, why: why}
}

// NewID mints an identifier where the derivation would have copied one.
//
// A conversion whose source already has an ID copies it, which is right for every shape that is
// the same row in another form. Where the destination is a different row, the identifier has to be
// new, and only the domain knows which of the two a conversion is.
func NewID(why string) Rule {
	return Rule{kind: ruleNewID, why: why}
}

// MapSlice converts a collection the destination names differently from the source.
//
// A collection named the same on both sides is derived and needs no rule; this is for the one case
// where it is not, and it exists so that a renamed collection is still checked for having a
// converter rather than being written out as a loop nobody reads.
func MapSlice(sourceField, converter, why string) Rule {
	return Rule{kind: ruleMapSlice, sourceField: sourceField, converter: converter, why: why}
}

// Skip leaves a destination field at its zero value.
//
// Every destination field has to be accounted for, so a field meant to stay empty says so here.
// The reason is required: "the caller sets this" and "this is not recoverable from the source" are
// different facts, and only one of them is a bug when it turns out to be false.
func Skip(why string) Rule {
	return Rule{kind: ruleSkip, why: why}
}

// elementPlaceholder is the name the generated loop gives the element being converted.
const elementPlaceholder = "item"
