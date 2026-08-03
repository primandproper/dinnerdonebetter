/*
Package dataprivacy wires this application's domains into platform-go's data
privacy registry.

This file is the whole cost of adding a domain to a subject access request: one
line in buildRegistry, naming a key and a collector that lives beside the domain
it collects. It replaces a struct that imported every domain package and that
every domain wrote into, where the same change meant editing a central type — the
file most likely to conflict — and where one domain's error aborted the entire
export.

It still imports every domain, and that is not the same thing. A list of
registrations has no shared state for two domains to fight over, no field order
for a merge to get wrong, and no way for a failure in one to reach another: the
Worker records a failed collector against its own key and delivers the rest of
the artifact with a manifest saying what is missing.
*/
package dataprivacy

import (
	"context"

	auditdomain "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/privacy"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediaprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhooksprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/privacy"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	"github.com/primandproper/platform-go/v9/database"
	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
	"github.com/primandproper/platform-go/v9/dataprivacy/auditerasure"
	platformdataprivacycfg "github.com/primandproper/platform-go/v9/dataprivacy/config"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterRegistry registers the collector and eraser registry with the injector.
//
// Prerequisites: every domain repository named in buildRegistry, plus *Config and
// database.Client for the audit eraser's policy.
func RegisterRegistry(i do.Injector) {
	do.Provide(i, buildRegistry)
}

// buildRegistry assembles the registry. Adding a domain to an export means adding
// a line here and a collector beside that domain.
func buildRegistry(i do.Injector) (*platformdataprivacy.Registry, error) {
	var (
		ctx            = do.MustInvoke[context.Context](i)
		logger         = do.MustInvoke[logging.Logger](i)
		tracerProvider = do.MustInvoke[tracing.TracerProvider](i)
		identityRepo   = do.MustInvoke[identity.Repository](i)
		registry       = platformdataprivacy.NewRegistry()
	)

	// Which accounts a subject appears in, which is the one question an
	// account-scoped collector cannot answer from its own domain. Resolved once and
	// shared, so five collectors asking it do not become five identical page walks
	// per collector — they still each call it, but through one implementation whose
	// cost is visible in one place.
	resolveAccounts := identityprivacy.ResolveAccountIDs(identityRepo)

	collectors := map[string]platformdataprivacy.Collector{
		ddbdataprivacy.CollectorKeyIdentity: identityprivacy.NewCollector(
			identityRepo, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyMealPlanning: mealplanningprivacy.NewCollector(
			do.MustInvoke[mealplanning.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyWebhooks: webhooksprivacy.NewCollector(
			do.MustInvoke[webhooks.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeySettings: settingsprivacy.NewCollector(
			do.MustInvoke[settings.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyNotifications: notificationsprivacy.NewCollector(
			do.MustInvoke[notifications.Repository](i), logger, tracerProvider),
		ddbdataprivacy.CollectorKeyPayments: paymentsprivacy.NewCollector(
			do.MustInvoke[payments.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyAuditLog: auditprivacy.NewCollector(
			do.MustInvoke[auditdomain.Repository](i), logger, tracerProvider),
		ddbdataprivacy.CollectorKeyIssueReports: issuereportsprivacy.NewCollector(
			do.MustInvoke[issuereports.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyUploadedMedia: uploadedmediaprivacy.NewCollector(
			do.MustInvoke[uploadedmedia.Repository](i), logger, tracerProvider),
		ddbdataprivacy.CollectorKeyWaitlists: waitlistsprivacy.NewCollector(
			do.MustInvoke[waitlists.Repository](i), logger, tracerProvider),
		ddbdataprivacy.CollectorKeyComments: commentsprivacy.NewCollector(
			do.MustInvoke[comments.Repository](i), logger, tracerProvider),
	}

	for key, collector := range collectors {
		if err := registry.RegisterCollector(key, collector); err != nil {
			return nil, platformerrors.Wrapf(err, "registering %q data privacy collector", key)
		}
	}

	// One application eraser, because every belongs_to_user and belongs_to_account
	// foreign key in this schema cascades from the user row. See
	// internal/domain/identity/privacy for what that covers and what would make a
	// second one worth writing.
	if err := registry.RegisterEraser(
		ddbdataprivacy.EraserKeyIdentity,
		identityprivacy.NewEraser(identityRepo, logger, tracerProvider),
	); err != nil {
		return nil, platformerrors.Wrap(err, "registering identity data privacy eraser")
	}

	// The audit log is the one store the cascade cannot reach, because a hash chain
	// cannot carry a foreign key that removes rows from the middle of it. Whether it
	// is erased at all is a policy question with a different answer per jurisdiction,
	// so platform-go makes it a config flag rather than a code change — and reports
	// which way it went, because "did this deployment erase audit records" gets asked
	// long afterwards.
	registered, err := platformdataprivacycfg.RegisterAuditEraser(
		ctx,
		prepareConfig(i),
		registry,
		auditerasure.WithScopeResolver(auditprivacy.ErasableScopeResolver(identityRepo)),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "registering audit data privacy eraser")
	}

	logging.EnsureLogger(logger).
		WithValue("collectors", len(registry.CollectorKeys())).
		WithValue("erasers", len(registry.EraserKeys())).
		WithValue("audit_erasure_enabled", registered).
		Info("data privacy registry assembled")

	return registry, nil
}

// RegisterWorker registers the fulfillment loop with the injector.
//
// Prerequisites: RegisterRegistry, and dataprivacycfg.RegisterArtifactStorage.
func RegisterWorker(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*platformdataprivacy.Worker, error) {
		workerOpts, _ := platformdataprivacycfg.EnsurePackaging(
			do.MustInvoke[dataprivacycfg.ArtifactCompressor](i).Compressor,
			do.MustInvoke[dataprivacycfg.ArtifactEncryptorDecryptor](i).EncryptorDecryptor,
		)

		return platformdataprivacycfg.NewWorker(
			do.MustInvoke[context.Context](i),
			prepareConfig(i),
			do.MustInvoke[platformdataprivacy.Store](i),
			do.MustInvoke[*platformdataprivacy.Registry](i),
			do.MustInvoke[dataprivacycfg.ArtifactUploadManager](i).UploadManager,
			// Artifacts are encrypted, so no signed URL can be minted for one: the
			// stored object is ciphertext and a subject following that link would get a
			// file they cannot open. Saying so here is what stops the Worker attaching
			// a broken download link to a completion notification.
			true,
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
			platformdataprivacycfg.WithWorkerOptions(workerOpts...),
		)
	})
}

// RegisterSweeper registers the expiry and retention sweep with the injector.
//
// It is the half of this package a deployment can most easily forget to run, and
// the one whose absence is invisible: without it every export artifact ever written
// stays in the bucket forever, and nothing about the request rows suggests
// otherwise. It is registered as a scheduled job rather than a loop of its own — see
// internal/build/jobs/scheduler.
//
// Prerequisites: RegisterArtifactStorage.
func RegisterSweeper(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*platformdataprivacy.Sweeper, error) {
		return platformdataprivacycfg.NewSweeper(
			do.MustInvoke[context.Context](i),
			prepareConfig(i),
			do.MustInvoke[platformdataprivacy.Store](i),
			do.MustInvoke[dataprivacycfg.ArtifactUploadManager](i).UploadManager,
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// prepareConfig resolves the platform config with the dialect and table prefixes
// pinned, the same way every other consumer in this process does.
func prepareConfig(i do.Injector) *platformdataprivacycfg.Config {
	return dataprivacycfg.PlatformConfig(
		do.MustInvoke[*dataprivacycfg.Config](i),
		do.MustInvoke[database.Client](i),
	)
}
