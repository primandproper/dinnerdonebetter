package main

import (
	"testing"

	"github.com/primandproper/primitives-go/v2/database/querygen"

	"github.com/cristalhq/builq"
	"github.com/stretchr/testify/assert"
)

func Test_applyToEach(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := []string{
			"things",
			"and",
			"stuff",
		}

		callCount := 0
		exampleFunc := func(_ int, x string) string {
			callCount += 1
			return x
		}

		expected := []string{
			"things",
			"and",
			"stuff",
		}
		actual := applyToEach(exampleInput, exampleFunc)

		assert.Len(t, exampleInput, callCount)
		assert.Equal(t, expected, actual)
	})
}

func Test_buildRawQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var whatever builq.Builder

		builder := whatever.Addf("SELECT * FROM things")

		expected := "SELECT * FROM things"
		actual := buildRawQuery(builder)

		assert.Equal(t, expected, actual)
	})
}

func Test_mergeColumns(T *testing.T) {
	T.Parallel()

	// Recipe prep tasks and their steps, which is the same parent-and-children shape the
	// webhooks tables had before migration 45 dropped them. What is under test is where the
	// second set lands, not either table's schema.
	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := []string{
			"recipe_prep_tasks.id",
			"recipe_prep_tasks.name",
			"recipe_prep_tasks.description",
			"recipe_prep_tasks.notes",
			"recipe_prep_tasks.optional",
			"recipe_prep_task_steps.id",
			"recipe_prep_task_steps.belongs_to_recipe_step",
			"recipe_prep_task_steps.belongs_to_recipe_prep_task",
			"recipe_prep_task_steps.satisfies_recipe_step",
			"recipe_prep_tasks.explicit_storage_instructions",
			"recipe_prep_tasks.minimum_time_buffer_before_recipe_in_seconds",
			"recipe_prep_tasks.maximum_time_buffer_before_recipe_in_seconds",
			"recipe_prep_tasks.storage_type",
			"recipe_prep_tasks.minimum_storage_temperature_in_celsius",
			"recipe_prep_tasks.maximum_storage_temperature_in_celsius",
			"recipe_prep_tasks.created_at",
			"recipe_prep_tasks.last_updated_at",
			"recipe_prep_tasks.archived_at",
			"recipe_prep_tasks.belongs_to_recipe",
		}

		actual := mergeColumns(
			applyToEach(recipePrepTasksColumns, func(_ int, s string) string {
				return querygen.Qualify(recipePrepTasksTableName, s)
			}),
			applyToEach(recipePrepTaskStepsColumns, func(_ int, s string) string {
				return querygen.Qualify(recipePrepTaskStepsTableName, s)
			}),
			5,
		)

		assert.Equal(t, expected, actual)
	})
}
