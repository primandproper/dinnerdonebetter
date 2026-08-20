// Package queuemessages holds the message envelopes this application puts on
// its own topics.
//
// platform-go v9 removed the application-specific fields from its message types —
// email.OutboundEmailMessage's UserID and TestID — as consumer-app leakage in a library every
// other consumer also pays for. They are this application's fields, so they are declared here,
// embedding the platform type so the wire format is unchanged.
//
// The search index request envelope that used to live here went with v10, which replaced the
// single search-index-requests topic and its index-type discriminator with one topic per index
// carrying searchsync.Event. The index a message belongs to is now the topic it arrived on.
package queuemessages

import (
	"github.com/primandproper/platform-go/v12/email"
)

type (
	// OutboundEmailMessage is an outbound email plus the fields the
	// outbound-emails topic carries.
	OutboundEmailMessage struct {
		email.OutboundEmailMessage

		// UserID is the user the email is about, recorded for correlation.
		UserID string `json:"userID,omitempty"`

		// TestID, when non-empty, marks the message as a queue-test probe: the
		// consumer acknowledges it and skips the business logic.
		TestID string `json:"testID,omitempty"`
	}
)
