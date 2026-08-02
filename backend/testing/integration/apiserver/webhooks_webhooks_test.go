package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkWebhookEquality(t *testing.T, expected, actual *webhooks.Webhook) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected Webhook to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected Webhook to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected Webhook Name")
	assert.Equal(t, expected.URL, actual.URL, "expected Webhook URL")
	assert.Equal(t, expected.Method, actual.Method, "expected Webhook Method")
	assert.Equal(t, expected.ContentType, actual.ContentType, "expected Webhook ContentType")
	assert.NotEmpty(t, actual.BelongsToAccount, "expected Webhook to have BelongsToAccount")

	require.Equal(t, len(expected.TriggerConfigs), len(actual.TriggerConfigs), "expected Webhook TriggerConfigs length")
	for i, expectedCfg := range expected.TriggerConfigs {
		if i >= len(actual.TriggerConfigs) {
			continue
		}
		actualCfg := actual.TriggerConfigs[i]
		assert.NotEmpty(t, actualCfg.ID, "expected Webhook TriggerConfig %d to have ID", i)
		assert.NotZero(t, actualCfg.CreatedAt, "expected Webhook TriggerConfig %d to have CreatedAt", i)
		assert.Equal(t, expectedCfg.TriggerEventID, actualCfg.TriggerEventID, "expected Webhook TriggerConfig %d TriggerEventID", i)
		assert.Equal(t, actual.ID, actualCfg.BelongsToWebhook, "expected Webhook TriggerConfig %d BelongsToWebhook", i)
	}
}

func createWebhookTriggerEventCatalogForTest(t *testing.T, ctx context.Context, testClient client.Client, name, description string) *webhookssvc.WebhookTriggerEvent {
	t.Helper()
	resp, err := testClient.CreateWebhookTriggerEvent(ctx, &webhookssvc.CreateWebhookTriggerEventRequest{
		Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{Name: name, Description: description},
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Created)
	return resp.Created
}

func createWebhookForTest(t *testing.T, testClient client.Client) *webhooks.Webhook {
	t.Helper()
	ctx := t.Context()

	catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, testClient, "test_trigger", "for integration test")
	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.TriggerConfigs[0].TriggerEventID = catalogEvent.Id
	exampleWebhookInput := converters.ConvertWebhookToWebhookCreationRequestInput(exampleWebhook)

	input := grpcconverters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(exampleWebhookInput)

	createdWebhook, err := testClient.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{Input: input})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCWebhookToWebhook(createdWebhook.Created)
	checkWebhookEquality(t, exampleWebhook, converted)

	retrievedWebhook, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedWebhook)

	webhook := grpcconverters.ConvertGRPCWebhookToWebhook(retrievedWebhook.Result)
	checkWebhookEquality(t, converted, webhook)

	return webhook
}

func TestWebhooks_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "webhooks", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleWebhookInput := &webhooks.WebhookCreationRequestInput{
			ContentType: "application/whatever",
			Method:      "UNRECOGNIZED",
			Name:        t.Name(),
			URL:         "invalid protocol :\\ neato.ai",
			Events:      []*webhooks.WebhookTriggerEventCreationRequestInput{},
		}

		input := grpcconverters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(exampleWebhookInput)

		_, err := testClient.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{Input: input})
		assert.Error(t, err)
	})
}

func TestWebhooks_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		retrieved, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{})
		assert.Error(t, err)
	})
}

func TestWebhooks_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdWebhooks := []*webhooks.Webhook{}
		for range exampleQuantity {
			createdWebhooks = append(createdWebhooks, createWebhookForTest(t, testClient))
		}

		results, err := testClient.GetWebhooks(ctx, &webhookssvc.GetWebhooksRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdWebhooks))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWebhooks(ctx, &webhookssvc.GetWebhooksRequest{})
		assert.Error(t, err)
	})
}

func TestWebhooks_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		_, err := testClient.ArchiveWebhook(ctx, &webhookssvc.ArchiveWebhookRequest{WebhookId: createdWebhook.ID})
		assert.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "webhooks", RelevantID: createdWebhook.ID},
		})
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.ArchiveWebhook(ctx, &webhookssvc.ArchiveWebhookRequest{WebhookId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWebhook(ctx, &webhookssvc.ArchiveWebhookRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Adding(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)
		catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, testClient, "webhook_archived", "when webhook is archived")

		addedConfig, err := testClient.AddWebhookTriggerConfig(ctx, &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: createdWebhook.ID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: createdWebhook.ID,
				TriggerEventId:   catalogEvent.Id,
			},
		})
		assert.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "webhook_trigger_configs", RelevantID: addedConfig.Created.Id},
		})
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.AddWebhookTriggerConfig(ctx, &webhookssvc.AddWebhookTriggerConfigRequest{WebhookId: nonexistentID, Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{}})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.AddWebhookTriggerConfig(ctx, &webhookssvc.AddWebhookTriggerConfigRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerConfigs_Removing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)
		catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, testClient, "webhook_archived", "when webhook is archived")

		createdTriggerConfig, err := testClient.AddWebhookTriggerConfig(ctx, &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: createdWebhook.ID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: createdWebhook.ID,
				TriggerEventId:   catalogEvent.Id,
			},
		})
		require.NoError(t, err)

		_, err = testClient.ArchiveWebhookTriggerConfig(ctx, &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              createdWebhook.ID,
			WebhookTriggerConfigId: createdTriggerConfig.Created.Id,
		})
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.ArchiveWebhookTriggerConfig(ctx, &webhookssvc.ArchiveWebhookTriggerConfigRequest{WebhookId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWebhookTriggerConfig(ctx, &webhookssvc.ArchiveWebhookTriggerConfigRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, testClient, "test_read_event", "for reading test")

		retrieved, err := testClient.GetWebhookTriggerEvent(ctx, &webhookssvc.GetWebhookTriggerEventRequest{Id: catalogEvent.Id})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, catalogEvent.Id, retrieved.Result.Id)
		assert.Equal(t, catalogEvent.Name, retrieved.Result.Name)
		assert.Equal(t, catalogEvent.Description, retrieved.Result.Description)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetWebhookTriggerEvent(ctx, &webhookssvc.GetWebhookTriggerEventRequest{Id: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWebhookTriggerEvent(ctx, &webhookssvc.GetWebhookTriggerEventRequest{Id: nonexistentID})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		for range exampleQuantity {
			createWebhookTriggerEventCatalogForTest(t, ctx, testClient, fmt.Sprintf("list_event_%d", time.Now().UnixNano()), "for listing test")
		}

		results, err := testClient.GetWebhookTriggerEvents(ctx, &webhookssvc.GetWebhookTriggerEventsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, len(results.Results) >= exampleQuantity)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetWebhookTriggerEvents(ctx, &webhookssvc.GetWebhookTriggerEventsRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, testClient, "test_update_event", "for updating test")

		_, err := testClient.UpdateWebhookTriggerEvent(ctx, &webhookssvc.UpdateWebhookTriggerEventRequest{
			Id: catalogEvent.Id,
			Input: &webhookssvc.WebhookTriggerEventUpdateRequestInput{
				Name:        "updated_event_name",
				Description: "updated description",
			},
		})
		assert.NoError(t, err)

		retrieved, err := testClient.GetWebhookTriggerEvent(ctx, &webhookssvc.GetWebhookTriggerEventRequest{Id: catalogEvent.Id})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "updated_event_name", retrieved.Result.Name)
		assert.Equal(t, "updated description", retrieved.Result.Description)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.UpdateWebhookTriggerEvent(ctx, &webhookssvc.UpdateWebhookTriggerEventRequest{
			Id: nonexistentID,
			Input: &webhookssvc.WebhookTriggerEventUpdateRequestInput{
				Name:        "doesn't matter",
				Description: "doesn't matter",
			},
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateWebhookTriggerEvent(ctx, &webhookssvc.UpdateWebhookTriggerEventRequest{
			Id: nonexistentID,
			Input: &webhookssvc.WebhookTriggerEventUpdateRequestInput{
				Name:        "doesn't matter",
				Description: "doesn't matter",
			},
		})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, testClient, "test_archive_event", "for archiving test")

		_, err := testClient.ArchiveWebhookTriggerEvent(ctx, &webhookssvc.ArchiveWebhookTriggerEventRequest{Id: catalogEvent.Id})
		assert.NoError(t, err)

		retrieved, err := testClient.GetWebhookTriggerEvent(ctx, &webhookssvc.GetWebhookTriggerEventRequest{Id: catalogEvent.Id})
		assert.Nil(t, retrieved)
		assert.Error(t, err)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveWebhookTriggerEvent(ctx, &webhookssvc.ArchiveWebhookTriggerEventRequest{Id: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveWebhookTriggerEvent(ctx, &webhookssvc.ArchiveWebhookTriggerEventRequest{})
		assert.Error(t, err)
	})
}
