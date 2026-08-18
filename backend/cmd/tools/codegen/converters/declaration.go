package main

// A conversion between two shapes of the same entity is almost always a field copy, and the
// interesting part is the handful of fields where it is not. The declaration below is built around
// that ratio: a conversion names its source and destination and nothing else, and every field that
// does not copy across by name carries a rule saying what happens instead — and, where the answer
// is not obvious, why.
//
// The property that makes this worth generating is the one a reviewer cannot supply: a destination
// field with no rule and no same-named source field is a build failure, not a zero value. A
// hand-written converter forgets a field silently and the symptom arrives much later, as a column
// that is empty for reasons nobody can reconstruct.

// Conversion declares one Convert* function.
type Conversion struct {

	// Fields carries a rule per destination field that does not copy across by name. A field
	// absent from this map is copied from the source field of the same name, and it is an
	// error for no such field to exist.
	Fields map[string]Rule

	// From is the parameter the fields are read from.
	From Param

	// Name is the generated function's name. It is spelled out rather than derived from the
	// two type names because the corpus this replaced did not derive it consistently —
	// ConvertAccountCreationInputToAccountDatabaseCreationInput takes an
	// AccountCreationRequestInput — and every one of these names is load-bearing at a call
	// site.
	Name string

	// Doc is the function's doc comment, minus the leading "Name ". It is carried in the
	// declaration rather than generated from the type names because several of them say
	// something a generator could not know: which fields the caller still has to set, what the
	// conversion loses.
	Doc string

	// To is the destination type, unqualified.
	To string

	// Extra are further parameters, in order, after From. They are only reachable from an
	// Expr rule — nothing is copied from them by name, since a parameter named userID has no
	// fields to match against.
	Extra []Param
}

// Param is one parameter of a generated function.
type Param struct {
	// Name is the parameter's identifier, which every rule's expression is written in terms of.
	Name string

	// Type is the parameter's type. An unqualified name is looked up in the domain package and
	// rendered as a pointer to it, which is what every conversion source is; anything else —
	// "string", "uint16" — is rendered as written.
	Type string
}

// ruleKind is what a field rule does.
type ruleKind uint8

const (
	// ruleCopy is the default and never appears in a declaration: it is what a field gets
	// when no rule is declared for it.
	ruleCopy ruleKind = iota
	// ruleRef takes the address of the source field, for a destination that holds a pointer
	// to what the source holds by value. This is also derived automatically, and declared
	// only when the source field has a different name.
	ruleRef
	// ruleDeref reads through a source pointer. It is never derived: a nil source field makes
	// it a panic, so it is a decision a human makes and signs with a reason.
	ruleDeref
	// ruleNestedID reads the ID of a nested entity — the shape a request input takes when the
	// full object the source carries is more than the destination needs.
	ruleNestedID
	// ruleOptionalNestedID is the same read, guarded on the nested entity being there.
	ruleOptionalNestedID
	// ruleNewID mints an identifier.
	ruleNewID
	// ruleMapSlice converts a slice of entities through another converter.
	ruleMapSlice
	// ruleExpr is an arbitrary expression, the escape hatch that keeps the rest of the
	// vocabulary small.
	ruleExpr
	// ruleSkip leaves a destination field at its zero value, deliberately.
	ruleSkip
)

// Rule is what happens to one destination field.
//
// It is a struct rather than an interface because the set is closed and small, and because a
// closed set is what lets the planner report "no rule and no matching field" as the single error
// it is.
type Rule struct {

	// sourceField names the field on From this rule reads, when that is not the destination
	// field's own name.
	sourceField string

	// converter is the function a ruleMapSlice element is passed through.
	converter string

	// belongsTo is the field on each converted element that is set to the destination's own
	// ID. Setting it here rather than leaving it to the caller is what makes a parent's
	// identifier reach its children in one pass.
	belongsTo string

	// expr is the rendered expression for ruleExpr.
	expr string

	// why is the reason this field is not a copy, rendered into the generated source above the
	// field. A rule whose reason is obvious from its kind leaves it empty.
	why string

	// callArgs is the element converter's full argument list, written with the element
	// itself as "item". It is the whole list rather than the extra arguments because two
	// converters take the element second, and a rule that could only append would have to
	// leave those hand-written.
	callArgs []string

	kind ruleKind

	// nilWhenEmpty leaves the destination slice nil rather than empty when the source has no
	// elements.
	nilWhenEmpty bool
}

// Expr assigns an arbitrary expression, written in terms of the parameter names.
//
// why is rendered as a comment above the field. It is a required argument rather than an optional
// one because this rule is the one that can encode anything, and an expression nobody can explain
// later is the thing the vocabulary above exists to avoid.
func Expr(expr, why string) Rule {
	return Rule{kind: ruleExpr, expr: expr, why: why}
}

// Ref takes the address of a differently named source field.
func Ref(sourceField string) Rule {
	return Rule{kind: ruleRef, sourceField: sourceField}
}

// Deref reads through a source pointer, which panics if it is nil.
func Deref(why string) Rule {
	return Rule{kind: ruleDeref, why: why}
}

// Rename copies from a differently named source field.
func Rename(sourceField string) Rule {
	return Rule{kind: ruleCopy, sourceField: sourceField}
}

// NestedID reads the ID of a nested entity the source carries in full.
//
// It reads straight through, which is a claim about the source: that the relation is always
// loaded on an entity that reaches this converter. Where that is not true, say so with
// OptionalNestedID — the difference between the two is a nil dereference.
func NestedID(sourceField string) Rule {
	return Rule{kind: ruleNestedID, sourceField: sourceField}
}

// OptionalNestedID reads the ID of a nested entity that may not be loaded, leaving the destination
// field nil when it is not.
//
// It needs a pointer destination, since a value one has nowhere to put the absence.
func OptionalNestedID(sourceField string) Rule {
	return Rule{kind: ruleOptionalNestedID, sourceField: sourceField}
}

// NewID mints an identifier for a row that does not have one yet.
func NewID() Rule {
	return Rule{kind: ruleNewID}
}

// MapSlice converts every element of a source slice through another converter.
func MapSlice(converter string, opts ...MapSliceOption) Rule {
	rule := Rule{kind: ruleMapSlice, converter: converter, callArgs: []string{elementPlaceholder}}
	for _, opt := range opts {
		opt(&rule)
	}

	return rule
}

// MapSliceOption varies what MapSlice emits.
type MapSliceOption func(*Rule)

// FromField reads a differently named source slice.
func FromField(sourceField string) MapSliceOption {
	return func(r *Rule) { r.sourceField = sourceField }
}

// WithArgs replaces the element converter's argument list. Write the element itself as "item",
// which is what the generated loop names it.
func WithArgs(args ...string) MapSliceOption {
	return func(r *Rule) { r.callArgs = args }
}

// BelongsTo sets the named field on every converted element to the destination's own ID.
//
// This is what forces the destination's identifier to be known before its children are built, so
// declaring it also decides that the ID is minted into a local rather than inline.
func BelongsTo(field string) MapSliceOption {
	return func(r *Rule) { r.belongsTo = field }
}

// NilWhenEmpty leaves the destination slice nil when the source has none, rather than allocating
// an empty one.
//
// The two are not interchangeable: a nil slice marshals to null and an empty one to [], and they
// compare unequal. Which one a conversion produces is therefore declared rather than left to
// whichever loop shape happened to be written.
func NilWhenEmpty() MapSliceOption {
	return func(r *Rule) { r.nilWhenEmpty = true }
}

// Because explains a mapping that is not self-evident.
func Because(why string) MapSliceOption {
	return func(r *Rule) { r.why = why }
}

// Skip leaves a destination field at its zero value.
//
// Every destination field has to be accounted for, so a field that is genuinely meant to stay
// empty says so here. The reason is required: "the caller sets this" and "this is never populated"
// are different facts, and only one of them is a bug when it turns out to be false.
func Skip(why string) Rule {
	return Rule{kind: ruleSkip, why: why}
}

// elementPlaceholder is the name the generated loop gives the element being converted, and so the
// name a declared argument list writes it as.
const elementPlaceholder = "item"
