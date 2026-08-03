package config

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// MetaSettings is primarily used for development.
type MetaSettings struct {
	RunMode runMode `env:"RUN_MODE" json:"runMode,omitempty"`
	Debug   bool    `env:"DEBUG"    json:"debug,omitempty"`
}

var _ validation.ValidatableWithContext = (*MetaSettings)(nil)

// ValidateWithContext validates a MetaSettings struct.
func (s *MetaSettings) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, s,
		validation.Field(&s.RunMode, validation.In(TestingRunMode, DevelopmentRunMode, ProductionRunMode)),
	)
}
