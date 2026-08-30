package main

import (
	goerrors "errors"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/types"
	"strings"
	"unicode"

	"github.com/primandproper/platform-go/v13/errors"
)

// plan is one conversion resolved against the types it reads and writes: every destination field
// accounted for, every local it needs decided.
type plan struct {
	Conversion *Conversion

	// Prelude holds the statements that run before the literal — the slice conversions and the
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
	// Why is rendered above the field.
	Why string
	// Skipped marks a field the literal deliberately leaves out. It is carried through so its
	// reason can be rendered in the field's place: an omission that leaves no trace is
	// indistinguishable from one nobody thought about.
	Skipped bool
}

// planner resolves conversions for one domain package.
type planner struct {
	index *structIndex
	// domain is the directory name of the domain package, which is also the qualifier every
	// domain type name needs on the way out.
	domain string
}

// Plan resolves one conversion.
//
// The whole of it is derivation. Every destination field is answered by asking the two structs a
// question — is there a field of this name, is this a pointer to what that one holds, is this the
// identifier of something the source carries whole, is this a slice of a shape of that element —
// and a field that none of those answer is an error naming it. Nothing is guessed: an adaptation
// with a plausible wrong answer is refused and declared instead.
func (p *planner) Plan(conversion *Conversion) (*plan, error) {
	destination, err := p.index.Lookup(p.domain, conversion.To)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving destination of %s", conversion.Name)
	}

	source, err := p.index.Lookup(p.domain, conversion.From)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving source of %s", conversion.Name)
	}

	if err = p.checkExceptionsAreReachable(conversion, destination); err != nil {
		return nil, err
	}

	resolved := &plan{Conversion: conversion}

	// Every field is attempted even after one fails, so that a conversion reports the whole
	// set of decisions it needs rather than the first: a generator that stops at field one
	// turns declaring five exceptions into five edit-and-rerun cycles.
	var unanswered []error

	for _, field := range destination.Fields {
		rule, declared := conversion.Fields[field.Name]
		if !declared {
			derived, deriveErr := p.derive(conversion, source, field)
			if deriveErr != nil {
				unanswered = append(unanswered, deriveErr)

				continue
			}

			rule = derived
		}

		assigned, applyErr := p.apply(conversion, source, field, &rule, resolved)
		if applyErr != nil {
			unanswered = append(unanswered, errors.Wrapf(applyErr, "%s: field %s", conversion.Name, field.Name))

			continue
		}

		if assigned != nil {
			resolved.Assignments = append(resolved.Assignments, *assigned)
		}
	}

	if len(unanswered) > 0 {
		return nil, goerrors.Join(unanswered...)
	}

	return resolved, nil
}

// checkExceptionsAreReachable refuses an exception for a field the destination does not have.
//
// A field renamed on the destination leaves its exception behind, and without this the exception
// is silently ignored and the field silently reverts to whatever the derivation makes of it.
func (p *planner) checkExceptionsAreReachable(conversion *Conversion, destination *structType) error {
	for name := range conversion.Fields {
		if _, ok := destination.Field(name); !ok {
			return errors.Newf("%s: exception for %s, which %s does not have", conversion.Name, name, conversion.To)
		}
	}

	return nil
}

// derive answers a destination field from the two structs, in order of how specific the question
// is. The first question that has an answer wins; a field none of them answer is an error.
func (p *planner) derive(conversion *Conversion, source *structType, field structField) (Rule, error) {
	if rule, ok := p.deriveFromSameName(source, field); ok {
		return rule, nil
	}

	if rule, ok := p.deriveNestedID(source, field); ok {
		return rule, nil
	}

	if rule, ok := p.deriveEntityIdentifier(source, field); ok {
		return rule, nil
	}

	if rule, ok := p.deriveMintedID(conversion, source, field); ok {
		return rule, nil
	}

	if sourceField, ok := source.Field(field.Name); ok {
		return Rule{}, errors.Newf(
			"%s: %s.%s is %s but %s.%s is %s; declare an exception",
			conversion.Name, conversion.To, field.Name, field.Type,
			conversion.From, sourceField.Name, sourceField.Type,
		)
	}

	return Rule{}, errors.Newf(
		"%s: nothing on %s answers %s.%s; declare an exception, or Skip it with a reason",
		conversion.Name, conversion.From, conversion.To, field.Name,
	)
}

// deriveFromSameName handles a destination field the source has under the same name: a copy, a
// value becoming a pointer to itself, or a slice of one shape becoming a slice of another.
//
// Reading through a source pointer is deliberately not here. A *string source and a string
// destination is a dereference that panics on nil, and the two spellings of that decision — read
// it and risk it, or guard it — are not something a generator can choose between.
func (p *planner) deriveFromSameName(source *structType, field structField) (Rule, bool) {
	sourceField, ok := source.Field(field.Name)
	if !ok {
		return Rule{}, false
	}

	switch {
	case sourceField.Type == field.Type:
		return Rule{kind: ruleCopy}, true
	case field.Type == "*"+sourceField.Type:
		return Rule{kind: ruleRef}, true
	}

	return p.deriveSliceConversion(sourceField, field)
}

// deriveSliceConversion maps a slice of one shape of an entity onto a slice of another.
//
// The converter it names is not checked for existence, because the enumeration that emits this
// conversion emits that one too: an entity's shapes are generated together. A slice whose element
// types are unrelated finds nothing to name and falls through to be declared.
func (p *planner) deriveSliceConversion(sourceField, field structField) (Rule, bool) {
	fromElement, ok := sliceElement(sourceField.Type)
	if !ok {
		return Rule{}, false
	}

	toElement, ok := sliceElement(field.Type)
	if !ok {
		return Rule{}, false
	}

	fromEntity, _ := entityOf(fromElement)
	toEntity, _ := entityOf(toElement)

	if fromEntity != toEntity || fromElement == toElement {
		return Rule{}, false
	}

	if !p.index.Declares(p.domain, fromElement) || !p.index.Declares(p.domain, toElement) {
		return Rule{}, false
	}

	return Rule{
		kind:        ruleMapSlice,
		converter:   converterFor(fromElement, toElement),
		elementType: field.Type,
	}, true
}

// deriveNestedID answers a destination that wants an identifier from a source that carries the
// whole entity — the shape a request input takes when the full object is more than it needs.
//
// Two spellings of the same relation are recognized, because the domain packages use both: a field
// named for the relation, where PurchasedMeasurementUnitID is answered by PurchasedMeasurementUnit,
// and a field named for the type, where ValidIngredientID is answered by whichever field holds a
// ValidIngredient. An ambiguous match is no match, so it is declared rather than picked.
func (p *planner) deriveNestedID(source *structType, field structField) (Rule, bool) {
	relation, isIdentifier := strings.CutSuffix(field.Name, "ID")
	if !isIdentifier || relation == "" {
		return Rule{}, false
	}

	nested, ok := source.Field(relation)
	if !ok {
		if nested, ok = p.soleFieldOfType(source, relation); !ok {
			return Rule{}, false
		}
	}

	if !p.index.Declares(p.domain, strings.TrimPrefix(nested.Type, "*")) {
		return Rule{}, false
	}

	if !p.hasIdentifier(nested) {
		return Rule{}, false
	}

	// A pointer destination is the case where absence is representable, so it is guarded; a
	// value destination has nowhere to put the absence and reads straight through.
	return Rule{
		kind:        ruleNestedID,
		sourceField: nested.Name,
		guarded:     strings.HasPrefix(field.Type, "*") && strings.HasPrefix(nested.Type, "*"),
	}, true
}

// deriveEntityIdentifier answers a destination that holds an identifier under the relation's own
// name, from a source that holds the whole entity there.
//
// AccountInvitation.FromUser is a User and AccountInvitationDatabaseCreationInput.FromUser is the
// string that identifies one. The convention is the same as a FooID field — reduce the entity to
// its identifier — spelled without the suffix, and the two spellings should not need two answers
// from a reader.
func (p *planner) deriveEntityIdentifier(source *structType, field structField) (Rule, bool) {
	if field.Type != "string" && field.Type != "*string" {
		return Rule{}, false
	}

	nested, ok := source.Field(field.Name)
	if !ok {
		return Rule{}, false
	}

	if !p.index.Declares(p.domain, strings.TrimPrefix(nested.Type, "*")) {
		return Rule{}, false
	}

	if !p.hasIdentifier(nested) {
		return Rule{}, false
	}

	return Rule{
		kind:        ruleNestedID,
		sourceField: nested.Name,
		guarded:     strings.HasPrefix(field.Type, "*") && strings.HasPrefix(nested.Type, "*"),
	}, true
}

// soleFieldOfType finds the one field of a struct whose type is the named one. Two candidates are
// no answer: which relation an identifier meant would be a guess.
func (p *planner) soleFieldOfType(source *structType, typeName string) (structField, bool) {
	var found structField
	matches := 0

	for _, candidate := range source.Fields {
		if strings.TrimPrefix(candidate.Type, "*") != typeName {
			continue
		}

		found = candidate
		matches++
	}

	return found, matches == 1
}

// hasIdentifier reports whether a nested entity has an ID to read.
func (p *planner) hasIdentifier(nested structField) bool {
	declared, err := p.index.Lookup(p.domain, strings.TrimPrefix(nested.Type, "*"))
	if err != nil {
		return false
	}

	_, ok := declared.Field("ID")

	return ok
}

// deriveMintedID answers the ID of a row that does not have one yet.
//
// A database creation input is the shape that is about to become a row, and a request input has no
// identifier to give it, so the identifier is minted here. Every other shape either carries an ID
// the source also has — which the same-name rule already answered — or does not have the field.
func (p *planner) deriveMintedID(conversion *Conversion, source *structType, field structField) (Rule, bool) {
	if field.Name != "ID" || field.Type != "string" {
		return Rule{}, false
	}

	if !strings.HasSuffix(conversion.To, string(databaseCreationInput)) {
		return Rule{}, false
	}

	if _, sourceHasID := source.Field("ID"); sourceHasID {
		return Rule{}, false
	}

	return Rule{kind: ruleNewID}, true
}

// apply renders one rule, appending to the plan's prelude where the rule needs a local. It returns
// nil for a field that is deliberately left unassigned.
func (p *planner) apply(
	conversion *Conversion,
	source *structType,
	field structField,
	rule *Rule,
	resolved *plan,
) (*assignment, error) {
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
		resolved.UsesIdentifiers = true

		return &assignment{Field: field.Name, Expr: "identifiers.New()", Why: rule.why}, nil
	case ruleCopy, ruleRef, ruleDeref:
		if _, ok := source.Field(sourceName); !ok {
			return nil, errors.Newf("%s has no field %s", conversion.From, sourceName)
		}

		return &assignment{
			Field: field.Name,
			Expr:  readOperator(rule.kind) + sourceParam + "." + sourceName,
			Why:   rule.why,
		}, nil
	case ruleNestedID:
		return p.applyNestedID(field, rule, resolved), nil
	case ruleMapSlice:
		return p.applyMapSlice(conversion, source, field, rule, resolved)
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

// applyNestedID reads the identifier of an entity the source carries whole. A guarded read needs a
// local, since a conditional has nowhere to live inside a literal.
func (p *planner) applyNestedID(field structField, rule *Rule, resolved *plan) *assignment {
	read := sourceParam + "." + rule.sourceField + ".ID"

	if !rule.guarded {
		if strings.HasPrefix(field.Type, "*") {
			read = "&" + read
		}

		return &assignment{Field: field.Name, Expr: read, Why: rule.why}
	}

	local := lowerFirstWord(field.Name)
	resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
		"var %s %s\nif %s.%s != nil {\n%s = &%s\n}",
		local, field.Type, sourceParam, rule.sourceField, local, read,
	))

	return &assignment{Field: field.Name, Expr: local, Why: rule.why}
}

// applyMapSlice converts a slice of entities element by element.
func (p *planner) applyMapSlice(
	conversion *Conversion,
	source *structType,
	field structField,
	rule *Rule,
	resolved *plan,
) (*assignment, error) {
	sourceName := rule.sourceField
	if sourceName == "" {
		sourceName = field.Name
	}

	if _, ok := source.Field(sourceName); !ok {
		return nil, errors.Newf("%s has no field %s", conversion.From, sourceName)
	}

	elementType := rule.elementType
	if elementType == "" {
		elementType = field.Type
	}

	sliceType, err := p.qualify(elementType)
	if err != nil {
		return nil, err
	}

	local := lowerFirstWord(field.Name)
	read := sourceParam + "." + sourceName

	resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
		"%s := make(%s, 0, len(%s))\nfor _, %s := range %s {\n%s = append(%s, %s(%s))\n}",
		local, sliceType, read, elementPlaceholder, read, local, local, rule.converter, elementPlaceholder,
	))

	return &assignment{Field: field.Name, Expr: local, Why: rule.why}, nil
}

// sliceElement returns the named element type of a slice of pointers to one.
func sliceElement(typeExpr string) (string, bool) {
	element, isSlice := strings.CutPrefix(typeExpr, "[]*")
	if !isSlice || strings.ContainsAny(element, "[]*. ") {
		return "", false
	}

	return element, true
}

// qualify rewrites a type as the domain package writes it into a type the generated file can name,
// by prefixing every identifier that names a type of that domain.
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
				typed.Name = p.domain + "." + typed.Name
			}

			return false
		default:
			return true
		}
	})

	return types.ExprString(parsed), nil
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
