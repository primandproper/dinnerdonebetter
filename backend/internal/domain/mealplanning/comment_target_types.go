package mealplanning

import (
	comments "github.com/primandproper/platform-go/v13/comments"
)

// The kinds of thing in this domain that a comment may be about.
//
// They are comments.TargetType rather than string so the set is discoverable by
// declared type: the comment catalog has to list every kind of thing the
// application accepts comments on, and a list kept by hand beside the constants
// that are its source of truth is a list that drifts.
const (
	CommentTargetTypeRecipes   comments.TargetType = "recipes"
	CommentTargetTypeMeals     comments.TargetType = "meals"
	CommentTargetTypeMealPlans comments.TargetType = "meal_plans"
)
