// Package queuemessages holds the message envelopes this application puts on
// its own topics.
//
// platform-go v9 removed the application-specific fields from its message
// types — email.OutboundEmailMessage's UserID and TestID, and
// textsearch.IndexRequest's TestID — as consumer-app leakage in a library every
// other consumer also pays for. They are this application's fields, so they are
// declared here, embedding the platform type so the wire format is unchanged.
package queuemessages

import (
	"github.com/primandproper/platform-go/v9/email"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
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

	// IndexRequest is a search index request plus the fields the
	// search-index-requests topic carries.
	IndexRequest struct {
		// TestID, when non-empty, marks the message as a queue-test probe: the
		// consumer acknowledges it and skips the business logic.
		TestID string `json:"testID,omitempty"`
		textsearch.IndexRequest
	}
)
