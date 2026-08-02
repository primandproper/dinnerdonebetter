package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderMemory is the in-memory session store provider.
	ProviderMemory = "memory"
	// ProviderPostgres is the PostgreSQL session store provider.
	ProviderPostgres = "postgres"
)

type (
	// Config is the configuration for the WebAuthn session store.
	Config struct {
		Provider string `env:"PROVIDER" json:"provider"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In("", ProviderMemory, ProviderPostgres)),
	)
}

// ProvideSessionStore provides a SessionStore based on the configured provider.
func ProvideSessionStore(cfg *Config, client database.Client, logger logging.Logger, tracerProvider tracing.TracerProvider) (webauthn.SessionStore, error) {
	provider := strings.TrimSpace(strings.ToLower(cfg.Provider))
	if provider == "" {
		provider = ProviderMemory
	}
	switch provider {
	case ProviderMemory:
		return webauthn.NewInMemorySessionStore(), nil
	case ProviderPostgres:
		if client == nil {
			return nil, fmt.Errorf("database client required for postgres session store provider")
		}
		return webauthn.NewPostgresSessionStore(client, logger, tracerProvider), nil
	default:
		return nil, fmt.Errorf("invalid session store provider: %q", cfg.Provider)
	}
}
