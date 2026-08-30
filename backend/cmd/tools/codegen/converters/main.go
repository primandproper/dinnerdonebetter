// Command converters generates the domain converters — the functions that carry an entity between
// its stored shape and the request and database input shapes of the same entity.
//
// There is no list of conversions anywhere in this tool. The domain packages already say which
// conversions exist, in the names of their types: an entity X is created from an
// XCreationRequestInput, stored through an XDatabaseCreationInput, and amended by an
// XUpdateRequestInput. shapes.go reads that convention, and every conversion it implies is
// generated — which is why a new entity needs no declaration at all, and why a conversion cannot
// exist for one entity and be missing for its neighbor.
//
// Fields are derived the same way. A destination field is answered by asking the two structs a
// question — is there a field of this name, is this a pointer to what that one holds, is this the
// identifier of something the source carries whole, is this a slice of another shape of the same
// element — and a field that none of those answer fails the build. That is the property worth
// having: a generated converter cannot forget a field, and the symptom of a forgotten copy is
// otherwise a column that is empty for reasons nobody can reconstruct months later.
//
// What is left is per-field knowledge no convention encodes: a field the source genuinely cannot
// answer, an expression particular to one entity. Those are in exceptions.go, each with its
// reason, and each reason is rendered into the generated source. Conversions whose bodies are not
// field assignments at all are listed there too, and stay hand-written in converters_manual.go
// beside the generated file.
package main

import (
	goerrors "errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/primandproper/platform-go/v13/errors"
)

// domainDir is where the domain packages and their converters live, relative to the backend root.
const domainDir = "internal/domain"

// generatedFileName is the file each domain's converters package gets. One file per package rather
// than one per entity: the split by entity was a convention of the hand-written corpus, and nobody
// navigates generated source by filename.
const generatedFileName = "converters_generated.go"

// domainConversions is one domain package's conversions, as enumerated.
type domainConversions struct {
	Domain      string
	Conversions []*Conversion
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

// generate enumerates, plans and writes every domain that has a converters package.
func generate(root string) error {
	index, err := buildStructIndex(filepath.Join(root, domainDir))
	if err != nil {
		return err
	}

	domains, err := domainsWithConverters(filepath.Join(root, domainDir))
	if err != nil {
		return err
	}

	// Every domain is attempted before anything is reported, for the same reason every field
	// is: the answer to "what does this need me to decide" should take one run to get.
	var (
		failures []error
		written  = map[string][]byte{}
	)

	for _, domain := range domains {
		rendered, renderErr := renderDomain(index, enumerate(index, domain))
		if renderErr != nil {
			failures = append(failures, renderErr)

			continue
		}

		written[generatedPath(domain)] = rendered
	}

	if len(failures) > 0 {
		return goerrors.Join(failures...)
	}

	for path, rendered := range written {
		if err = os.WriteFile(filepath.Join(root, path), rendered, 0o0600); err != nil {
			return errors.Wrapf(err, "writing %s", path)
		}
	}

	return nil
}

// domainsWithConverters lists the domains that have a converters package.
//
// The set is read from the tree rather than declared, for the same reason the conversions are: a
// domain that grows a converters package gets its conversions by having one, and a list would be
// one more thing to forget.
func domainsWithConverters(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", root)
	}

	var domains []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if _, statErr := os.Stat(filepath.Join(root, entry.Name(), "converters")); statErr != nil {
			continue
		}

		domains = append(domains, entry.Name())
	}

	return domains, nil
}

// enumerate is the whole specification of which conversions exist: every entity a domain declares,
// crossed with every shape rule whose two types that entity has.
func enumerate(index *structIndex, domain string) *domainConversions {
	enumerated := &domainConversions{Domain: domain}

	for _, typeName := range index.TypeNames(domain) {
		entityName, isEntity := entityOf(typeName)
		if !isEntity {
			continue
		}

		for shapeIndex, shape := range conversionShapes {
			from := shape.From.typeName(entityName)
			to := shape.To.typeName(entityName)

			if !index.Declares(domain, from) || !index.Declares(domain, to) {
				continue
			}

			name := conversionName(from, to)
			if _, manual := handWritten[name]; manual {
				continue
			}

			enumerated.Conversions = append(enumerated.Conversions, &Conversion{
				Entity:     entityName,
				From:       from,
				To:         to,
				Name:       name,
				Fields:     fieldExceptions[name],
				shapeIndex: shapeIndex,
			})
		}
	}

	sortConversions(enumerated.Conversions)

	return enumerated
}

// renderDomain plans every conversion of one domain and renders the file it belongs in.
func renderDomain(index *structIndex, enumerated *domainConversions) ([]byte, error) {
	if len(enumerated.Conversions) == 0 {
		return nil, errors.Newf("domain %s enumerated no conversions", enumerated.Domain)
	}

	resolver := &planner{index: index, domain: enumerated.Domain}
	plans := make([]*plan, 0, len(enumerated.Conversions))

	var unplanned []error

	for _, conversion := range enumerated.Conversions {
		resolved, err := resolver.Plan(conversion)
		if err != nil {
			unplanned = append(unplanned, err)

			continue
		}

		plans = append(plans, resolved)
	}

	if len(unplanned) > 0 {
		return nil, goerrors.Join(unplanned...)
	}

	return render(enumerated.Domain, plans)
}

// generatedPath is where a domain's generated file lives, relative to the backend root. The
// staleness test needs the same answer this tool acts on, and deriving it twice is how the two
// drift.
func generatedPath(domain string) string {
	return fmt.Sprintf("%s/%s/converters/%s", domainDir, domain, generatedFileName)
}
