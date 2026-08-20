package main

import (
	goast "go/ast"
	"go/types"
	"path"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/reflection/ast"
)

// structIndex is every struct type the domain packages declare, keyed by domain and type name.
//
// It is read out of the source rather than out of the compiled packages because a generator that
// needs its own output to typecheck cannot use a typechecker to decide what to emit: the first
// time a field is added to a domain type, the converters that ought to copy it do not yet, and
// loading the package would fail on nothing more than that.
type structIndex struct {
	// byDomain maps a domain directory name to its declared struct types.
	byDomain map[string]map[string]*structType
}

// structType is one struct declaration, with its fields in declaration order.
type structType struct {
	byName map[string]structField
	Name   string
	// Fields are in declaration order, which is the order the generated literal writes them
	// in. Ordering by the destination rather than alphabetically is what makes a generated
	// literal read like the struct it fills.
	Fields []structField
}

// structField is one field of a struct.
type structField struct {
	Name string
	// Type is the field's type exactly as the domain package writes it, so an unqualified
	// name in it refers to that package.
	Type string
}

// Field returns the named field, and whether the struct has one.
func (s *structType) Field(name string) (structField, bool) {
	field, ok := s.byName[name]

	return field, ok
}

// buildStructIndex walks the domain tree and indexes every struct type each domain package
// declares at its top level.
//
// Only the top level: a domain's subpackages are its converters, its fakes, its manager and its
// mocks, and none of them declare a type a conversion reads or writes. Indexing them would
// introduce name collisions — every domain has a mock and a manager — for no gain.
func buildStructIndex(domainDir string) (*structIndex, error) {
	index := &structIndex{byDomain: map[string]map[string]*structType{}}

	if err := ast.WalkModule(domainDir, func(file *goast.File, relDir string) error {
		domain, rest := splitDomain(relDir)
		if domain == "" || rest != "" {
			return nil
		}

		for _, declared := range structsInFile(file) {
			if index.byDomain[domain] == nil {
				index.byDomain[domain] = map[string]*structType{}
			}

			index.byDomain[domain][declared.Name] = declared
		}

		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "indexing domain structs under %q", domainDir)
	}

	return index, nil
}

// Lookup returns the named struct declared by a domain.
func (i *structIndex) Lookup(domain, name string) (*structType, error) {
	declaredTypes, ok := i.byDomain[domain]
	if !ok {
		return nil, errors.Newf("no domain package %q", domain)
	}

	declared, ok := declaredTypes[name]
	if !ok {
		return nil, errors.Newf("domain %q declares no type %q", domain, name)
	}

	return declared, nil
}

// TypeNames lists every struct type a domain declares, in a stable order so that the enumeration
// it drives produces the same file on every run.
func (i *structIndex) TypeNames(domain string) []string {
	names := make([]string, 0, len(i.byDomain[domain]))
	for name := range i.byDomain[domain] {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// Declares reports whether a domain declares the named type, which is what decides whether an
// identifier appearing in a field's type needs qualifying on the way out.
func (i *structIndex) Declares(domain, name string) bool {
	_, ok := i.byDomain[domain][name]

	return ok
}

// splitDomain separates the domain directory from whatever lies beneath it.
func splitDomain(relDir string) (domain, rest string) {
	relDir = path.Clean(relDir)
	if relDir == "." || relDir == "" {
		return "", ""
	}

	domain, rest, _ = strings.Cut(relDir, "/")

	return domain, rest
}

// structsInFile returns every struct type declared in a file, including those inside a grouped
// type declaration, which is how the domain packages write nearly all of them.
func structsInFile(file *goast.File) []*structType {
	var found []*structType

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*goast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, isType := spec.(*goast.TypeSpec)
			if !isType {
				continue
			}

			structDecl, isStruct := typeSpec.Type.(*goast.StructType)
			if !isStruct {
				continue
			}

			found = append(found, newStructType(typeSpec.Name.Name, structDecl))
		}
	}

	return found
}

// newStructType records a struct's exported fields in declaration order.
//
// The blank sentinel field every domain type opens with — `_ struct{}`, which forbids
// construction without field names — is not a field a conversion assigns, and neither is an
// unexported one, so neither is indexed.
func newStructType(name string, decl *goast.StructType) *structType {
	declared := &structType{Name: name, byName: map[string]structField{}}

	for _, field := range decl.Fields.List {
		fieldType := types.ExprString(field.Type)

		names := field.Names
		if len(names) == 0 {
			// An embedded field is reached by its type's base name.
			if embedded := ast.EmbeddedFieldName(field.Type); embedded != "" {
				declared.add(embedded, fieldType)
			}

			continue
		}

		for _, ident := range names {
			if !ident.IsExported() {
				continue
			}

			declared.add(ident.Name, fieldType)
		}
	}

	return declared
}

func (s *structType) add(name, fieldType string) {
	field := structField{Name: name, Type: fieldType}
	s.Fields = append(s.Fields, field)
	s.byName[name] = field
}
