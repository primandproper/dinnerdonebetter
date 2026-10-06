// Package queuescfg names the message queue topics this application publishes to
// and consumes from.
//
// These names lived in platform-go's messagequeue/config until v9, which removed
// them: they are one application's topics, and every other consumer of that module
// had to invent values for topics it does not have. They are application
// configuration, so they live here.
package queuescfg

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// DefaultDataChangesTopicName is the topic every data change event is published
// on when no configuration names one.
//
// It is one of the two topic names a process may need without a queues config: a
// process with no broker still writes its events to the outbox, under this
// topic, for the worker that has one to relay. See
// internal/recordingspine, and DefaultQueuedMailTopicName for the other.
const DefaultDataChangesTopicName = "data_changes"

// DefaultQueuedMailTopicName is the topic platform's QueuedMailer enqueues the mail of the
// identity, sign-in, password reset and waitlist doors on, when no configuration names one.
//
// It is needed without a queues config for DefaultDataChangesTopicName's reason. It must be a
// topic the mail Drainer alone consumes: every message on it carries a live credential. See
// internal/build/queuedmail.
const DefaultQueuedMailTopicName = "queued_mail"

type (
	// Config contains the various queue names.
	Config struct {
		_ struct{} `json:"-" yaml:"-"`

		DataChangesTopicName         string `env:"DATA_CHANGES_TOPIC_NAME"          json:"dataChangesTopicName,omitempty"         yaml:"dataChangesTopicName,omitempty"`
		OutboundEmailsTopicName      string `env:"OUTBOUND_EMAILS_TOPIC_NAME"       json:"outboundEmailsTopicName,omitempty"      yaml:"outboundEmailsTopicName,omitempty"`
		SearchIndexRequestsTopicName string `env:"SEARCH_INDEX_REQUESTS_TOPIC_NAME" json:"searchIndexRequestsTopicName,omitempty" yaml:"searchIndexRequestsTopicName,omitempty"`
		MobileNotificationsTopicName string `env:"MOBILE_NOTIFICATIONS_TOPIC_NAME"  json:"mobileNotificationsTopicName,omitempty" yaml:"mobileNotificationsTopicName,omitempty"`
		// QueuedMailTopicName is the mail Drainer's topic, and nothing else's: every message on
		// it carries the secret its mail exists to deliver.
		QueuedMailTopicName string `env:"QUEUED_MAIL_TOPIC_NAME" json:"queuedMailTopicName,omitempty" yaml:"queuedMailTopicName,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.DataChangesTopicName, validation.Required),
		validation.Field(&c.OutboundEmailsTopicName, validation.Required),
		validation.Field(&c.SearchIndexRequestsTopicName, validation.Required),
		validation.Field(&c.MobileNotificationsTopicName, validation.Required),
		validation.Field(&c.QueuedMailTopicName, validation.Required),
	)
}
