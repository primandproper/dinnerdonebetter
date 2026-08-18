package converters

import (
	"log"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	"github.com/primandproper/platform-go/v11/encoding"
	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertStringToWebhookContentType converts a content type string to its proto enum.
//
// JSON is the only value either side has. An unknown string is reported and mapped to JSON
// rather than rejected here, because this runs on the read path over rows written before XML was
// retired — refusing would make an old row unreadable rather than merely mislabeled. Writes are
// refused, by validation, before they get this far.
func ConvertStringToWebhookContentType(s string) webhookssvc.WebhookContentType {
	if s != encoding.ContentTypeJSON.String() {
		log.Printf("unsupported webhook content type: %q", s)
	}

	return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON
}

// ConvertWebhookContentTypeToString converts the proto enum to a content type string.
func ConvertWebhookContentTypeToString(s webhookssvc.WebhookContentType) string {
	if s != webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON {
		log.Printf("unsupported webhook content type: %q", s)
	}

	return encoding.ContentTypeJSON.String()
}

// ConvertStringToWebhookMethod converts an HTTP method string to its proto enum.
//
// Same asymmetry as the content type: rows written before the other methods were retired still
// read, as POST, because that is what they will actually be delivered with. A request asking for
// one of the others is refused by validation, not silently coerced here.
func ConvertStringToWebhookMethod(s string) webhookssvc.WebhookMethod {
	if s != webhooks.DeliveryMethod {
		log.Printf("unsupported webhook method: %q", s)
	}

	return webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST
}

// ConvertWebhookMethodToString converts the proto enum to an HTTP method string.
func ConvertWebhookMethodToString(s webhookssvc.WebhookMethod) string {
	if s != webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST {
		log.Printf("unsupported webhook method: %q", s)
	}

	return webhooks.DeliveryMethod
}

func ConvertWebhookToGRPCWebhook(webhook *webhooks.Webhook) *webhookssvc.Webhook {
	converted := &webhookssvc.Webhook{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(webhook.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(webhook.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(webhook.LastUpdatedAt),
		Name:             webhook.Name,
		Url:              webhook.URL,
		Method:           ConvertStringToWebhookMethod(webhook.Method),
		Id:               webhook.ID,
		BelongsToAccount: webhook.BelongsToAccount,
		ContentType:      ConvertStringToWebhookContentType(webhook.ContentType),
		CreatedByUser:    webhook.CreatedByUser,
	}
	for _, cfg := range webhook.TriggerConfigs {
		converted.TriggerConfigs = append(converted.TriggerConfigs, ConvertWebhookTriggerConfigToGRPCWebhookTriggerConfig(cfg))
	}
	return converted
}

// ConvertWebhookTriggerConfigToGRPCWebhookTriggerConfig converts a domain WebhookTriggerConfig to proto.
func ConvertWebhookTriggerConfigToGRPCWebhookTriggerConfig(z *webhooks.WebhookTriggerConfig) *webhookssvc.WebhookTriggerConfig {
	if z == nil {
		return nil
	}
	return &webhookssvc.WebhookTriggerConfig{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(z.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(z.ArchivedAt),
		Id:               z.ID,
		BelongsToWebhook: z.BelongsToWebhook,
		EventType:        z.EventType,
	}
}

// ConvertWebhookEventTypeToGRPCWebhookEventType converts a catalog entry to proto.
func ConvertWebhookEventTypeToGRPCWebhookEventType(z *webhooks.WebhookEventType) *webhookssvc.WebhookEventType {
	if z == nil {
		return nil
	}
	return &webhookssvc.WebhookEventType{
		Type:        z.Type,
		Description: z.Description,
	}
}

func ConvertGRPCWebhookToWebhook(webhook *webhookssvc.Webhook) *webhooks.Webhook {
	converted := &webhooks.Webhook{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(webhook.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(webhook.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertPBTimestampToTimePointer(webhook.LastUpdatedAt),
		Name:             webhook.Name,
		URL:              webhook.Url,
		Method:           ConvertWebhookMethodToString(webhook.Method),
		ContentType:      ConvertWebhookContentTypeToString(webhook.ContentType),
		ID:               webhook.Id,
		BelongsToAccount: webhook.BelongsToAccount,
		CreatedByUser:    webhook.CreatedByUser,
	}
	for _, cfg := range webhook.TriggerConfigs {
		converted.TriggerConfigs = append(converted.TriggerConfigs, ConvertGRPCWebhookTriggerConfigToWebhookTriggerConfig(cfg))
	}
	return converted
}

// ConvertGRPCWebhookTriggerConfigToWebhookTriggerConfig converts a proto WebhookTriggerConfig to domain.
func ConvertGRPCWebhookTriggerConfigToWebhookTriggerConfig(z *webhookssvc.WebhookTriggerConfig) *webhooks.WebhookTriggerConfig {
	if z == nil {
		return nil
	}
	return &webhooks.WebhookTriggerConfig{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(z.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(z.ArchivedAt),
		ID:               z.Id,
		BelongsToWebhook: z.BelongsToWebhook,
		EventType:        z.EventType,
	}
}

func ConvertGRPCWebhookCreationRequestInputToWebhookCreationRequestInput(input *webhookssvc.WebhookCreationRequestInput) *webhooks.WebhookCreationRequestInput {
	if input == nil {
		return nil
	}
	return &webhooks.WebhookCreationRequestInput{
		Name:        input.Name,
		ContentType: ConvertWebhookContentTypeToString(input.ContentType),
		URL:         input.Url,
		Method:      ConvertWebhookMethodToString(input.Method),
		Events:      input.GetEventTypes(),
	}
}

func ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(input *webhooks.WebhookCreationRequestInput) *webhookssvc.WebhookCreationRequestInput {
	return &webhookssvc.WebhookCreationRequestInput{
		Name:        input.Name,
		ContentType: ConvertStringToWebhookContentType(input.ContentType),
		Url:         input.URL,
		Method:      ConvertStringToWebhookMethod(input.Method),
		EventTypes:  input.Events,
	}
}

// ConvertGRPCWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput converts proto AddWebhookTriggerConfig input to domain DB input.
func ConvertGRPCWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput(input *webhookssvc.WebhookTriggerConfigCreationRequestInput) *webhooks.WebhookTriggerConfigDatabaseCreationInput {
	if input == nil {
		return nil
	}
	return &webhooks.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: input.BelongsToWebhook,
		EventType:        input.EventType,
	}
}

// ConvertUserDataCollectionToGRPCDataCollection converts a domain webhooks UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *webhooks.UserDataCollection) *webhookssvc.DataCollection {
	result := &webhookssvc.DataCollection{
		Webhooks: make(map[string]*webhookssvc.WebhookList),
	}

	for accountID, webhookList := range input.Data {
		var grpcWebhooks []*webhookssvc.Webhook
		for i := range webhookList {
			grpcWebhooks = append(grpcWebhooks, ConvertWebhookToGRPCWebhook(&webhookList[i]))
		}
		result.Webhooks[accountID] = &webhookssvc.WebhookList{Webhooks: grpcWebhooks}
	}

	return result
}
