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
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
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

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	commentsprivacy "github.com/primandproper/platform-go/v13/comments/privacy"
	"github.com/primandproper/platform-go/v13/database"
	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/dataprivacy/auditerasure"
	platformdataprivacycfg "github.com/primandproper/platform-go/v13/dataprivacy/config"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/operations"

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
		tracerProvider = do.MustInvoke[tracing.Provider](i)
		identityRepo   = do.MustInvoke[identity.Repository](i)
		registry       = platformdataprivacy.NewRegistry()
	)

	// Which accounts a subject appears in, which is the one question an
	// account-scoped collector cannot answer from its own domain. Resolved once and
	// shared, so five collectors asking it do not become five identical page walks
	// per collector — they still each call it, but through one implementation whose
	// cost is visible in one place.
	resolveAccounts := identityprivacy.ResolveAccountIDs(identityRepo)

	// Named rather than plain err: the registration blocks below each take an err of
	// their own, and an outer one declared here would make every one of them a shadow.
	commentsCollector, commentsEraser, commentsErr := commentsPrivacy(i)
	if commentsErr != nil {
		return nil, commentsErr
	}

	// Four of these take a repository and nothing else, because a collector whose
	// whole body is "page one list read and encode the rows" is
	// platformdataprivacy.CollectorFor and has no observability of its own to do.
	// The Fulfiller already opens a span per section, tags it with the section key
	// and the subject, times it, and records the error, so a second span inside the
	// collector added a name and no information. What still carries a logger and a
	// tracer is the collector that has something to say between reads — several
	// reads to attribute an error to, an account hop, a user record whose absence is
	// a different failure from an empty section.
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
		ddbdataprivacy.CollectorKeyAuditLog: auditprivacy.NewCollector(do.MustInvoke[auditdomain.Repository](i)),
		ddbdataprivacy.CollectorKeyIssueReports: issuereportsprivacy.NewCollector(
			do.MustInvoke[issuereports.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyUploadedMedia: uploadedmediaprivacy.NewCollector(do.MustInvoke[uploadedmedia.Repository](i)),
		ddbdataprivacy.CollectorKeyWaitlists:     waitlistsprivacy.NewCollector(do.MustInvoke[waitlists.Repository](i)),
		ddbdataprivacy.CollectorKeyComments:      commentsCollector,
	}

	for key, collector := range collectors {
		if err := registry.RegisterCollector(key, collector); err != nil {
			return nil, platformerrors.Wrapf(err, "registering %q data privacy collector", key)
		}
	}

	// Comments erase through their own eraser rather than through the cascade,
	// because platform-go's comment table has no foreign key to cascade from — see
	// ddbdataprivacy.EraserKeyComments. Both halves resolve the same single scope
	// this deployment files every comment under.
	if err := registry.RegisterEraser(
		ddbdataprivacy.EraserKeyComments,
		commentsEraser,
	); err != nil {
		return nil, platformerrors.Wrap(err, "registering comments data privacy eraser")
	}

	// One application eraser for the cascading tables, because every belongs_to_user
	// and belongs_to_account foreign key in this schema cascades from the user row. See
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

// RegisterOperationsRegistry registers the *operations.Registry this application's operations
// are looked up in, with the data privacy kinds already registered into it.
//
// platform-go v10 fulfills privacy requests as operations: the fulfillment loop is no longer a
// worker of its own but a set of runners registered under operation kinds, which an
// operations.Worker claims and runs. Building the Fulfiller is what performs that registration.
//
// It happens inside the registry's own provider rather than beside it because the ordering is
// load-bearing and invisible when wrong. samber/do resolves lazily, so a Registry resolved
// before anything built the Fulfiller is an empty one — and an empty registry does not fail
// loudly, it makes Service.Start refuse every privacy request with ErrUnknownKind and
// Worker.Run reject every claim the same way. Depending on the Fulfiller here makes the
// registration a precondition of holding the registry at all.
//
// Both process roles need this, not just the one that runs the work: Start looks the kind up in
// the registry of the process calling it, so an API server that only submits requests still has
// to know the kinds exist.
//
// Prerequisites: RegisterRegistry, and dataprivacycfg.RegisterArtifactStorage.
func RegisterOperationsRegistry(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*operations.Registry, error) {
		registry := operations.NewRegistry()

		fulfillerOpts, _ := platformdataprivacycfg.EnsurePackaging(
			do.MustInvoke[dataprivacycfg.ArtifactCompressor](i).Compressor,
			do.MustInvoke[dataprivacycfg.ArtifactEncryptorDecryptor](i).EncryptorDecryptor,
		)

		// The Fulfiller is discarded on purpose. Its whole effect here is the registration
		// it performs into registry; nothing calls it directly afterwards, because the
		// operations.Worker runs it through the kinds it registered.
		if _, err := platformdataprivacycfg.NewFulfiller(
			do.MustInvoke[context.Context](i),
			prepareConfig(i),
			do.MustInvoke[platformdataprivacy.Store](i),
			do.MustInvoke[*platformdataprivacy.Registry](i),
			registry,
			do.MustInvoke[dataprivacycfg.ArtifactUploadManager](i).UploadManager,
			// Artifacts are encrypted, so no signed URL can be minted for one: the
			// stored object is ciphertext and a subject following that link would get a
			// file they cannot open. Saying so here is what stops a completion
			// notification carrying a broken download link.
			true,
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
			platformdataprivacycfg.WithFulfillerOptions(fulfillerOpts...),
		); err != nil {
			return nil, platformerrors.Wrap(err, "registering data privacy operation kinds")
		}

		return registry, nil
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
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
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

// commentsPrivacy builds the comment collector and eraser, which are
// platform-go's over platform-go's store.
//
// Both take a scope resolver rather than reading the subject's scope, because a
// deployment that files comments per tenant has to be told which tenants to walk.
// This one files them all in the single scope ddbcomments.Scope names, so the
// resolver is fixed and shared — neither half can drift from the other about
// which rows a subject's comments are.
func commentsPrivacy(i do.Injector) (platformdataprivacy.Collector, platformdataprivacy.Eraser, error) {
	store := do.MustInvoke[platformcomments.Store](i)
	resolveScopes := commentsprivacy.FixedScopes(ddbcomments.Scope())

	collector, err := commentsprivacy.NewCollector(store, resolveScopes)
	if err != nil {
		return nil, nil, platformerrors.Wrap(err, "building the comments data privacy collector")
	}

	eraser, err := commentsprivacy.NewEraser(store, resolveScopes)
	if err != nil {
		return nil, nil, platformerrors.Wrap(err, "building the comments data privacy eraser")
	}

	return collector, eraser, nil
}
