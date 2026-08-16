/*
Package entitydecl is the vocabulary a domain uses to describe its own entities to code
generators.

A domain package declares `var Entities` beside the types it declares, and generators read it
rather than guessing. Today the only reader is cmd/tools/codegen/fakes; the shape is meant to
grow a section per generator rather than a declaration per generator, so that a type gains its
fake, its converters and whatever comes next from one place.

# Why the declaration lives beside the type

The alternative is a table inside the generator, which is where this repository keeps its query
declarations. That works when the declaration is about the generator's output — a SQL query is
not a property of the Go type. A fake is: it says what a valid instance of this type looks like,
which is knowledge the type's author has and the generator does not. Putting it next to the type
is what makes it maintainable by whoever adds a field.

# Why expressions are strings

An override carries Go source, not a Go value:

	{Name: "URL", Expr: `"https://192.0.2.1/webhook"`}

A value would have to be constructed here, in the domain package, and the useful ones cannot be:
they call into the fakes package's own helpers (BuildFakeID), or into converters, which import
the domain package and so cannot be imported by it. Source text has neither problem, is what the
generator needs to emit anyway, and keeps this package free of every dependency but the standard
library.

The cost is that a typo in an expression is a compile error in the generated file rather than in
this one. That is an acceptable trade for a file that `make fakes` rewrites on every run.
*/
package entitydecl

// Domain is one domain package's entity declarations.
//
// Declared as `var Entities` in the domain package, e.g. internal/domain/webhooks/entities.go.
type Domain struct {
	// Entities is every entity the domain declares, in the order generators emit them.
	//
	// Order is the declaration's, not the type system's, because generated files are read by
	// people: keeping related types adjacent is worth more than an alphabetical sort nobody
	// asked for.
	Entities []Entity
}

// Entity is one type, plus what generators need to know about it beyond its fields.
type Entity struct {
	// Type is a zero value of the entity type — `Webhook{}`, not `&Webhook{}` and not a
	// reflect.Type. Generators reflect over it for the field list, so the fields themselves
	// never appear here and adding one to the struct needs no edit in this file unless the
	// mechanical default is wrong for it.
	Type any

	// Fake declares how to build a fake of Type. The zero value asks for a builder with
	// nothing but mechanical defaults, which is right for most types.
	Fake Fake
}

// Fake is the fake-builder declaration for one entity.
//
// The field order is the one `make format` settles on, not a reading order; start at Fields.
type Fake struct {
	// List asks for a BuildFake<Type>sList builder over filtering.QueryFilteredResult.
	List *List

	// BespokeWhy is why the builder is Bespoke, and belongs beside every Bespoke. It is not
	// emitted anywhere; it is for the next person deciding whether the opt-out still holds.
	BespokeWhy string

	// Doc is the builder's doc comment sentence, without the leading function name.
	// Empty means "builds a faked <type>.".
	Doc string

	// Locals are statements emitted above the composite literal, in order, verbatim.
	//
	// They exist for fields that cannot be independent: a min/max pair where the max must
	// exceed the min, a parent ID that several children have to agree on.
	Locals []Local

	// Fields overrides the mechanical default for named fields. A field not named here gets
	// the default for its name and type; see cmd/tools/codegen/fakes.
	Fields []Field

	// Inputs are the request/database input builders derived from this entity's fake by a
	// converter.
	Inputs []Input

	// Bespoke opts the whole builder out of generation, leaving the hand-written one in
	// place. The entity stays declared so other generators still see it.
	//
	// This is for types whose fake is a procedure rather than a literal — one that consults a
	// converter, mutates what it built, or has to satisfy an invariant spanning several
	// fields. Expressing those here would mean growing this vocabulary until it was a
	// programming language, and a worse one than Go.
	Bespoke bool
}

// Local is one statement emitted above a fake's composite literal.
type Local struct {
	// Code is the statement, verbatim. No trailing semicolon or newline.
	Code string

	// Why is a comment emitted above it. Optional, but a local that needs no explanation is
	// usually a field override in disguise.
	Why string
}

// Field is a per-field override.
type Field struct {
	// Name is the struct field's name.
	Name string

	// Expr is the Go expression assigned to it, verbatim.
	Expr string

	// Why is emitted as a comment above the field in the generated source.
	//
	// This is the reason the override exists at all. An override without one reads, six
	// months later, as an arbitrary choice somebody is free to "simplify" back into a random
	// value that fails validation — which is the failure this whole mechanism exists to
	// prevent. Write the constraint down.
	Why string
}

// List declares a paginated-list builder.
type List struct {
	// Name overrides the builder's name. Empty means "BuildFake<Type>sList".
	Name string

	// Doc is the doc comment sentence. Empty means "builds a faked <type> list.".
	Doc string
}

// Input declares a builder that converts this entity's fake into an input type.
type Input struct {
	// Type is a zero value of the input type, e.g. `WebhookCreationRequestInput{}`.
	Type any

	// Converter is the unqualified name of the function in the domain's converters package
	// that produces Type from the entity, e.g.
	// "ConvertWebhookToWebhookCreationRequestInput".
	Converter string

	// Name overrides the builder's name. Empty means "BuildFake<input type>".
	Name string

	// Doc is the doc comment sentence. Empty means "builds a faked <input type>.".
	Doc string
}
