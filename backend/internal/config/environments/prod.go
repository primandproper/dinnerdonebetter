package environments

import (
	"encoding/base64"
	"time"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	authservice "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/config"
	mealplanningcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/config"
	oauthcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/config"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	analyticscfg "github.com/primandproper/platform-go/v13/analytics/config"
	analyticsposthog "github.com/primandproper/platform-go/v13/analytics/posthog"
	oauth2servercfg "github.com/primandproper/platform-go/v13/authentication/oauth2server/config"
	oauth2database "github.com/primandproper/platform-go/v13/authentication/oauth2server/database"
	tokenscfg "github.com/primandproper/platform-go/v13/authentication/tokens/config"
	platformwebauthn "github.com/primandproper/platform-go/v13/authentication/webauthn"
	webauthncfg "github.com/primandproper/platform-go/v13/authentication/webauthn/config"
	capitalismcfg "github.com/primandproper/platform-go/v13/capitalism/config"
	circuitbreakingcfg "github.com/primandproper/platform-go/v13/circuitbreaking/config"
	encryptioncfg "github.com/primandproper/platform-go/v13/cryptography/encryption/config"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	emailcfg "github.com/primandproper/platform-go/v13/email/config"
	"github.com/primandproper/platform-go/v13/email/resend"
	"github.com/primandproper/platform-go/v13/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v13/featureflags/config"
	"github.com/primandproper/platform-go/v13/featureflags/posthog"
	msgconfig "github.com/primandproper/platform-go/v13/messagequeue/config"
	"github.com/primandproper/platform-go/v13/messagequeue/pubsub"
	"github.com/primandproper/platform-go/v13/notifications/mobile/apns"
	notificationscfg "github.com/primandproper/platform-go/v13/notifications/mobile/config"
	"github.com/primandproper/platform-go/v13/notifications/mobile/fcm"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v13/observability/logging/config"
	logotelgrpc "github.com/primandproper/platform-go/v13/observability/logging/otelgrpc"
	metricscfg "github.com/primandproper/platform-go/v13/observability/metrics/config"
	"github.com/primandproper/platform-go/v13/observability/metrics/otelgrpc"
	profilingcfg "github.com/primandproper/platform-go/v13/observability/profiling/config"
	"github.com/primandproper/platform-go/v13/observability/profiling/pyroscope"
	tracingcfg "github.com/primandproper/platform-go/v13/observability/tracing/config"
	"github.com/primandproper/platform-go/v13/observability/tracing/oteltrace"
	"github.com/primandproper/platform-go/v13/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v13/routing/config"
	"github.com/primandproper/platform-go/v13/search/text/algolia"
	textsearchcfg "github.com/primandproper/platform-go/v13/search/text/config"
	"github.com/primandproper/platform-go/v13/server/grpc"
	"github.com/primandproper/platform-go/v13/server/http"
	uploadscfg "github.com/primandproper/platform-go/v13/uploads/config"
	"github.com/primandproper/platform-go/v13/uploads/objectstorage"
)

const (
	prodGCPProject            = "dinner-done-better-prod"
	prodMediaBucket           = "media.dinnerdonebetter.com"
	prodUserDataBucket        = "userdata.dinnerdonebetter.com"
	prodOtelCollectorEndpoint = "otel-collector-svc.prod.svc.cluster.local:4317"
	// prodAPIPublicURL is where this API server answers. It is the OAuth2 issuer, the
	// resource an access token names, and the audience a session JWT is minted for — one
	// address under three protocol names, so it is spelled once.
	prodAPIPublicURL   = "https://http-api.dinnerdonebetter.com"
	prodTokensAudience = prodAPIPublicURL
	iosTeamID          = "K8R2Q5UWQS"
	iosBundleID        = "com.dinnerdonebetter.ios"
)

// BuildProdConfig returns the configuration the production environment runs with.
func BuildProdConfig() *config.APIServiceConfig {
	gcpMediaStorage := objectstorage.Config{
		Provider:     objectstorage.GCPCloudStorageProvider,
		BucketName:   prodMediaBucket,
		BucketPrefix: "avatars/",
	}

	gcpUserDataStorage := objectstorage.Config{
		Provider:   objectstorage.GCPCloudStorageProvider,
		BucketName: prodUserDataBucket,
	}

	pubsubConfig := msgconfig.MessageQueueConfig{
		Provider: msgconfig.ProviderPubSub,
		PubSub: pubsub.Config{
			ProjectID: prodGCPProject,
		},
	}

	prodObservabilityConfig := observability.Config{
		Logging: loggingcfg.Config{
			ServiceName: otelServiceName,
			Level:       logging.InfoLevel,
			Provider:    loggingcfg.ProviderOtelSlog,
			OtelSlog: &logotelgrpc.Config{
				CollectorEndpoint: prodOtelCollectorEndpoint,
				Insecure:          true,
				Timeout:           2 * time.Second,
			},
		},
		Metrics: metricscfg.Config{
			ServiceName: otelServiceName,
			Otel: &otelgrpc.Config{
				Insecure:             true,
				CollectorEndpoint:    prodOtelCollectorEndpoint,
				CollectionInterval:   30 * time.Second,
				EnableRuntimeMetrics: true,
				EnableHostMetrics:    true,
			},
			Provider: metricscfg.ProviderOtel,
			Enabled:  true,
		},
		Tracing: tracingcfg.Config{
			Provider:                  tracingcfg.ProviderOtel,
			ServiceName:               otelServiceName,
			SpanCollectionProbability: 1.0,
			Otel: &oteltrace.Config{
				Insecure:          true,
				CollectorEndpoint: prodOtelCollectorEndpoint,
			},
		},
		Profiling: profilingcfg.Config{
			ServiceName: otelServiceName,
			Provider:    profilingcfg.ProviderPyroscope,
			Pyroscope: &pyroscope.Config{
				ServerAddress: "https://profiles-prod-001.grafana.net",
				UploadRate:    15 * time.Second,
			},
		},
	}

	return &config.APIServiceConfig{
		Webhooks:     buildWebhooksConfig(),
		Metering:     config.DefaultMeteringConfig(),
		Entitlements: config.DefaultEntitlementsConfig(),
		Operations:   config.DefaultOperationsConfig(),
		Routing: routingcfg.Config{
			Provider: routingcfg.ProviderChi,
			Chi: &chi.Config{
				ServiceName:            otelServiceName,
				EnableCORSForLocalhost: false,
				SilenceRouteLogging:    false,
			},
		},
		// Off until prod has a shared record store. The interceptor needs one: with the
		// memory provider each replica keeps its own records, so a retry that lands on a
		// different pod re-executes and two concurrent requests can both claim the same
		// key. That is the failure this exists to prevent, so shipping it that way would
		// be worse than shipping nothing. Provision Redis, point Manager.Cache at it, and
		// flip Enabled.
		Idempotency: config.IdempotencyConfig{
			Enabled: false,
		},
		Queues: queuescfg.Config{
			DataChangesTopicName:         dataChangesTopicName,
			OutboundEmailsTopicName:      outboundEmailsTopicName,
			SearchIndexRequestsTopicName: searchIndexRequestsTopicName,
			MobileNotificationsTopicName: mobileNotificationsTopicName,
		},
		Meta: config.MetaSettings{
			Debug:   false,
			RunMode: "production",
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Events: msgconfig.Config{
			Consumer:  pubsubConfig,
			Publisher: pubsubConfig,
		},
		GRPCServer: grpc.Config{
			Port: defaultGRPCPort,
		},
		HTTPServer: http.Config{
			Port:            defaultHTTPPort,
			StartupDeadline: 60 * time.Second,
			AppleAppSiteAssociation: &http.AppleAppSiteAssociationConfig{
				TeamID:   appleTeamID,
				BundleID: appleBundleID,
			},
		},
		Database: dbcfg.Config{
			Config: databasecfg.Config{
				Provider:        databasecfg.ProviderPostgres,
				Debug:           false,
				RunMigrations:   true,
				LogQueries:      false,
				MaxPingAttempts: maxAttempts,
				PingWaitPeriod:  time.Second,
				MaxIdleConns:    5,
				MaxOpenConns:    7,
				ConnMaxLifetime: 30 * time.Minute,
				ReadConnection: databasecfg.ConnectionDetails{
					Username:   "api_db_user",
					Password:   replaceAtDeploy, /* #nosec G101 */
					Database:   serviceName,
					Host:       replaceAtDeploy,
					Port:       5432,
					DisableSSL: false,
				},
				WriteConnection: databasecfg.ConnectionDetails{
					Username:   "api_db_user",
					Password:   replaceAtDeploy, /* #nosec G101 */
					Database:   serviceName,
					Host:       replaceAtDeploy,
					Port:       5432,
					DisableSSL: false,
				},
			},
		},
		Observability: prodObservabilityConfig,
		Email: emailcfg.Config{
			Provider: emailcfg.ProviderResend,
			Resend: &resend.Config{
				APIToken: placeholderValue, // overridden by env from api-service-config secret
			},
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "prod_emailer",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Analytics: analyticscfg.Config{
			ProxySources: analyticscfg.ProxySourcesConfig{
				iosPlatform: {
					Provider: analyticscfg.ProviderPostHog,
					Posthog:  &analyticsposthog.Config{APIKey: placeholderValue}, // overridden by env from api-service-config secret
					CircuitBreaker: circuitbreakingcfg.Config{
						Name:                   iosAnalyticsSource,
						ErrorRate:              .5,
						MinimumSampleThreshold: 100,
					},
				},
				webPlatform: {
					Provider: analyticscfg.ProviderPostHog,
					Posthog:  &analyticsposthog.Config{APIKey: placeholderValue}, // overridden by env from api-service-config secret
					CircuitBreaker: circuitbreakingcfg.Config{
						Name:                   webAnalyticsSource,
						ErrorRate:              .5,
						MinimumSampleThreshold: 100,
					},
				},
			},
			SourceConfig: analyticscfg.SourceConfig{
				Provider: analyticscfg.ProviderPostHog,
				Posthog:  &analyticsposthog.Config{APIKey: placeholderValue}, // overridden by env from api-service-config secret
				CircuitBreaker: circuitbreakingcfg.Config{
					Name:                   "api_analytics",
					ErrorRate:              .5,
					MinimumSampleThreshold: 100,
				},
			},
		},
		TextSearch: textsearchcfg.Config{
			Provider: textsearchcfg.AlgoliaProvider,
			Algolia: &algolia.Config{
				AppID:  placeholderValue, // overridden by env from the algolia-credentials secret
				APIKey: placeholderValue,
			},
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "prod_text_searcher",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		FeatureFlags: featureflagscfg.Config{
			Provider: featureflagscfg.ProviderPostHog,
			// Both keys are placeholders, overridden by env from the CSI secret. v10 made
			// PersonalAPIKey required, having found that the SDK refuses every flag
			// evaluation without one — so a config naming only the project key used to
			// validate clean and then fail to serve a single flag. It is a real secret this
			// deployment now has to supply.
			PostHog: &posthog.Config{ProjectAPIKey: placeholderValue, PersonalAPIKey: placeholderValue},
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   featureFlaggerSource,
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Auth: authcfg.Config{
			Sessions: authcfg.SessionsConfig{
				AbsoluteTimeout: sessionAbsoluteTimeout,
				IdleTimeout:     sessionIdleTimeout,
				TouchInterval:   sessionTouchInterval,
			},
			Passkey: webauthncfg.Config{
				// The table, named rather than left to the default. Ceremony state has to
				// outlive the replica that issued the challenge, and a passkey login that
				// lands on a second pod is the normal case here, not the edge one.
				Provider: webauthncfg.ProviderDatabase,
				RelyingParty: platformwebauthn.Config{
					// The bare effective domain, which covers every subdomain a web app is
					// served from. It is also the scope of every passkey registered under it:
					// changing it invalidates all of them, because an authenticator will not
					// answer for a domain it did not register against.
					RPID:          branding.PublicDomain,
					RPDisplayName: branding.CompanyName,
					RPOrigins:     branding.WebAppOrigins(),
					// One number, in the three places the ceremony's deadline used to be
					// configured separately: what the browser is asked for, what the library
					// enforces when the response comes back, and how long the row lives.
					CeremonyTimeout: passkeyCeremonyTimeout,
				},
			},
			Tokens: authcfg.TokensConfig{
				Config: tokenscfg.Config{
					Provider:                tokenscfg.ProviderPASETO,
					Issuer:                  serviceName,
					Audience:                prodTokensAudience,
					Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
				},
			},
			Debug:                 false,
			EnableUserSignup:      true,
			MinimumUsernameLength: 3,
			MinimumPasswordLength: 8,
		},
		Services: config.ServicesConfig{
			// Both payment providers are named rather than left empty — the web checkout's
			// through platform-go's own config, the mobile store's through ours — because an
			// unset provider is an error precisely so that "we forgot to configure billing"
			// cannot masquerade as "we chose not to bill".
			Payments: paymentscfg.Config{
				Capitalism: capitalismcfg.Config{
					Provider: capitalismcfg.NoopProvider,
				},
				MobileProvider: capitalismcfg.NoopProvider,
			},
			Auth: authservice.Config{
				OAuth2: oauth2servercfg.Config{
					Provider: oauth2servercfg.ProviderDatabase,
					// The table prefix has to be the one migration 33 created the tables
					// under: a prefix that differs between the DDL and the store is a
					// server that comes up clean and cannot find a table.
					Database: oauth2database.Config{TablePrefix: ddboauth.TablePrefix},
					// The issuer is this API server's own public address, not the web app's.
					// Every endpoint in the discovery document is derived from it, and a
					// client compares it against the "iss" on an authorization response — so
					// naming the consumer web app here, as the redirect-domain setting this
					// replaces did, would advertise four addresses nothing is listening at.
					Issuer: prodAPIPublicURL,
					// The audience every access token carries, and the identifier the gRPC
					// interceptor checks a presented token against. Named explicitly rather
					// than defaulted so that a token minted for the MCP server — which
					// shares this database, and therefore this store — is not spendable
					// here.
					Resources: []string{prodAPIPublicURL},
					// Meant as "no sweeper of its own": `ddb job db-cleaner` calls Sweep
					// instead, one pass for the fleet rather than one per replica, each
					// running the same full-table delete on its own timer.
					//
					// It does not take effect. The field documents a non-positive value
					// as no sweeper, but oauth2servercfg.EnsureDefaults rewrites this
					// zero to ten minutes before it reaches WithSweeper, so every replica
					// sweeps as well — see platform-go#456. Left at zero rather than
					// worked around with a negative duration, which is undocumented
					// behavior the fix upstream may well remove.
					SweepInterval: 0,
				},
				Debug:                 false,
				EnableUserSignup:      true,
				MinimumUsernameLength: 3,
				MinimumPasswordLength: 8,
				TokenLifetime:         5 * time.Minute,
				Tokens: authcfg.TokensConfig{
					Config: tokenscfg.Config{
						Provider:                tokenscfg.ProviderPASETO,
						Issuer:                  serviceName,
						Audience:                prodTokensAudience,
						Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
					},
				},
			},
			DataPrivacy: dataprivacycfg.Config{
				Uploads: uploadscfg.Config{
					Storage: gcpUserDataStorage,
					Debug:   false,
				},
				Encryption: encryptioncfg.Config{Provider: encryptioncfg.ProviderAES, CurrentKeyID: "v1"},
				// Supplied from the environment, like every other secret in this file.
				ArtifactEncryptionKey: "",
			},
			Users: identitycfg.Config{
				PublicMediaURLPrefix: "https://" + prodMediaBucket + "/avatars",
				Uploads: uploadscfg.Config{
					Storage: gcpMediaStorage,
					Debug:   false,
				},
			},
			UploadedMedia: uploadedmediacfg.Config{
				Uploads: uploadscfg.Config{
					Storage: gcpMediaStorage,
					Debug:   false,
				},
			},
			MealPlanning: mealplanningcfg.Config{
				UseSearchService: true,
			},
			OAuth2Clients: oauthcfg.Config{
				OAuth2ClientCreationDisabled: true,
			},
		},
		PushNotifications: notificationscfg.Config{
			Provider: notificationscfg.ProviderAPNsFCM,
			APNs: &apns.Config{
				AuthKeyPath: "/mnt/apns/apns-auth-key.p8", // mounted from K8s secret apns-credentials
				TeamID:      iosTeamID,
				BundleID:    iosBundleID,
				Production:  true,
			},
			FCM: &fcm.Config{
				// CredentialsPath empty: uses Application Default Credentials (GCP workload identity)
			},
		},
	}
}
