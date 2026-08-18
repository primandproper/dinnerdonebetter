// Command converters generates the domain converters — the functions that carry an entity between
// its stored shape and the request and database input shapes of the same entity.
//
// Nearly all of that is a field copy, and the reason to generate it is not the line count: a
// generated converter cannot forget a field. A destination field with no same-named source field
// and no declared rule fails this tool rather than becoming a zero value, which is the one failure
// mode a reviewer of a hand-written converter cannot reliably catch — the symptom of a forgotten
// copy is a column that is empty, arriving much later, with nothing in it to say what went wrong.
//
// What is left is the quarter of the assignments that are not copies: an ID minted, a nested
// entity reduced to its identifier, a slice mapped through another converter, an expression that
// is genuinely bespoke. Those are declared beside the conversion, with the reason, and the reason
// is rendered into the generated source where a reader of the output meets it.
//
// Conversions this vocabulary cannot express stay hand-written in the same package, beside the
// generated file. See converters_manual.go in each domain for what that is and why.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/primandproper/platform-go/v11/errors"
)

// domainDir is where the domain packages and their converters live, relative to the backend root.
const domainDir = "internal/domain"

// generatedFileName is the file each domain's converters package gets. One file per package rather
// than one per entity, because the split by entity is a convention of the hand-written corpus and
// a generated tree does not need it: nobody navigates generated source by filename.
const generatedFileName = "converters_generated.go"

// domainConversions is one domain package's declared conversions.
type domainConversions struct {
	// Domain is the directory under internal/domain, which is also the domain package's name.
	Domain string
	// Conversions are declared in the order they are rendered, so the generated file's order
	// is a decision in the declaration rather than a property of a map iteration.
	Conversions []*Conversion
}

// registered is every domain's declarations, populated by the per-domain files' init functions.
var registered []*domainConversions

// register records a domain's conversions. It is called from an init in each declaration file,
// which is what lets a domain be added by adding one file.
func register(domain string, conversions []*Conversion) {
	registered = append(registered, &domainConversions{Domain: domain, Conversions: conversions})
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if err = generate(root); err != nil {
		log.Fatal(err)
	}
}

// generate plans and writes every registered domain.
func generate(root string) error {
	index, err := buildStructIndex(filepath.Join(root, domainDir))
	if err != nil {
		return err
	}

	for _, declarations := range registered {
		rendered, renderErr := renderDomain(index, declarations)
		if renderErr != nil {
			return renderErr
		}

		path := filepath.Join(root, domainDir, declarations.Domain, "converters", generatedFileName)
		if err = os.WriteFile(path, rendered, 0o0600); err != nil {
			return errors.Wrapf(err, "writing %s", path)
		}
	}

	return nil
}

// renderDomain plans every conversion of one domain and renders the file it belongs in.
func renderDomain(index *structIndex, declarations *domainConversions) ([]byte, error) {
	if len(declarations.Conversions) == 0 {
		return nil, errors.Newf("domain %s registered no conversions", declarations.Domain)
	}

	resolver := &planner{index: index, domain: declarations.Domain, pkg: declarations.Domain}

	seen := map[string]struct{}{}
	plans := make([]*plan, 0, len(declarations.Conversions))

	for _, conversion := range declarations.Conversions {
		if _, duplicate := seen[conversion.Name]; duplicate {
			return nil, errors.Newf("%s is declared twice in %s", conversion.Name, declarations.Domain)
		}

		seen[conversion.Name] = struct{}{}

		resolved, err := resolver.Plan(conversion)
		if err != nil {
			return nil, err
		}

		plans = append(plans, resolved)
	}

	return render(declarations.Domain, plans)
}

// generatedPath is where a domain's generated file lives, relative to the backend root. The
// staleness test needs the same answer this tool acts on, and deriving it twice is how the two
// drift.
func generatedPath(domain string) string {
	return fmt.Sprintf("%s/%s/converters/%s", domainDir, domain, generatedFileName)
}
