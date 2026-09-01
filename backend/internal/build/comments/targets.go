/*
Package comments assembles the catalog of things this application accepts
comments on.

It lives in the build layer because it is the one place that may know both
halves. platform-go's comment store cannot see what a comment is about — the
rows live in tables it has never been shown — so it takes the vocabulary as a
parameter. The domains that own those rows do not know about comments either,
and should not: comments is generic machinery, and a target type belongs to
whoever is being commented on. The injector already holds both, so the catalog is
assembled here rather than by making one side import the other.
*/
package comments

import (
	"context"
	"database/sql"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/samber/do/v2"
)

// Catalog is what this application accepts comments on, with no existence checks.
//
// It is the catalog for a process that reads and erases comments but never
// writes one — the scheduler fulfilling a data privacy request, or the data
// change handler. The catalog gates writes rather than reads, so a hookless one
// is exactly right there, and it still refuses a misspelled type should a write
// path arrive later.
func Catalog() platformcomments.Targets {
	return platformcomments.Targets{
		mealplanning.CommentTargetTypeRecipes:      {Description: "A recipe."},
		mealplanning.CommentTargetTypeMeals:        {Description: "A meal."},
		mealplanning.CommentTargetTypeMealPlans:    {Description: "A meal plan."},
		issuereports.CommentTargetTypeIssueReports: {Description: "An issue report."},
	}
}

// CatalogWithChecks is Catalog with an existence check on every type whose owning
// domain can answer "is this there" from the scope and the ID alone.
//
// Two types are deliberately left unchecked, for the same reason from opposite
// directions: the check platform runs is handed the comment's scope and nothing
// else. Reading a meal plan takes an owner as well as an ID, and reading an issue
// report takes the account it was filed under — while comments in this deployment
// are all filed globally, so the scope the hook receives is not one either read
// can use. Both services read their target as the caller before they delegate,
// which is a stronger check than this one would be rather than a missing one.
//
// A check narrows the window in which a comment can be written about something
// that is not there; it does not close it. A target deleted between the check and
// the insert is still a comment about nothing.
func CatalogWithChecks(
	mealPlanning mealplanningmanagers.MealPlanningManager,
) platformcomments.Targets {
	catalog := Catalog()

	catalog[mealplanning.CommentTargetTypeRecipes] = withCheck(
		catalog[mealplanning.CommentTargetTypeRecipes],
		func(ctx context.Context, targetID string) error {
			_, err := mealPlanning.ReadRecipe(ctx, targetID)

			return err
		},
	)

	catalog[mealplanning.CommentTargetTypeMeals] = withCheck(
		catalog[mealplanning.CommentTargetTypeMeals],
		func(ctx context.Context, targetID string) error {
			_, err := mealPlanning.ReadMeal(ctx, targetID)

			return err
		},
	)

	return catalog
}

// withCheck turns a domain read into the existence hook platform's catalog takes.
//
// A read that failed because the row is not there is "absent"; any other failure
// is an error, and the two are kept apart deliberately. Platform's hook says an
// error is not absent, because a hook that decided an unreachable table meant a
// missing target would refuse writes the caller should have been told to retry.
func withCheck(definition platformcomments.TargetDefinition, read func(ctx context.Context, targetID string) error) platformcomments.TargetDefinition {
	definition.Exists = func(ctx context.Context, _ tenancy.Scope, targetID string) (bool, error) {
		if err := read(ctx, targetID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}

			return false, err
		}

		return true, nil
	}

	return definition
}

// RegisterTargets registers the catalog carrying existence checks, for a process
// that writes comments.
func RegisterTargets(i do.Injector) {
	do.Provide[platformcomments.Targets](i, func(i do.Injector) (platformcomments.Targets, error) {
		return CatalogWithChecks(
			do.MustInvoke[mealplanningmanagers.MealPlanningManager](i),
		), nil
	})
}

// RegisterReadOnlyTargets registers the catalog without existence checks, for a
// process that reads and erases comments but never writes one.
func RegisterReadOnlyTargets(i do.Injector) {
	do.Provide[platformcomments.Targets](i, func(do.Injector) (platformcomments.Targets, error) {
		return Catalog(), nil
	})
}
