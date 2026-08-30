package internalops

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	notifications "github.com/primandproper/platform-go/v13/notifications/mobile"
)

// testMessageMarker is the title, body and request type of the probe message this
// queue check enqueues; nothing reads it back apart from the check itself.
const testMessageMarker = "test"

// BuildQueueTestMessage returns a message with TestID set for the given topic. Non-empty TestID triggers queue test behavior.
func BuildQueueTestMessage(topicName, testID, userID string) (any, error) {
	switch topicName {
	case "data_changes":
		return &audit.DataChangeMessage{TestID: testID, UserID: userID}, nil
	case "outbound_emails":
		return &queuemessages.OutboundEmailMessage{TestID: testID, UserID: userID}, nil
	// There is no search_index_requests topic, no webhook_execution_requests topic and no
	// user_data_aggregation topic any more. A webhook delivery and a data privacy export are
	// both rows a worker claims rather than messages on a broker, and platform-go v10 split
	// search indexing into one topic per index carrying searchsync.Event — a payload with
	// nowhere to put a TestID, since a Syncer reads the row named by the event rather than
	// anything the message carries. Probing those means probing an index, not a queue.
	case "mobile_notifications":
		return &notifications.MobileNotificationRequest{TestID: testID, Title: testMessageMarker, Body: testMessageMarker, RequestType: testMessageMarker}, nil
	default:
		return nil, fmt.Errorf("unknown queue: %s", topicName)
	}
}
