package internalops

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
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
	// There is no webhook_execution_requests topic and no user_data_aggregation topic any
	// more. A webhook delivery and a data privacy export are both rows a worker claims rather
	// than messages on a broker, so there is no queue here to probe.
	case "mobile_notifications":
		return &notifications.MobileNotificationRequest{TestID: testID, Title: "test", Body: "test", RequestType: "test"}, nil
	default:
		return nil, fmt.Errorf("unknown queue: %s", topicName)
	}
}
