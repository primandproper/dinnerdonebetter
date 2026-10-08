package main

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/primandproper/primitives-go/v2/errors"
)

// The gRPC axis carries an entity between its domain type and the protobuf message of the same
// name, in whichever direction a handler needs. It is the same problem as the domain axis — a
// mechanical whole-struct copy with a handful of per-field decisions — and it is answered with the
// same vocabulary: every destination field is derived from the two structs, and a field the
// derivation cannot answer, or answers differently from what the converter it replaced did, is an
// exception in grpc_exceptions.go with its reason rendered into the generated source.
//
// What differs is the enumeration. The domain axis has a naming convention that says which
// conversions exist; the gRPC axis does not — a message is converted in the directions a handler
// or a test uses it, and nothing about the two type names says which those are. So the list below
// is declared, and is the one thing on this axis that is.

const (
	// grpcDomain is the domain package the gRPC converters read and write.
	grpcDomain = "mealplanning"

	// grpcConvertersDir is where the gRPC converters live, relative to the backend root.
	grpcConvertersDir = "internal/services/mealplanning/grpc/converters"

	// grpcMessagesDir is the generated protobuf package the converters read and write.
	grpcMessagesDir = "internal/grpc/generated/services/mealplanning"

	// grpcMessagesQualifier is what the generated file calls the protobuf package.
	grpcMessagesQualifier = "mealplanningsvc"
)

// grpcDirection is which way a gRPC conversion goes.
type grpcDirection uint8

const (
	// toGRPCDirection reads a domain type and writes a message.
	toGRPCDirection grpcDirection = iota
	// fromGRPCDirection reads a message and writes a domain type.
	fromGRPCDirection
)

// grpcDeclaration is one declared gRPC conversion: a domain type, the message it corresponds to,
// and the direction. The name is derived from the three.
type grpcDeclaration struct {
	// Doc is an optional paragraph for the generated doc comment, for a conversion whose
	// purpose is not obvious from the two type names — a summary that drops what a full
	// message carries, say.
	Doc       string
	Domain    string
	Message   string
	Direction grpcDirection
}

// Name is the generated function's name. The domain side is named as itself and the message side
// with a GRPC prefix, which is how the converters this replaced spelled all but sixteen of their
// names; those sixteen were renamed to match rather than declared as exceptions.
func (d grpcDeclaration) Name() string {
	if d.Direction == toGRPCDirection {
		return fmt.Sprintf("Convert%sToGRPC%s", d.Domain, d.Message)
	}

	return fmt.Sprintf("ConvertGRPC%sTo%s", d.Message, d.Domain)
}

// From and To are the qualified source and destination types.
func (d grpcDeclaration) From() string {
	if d.Direction == toGRPCDirection {
		return grpcDomain + "." + d.Domain
	}

	return grpcMessagesQualifier + "." + d.Message
}

func (d grpcDeclaration) To() string {
	if d.Direction == toGRPCDirection {
		return grpcMessagesQualifier + "." + d.Message
	}

	return grpcDomain + "." + d.Domain
}

// both declares a conversion each way between a domain type and the message of the same name.
func both(name string) []grpcDeclaration {
	return []grpcDeclaration{toGRPC(name)[0], fromGRPC(name)[0]}
}

// toGRPC declares a conversion from a domain type to the message of the same name.
func toGRPC(name string) []grpcDeclaration {
	return toGRPCAs(name, name)
}

// fromGRPC declares a conversion from a message to the domain type of the same name.
func fromGRPC(name string) []grpcDeclaration {
	return fromGRPCAs(name, name)
}

// toGRPCAs declares a conversion from a domain type to a message named differently.
func toGRPCAs(domain, message string) []grpcDeclaration {
	return []grpcDeclaration{{Domain: domain, Message: message, Direction: toGRPCDirection}}
}

// fromGRPCAs declares a conversion from a message to a domain type named differently.
func fromGRPCAs(message, domain string) []grpcDeclaration {
	return []grpcDeclaration{{Domain: domain, Message: message, Direction: fromGRPCDirection}}
}

// documented attaches a doc paragraph to every conversion in a declaration.
func documented(doc string, declarations []grpcDeclaration) []grpcDeclaration {
	for i := range declarations {
		declarations[i].Doc = doc
	}

	return declarations
}

// grpcIndex is what the gRPC planner reads: the domain package's structs, the message package's
// structs, and the message package's enums. Every field type in both is held fully qualified, so
// that a planner comparing a domain field to a message field compares two spellings of one
// namespace rather than two namespaces.
type grpcIndex struct {
	domain   map[string]*structType
	messages map[string]*structType
	enums    map[string]struct{}
}

// buildGRPCIndex reads the domain package out of the domain index and the message package out of
// its generated source.
func buildGRPCIndex(root string, domainIndex *structIndex) (*grpcIndex, error) {
	index := &grpcIndex{
		domain:   map[string]*structType{},
		messages: map[string]*structType{},
		enums:    map[string]struct{}{},
	}

	domainTypes, ok := domainIndex.byDomain[grpcDomain]
	if !ok {
		return nil, errors.Newf("no domain package %q", grpcDomain)
	}

	messageTypes, enums, err := readMessagePackage(filepath.Join(root, grpcMessagesDir))
	if err != nil {
		return nil, err
	}

	for name := range enums {
		index.enums[grpcMessagesQualifier+"."+name] = struct{}{}
	}

	for name, declared := range domainTypes {
		index.domain[name] = qualifyStruct(declared, grpcDomain)
	}

	for name, declared := range messageTypes {
		index.messages[name] = qualifyStruct(declared, grpcMessagesQualifier)
	}

	return index, nil
}

// readMessagePackage parses the generated protobuf source for its message structs and its enums.
//
// An enum is a named int32, which is how protoc-gen-go spells every one of them. The unexported
// fields every message carries — its state, its size cache, its unknown fields — are dropped by
// the same rule that drops them from a domain struct, so a message's indexed fields are exactly
// the ones a converter assigns.
func readMessagePackage(dir string) (messages map[string]*structType, enums map[string]struct{}, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "reading %s", dir)
	}

	messages = map[string]*structType{}
	enums = map[string]struct{}{}
	fileSet := token.NewFileSet()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pb.go") {
			continue
		}

		file, parseErr := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return nil, nil, errors.Wrapf(parseErr, "parsing %s", entry.Name())
		}

		for _, declared := range structsInFile(file) {
			messages[declared.Name] = declared
		}

		for name := range enumsInFile(file) {
			enums[name] = struct{}{}
		}
	}

	return messages, enums, nil
}

// enumsInFile returns every type a file declares as a named int32.
func enumsInFile(file *goast.File) map[string]struct{} {
	found := map[string]struct{}{}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*goast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, isType := spec.(*goast.TypeSpec)
			if !isType {
				continue
			}

			if ident, isIdent := typeSpec.Type.(*goast.Ident); isIdent && ident.Name == "int32" {
				found[typeSpec.Name.Name] = struct{}{}
			}
		}
	}

	return found
}

// qualifyStruct copies a struct with every type its package declares qualified by that package.
func qualifyStruct(declared *structType, qualifier string) *structType {
	qualified := &structType{Name: declared.Name, byName: map[string]structField{}}

	for _, field := range declared.Fields {
		qualified.add(field.Name, qualifyType(field.Type, qualifier))
	}

	return qualified
}

// qualifyType prefixes every identifier in a type expression that names a type of the package.
//
// An exported identifier that is not already behind a selector can only name a type the package
// itself declares — Go's predeclared types are all lowercase — so that is the whole test.
func qualifyType(typeExpr, qualifier string) string {
	var out strings.Builder

	for i := 0; i < len(typeExpr); {
		if !isIdentStart(typeExpr[i]) {
			out.WriteByte(typeExpr[i])
			i++

			continue
		}

		start := i
		for i < len(typeExpr) && isIdentPart(typeExpr[i]) {
			i++
		}

		ident := typeExpr[start:i]
		alreadyQualified := start > 0 && typeExpr[start-1] == '.'
		selector := i < len(typeExpr) && typeExpr[i] == '.'

		if !alreadyQualified && !selector && goast.IsExported(ident) {
			out.WriteString(qualifier + ".")
		}

		out.WriteString(ident)
	}

	return out.String()
}

func isIdentStart(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isIdentPart(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

// lookup finds a qualified struct type on whichever side of the boundary it lives.
func (i *grpcIndex) lookup(qualified string) (*structType, bool) {
	qualifier, name, ok := strings.Cut(qualified, ".")
	if !ok {
		return nil, false
	}

	var declared *structType

	switch qualifier {
	case grpcDomain:
		declared, ok = i.domain[name]
	case grpcMessagesQualifier:
		declared, ok = i.messages[name]
	default:
		return nil, false
	}

	return declared, ok
}

// isEnum reports whether a qualified type is a protobuf enum.
func (i *grpcIndex) isEnum(qualified string) bool {
	_, ok := i.enums[qualified]

	return ok
}

// grpcConversionList flattens the declarations into the order the generated file lists them in:
// by domain type, then message, then direction.
func grpcConversionList() []grpcDeclaration {
	var all []grpcDeclaration
	for _, group := range grpcConversions {
		all = append(all, group...)
	}

	slices.SortStableFunc(all, func(a, b grpcDeclaration) int {
		if c := strings.Compare(a.Domain, b.Domain); c != 0 {
			return c
		}

		if c := strings.Compare(a.Message, b.Message); c != 0 {
			return c
		}

		return int(a.Direction) - int(b.Direction)
	})

	return all
}

// grpcGeneratedPath is where the gRPC converters' generated file lives, relative to the backend
// root.
func grpcGeneratedPath() string {
	return grpcConvertersDir + "/" + generatedFileName
}
