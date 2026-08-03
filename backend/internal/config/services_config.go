package config

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/config"
	mealplanningcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/config"
	oauthcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/config"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"

	"github.com/hashicorp/go-multierror"
)

type (
	// ServicesConfig collects the various service configurations.
	ServicesConfig struct {
		_ struct{} `json:"-"`

		Payments      paymentscfg.Config      `envPrefix:"PAYMENTS_"       json:"payments,omitzero"`
		Users         identitycfg.Config      `envPrefix:"USERS_"          json:"users,omitzero"`
		UploadedMedia uploadedmediacfg.Config `envPrefix:"UPLOADED_MEDIA_" json:"uploadedMedia,omitzero"`
		MealPlanning  mealplanningcfg.Config  `envPrefix:"MEAL_PLANNING_"  json:"mealPlanning,omitzero"`
		Auth          authentication.Config   `envPrefix:"AUTH_"           json:"auth,omitzero"`
		DataPrivacy   dataprivacycfg.Config   `envPrefix:"DATA_PRIVACY_"   json:"dataPrivacy,omitzero"`
		OAuth2Clients oauthcfg.Config         `envPrefix:"OAUTH2_CLIENTS_" json:"oauth2Clients,omitzero"`
	}
)

// ValidateWithContext validates a APIServiceConfig struct.
func (cfg *ServicesConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validatorsToRun := map[string]func(context.Context) error{
		"Users":         cfg.Users.ValidateWithContext,
		"DataPrivacy":   cfg.DataPrivacy.ValidateWithContext,
		"MealPlanning":  cfg.MealPlanning.ValidateWithContext,
		"OAuth2Clients": cfg.OAuth2Clients.ValidateWithContext,
	}

	for name, validator := range validatorsToRun {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
