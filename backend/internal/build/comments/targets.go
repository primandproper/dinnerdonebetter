/*
Package comments assembles the catalog of things this application accepts
comments on.

It lives in the build layer because it is the one place that may know both
halves. platform-go's comment store cannot see what a comment is about — the
rows live in tables it has never been shown — so it takes the vocabulary as a
parameter. The domains that own those rows do not know about the store either,
and should not: comments is generic machinery, and a target type belongs to
whoever is being commented on. So each domain contributes its own entry — the
types it accepts comments on, and the existence check behind each where it can
answer one — and this package merges them.

# Tenancy

Every comment is filed under tenancy.Global(), and that is a decision rather
than a default. A comment is about a recipe, a meal, a meal plan or an issue
report, and the first of those is readable across accounts — so scoping
comments by account would make one recipe's discussion look different depending
on who was reading it, which is not what a discussion is. Household-private
targets are protected by the checks the owning service already runs before it
delegates here, not by the column.
*/
package comments

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"

	platformcomments "github.com/primandproper/platform-go/v15/comments"

	"github.com/samber/do/v2"
)

// contribution is one domain's entry in the catalog: its targets with no existence checks,
// and, for a process that writes comments, the same targets with the checks it can make.
//
// checked is nil for a domain that checks nothing, and its unchecked targets are used for
// both.
type contribution struct {
	targets func() platformcomments.Targets
	checked func(i do.Injector) platformcomments.Targets
}

// contributions is every domain's entry, and the one list both catalogs are built from.
func contributions() []contribution {
	return []contribution{
		{targets: issuereports.CommentTargets},
		// Domain: mealplanning
		{targets: mealplanningregistration.CommentTargets, checked: mealplanningregistration.CheckedCommentTargets},
	}
}

// Catalog is what this application accepts comments on, with no existence checks.
//
// It is the catalog for a process that reads and erases comments but never
// writes one — the scheduler fulfilling a data privacy request, or the data
// change handler. The catalog gates writes rather than reads, so a hookless one
// is exactly right there, and it still refuses a misspelled type should a write
// path arrive later.
func Catalog() (platformcomments.Targets, error) {
	catalog := platformcomments.Targets{}

	for _, c := range contributions() {
		if err := merge(catalog, c.targets()); err != nil {
			return nil, err
		}
	}

	return catalog, nil
}

// CatalogWithChecks is Catalog with an existence check on every type whose owning
// domain can answer "is this there" from the scope and the ID alone, resolved from i.
//
// A check narrows the window in which a comment can be written about something
// that is not there; it does not close it. A target deleted between the check and
// the insert is still a comment about nothing.
func CatalogWithChecks(i do.Injector) (platformcomments.Targets, error) {
	catalog := platformcomments.Targets{}

	for _, c := range contributions() {
		targets := c.targets
		if c.checked != nil {
			targets = func() platformcomments.Targets { return c.checked(i) }
		}

		if err := merge(catalog, targets()); err != nil {
			return nil, err
		}
	}

	return catalog, nil
}

// merge adds targets to catalog, refusing a type two domains both claim: a comment about one
// is a comment about something, and which something cannot be settled by list order.
func merge(catalog, targets platformcomments.Targets) error {
	for targetType, definition := range targets {
		if _, taken := catalog[targetType]; taken {
			return fmt.Errorf("comment target type %q is contributed twice", targetType)
		}

		catalog[targetType] = definition
	}

	return nil
}

// RegisterTargets registers the catalog carrying existence checks, for a process
// that writes comments.
func RegisterTargets(i do.Injector) {
	do.Provide[platformcomments.Targets](i, CatalogWithChecks)
}

// RegisterReadOnlyTargets registers the catalog without existence checks, for a
// process that reads and erases comments but never writes one.
func RegisterReadOnlyTargets(i do.Injector) {
	do.Provide[platformcomments.Targets](i, func(do.Injector) (platformcomments.Targets, error) {
		return Catalog()
	})
}
