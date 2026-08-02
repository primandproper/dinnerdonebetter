package converters

import (
	"log"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/identifiers"
)

func ConvertStringToWebhookContentType(s string) webhookssvc.WebhookContentType {
	switch s {
	case encoding.ContentTypeXML.String():
		return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_XML
	case encoding.ContentTypeJSON.String():
		return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON
	default:
		log.Printf("unknown content type: %q", s)
		return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON
	}
}

func ConvertWebhookContentTypeToString(s webhookssvc.WebhookContentType) string {
	switch s {
	case webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_XML:
		return encoding.ContentTypeXML.String()
	case webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON:
		return encoding.ContentTypeJSON.String()
	default:
		log.Printf("unknown content type: %q", s)
		return encoding.ContentTypeJSON.String()
	}
}

func ConvertStringToWebhookMethod(s string) webhookssvc.WebhookMethod {
	switch s {
	case http.MethodGet:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_GET
	case http.MethodPut:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_PUT
	case http.MethodPatch:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_PATCH
	case http.MethodDelete:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_DELETE
	case http.MethodPost:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST
	default:
		log.Printf("unknown webhook method: %q", s)
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST
	}
}

func ConvertWebhookMethodToString(s webhookssvc.WebhookMethod) string {
	switch s {
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_GET:
		return http.MethodGet
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_PUT:
		return http.MethodPut
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_PATCH:
		return http.MethodPatch
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_DELETE:
		return http.MethodDelete
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST:
		return http.MethodPost
	default:
		log.Printf("unknown webhook method: %q", s)
		return http.MethodPost
	}
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

// ConvertWebhookTriggerConfigToGRPCWebhookTriggerConfig converts domain join-table WebhookTriggerConfig to proto.
func ConvertWebhookTriggerConfigToGRPCWebhookTriggerConfig(z *webhooks.WebhookTriggerConfig) *webhookssvc.WebhookTriggerConfig {
	if z == nil {
		return nil
	}
	return &webhookssvc.WebhookTriggerConfig{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(z.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(z.ArchivedAt),
		Id:               z.ID,
		BelongsToWebhook: z.BelongsToWebhook,
		TriggerEventId:   z.TriggerEventID,
	}
}

// ConvertWebhookTriggerEventCatalogToGRPCWebhookTriggerEvent converts domain catalog WebhookTriggerEvent to proto.
func ConvertWebhookTriggerEventCatalogToGRPCWebhookTriggerEvent(z *webhooks.WebhookTriggerEvent) *webhookssvc.WebhookTriggerEvent {
	if z == nil {
		return nil
	}
	return &webhookssvc.WebhookTriggerEvent{
		Id:            z.ID,
		Name:          z.Name,
		Description:   z.Description,
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(z.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(z.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(z.ArchivedAt),
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

// ConvertGRPCWebhookTriggerConfigToWebhookTriggerConfig converts proto WebhookTriggerConfig to domain join-table type.
func ConvertGRPCWebhookTriggerConfigToWebhookTriggerConfig(z *webhookssvc.WebhookTriggerConfig) *webhooks.WebhookTriggerConfig {
	if z == nil {
		return nil
	}
	return &webhooks.WebhookTriggerConfig{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(z.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(z.ArchivedAt),
		ID:               z.Id,
		BelongsToWebhook: z.BelongsToWebhook,
		TriggerEventID:   z.TriggerEventId,
	}
}

func ConvertGRPCWebhookCreationRequestInputToWebhookCreationRequestInput(input *webhookssvc.WebhookCreationRequestInput) *webhooks.WebhookCreationRequestInput {
	if input == nil {
		return nil
	}
	events := make([]*webhooks.WebhookTriggerEventCreationRequestInput, 0, len(input.GetEvents()))
	for _, pev := range input.GetEvents() {
		if pev == nil {
			continue
		}
		events = append(events, convertGRPCWebhookTriggerEventCreationRequestInputToDomain(pev))
	}
	return &webhooks.WebhookCreationRequestInput{
		Name:        input.Name,
		ContentType: ConvertWebhookContentTypeToString(input.ContentType),
		URL:         input.Url,
		Method:      ConvertWebhookMethodToString(input.Method),
		Events:      events,
	}
}

func convertGRPCWebhookTriggerEventCreationRequestInputToDomain(pev *webhookssvc.WebhookTriggerEventCreationRequestInput) *webhooks.WebhookTriggerEventCreationRequestInput {
	if pev == nil {
		return nil
	}
	out := &webhooks.WebhookTriggerEventCreationRequestInput{
		ID:          identifiers.New(),
		Name:        pev.GetName(),
		Description: pev.GetDescription(),
	}
	if id := pev.GetId(); id != "" {
		out.ID = id
	}
	return out
}

func ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(input *webhooks.WebhookCreationRequestInput) *webhookssvc.WebhookCreationRequestInput {
	events := make([]*webhookssvc.WebhookTriggerEventCreationRequestInput, 0, len(input.Events))
	for _, ev := range input.Events {
		if ev == nil {
			continue
		}
		events = append(events, convertDomainWebhookTriggerEventCreationRequestInputToGRPC(ev))
	}
	return &webhookssvc.WebhookCreationRequestInput{
		Name:        input.Name,
		ContentType: ConvertStringToWebhookContentType(input.ContentType),
		Url:         input.URL,
		Method:      ConvertStringToWebhookMethod(input.Method),
		Events:      events,
	}
}

func convertDomainWebhookTriggerEventCreationRequestInputToGRPC(ev *webhooks.WebhookTriggerEventCreationRequestInput) *webhookssvc.WebhookTriggerEventCreationRequestInput {
	if ev == nil {
		return nil
	}
	out := &webhookssvc.WebhookTriggerEventCreationRequestInput{
		Name:        ev.Name,
		Description: ev.Description,
	}
	if ev.ID != "" {
		out.Id = &ev.ID
	}
	return out
}

// ConvertGRPCWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput converts proto AddWebhookTriggerConfig input to domain DB input.
func ConvertGRPCWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput(input *webhookssvc.WebhookTriggerConfigCreationRequestInput) *webhooks.WebhookTriggerConfigDatabaseCreationInput {
	if input == nil {
		return nil
	}
	return &webhooks.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: input.BelongsToWebhook,
		TriggerEventID:   input.TriggerEventId,
	}
}

// ConvertGRPCWebhookTriggerEventCreationRequestInputToWebhookTriggerEventCreationRequestInput converts proto catalog CreateWebhookTriggerEvent input to domain request input.
func ConvertGRPCWebhookTriggerEventCreationRequestInputToWebhookTriggerEventCreationRequestInput(input *webhookssvc.WebhookTriggerEventCreationRequestInput) *webhooks.WebhookTriggerEventCreationRequestInput {
	if input == nil {
		return nil
	}
	return &webhooks.WebhookTriggerEventCreationRequestInput{
		ID:          identifiers.New(),
		Name:        input.Name,
		Description: input.Description,
	}
}

// ConvertGRPCWebhookTriggerEventUpdateRequestInputToWebhookTriggerEventUpdateRequestInput converts proto catalog Update input to domain.
func ConvertGRPCWebhookTriggerEventUpdateRequestInputToWebhookTriggerEventUpdateRequestInput(input *webhookssvc.WebhookTriggerEventUpdateRequestInput) *webhooks.WebhookTriggerEventUpdateRequestInput {
	if input == nil {
		return nil
	}
	return &webhooks.WebhookTriggerEventUpdateRequestInput{
		Name:        input.Name,
		Description: input.Description,
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
