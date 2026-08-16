// Command fakes generates the domain fake builders from the entity declarations the domains keep
// beside their types.
//
// A fake builder is a struct literal with one line per field, and there are ninety of them. Almost
// every line is mechanical — a string field gets a string, a bool field gets a bool — and the
// mechanical lines were the entire maintenance cost: a field added to a type is a field missing
// from its fake until somebody notices, and nothing notices except a test that happens to read it.
//
// The lines that are not mechanical are the reason this is a generator over a declaration rather
// than a call to a reflective faker. A webhook's URL cannot be random, because registration
// resolves it. Those are per-field overrides in internal/domain/*/entities.go, each carrying the
// sentence explaining it, and both the expression and the sentence are emitted into the generated
// source. A generator that only did the mechanical part would produce fakes that fail validation
// everywhere, which is a worse outcome than the duplication it removed.
//
// # What counts as mechanical
//
// More than the name suggests, and deliberately: every field the generator cannot answer becomes a
// line somebody writes by hand, and enough of those and the declarations cost what the fakes did.
// Four rules do most of the work, and each one was a category of hand-written override before it:
//
//   - Enumerated fields. Almost none of them are typed — `MealPlan.Status` is a `string` — so the
//     type says nothing about which strings it may hold. The rule the type declares about itself
//     does: `validation.Field(&x.Shape, validation.In(...))` is the list, it is authoritative
//     because it is what rejects a bad value at runtime, and enums.go reads it out of the source.
//     All of its members are offered, at random, rather than the one somebody picked.
//   - Optional fields, which are filled rather than left nil. A nil optional makes every assertion
//     about it vacuous. `ArchivedAt` is excepted: there, nil is the field's meaning.
//   - Child collections, which are built and wired to their parent's ID by children.go. This was
//     the same loop written once per parent, nineteen times in mealplanning alone.
//   - Overrides that restate any of the above, which are rejected rather than emitted, so that an
//     improvement here actually shrinks the declarations instead of being shadowed by them.
//
// Run with `make fakes`.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
)

// outputName is the generated file in each fakes package.
//
// One file per package rather than one per entity, because the split into recipe_step.go,
// recipe_step_product.go and so on carried information only while the files were written by hand.
const outputName = "generated.go"

// domain pairs a domain's declarations with where its fakes package lives.
type domain struct {
	// name is the domain package's name, which is also its directory under internal/domain.
	name string
	// decl is the domain's entity declarations.
	decl entitydecl.Domain
}

// domains is every domain with a fakes package.
//
// Hand-maintained, and deliberately so: the alternative is walking internal/domain and looking
// for an Entities symbol, which trades thirteen lines here for a reflective search that fails
// silently when a domain is renamed. A domain missing from this list has a stale fakes package
// and no signal saying so, so the list is checked against the tree by TestDomainsAreComplete.
var domains = []domain{
	{name: "audit", decl: audit.Entities},
	{name: "auth", decl: auth.Entities},
	{name: "comments", decl: comments.Entities},
	{name: "identity", decl: identity.Entities},
	{name: "issuereports", decl: issuereports.Entities},
	{name: "mealplanning", decl: mealplanning.Entities},
	{name: "notifications", decl: notifications.Entities},
	{name: "oauth", decl: oauth.Entities},
	{name: "payments", decl: payments.Entities},
	{name: "settings", decl: settings.Entities},
	{name: "uploadedmedia", decl: uploadedmedia.Entities},
	{name: "waitlists", decl: waitlists.Entities},
	{name: "webhooks", decl: webhooks.Entities},
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	enums := newEnumIndex(root)

	for _, d := range domains {
		dir := filepath.Join(root, packageDir(d.name))

		rendered, renderErr := renderDomain(d, dir, enums)
		if renderErr != nil {
			log.Fatalf("generating fakes for %s: %v", d.name, renderErr)
		}

		if err = os.WriteFile(filepath.Join(dir, outputName), rendered, 0o0600); err != nil {
			log.Fatal(err)
		}
	}
}

// packageDir is a domain's fakes package, relative to the backend root.
func packageDir(name string) string {
	return filepath.Join("internal", "domain", name, "fakes")
}

// domainImportPath is a domain package's import path.
func domainImportPath(name string) string {
	return fmt.Sprintf("github.com/primandproper/dinnerdonebetter/backend/internal/domain/%s", name)
}
