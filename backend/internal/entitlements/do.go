package entitlements

import (
	"github.com/primandproper/platform-go/v13/billing"
	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"
	entitlementscfg "github.com/primandproper/platform-go/v13/entitlements/config"
	platformmetering "github.com/primandproper/platform-go/v13/metering"

	"github.com/samber/do/v2"
)

// RegisterFeatures registers this application's feature declarations with the injector.
//
// The catalog reads them from there rather than calling Features itself, because the platform's
// catalog constructor takes them as an argument — features are code and plans are configuration,
// and the two arrive by different routes on purpose. See Features.
func RegisterFeatures(i do.Injector) {
	do.Provide[[]platformentitlements.Feature](i, func(do.Injector) ([]platformentitlements.Feature, error) {
		return Features(), nil
	})
}

// RegisterPlanSource registers the account-to-plan lookup with the injector.
//
// Prerequisites: a billing.Store, which is the only thing here that knows what an account has
// bought.
func RegisterPlanSource(i do.Injector) {
	do.Provide[platformentitlements.PlanSource](i, func(i do.Injector) (platformentitlements.PlanSource, error) {
		// Built into a variable and returned only once err is known to be nil: returning the
		// constructor's result straight through would register a non-nil PlanSource wrapping
		// a nil pointer whenever construction failed.
		source, err := NewPlanSource(do.MustInvoke[billing.Store](i))
		if err != nil {
			return nil, err
		}

		return source, nil
	})
}

// RegisterCatalog registers the feature and plan catalog with the injector.
//
// Prerequisites: *entitlementscfg.Config and the features RegisterFeatures provides.
func RegisterCatalog(i do.Injector) {
	entitlementscfg.RegisterCatalog(i)
}

// RegisterQuotaSource registers the catalog's limits as the limits metering enforces.
//
// It registers the same object twice, and the second registration is the point of the whole
// arrangement. metering asks a QuotaSource what a subject may consume and enforces whatever it
// is told; the catalog is what knows. Registered as metering.QuotaSource, the limit an account
// is shown by a Checker and the limit an Enforcer applies to it are by construction one number.
// Registered any other way there are two, and they will disagree.
//
// The platform's own registration deliberately stops short of the alias, because importing a
// package should not silently change what an enforcer elsewhere in the same container enforces.
// Here there is one enforcer and this is what it enforces, so the decision is made once.
//
// Prerequisites: *entitlementscfg.Config, the catalog, the plan source, and a
// *metering.Registry.
func RegisterQuotaSource(i do.Injector) {
	entitlementscfg.RegisterQuotaSource(i)

	do.Provide[platformmetering.QuotaSource](i, func(i do.Injector) (platformmetering.QuotaSource, error) {
		source, err := do.Invoke[*platformentitlements.QuotaSource](i)
		if err != nil {
			return nil, err
		}

		return source, nil
	})
}

// RegisterChecker registers the read path — "may this account use this feature, and how much is
// left" — with the injector.
//
// Nothing calls it yet, for the same reason nothing calls the metering enforcer: every plan
// grants every feature without a bound, so there is no decision to make. It is registered so
// that the day a limit goes into the configured plans, the change is a Check at one call site
// rather than a wiring exercise.
//
// One thing has to change with that first limit. No cache.Cache[entitlements.Assignment] is
// registered, so every Check resolves the account's plan through the billing store — a
// durable read on a request path, which is what the cache exists to avoid. The feature flag
// manager and the metering enforcer are both picked up if registered, and both are.
//
// Prerequisites: *entitlementscfg.Config, the catalog, and the plan source.
func RegisterChecker(i do.Injector) {
	entitlementscfg.RegisterChecker(i)
}
