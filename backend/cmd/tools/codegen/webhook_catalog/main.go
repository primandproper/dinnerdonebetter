// Command webhook_catalog generates the webhook event catalog from the service event type
// constants the domains already declare.
//
// The catalog is what platform-go's webhooks package gates both subscription and dispatch on: an
// event type outside it cannot be subscribed to and cannot be dispatched. That makes it the
// authoritative list of which events this application publishes, and the one thing that must not
// drift from the constants themselves — a hand-maintained copy beside 171 constants spread over
// two dozen files drifts the first time someone adds a domain.
//
// So it is derived rather than written. Every `const` in internal/domain whose name ends in
// ServiceEventType contributes its value as the event type and its doc comment as the
// description an endpoint-management UI renders beside the checkbox.
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// domainDir is the only tree searched. Event types are a domain concept; one declared in a
// repository or a service would be a layering mistake, and finding it here would legitimize it.
const domainDir = "internal/domain"

// constSuffix identifies the constants this tool collects.
//
// EventType rather than ServiceEventType, which is what most of them are named. The identity
// domain has a dozen — password_changed, user_session_revoked, the email verification pair —
// that drop the Service infix, and they are published exactly like the rest. Collecting only the
// longer suffix left them outside the catalog, which would have made every one of them fail the
// dispatch gate and take its transaction down with it.
const constSuffix = "EventType"

// outputPath is the generated file, relative to the backend root.
var outputPath = filepath.Join("internal", "domain", "webhooks", "catalog", "catalog.go")

// event is one catalog entry.
type event struct {
	// eventType is the constant's string value — what travels on the wire.
	eventType string
	// description is prose from the constant's doc comment.
	description string
	// constName is retained for error messages, which are otherwise unable to say where a
	// duplicate came from.
	constName string
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	events, err := collect(filepath.Join(dir, domainDir))
	if err != nil {
		log.Fatal(err)
	}

	if len(events) == 0 {
		// An empty catalog rejects every event type, which would silently disable webhook
		// delivery entirely. Far better to fail the build than to generate that.
		log.Fatalf("no %s constants found under %s", constSuffix, domainDir)
	}

	rendered, err := render(events)
	if err != nil {
		log.Fatal(err)
	}

	if err = os.WriteFile(filepath.Join(dir, outputPath), rendered, 0o0600); err != nil {
		log.Fatal(err)
	}
}

// collect walks root and returns every service event type constant, sorted by event type.
func collect(root string) ([]event, error) {
	byType := map[string]event{}

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Test files and generated fakes declare no event types, and parsing them only
		// creates opportunities for a fixture constant to reach the wire.
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		node, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parsing %s: %w", path, parseErr)
		}

		for _, found := range eventsInFile(node) {
			// Two constants with the same value are the same event under two names, which
			// means one of them is dead and nobody knows which. Refuse rather than pick.
			if existing, ok := byType[found.eventType]; ok && existing.constName != found.constName {
				return fmt.Errorf(
					"event type %q is declared by both %s and %s",
					found.eventType, existing.constName, found.constName,
				)
			}

			byType[found.eventType] = found
		}

		return nil
	}); err != nil {
		return nil, err
	}

	events := make([]event, 0, len(byType))
	for _, e := range byType {
		events = append(events, e)
	}

	slices.SortFunc(events, func(a, b event) int {
		return strings.Compare(a.eventType, b.eventType)
	})

	return events, nil
}

// eventsInFile returns the service event type constants declared in one file.
func eventsInFile(node *ast.File) []event {
	var events []event

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, isValueSpec := spec.(*ast.ValueSpec)
			if !isValueSpec || len(valueSpec.Names) != 1 || len(valueSpec.Values) != 1 {
				continue
			}

			name := valueSpec.Names[0].Name
			if !strings.HasSuffix(name, constSuffix) {
				continue
			}

			// Only an untyped string literal is usable: a constant built from an
			// expression has no value this tool can read without type checking, and one
			// has never appeared here.
			lit, isLit := valueSpec.Values[0].(*ast.BasicLit)
			if !isLit || lit.Kind != token.STRING {
				continue
			}

			value, err := strconv.Unquote(lit.Value)
			if err != nil || value == "" {
				continue
			}

			events = append(events, event{
				eventType:   value,
				description: describe(name, valueSpec.Doc),
				constName:   name,
			})
		}
	}

	return events
}

// describe renders a constant's doc comment as endpoint-facing prose.
//
// The comments follow Go's convention of leading with the identifier — "MealPlanCreated‐
// ServiceEventType indicates a meal plan was created" — and that prefix is noise to a subscriber
// choosing events in a UI. It is stripped, leaving the sentence that was always the useful part.
func describe(name string, doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	text := strings.TrimSpace(doc.Text())
	if text == "" {
		return ""
	}

	// Collapse the comment to a single line: descriptions render in one cell of a
	// subscription UI, and a wrapped comment would otherwise carry its newlines there.
	text = strings.Join(strings.Fields(text), " ")

	// Linter directives share the doc comment with the prose but are not prose. Left in,
	// "#nosec G101" is shipped to every subscriber browsing the event list.
	if idx := strings.Index(text, "#nosec"); idx >= 0 {
		text = strings.TrimSpace(text[:idx])
	}

	for _, prefix := range []string{name + " indicates ", name + " is ", name + " "} {
		if after, found := strings.CutPrefix(text, prefix); found {
			text = after
			break
		}
	}

	text = strings.TrimSuffix(text, ".")
	if text == "" {
		return ""
	}

	return strings.ToUpper(text[:1]) + text[1:] + "."
}

// render emits the generated file, gofmt'd.
//
// Formatting here rather than in a follow-up shell step is what lets the staleness test compare
// this output to the file on disk byte for byte: a separately formatted file would differ from
// unformatted output on every run and the test could only ever check something weaker.
func render(events []event) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(`// Code generated by cmd/tools/codegen/webhook_catalog. DO NOT EDIT.

// Package catalog is the set of event types this application publishes, and the only source of
// which events a webhook endpoint may subscribe to.
//
// It is generated from the ServiceEventType constants the domains declare, so an event that
// exists is subscribable and one that does not cannot be subscribed to by typo. Both halves
// matter: platform-go's webhooks package rejects an unknown event type at registration and again
// at dispatch, and a subscription to "reciped.created" accepted silently produces an endpoint
// that never fires and no signal explaining why.
//
// Add an event by declaring its constant in internal/domain and running ` + "`make webhook_catalog`" + `.
package catalog

import (
	"github.com/primandproper/platform-go/v12/webhooks"
)

// definitions is the generated catalog.
var definitions = webhooks.Catalog{
`)

	for _, e := range events {
		description := e.description
		if description == "" {
			// A missing description is not worth failing the build over, but it is worth
			// being visible: this string is what a subscriber reads.
			description = "No description available."
		}

		fmt.Fprintf(&sb, "\t%s: {Description: %s},\n", strconv.Quote(e.eventType), strconv.Quote(description))
	}

	sb.WriteString(`}

// Catalog returns the event types a webhook may subscribe to.
//
// That is everything this application publishes, less the exclusions in excluded.go — events
// describing an account's security activity rather than its contents. Filtering here rather than
// at each call site is what makes the exclusion hold at registration and at dispatch from one
// decision.
//
// A fresh map each call, because the caller hands it to a dispatcher that retains it: a shared
// map would let one consumer's mutation change what every other consumer considers dispatchable.
func Catalog() webhooks.Catalog {
	subscribable := make(webhooks.Catalog, len(definitions))
	for eventType, definition := range definitions {
		if Excluded(eventType.String()) {
			continue
		}

		subscribable[eventType] = definition
	}

	return subscribable
}

// Known reports whether eventType may be delivered to a webhook.
//
// An excluded event is not known here even though the application publishes it, so a dispatch of
// one is refused even if a subscription to it somehow existed.
//
// The parameter is a plain string rather than a webhooks.EventType because the domains name their
// events with untyped constants and this is where the two vocabularies meet. Converting here keeps
// that conversion in one place instead of at every call site.
func Known(eventType string) bool {
	if Excluded(eventType) {
		return false
	}

	_, ok := definitions[webhooks.EventType(eventType)]

	return ok
}

// Published reports whether eventType is one this application emits at all, deliverable or not.
func Published(eventType string) bool {
	_, ok := definitions[webhooks.EventType(eventType)]

	return ok
}
`)

	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		return nil, fmt.Errorf("formatting generated catalog: %w", err)
	}

	return formatted, nil
}
