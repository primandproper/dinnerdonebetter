package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// knownImports is every package a generated fakes file may refer to, keyed by the qualifier it
// refers to it by.
//
// A closed list rather than goimports, because this file's bytes are compared against the ones on
// disk to detect drift: a resolver that picks an alias by searching the module is one upgrade away
// from picking a different one, and the failure would look like a stale generated file.
//
// Adding to it is the price of an override that reaches for a new package, and the generator says
// so by name when one does.
func knownImports(domainName string) map[string]string {
	domainPath := domainImportPath(domainName)

	imports := map[string]string{
		"base32":    "encoding/base32",
		"fmt":       "fmt",
		"http":      "net/http",
		"log":       "log",
		"math":      "math",
		packageTime: packageTime,

		"authorization": backendPrefix + "/internal/authorization",
		"sessions":      backendPrefix + "/internal/authentication/sessions",
		"identity":      backendPrefix + "/internal/domain/identity",

		"filtering":   "github.com/primandproper/platform-go/v10/filtering",
		"identifiers": "github.com/primandproper/platform-go/v10/identifiers",
		"pointer":     "github.com/primandproper/platform-go/v10/pointer",

		"fake": "github.com/brianvoe/gofakeit/v7",
		"totp": "github.com/pquerna/otp/totp",

		domainAlias:  domainPath,
		"converters": domainPath + "/converters",
		"catalog":    domainPath + "/catalog",
	}

	// identity is the one domain that is also a package other domains' fakes import by its own
	// name. Inside its own fakes package it is `types` like every other domain, and leaving
	// both bindings in place would import it twice.
	if domainPath == imports["identity"] {
		delete(imports, "identity")
	}

	return imports
}

// qualifiers returns the identifiers that appear before a dot in Go source.
//
// Tokenised rather than pattern-matched, so that a package name inside a comment — and the
// override comments quote plenty of code — does not pull in an import nothing uses.
func qualifiers(source string) map[string]struct{} {
	found := map[string]struct{}{}

	var s scanner.Scanner

	file := token.NewFileSet().AddFile("", -1, len(source))
	s.Init(file, []byte(source), nil, 0)

	previous := ""
	previousWasIdent := false

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		if tok == token.PERIOD && previousWasIdent {
			found[previous] = struct{}{}
		}

		previousWasIdent = tok == token.IDENT
		previous = lit
	}

	return found
}

// identifiersInPackage returns every identifier mentioned by a fakes package: the ones in the
// body about to be generated, and the ones in the hand-written files that will sit beside it.
//
// The hand-written half is what decides which helpers survive. buildFakePassword exists in a
// package because something calls it, and after this generator takes over the builders that
// something may be a bespoke builder or a test. Emitting the helper regardless would leave dead
// code in four packages; emitting it only for generated callers would break the others.
func identifiersInPackage(dir, body string) (map[string]struct{}, error) {
	used := map[string]struct{}{}

	var s scanner.Scanner

	file := token.NewFileSet().AddFile("", -1, len(body))
	s.Init(file, []byte(body), nil, 0)

	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		if tok == token.IDENT {
			used[lit] = struct{}{}
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || entry.Name() == outputName {
			continue
		}

		parsed, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, 0)
		if parseErr != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), parseErr)
		}

		ast.Inspect(parsed, func(node ast.Node) bool {
			if identifier, ok := node.(*ast.Ident); ok {
				used[identifier.Name] = struct{}{}
			}

			return true
		})
	}

	return used, nil
}
