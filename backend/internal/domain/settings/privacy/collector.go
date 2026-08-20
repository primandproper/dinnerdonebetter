// Package privacy is the settings domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"

	platformdataprivacy "github.com/primandproper/platform-go/v12/dataprivacy"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const o11yName = "settings_privacy_collector"

// Collector collects setting configurations about a subject.
type Collector struct {
	repo            settings.Repository
	resolveAccounts dataprivacy.AccountIDResolver
	tracer          tracing.Tracer
	logger          logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the settings collector.
func NewCollector(
	repo settings.Repository,
	resolveAccounts dataprivacy.AccountIDResolver,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) *Collector {
	return &Collector{
		repo:            repo,
		resolveAccounts: resolveAccounts,
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:          logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span)

	userSettings, err := dataprivacy.CollectAllValues(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
		return c.repo.GetServiceSettingConfigurationsForUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user settings")
	}

	accountIDs, err := c.resolveAccounts(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "resolving accounts")
	}

	accountSettings, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetServiceSettingConfigurationsForAccount)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account settings")
	}

	held := len(userSettings) > 0 || len(accountSettings) > 0

	return dataprivacy.Fragment(held, &settings.UserDataCollection{
		UserSettings:    userSettings,
		AccountSettings: accountSettings,
	})
}
