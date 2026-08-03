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

type (
	// Config contains the various queue names.
	Config struct {
		_ struct{} `json:"-" yaml:"-"`

		DataChangesTopicName         string `env:"DATA_CHANGES_TOPIC_NAME"          json:"dataChangesTopicName,omitempty"         yaml:"dataChangesTopicName,omitempty"`
		OutboundEmailsTopicName      string `env:"OUTBOUND_EMAILS_TOPIC_NAME"       json:"outboundEmailsTopicName,omitempty"      yaml:"outboundEmailsTopicName,omitempty"`
		SearchIndexRequestsTopicName string `env:"SEARCH_INDEX_REQUESTS_TOPIC_NAME" json:"searchIndexRequestsTopicName,omitempty" yaml:"searchIndexRequestsTopicName,omitempty"`
		MobileNotificationsTopicName string `env:"MOBILE_NOTIFICATIONS_TOPIC_NAME"  json:"mobileNotificationsTopicName,omitempty" yaml:"mobileNotificationsTopicName,omitempty"`
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
	)
}
