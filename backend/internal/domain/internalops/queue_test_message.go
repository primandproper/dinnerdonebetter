package internalops

import (
	"encoding/json"
	"fmt"

	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	"github.com/primandproper/platform-go/v15/webhooks"
	notifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
)

// testMessageMarker is the title, body and request type of the probe message this
// queue check enqueues; nothing reads it back apart from the check itself.
const testMessageMarker = "test"

// QueueTestProbe is the event type of the probe the queue check publishes on the data changes
// topic. Everything on that topic is a webhooks.Envelope naming its event, so the probe is one
// too, and the consumer recognizes it by this name and acknowledges it by the envelope's ID.
//
// It is not a webhook event: nothing emits it through the recording spine, it is in no catalog,
// and its name is deliberately not an *EventType constant, which is what the catalog generator
// collects.
const QueueTestProbe webhooks.EventType = "internalops.queue_test"

// BuildQueueTestMessage returns the probe message for the given topic, carrying testID where
// that topic's consumer looks for it.
func BuildQueueTestMessage(topicName, testID, userID string) (any, error) {
	switch topicName {
	case "data_changes":
		return &webhooks.Envelope{EventType: QueueTestProbe, ID: testID, Payload: json.RawMessage(`{}`)}, nil
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
