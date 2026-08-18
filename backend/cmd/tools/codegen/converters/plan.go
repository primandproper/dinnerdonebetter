package main

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/types"
	"strings"
	"unicode"

	"github.com/primandproper/platform-go/v11/errors"
)

// plan is one conversion resolved against the types it reads and writes: every destination field
// accounted for, every local it needs decided.
//
// Resolving and rendering are separate because the interesting failure — a destination field that
// nothing fills — is a property of the resolution, and reporting it from here means it is reported
// once per field with the field's name in it, rather than as a compile error in generated source.
type plan struct {
	Conversion *Conversion

	// Prelude holds the statements that run before the literal: the slice conversions and the
	// guarded reads that cannot be written as an expression.
	Prelude []string

	// Assignments are the literal's fields, in destination declaration order.
	Assignments []assignment

	// UsesIdentifiers reports whether the rendered function mints an ID, which is what decides
	// the generated file's imports.
	UsesIdentifiers bool
}

// assignment is one field of the generated struct literal.
type assignment struct {
	Field string
	Expr  string
	// Why is rendered above the field. It carries the reason a field is not a plain copy into
	// the generated source, so the reasoning survives where a reader of the output meets it.
	Why string
	// Skipped marks a field the literal deliberately leaves out. It is still carried through
	// as an assignment so that its reason can be rendered in the field's place: a reader of
	// the generated source needs to see that the omission was decided, and an omission that
	// leaves no trace is indistinguishable from one nobody thought about.
	Skipped bool
}

// planner resolves conversions for one domain package.
type planner struct {
	index *structIndex
	// domain is the directory name of the domain package, which is also how the index is keyed.
	domain string
	// pkg is the name the generated file imports the domain package under, and so the
	// qualifier every domain type name needs on the way out.
	pkg string
}

// Plan resolves one conversion.
func (p *planner) Plan(conversion *Conversion) (*plan, error) {
	destination, err := p.index.Lookup(p.domain, conversion.To)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving destination of %s", conversion.Name)
	}

	source, err := p.index.Lookup(p.domain, conversion.From.Type)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving source of %s", conversion.Name)
	}

	if err = p.checkRulesAreReachable(conversion, destination); err != nil {
		return nil, err
	}

	resolved := &plan{Conversion: conversion}

	// An ID the children have to carry has to be readable before the children are built, so
	// what they read is decided here rather than declared. A minted one is hoisted into a
	// local; a copied one is already an expression both can use.
	parentID, err := p.parentIDFor(conversion, source, destination)
	if err != nil {
		return nil, err
	}

	idLocal := ""
	if parentID.hoist {
		idLocal = parentID.expr
	}

	for _, field := range destination.Fields {
		rule, declared := conversion.Fields[field.Name]
		if !declared {
			if rule, err = p.deriveRule(conversion, source, field); err != nil {
				return nil, err
			}
		}

		assigned, applyErr := p.applyRule(conversion, source, field, &rule, resolved, parentID)
		if applyErr != nil {
			return nil, errors.Wrapf(applyErr, "%s: field %s", conversion.Name, field.Name)
		}

		if assigned != nil {
			resolved.Assignments = append(resolved.Assignments, *assigned)
		}
	}

	if parentID.hoist {
		resolved.Prelude = append([]string{fmt.Sprintf("%s := identifiers.New()", idLocal)}, resolved.Prelude...)
		resolved.UsesIdentifiers = true
	}

	return resolved, nil
}

// parentIDRead is what a converted child reads to learn the ID of the entity it belongs to.
type parentIDRead struct {
	// expr is the expression the children read. It is a local when hoist is set, and the
	// destination ID's own expression otherwise.
	expr string
	// hoist reports whether expr names a local this plan has to declare.
	hoist bool
}

// parentIDFor decides what a BelongsTo child reads.
//
// A destination that mints its ID has to mint it once, before the children, so the same value
// reaches both. A destination that copies one already has an expression the children can read
// directly. Anything else — an ID computed from a guarded read, or from another prelude local —
// would need an ordering this does not establish, so it is refused rather than silently reading a
// different value than the destination stores.
func (p *planner) parentIDFor(conversion *Conversion, source, destination *structType) (parentIDRead, error) {
	if !p.destinationIDIsSharedWithChildren(conversion) {
		return parentIDRead{}, nil
	}

	field, ok := destination.Field("ID")
	if !ok {
		return parentIDRead{}, errors.Newf("%s: BelongsTo needs %s to have an ID", conversion.Name, conversion.To)
	}

	rule, declared := conversion.Fields["ID"]
	if !declared {
		derived, err := p.deriveRule(conversion, source, field)
		if err != nil {
			return parentIDRead{}, err
		}

		rule = derived
	}

	switch rule.kind {
	case ruleNewID:
		return parentIDRead{expr: entityLocalName(conversion.To) + "ID", hoist: true}, nil
	case ruleCopy:
		sourceName := rule.sourceField
		if sourceName == "" {
			sourceName = "ID"
		}

		return parentIDRead{expr: conversion.From.Name + "." + sourceName}, nil
	default:
		return parentIDRead{}, errors.Newf(
			"%s: BelongsTo reads %s.ID, which is neither copied nor minted", conversion.Name, conversion.To,
		)
	}
}

// checkRulesAreReachable refuses a rule for a field the destination does not have.
//
// A rule that names nothing is the failure mode this whole exercise is meant to remove, one step
// removed: a field renamed on the destination leaves its override behind, and without this the
// override is silently ignored and the field silently becomes a plain copy.
func (p *planner) checkRulesAreReachable(conversion *Conversion, destination *structType) error {
	for name := range conversion.Fields {
		if _, ok := destination.Field(name); !ok {
			return errors.Newf("%s: rule for %s, which %s does not have", conversion.Name, name, conversion.To)
		}
	}

	return nil
}

// destinationIDIsSharedWithChildren reports whether any converted child is stamped with the
// destination's own ID.
func (p *planner) destinationIDIsSharedWithChildren(conversion *Conversion) bool {
	for _, rule := range conversion.Fields {
		if rule.kind == ruleMapSlice && rule.belongsTo != "" {
			return true
		}
	}

	return false
}

// deriveRule is what happens to a destination field nobody declared a rule for.
//
// Two shapes are derived and no more: the same type copies, and a value becomes a pointer to
// itself. Everything else is an error naming both types, because every other adaptation — reading
// through a pointer, reaching into a nested entity, converting a slice — is a decision with a
// wrong answer, and a generator that guesses is worse than the hand-written code it replaced.
func (p *planner) deriveRule(conversion *Conversion, source *structType, field structField) (Rule, error) {
	sourceField, ok := source.Field(field.Name)
	if !ok {
		return Rule{}, errors.Newf(
			"%s: %s.%s has no counterpart on %s and no rule; declare one, or Skip it with a reason",
			conversion.Name, conversion.To, field.Name, conversion.From.Type,
		)
	}

	switch {
	case sourceField.Type == field.Type:
		return Rule{kind: ruleCopy}, nil
	case field.Type == "*"+sourceField.Type:
		return Rule{kind: ruleRef}, nil
	default:
		return Rule{}, errors.Newf(
			"%s: %s.%s is %s but %s.%s is %s; declare a rule",
			conversion.Name, conversion.To, field.Name, field.Type,
			conversion.From.Type, sourceField.Name, sourceField.Type,
		)
	}
}

// applyRule renders one rule, appending to the plan's prelude where the rule needs a local. It
// returns nil for a field that is deliberately left unassigned.
func (p *planner) applyRule(
	conversion *Conversion,
	source *structType,
	field structField,
	rule *Rule,
	resolved *plan,
	parentID parentIDRead,
) (*assignment, error) {
	from := conversion.From.Name

	sourceName := rule.sourceField
	if sourceName == "" {
		sourceName = field.Name
	}

	switch rule.kind {
	case ruleSkip:
		return &assignment{Field: field.Name, Why: rule.why, Skipped: true}, nil
	case ruleExpr:
		return &assignment{Field: field.Name, Expr: rule.expr, Why: rule.why}, nil
	case ruleNewID:
		if parentID.hoist && field.Name == "ID" {
			return &assignment{Field: field.Name, Expr: parentID.expr, Why: rule.why}, nil
		}

		resolved.UsesIdentifiers = true

		return &assignment{Field: field.Name, Expr: "identifiers.New()", Why: rule.why}, nil
	case ruleCopy, ruleRef, ruleDeref:
		if _, ok := source.Field(sourceName); !ok {
			return nil, errors.Newf("%s has no field %s", conversion.From.Type, sourceName)
		}

		return &assignment{
			Field: field.Name,
			Expr:  readOperator(rule.kind) + from + "." + sourceName,
			Why:   rule.why,
		}, nil
	case ruleNestedID, ruleOptionalNestedID:
		return p.applyNestedID(conversion, source, field, rule, resolved)
	case ruleMapSlice:
		return p.applyMapSlice(conversion, source, field, rule, resolved, parentID)
	default:
		return nil, errors.Newf("unhandled rule kind %d", rule.kind)
	}
}

// readOperator is what a copy, an address-of and a dereference put in front of the source field.
func readOperator(kind ruleKind) string {
	switch kind {
	case ruleRef:
		return "&"
	case ruleDeref:
		return "*"
	default:
		return ""
	}
}

// applyNestedID reads the identifier of an entity the source carries in full.
//
// Whether the read is guarded is declared rather than derived from the destination's pointer-ness,
// because both spellings appear in the corpus this replaced against the same shapes, and picking
// one would be changing behavior rather than generating it. A guarded read needs a local, since a
// conditional has nowhere to live inside a literal.
func (p *planner) applyNestedID(
	conversion *Conversion,
	source *structType,
	field structField,
	rule *Rule,
	resolved *plan,
) (*assignment, error) {
	nested, ok := source.Field(rule.sourceField)
	if !ok {
		return nil, errors.Newf("%s has no field %s", conversion.From.Type, rule.sourceField)
	}

	read := conversion.From.Name + "." + nested.Name + ".ID"

	pointerDestination := strings.HasPrefix(field.Type, "*")

	if rule.kind == ruleNestedID {
		if !pointerDestination {
			return &assignment{Field: field.Name, Expr: read, Why: rule.why}, nil
		}

		return &assignment{Field: field.Name, Expr: "&" + read, Why: rule.why}, nil
	}

	if !pointerDestination {
		return nil, errors.Newf(
			"OptionalNestedID needs a pointer destination, and %s.%s is %s",
			conversion.To, field.Name, field.Type,
		)
	}

	local := localName(field.Name)
	resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
		"var %s %s\nif %s.%s != nil {\n%s = &%s\n}",
		local, field.Type, conversion.From.Name, nested.Name, local, read,
	))

	return &assignment{Field: field.Name, Expr: local, Why: rule.why}, nil
}

// applyMapSlice converts a slice of entities element by element.
func (p *planner) applyMapSlice(
	conversion *Conversion,
	source *structType,
	field structField,
	rule *Rule,
	resolved *plan,
	parentID parentIDRead,
) (*assignment, error) {
	sourceName := rule.sourceField
	if sourceName == "" {
		sourceName = field.Name
	}

	if _, ok := source.Field(sourceName); !ok {
		return nil, errors.Newf("%s has no field %s", conversion.From.Type, sourceName)
	}

	sliceType, err := p.qualify(field.Type)
	if err != nil {
		return nil, err
	}

	local := localName(field.Name)
	read := conversion.From.Name + "." + sourceName
	args := strings.Join(rule.callArgs, ", ")

	var body string
	if rule.belongsTo == "" {
		body = fmt.Sprintf("%s = append(%s, %s(%s))", local, local, rule.converter, args)
	} else {
		body = fmt.Sprintf(
			"converted := %s(%s)\nconverted.%s = %s\n%s = append(%s, converted)",
			rule.converter, args, rule.belongsTo, parentID.expr, local, local,
		)
	}

	declaration := fmt.Sprintf("%s := make(%s, 0, len(%s))", local, sliceType, read)
	if rule.nilWhenEmpty {
		declaration = fmt.Sprintf("var %s %s", local, sliceType)
	}

	resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
		"%s\nfor _, %s := range %s {\n%s\n}",
		declaration, elementPlaceholder, read, body,
	))

	return &assignment{Field: field.Name, Expr: local, Why: rule.why}, nil
}

// qualify rewrites a type as the domain package writes it into a type the generated file can
// name, by prefixing every identifier that names a type of that domain.
func (p *planner) qualify(typeExpr string) (string, error) {
	parsed, err := parser.ParseExpr(typeExpr)
	if err != nil {
		return "", errors.Wrapf(err, "parsing type %q", typeExpr)
	}

	goast.Inspect(parsed, func(node goast.Node) bool {
		switch typed := node.(type) {
		case *goast.SelectorExpr:
			// Already qualified by another package; neither half is ours to rewrite.
			return false
		case *goast.Ident:
			if p.index.Declares(p.domain, typed.Name) {
				typed.Name = p.pkg + "." + typed.Name
			}

			return false
		default:
			return true
		}
	})

	return types.ExprString(parsed), nil
}

// localName is the variable a field's value is computed into.
func localName(field string) string {
	return lowerFirstWord(field)
}

// entityLocalName strips the input-shape suffix from a type name, so the identifier local of a
// WebhookDatabaseCreationInput reads webhookID rather than webhookDatabaseCreationInputID.
func entityLocalName(typeName string) string {
	for _, suffix := range []string{
		"DatabaseCreationInput",
		"CreationRequestInput",
		"CreationInput",
		"UpdateRequestInput",
	} {
		if trimmed, found := strings.CutSuffix(typeName, suffix); found {
			typeName = trimmed

			break
		}
	}

	return lowerFirstWord(typeName)
}

// lowerFirstWord lowercases the leading word of an identifier, keeping an initialism whole: ID
// becomes id, URLPath becomes urlPath, TriggerConfigs becomes triggerConfigs.
func lowerFirstWord(name string) string {
	if name == "" {
		return name
	}

	runes := []rune(name)

	upper := 0
	for upper < len(runes) && unicode.IsUpper(runes[upper]) {
		upper++
	}

	switch upper {
	case 0:
		return name
	case 1:
		runes[0] = unicode.ToLower(runes[0])
	case len(runes):
		for i := range runes {
			runes[i] = unicode.ToLower(runes[i])
		}
	default:
		// The last capital of a run starts the next word: URLPath is url + Path.
		for i := range upper - 1 {
			runes[i] = unicode.ToLower(runes[i])
		}
	}

	return string(runes)
}
