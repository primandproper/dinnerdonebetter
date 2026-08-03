package internalops

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	notifications "github.com/primandproper/platform-go/v9/notifications/mobile"
)

// BuildQueueTestMessage returns a message with TestID set for the given topic. Non-empty TestID triggers queue test behavior.
func BuildQueueTestMessage(topicName, testID, userID string) (any, error) {
	switch topicName {
	case "data_changes":
		return &audit.DataChangeMessage{TestID: testID, UserID: userID}, nil
	case "outbound_emails":
		return &queuemessages.OutboundEmailMessage{TestID: testID, UserID: userID}, nil
	case "search_index_requests":
		return &queuemessages.IndexRequest{TestID: testID}, nil
	// There is no webhook_execution_requests topic any more. Webhook deliveries are rows
	// written inside the transaction that caused them and claimed by a worker, not messages on
	// a broker, so there is no queue here to probe.
	case "user_data_aggregation", "user_data_aggregation_requests":
		return &dataprivacy.UserDataAggregationRequest{TestID: testID, UserID: userID}, nil
	case "mobile_notifications":
		return &notifications.MobileNotificationRequest{TestID: testID, Title: "test", Body: "test", RequestType: "test"}, nil
	default:
		return nil, fmt.Errorf("unknown queue: %s", topicName)
	}
}
