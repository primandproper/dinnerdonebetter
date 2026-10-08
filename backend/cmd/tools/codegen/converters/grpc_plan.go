package main

import (
	goerrors "errors"
	"fmt"
	goast "go/ast"
	"go/token"
	"strings"

	"github.com/primandproper/primitives-go/v2/errors"
)

// grpcPlanner resolves gRPC conversions against the two packages they cross between.
type grpcPlanner struct {
	index *grpcIndex
	// converters maps a source and destination pointer type, both qualified, to the declared
	// conversion between them — which is how a nested message or a collection of them finds
	// the function that converts it.
	converters map[[2]string]string
}

func newGRPCPlanner(index *grpcIndex, declarations []grpcDeclaration) *grpcPlanner {
	converters := map[[2]string]string{}
	for _, declared := range declarations {
		// Keyed by pointer types, because that is what every converter takes and returns.
		converters[[2]string{"*" + declared.From(), "*" + declared.To()}] = declared.Name()
	}

	return &grpcPlanner{index: index, converters: converters}
}

// Plan resolves one gRPC conversion. As on the domain axis, every destination field is attempted
// even after one fails, so that one run reports every decision a conversion needs.
func (p *grpcPlanner) Plan(declared grpcDeclaration, rules map[string]Rule) (*plan, error) {
	conversion := &Conversion{Name: declared.Name(), From: declared.From(), To: declared.To(), Fields: rules}

	destination, ok := p.index.lookup(conversion.To)
	if !ok {
		return nil, errors.Newf("%s: no struct %s", conversion.Name, conversion.To)
	}

	source, ok := p.index.lookup(conversion.From)
	if !ok {
		return nil, errors.Newf("%s: no struct %s", conversion.Name, conversion.From)
	}

	for name := range rules {
		if _, has := destination.Field(name); !has {
			return nil, errors.Newf("%s: exception for %s, which %s does not have", conversion.Name, name, conversion.To)
		}
	}

	resolved := &plan{Conversion: conversion}

	var unanswered []error

	for _, field := range destination.Fields {
		assigned, err := p.field(conversion, source, field, resolved)
		if err != nil {
			unanswered = append(unanswered, errors.Wrapf(err, "%s: field %s", conversion.Name, field.Name))

			continue
		}

		resolved.Assignments = append(resolved.Assignments, *assigned)
	}

	if len(unanswered) > 0 {
		return nil, goerrors.Join(unanswered...)
	}

	return resolved, nil
}

// field answers one destination field, from its exception if it has one and from the two structs
// if it does not.
func (p *grpcPlanner) field(conversion *Conversion, source *structType, field structField, resolved *plan) (*assignment, error) {
	rule, declared := conversion.Fields[field.Name]

	switch rule.kind {
	case ruleSkip:
		if declared {
			return &assignment{Field: field.Name, Why: rule.why, Skipped: true}, nil
		}
	case ruleExpr:
		return &assignment{Field: field.Name, Expr: rule.expr, Why: rule.why}, nil
	case ruleCopy, ruleRef, ruleDeref, ruleNestedID, ruleNewID, ruleMapSlice:
		if declared {
			return nil, errors.Newf("rule kind %d is not used on the gRPC axis", rule.kind)
		}
	case ruleFrom, ruleUnguarded, ruleEmptySlice, ruleDetached:
	}

	sourceName := rule.sourceField
	sourceField, ok := source.Field(sourceName)
	if sourceName == "" {
		sourceField, ok = sameNamedField(source, field.Name)
	}

	if !ok && sourceName == "" {
		if read, isNestedID := p.nestedID(source, field); isNestedID {
			return &assignment{Field: field.Name, Expr: read, Why: rule.why}, nil
		}
	}

	if !ok {
		return nil, errors.Newf("nothing on %s answers it; declare an exception, or Skip it with a reason", conversion.From)
	}

	read := sourceParam + "." + sourceField.Name
	local := localName(field.Name)

	expr, err := p.carry(sourceField.Type, field.Type, read, local, rule.kind, resolved)
	if err != nil {
		return nil, err
	}

	return &assignment{Field: field.Name, Expr: expr, Why: rule.why}, nil
}

// sameNamedField finds the source field a destination field corresponds to. Names are compared
// without regard to case, because protoc-gen-go spells an initialism the way it spells any other
// word: the domain's ID is the message's Id, and SourceISBN is SourceIsbn.
func sameNamedField(source *structType, name string) (structField, bool) {
	if field, ok := source.Field(name); ok {
		return field, true
	}

	for _, candidate := range source.Fields {
		if strings.EqualFold(candidate.Name, name) {
			return candidate, true
		}
	}

	return structField{}, false
}

// nestedID answers an identifier the destination holds on its own from a source that holds the
// whole entity by value — a list item's MealId from its Meal. Only a value is read this way: it
// cannot be nil, so the read cannot panic, and an entity held by pointer is left to be declared.
func (p *grpcPlanner) nestedID(source *structType, field structField) (string, bool) {
	if field.Type != stringType {
		return "", false
	}

	relation, isIdentifier := strings.CutSuffix(field.Name, "Id")
	if !isIdentifier {
		relation, isIdentifier = strings.CutSuffix(field.Name, "ID")
	}

	if !isIdentifier || relation == "" {
		return "", false
	}

	nested, ok := sameNamedField(source, relation)
	if !ok {
		return "", false
	}

	declared, ok := p.index.lookup(nested.Type)
	if !ok {
		return "", false
	}

	identifier, ok := sameNamedField(declared, "ID")
	if !ok || identifier.Type != stringType {
		return "", false
	}

	return fmt.Sprintf("%s.%s.%s", sourceParam, nested.Name, identifier.Name), true
}

// carry renders the expression that takes a source field of one type into a destination field of
// another, appending to the plan's prelude where that takes a statement rather than an expression.
func (p *grpcPlanner) carry(from, to, read, local string, kind ruleKind, resolved *plan) (string, error) {
	fromElement, fromIsSlice := strings.CutPrefix(from, "[]")
	toElement, toIsSlice := strings.CutPrefix(to, "[]")

	switch {
	case kind == ruleDetached:
		if from != to || !strings.HasPrefix(from, "*") {
			return "", errors.Newf("Detached needs the same pointer type on both sides, not %s and %s", from, to)
		}

		resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
			"var %s %s\nif %s != nil {\n%s = new(*%s)\n}", local, to, read, local, read,
		))

		return local, nil
	case from == to:
		if kind == ruleEmptySlice || kind == ruleUnguarded {
			return "", errors.Newf("the exception does not apply to a field that is copied")
		}

		return read, nil
	case fromIsSlice && toIsSlice:
		return p.carrySlice(fromElement, toElement, to, read, local, kind, resolved)
	}

	if kind == ruleEmptySlice {
		return "", errors.Newf("EmptySlice applies only to a collection")
	}

	if expr, ok := p.adapt(from, to, read); ok {
		if kind == ruleUnguarded {
			return "", errors.Newf("Unguarded applies only to a nested message held by pointer on both sides")
		}

		return expr, nil
	}

	// A pointer on both sides whose pointees can be carried across is an optional value: absent
	// stays absent, and present is converted into a new pointer.
	fromPointee, fromIsPointer := strings.CutPrefix(from, "*")
	toPointee, toIsPointer := strings.CutPrefix(to, "*")

	if !fromIsPointer || !toIsPointer {
		return "", errors.Newf("%s cannot be carried into %s; declare an exception", from, to)
	}

	if converter, isMessage := p.converters[[2]string{from, to}]; isMessage {
		if kind == ruleUnguarded {
			return fmt.Sprintf("%s(%s)", converter, read), nil
		}

		resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
			"var %s %s\nif %s != nil {\n%s = %s(%s)\n}", local, to, read, local, converter, read,
		))

		return local, nil
	}

	if kind == ruleUnguarded {
		return "", errors.Newf("Unguarded applies only to a nested message held by pointer on both sides")
	}

	inner, ok := p.adapt(fromPointee, toPointee, "*"+read)
	if !ok {
		return "", errors.Newf("%s cannot be carried into %s; declare an exception", from, to)
	}

	resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
		"var %s %s\nif %s != nil {\n%s = new(%s)\n}", local, to, read, local, inner,
	))

	return local, nil
}

// carrySlice converts a collection element by element.
func (p *grpcPlanner) carrySlice(fromElement, toElement, sliceType, read, local string, kind ruleKind, resolved *plan) (string, error) {
	if kind == ruleUnguarded {
		return "", errors.Newf("Unguarded does not apply to a collection")
	}

	element, ok := p.adapt(fromElement, toElement, elementPlaceholder)
	if !ok {
		converter, isMessage := p.converters[[2]string{fromElement, toElement}]
		if !isMessage {
			return "", errors.Newf("no conversion from %s to %s; declare one", fromElement, toElement)
		}

		element = fmt.Sprintf("%s(%s)", converter, elementPlaceholder)
	}

	initial := fmt.Sprintf("var %s %s", local, sliceType)
	if kind == ruleEmptySlice {
		initial = fmt.Sprintf("%s := %s{}", local, sliceType)
	}

	resolved.Prelude = append(resolved.Prelude, fmt.Sprintf(
		"%s\nfor _, %s := range %s {\n%s = append(%s, %s)\n}",
		initial, elementPlaceholder, read, local, local, element,
	))

	return local, nil
}

// numericTypes are the integer types the two sides disagree on. Protobuf has no integer narrower
// than 32 bits, so every narrower domain integer is widened on the way out and narrowed on the way
// back in.
var numericTypes = map[string]struct{}{
	"int8": {}, "int16": {}, "int32": {}, "int64": {},
	"uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
}

// The wire types the helpers below convert to and from.
const (
	wireTimestamp     = "*timestamppb.Timestamp"
	wireUint32Pointer = "*uint32"
)

// helpers are the conversions between a domain type and its wire type that the shared gRPC
// converters package already has a function for.
var helpers = map[[2]string]string{
	{"time.Time", wireTimestamp}:                         "converters.ConvertTimeToPBTimestamp",
	{"*time.Time", wireTimestamp}:                        "converters.ConvertTimePointerToPBTimestamp",
	{wireTimestamp, "time.Time"}:                         "converters.ConvertPBTimestampToTime",
	{wireTimestamp, "*time.Time"}:                        "converters.ConvertPBTimestampToTimePointer",
	{"*uint16", wireUint32Pointer}:                       "converters.ConvertUint16PointerToUint32Pointer",
	{wireUint32Pointer, "*uint16"}:                       "converters.ConvertUint32PointerToUint16Pointer",
	{"*uint8", wireUint32Pointer}:                        "converters.ConvertUint8PointerToUint32Pointer",
	{wireUint32Pointer, "*uint8"}:                        "converters.ConvertUint32PointerToUint8Pointer",
	{"*mediaregistry.Object", "*mediaregistrypb.Object"}: "mediaregistrygrpc.ObjectToProto",
}

// adapt renders the expression that carries a value of one type into another, where that is an
// expression: a helper, a numeric conversion, an enum, a nested message. It reports false for a
// pair it has no answer for, which the caller either answers another way or refuses.
func (p *grpcPlanner) adapt(from, to, read string) (string, bool) {
	if from == to {
		return read, true
	}

	if helper, ok := helpers[[2]string{from, to}]; ok {
		return fmt.Sprintf("%s(%s)", helper, read), true
	}

	_, fromIsNumber := numericTypes[from]
	_, toIsNumber := numericTypes[to]

	if fromIsNumber && toIsNumber {
		return fmt.Sprintf("%s(%s)", to, read), true
	}

	// An enum crosses as a string on the domain side, through the hand-written pair of
	// functions every enum has.
	if from == stringType && p.index.isEnum(to) {
		return fmt.Sprintf("ConvertStringTo%s(%s)", unqualified(to), read), true
	}

	if to == stringType && p.index.isEnum(from) {
		return fmt.Sprintf("Convert%sToString(%s)", unqualified(from), read), true
	}

	// A nested message the domain holds by value: the converter wants a pointer on the way
	// out, and hands one back on the way in. The way in reads straight through, because a
	// value destination has nowhere to put an absent message.
	if converter, ok := p.converters[[2]string{"*" + from, to}]; ok && !strings.HasPrefix(from, "*") {
		return fmt.Sprintf("%s(&%s)", converter, read), true
	}

	if converter, ok := p.converters[[2]string{from, "*" + to}]; ok && !strings.HasPrefix(to, "*") {
		return fmt.Sprintf("*%s(%s)", converter, read), true
	}

	// A value the destination holds by pointer is pointed at where it is, and a pointer the
	// destination holds by value is read through safely.
	if to == "*"+from {
		return "&" + read, true
	}

	if from == "*"+to {
		return fmt.Sprintf("pointer.Dereference(%s)", read), true
	}

	return "", false
}

// unqualified drops the package from a qualified type name.
func unqualified(qualified string) string {
	_, name, found := strings.Cut(strings.TrimPrefix(qualified, "*"), ".")
	if !found {
		return qualified
	}

	return name
}

// localName is the variable a field's converted value is held in before the literal. It is the
// field's name lowercased, except where that would be a keyword or would shadow a name the
// generated function already uses.
func localName(field string) string {
	name := lowerFirstWord(field)

	if token.IsKeyword(name) || name == sourceParam || name == elementPlaceholder || !goast.IsExported(field) {
		return name + "Value"
	}

	return name
}
