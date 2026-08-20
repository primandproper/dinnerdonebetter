package main

import (
	"testing"

	"github.com/primandproper/platform-go/v12/database/querygen"

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

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := []string{
			"webhooks.id",
			"webhooks.name",
			"webhooks.content_type",
			"webhooks.url",
			"webhooks.method",
			"webhook_trigger_configs.id",
			"webhook_trigger_configs.trigger_event",
			"webhook_trigger_configs.belongs_to_webhook",
			"webhook_trigger_configs.created_at",
			"webhook_trigger_configs.archived_at",
			"webhooks.created_at",
			"webhooks.last_updated_at",
			"webhooks.archived_at",
			"webhooks.created_by_user",
			"webhooks.belongs_to_account",
		}

		actual := mergeColumns(
			applyToEach(webhooksColumns, func(_ int, s string) string {
				return querygen.Qualify(webhooksTableName, s)
			}),
			applyToEach(webhookTriggerConfigsColumns, func(_ int, s string) string {
				return querygen.Qualify(webhookTriggerConfigsTableName, s)
			}),
			5,
		)

		assert.Equal(t, expected, actual)
	})
}
