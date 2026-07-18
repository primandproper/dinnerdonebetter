package internalops

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v5/email"
	notifications "github.com/primandproper/platform-go/v5/notifications/mobile"
	textsearch "github.com/primandproper/platform-go/v5/search/text"
)

// BuildQueueTestMessage returns a message with TestID set for the given topic. Non-empty TestID triggers queue test behavior.
func BuildQueueTestMessage(topicName, testID, userID string) (any, error) {
	switch topicName {
	case "data_changes":
		return &audit.DataChangeMessage{TestID: testID, UserID: userID}, nil
	case "outbound_emails":
		return &email.OutboundEmailMessage{TestID: testID, UserID: userID}, nil
	case "search_index_requests":
		return &textsearch.IndexRequest{TestID: testID}, nil
	case "webhook_execution_requests":
		return &webhooks.WebhookExecutionRequest{TestID: testID}, nil
	case "user_data_aggregation", "user_data_aggregation_requests":
		return &dataprivacy.UserDataAggregationRequest{TestID: testID, UserID: userID}, nil
	case "mobile_notifications":
		return &notifications.MobileNotificationRequest{TestID: testID, Title: "test", Body: "test", RequestType: "test"}, nil
	default:
		return nil, fmt.Errorf("unknown queue: %s", topicName)
	}
}
