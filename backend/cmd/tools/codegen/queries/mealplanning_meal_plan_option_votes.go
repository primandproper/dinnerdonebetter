package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	mealPlanOptionVotesTableName = "meal_plan_option_votes"

	mealPlanOptionVoteIDColumn    = "meal_plan_option_vote_id"
	belongsToMealPlanOptionColumn = "belongs_to_meal_plan_option"
	abstainColumn                 = "abstain"
	byUserColumn                  = "by_user"
)

var mealPlanOptionVotesColumns = []string{
	idColumn,
	"rank",
	abstainColumn,
	notesColumn,
	byUserColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToMealPlanOptionColumn,
}

func buildMealPlanOptionVotesQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(mealPlanOptionVotesTableName, mealPlanOptionVotesColumns,
				querygen.WithEntity("MealPlanOptionVote", "MealPlanOptionVotes"),
				querygen.WithOwnership(belongsToMealPlanOptionColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckMealPlanOptionVoteExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		JOIN %s ON %s.%s=%s.%s
		JOIN %s ON %s.%s=%s.%s
		JOIN %s ON %s.%s=%s.%s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
						mealPlanOptionVotesTableName, idColumn,
						mealPlanOptionVotesTableName,
						mealPlanOptionsTableName, mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
						mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
						mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
						mealPlanOptionVotesTableName, archivedAtColumn,
						mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn,
						mealPlanOptionVotesTableName, idColumn, mealPlanOptionVoteIDColumn,
						mealPlanOptionsTableName, archivedAtColumn,
						mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventIDColumn,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
						mealPlanOptionsTableName, idColumn, mealPlanOptionIDColumn,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, idColumn, mealPlanIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealPlanOptionVotesForMealPlanOption",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(mealPlanOptionVotesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlanOptionVotesTableName, s)
						}), ",\n\t"),
						mealPlanOptionVotesTableName,
						mealPlanOptionsTableName, mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
						mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
						mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
						mealPlanOptionVotesTableName, archivedAtColumn,
						mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn,
						mealPlanOptionsTableName, archivedAtColumn,
						mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventIDColumn,
						mealPlanOptionsTableName, idColumn, mealPlanOptionIDColumn,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
						mealPlanEventsTableName, idColumn, mealPlanEventIDColumn,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, idColumn, mealPlanIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealPlanOptionVotes",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT 
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s
GROUP BY
	%s.%s,
	%s.%s,
	%s.%s,
	%s.%s
%s;`,
						strings.Join(applyToEach(mealPlanOptionVotesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlanOptionVotesTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(mealPlanOptionVotesTableName, mealPlanOptionVotesColumns, []string{}, "meal_plan_option_votes.belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)"),
						pgGen.TotalCountSelect(mealPlanOptionVotesTableName, mealPlanOptionVotesColumns, []string{}),
						mealPlanOptionVotesTableName,
						mealPlanOptionsTableName,
						mealPlanOptionVotesTableName,
						belongsToMealPlanOptionColumn,
						mealPlanOptionsTableName,
						idColumn,
						mealPlanEventsTableName,
						mealPlanOptionsTableName,
						belongsToMealPlanEventColumn,
						mealPlanEventsTableName,
						idColumn,
						mealPlansTableName,
						mealPlanEventsTableName,
						belongsToMealPlanColumn,
						mealPlansTableName,
						idColumn,
						pgGen.FilterConditions(mealPlanOptionVotesTableName, mealPlanOptionVotesColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn),
							"meal_plan_options.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanOptionsTableName, idColumn, mealPlanOptionIDColumn),
							"meal_plan_events.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanEventsTableName, idColumn, mealPlanEventIDColumn),
							"meal_plans.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlansTableName, idColumn, mealPlanIDColumn),
						),
						mealPlanOptionVotesTableName,
						idColumn,
						mealPlanOptionsTableName,
						idColumn,
						mealPlanEventsTableName,
						idColumn,
						mealPlansTableName,
						idColumn,
						pgGen.CursorLimitClause(mealPlanOptionVotesTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealPlanOptionVote",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(mealPlanOptionVotesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlanOptionVotesTableName, s)
						}), ",\n\t"),
						mealPlanOptionVotesTableName,
						mealPlanOptionsTableName, mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
						mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
						mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
						mealPlanOptionVotesTableName, archivedAtColumn,
						mealPlanOptionVotesTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn,
						mealPlanOptionVotesTableName, idColumn, mealPlanOptionVoteIDColumn,
						mealPlanOptionsTableName, archivedAtColumn,
						mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventIDColumn,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
						mealPlanOptionsTableName, idColumn, mealPlanOptionIDColumn,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, idColumn, mealPlanIDColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
