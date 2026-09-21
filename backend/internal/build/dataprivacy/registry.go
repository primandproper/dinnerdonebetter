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
	identityprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/privacy"
	issuereportsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/privacy"
	paymentsprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/privacy"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	oauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v14/authentication/passkeys"
	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/platform-go/v14/billing"
	platformcomments "github.com/primandproper/platform-go/v14/comments"
	commentsprivacy "github.com/primandproper/platform-go/v14/comments/privacy"
	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/platform-go/v14/dataprivacy/auditerasure"
	platformdataprivacycfg "github.com/primandproper/platform-go/v14/dataprivacy/config"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	issuereports "github.com/primandproper/platform-go/v14/issuereports"
	uploadsregistry "github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/platform-go/v14/operations"
	"github.com/primandproper/platform-go/v14/privacyadapters"
	platformsettings "github.com/primandproper/platform-go/v14/settings"
	platformwaitlists "github.com/primandproper/platform-go/v14/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

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
		identityStore  = do.MustInvoke[platformidentity.Store](i)
		registry       = platformdataprivacy.NewRegistry()
	)

	// Every collector below takes a read executor at construction, because
	// dataprivacy.Collector.Collect is handed none: an export is a read, and it
	// runs outside the erasure transaction by design. The erasers, which do run
	// inside it, are handed the request's database.Tx per call.
	reader := do.MustInvoke[database.Client](i).Reader()

	// Which accounts a subject appears in, which is the one question an
	// account-scoped collector cannot answer from its own domain. Resolved once and
	// shared, so five collectors asking it do not become five identical page walks
	// per collector — they still each call it, but through one implementation whose
	// cost is visible in one place.
	resolveAccounts := identityprivacy.ResolveAccountIDs(identityStore, reader)

	// The nine adapters platform ships that this deployment runs, registered in one call
	// under each package's own DefaultKey.
	//
	// This replaces nine hand-written constructor calls, and the reason to prefer the call
	// is not that it is shorter. Register is all-or-nothing: every adapter is built before
	// any is registered, and the keys are checked against what the registry already holds
	// first, so a nil store in the last field cannot leave a registry holding eight of
	// nine. A half-registered registry is exactly the state that produces an export that
	// is well-formed, reports success, and is missing a domain — which is the failure this
	// application shipped for months and closed by hand two days ago.
	//
	// It also fails upstream when platform adds a twelfth adapter: privacyadapters' own
	// roster test requires every key the module ships to come back from Register, so a new
	// domain is a compile or a test failure rather than a section nobody notices is absent.
	//
	// Three fields are deliberately nil, and nil means "this deployment does not run it",
	// so each is a claim worth defending:
	//
	//   - Identity, because this application's eraser is not platform's. It is platform's
	//     with the succession rule in front of it — the households a departing owner leaves
	//     behind are transferred to their longest-tenured member before the user row goes.
	//     IdentityAdapter takes a Store and a resolver and builds both halves itself, so
	//     there is nowhere to hand it a decorated eraser. Registered by hand below.
	//   - Notifications, because this application's collector reads its own Repository
	//     rather than platform's Inbox and Registry, and answers as one section where
	//     platform answers as two.
	//   - AuditErasure, because whether the audit log is erased at all is a config flag
	//     this deployment sets through dataprivacycfg.RegisterAuditEraser — which
	//     privacyadapters' own documentation names as the deliberate alternative.
	//
	// The first of those is the one to revisit: identity is the domain whose absence from
	// the roster guarantee matters most, and the only reason it is absent is an adapter
	// that cannot take an eraser somebody else built.
	credentialScopes := identityprivacy.Scopes()

	successionStep, successionErr := identityprivacy.SuccessionStep(identityStore)
	if successionErr != nil {
		return nil, platformerrors.Wrap(successionErr, "building the household succession rule")
	}

	adopted, adoptedErr := privacyadapters.Register(registry, &privacyadapters.Adapters{
		Reader: reader,

		Comments: &privacyadapters.CommentsAdapter{
			Store:   do.MustInvoke[platformcomments.Store](i),
			Resolve: platformdataprivacy.FixedScopes(ddbcomments.Scope()),
		},
		// Issue reports are filed per account, so the resolver turns this application's
		// account ids into the scopes those rows live under. Same conversion the
		// hand-written constructor made internally, now spelled at the wiring.
		IssueReports: &privacyadapters.IssueReportsAdapter{
			Store:   do.MustInvoke[issuereports.Store](i),
			Resolve: issuereportsprivacy.AccountScopes(resolveAccounts),
		},
		Settings: &privacyadapters.SettingsAdapter{
			Store:   do.MustInvoke[platformsettings.Store](i),
			Resolve: platformdataprivacy.FixedScopes(ddbsettings.Scope()),
		},
		Waitlists: &privacyadapters.WaitlistsAdapter{
			Store:   do.MustInvoke[platformwaitlists.Store](i),
			Resolve: platformdataprivacy.FixedScopes(ddbwaitlists.Scope()),
		},
		MediaRegistry: &privacyadapters.MediaRegistryAdapter{
			Store:   do.MustInvoke[uploadsregistry.Store](i),
			Resolve: platformdataprivacy.FixedScopes(ddbuploadedmedia.Scope()),
		},

		// The three credential domains, all under the one scope this directory has.
		OAuth2Clients: &privacyadapters.OAuth2ClientsAdapter{
			Store:   do.MustInvoke[oauth2clients.Store](i),
			Resolve: credentialScopes,
		},
		Passkeys: &privacyadapters.PasskeysAdapter{
			Store:   do.MustInvoke[passkeys.Store](i),
			Resolve: credentialScopes,
		},
		PasswordReset: &privacyadapters.PasswordResetAdapter{
			Store:   do.MustInvoke[passwordreset.Store](i),
			Resolve: credentialScopes,
		},

		// Identity, with the succession rule running ahead of platform's eraser. That
		// seam is why this is here rather than hand-registered below: the adapter builds
		// both halves and there was no way to put anything in front of the eraser, so
		// identity — the domain whose absence from the roster matters most — sat outside
		// it. platform added BeforeErase for exactly this.
		Identity: &privacyadapters.IdentityAdapter{
			Store:       identityStore,
			Resolve:     credentialScopes,
			BeforeErase: successionStep,
		},

		// Billing takes a resolver of its own shape — accounts rather than scopes —
		// because what it pages is filed per account. The conversion is this
		// application's tenancy model and lives beside the payments domain.
		Billing: &privacyadapters.BillingAdapter{
			Store:   do.MustInvoke[billing.Store](i),
			Resolve: paymentsprivacy.AccountResolver(resolveAccounts),
		},
	})
	if adoptedErr != nil {
		return nil, platformerrors.Wrap(adoptedErr, "registering platform's privacy adapters")
	}

	logger.WithValue("keys", adopted).Info("registered platform's privacy adapters")

	// And the three this application answers for itself, none of which platform ships a
	// counterpart for that this deployment uses: meal planning is the domain this
	// application is, the audit log is a hash chain nothing else models, and the
	// notifications collector reads one repository and answers as one section where
	// platform's reads an inbox and a device registry and answers as two.
	//
	// A collector whose whole body is "page one list read and encode the rows" is
	// platformdataprivacy.CollectorFor and has no observability of its own to do: the
	// Fulfiller already opens a span per section, tags it with the section key and the
	// subject, times it, and records the error. What still carries a logger and a tracer
	// is the collector with something to say between reads — several reads to attribute an
	// error to, an account hop, a user record whose absence is a different failure from an
	// empty section.
	collectors := map[string]platformdataprivacy.Collector{
		ddbdataprivacy.CollectorKeyMealPlanning: mealplanningprivacy.NewCollector(
			do.MustInvoke[mealplanning.Repository](i), resolveAccounts, logger, tracerProvider),
		ddbdataprivacy.CollectorKeyNotifications: notificationsprivacy.NewCollector(
			do.MustInvoke[notifications.Repository](i), logger, tracerProvider),
		ddbdataprivacy.CollectorKeyAuditLog: auditprivacy.NewCollector(do.MustInvoke[auditdomain.Repository](i)),
	}

	for key, collector := range collectors {
		if err := registry.RegisterCollector(key, collector); err != nil {
			return nil, platformerrors.Wrapf(err, "registering %q data privacy collector", key)
		}
	}

	// Three erasers this file used to register by hand — comments, waitlist signups and
	// registered OAuth2 clients — are registered by the call above, because each is
	// platform's over a platform store and the adapter builds both halves. Why each needs
	// an eraser at all rather than riding the identity cascade is unchanged and is
	// documented in docs/data-privacy.md: comments have no key to cascade from, waitlist
	// signups cannot have one because withdrawal blanks the column, and the client
	// registry's column mostly does not name a user.

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
		platformdataprivacycfg.WithAuditEraserOptions(
			auditerasure.WithScopeResolver(auditprivacy.ErasableScopeResolver(identityStore, reader)),
		),
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

		// The Fulfiller is discarded on purpose. Its whole effect here is the registration
		// it performs into registry; nothing calls it directly afterwards, because the
		// operations.Worker runs it through the kinds it registered.
		if _, err := platformdataprivacycfg.NewFulfiller(
			do.MustInvoke[context.Context](i),
			prepareConfig(i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformdataprivacy.Store](i),
			do.MustInvoke[*platformdataprivacy.Registry](i),
			registry,
			do.MustInvoke[dataprivacycfg.ArtifactUploadManager](i).UploadManager,
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
			// Artifacts are encrypted, so no signed URL can be minted for one: the
			// stored object is ciphertext and a subject following that link would get a
			// file they cannot open. v14 reads that off the encryptor rather than off a
			// flag beside it, which is what stops the two disagreeing — so naming the
			// encryptor here is also what stops a completion notification carrying a
			// broken download link.
			platformdataprivacycfg.WithCompressor(do.MustInvoke[dataprivacycfg.ArtifactCompressor](i).Compressor),
			platformdataprivacycfg.WithEncryptor(do.MustInvoke[dataprivacycfg.ArtifactEncryptorDecryptor](i).EncryptorDecryptor),
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

	collector, err := commentsprivacy.NewCollector(store, do.MustInvoke[database.Client](i).Reader(), resolveScopes)
	if err != nil {
		return nil, nil, platformerrors.Wrap(err, "building the comments data privacy collector")
	}

	eraser, err := commentsprivacy.NewEraser(store, resolveScopes)
	if err != nil {
		return nil, nil, platformerrors.Wrap(err, "building the comments data privacy eraser")
	}

	return collector, eraser, nil
}
