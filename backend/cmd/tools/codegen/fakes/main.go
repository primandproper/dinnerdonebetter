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
// resolves it; a webhook's event type cannot be random, because the catalog rejects anything it
// does not know. Those are per-field overrides in internal/domain/*/entities.go, each carrying the
// sentence explaining it, and both the expression and the sentence are emitted into the generated
// source. A generator that only did the mechanical part would produce fakes that fail validation
// everywhere, which is a worse outcome than the duplication it removed.
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

	for _, d := range domains {
		dir := filepath.Join(root, packageDir(d.name))

		rendered, renderErr := renderDomain(d, dir)
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
