package main

import (
	"encoding/base64"
	"time"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	authservice "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/config"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	tokenscfg "github.com/primandproper/platform-go/v9/authentication/tokens/config"
	cachecfg "github.com/primandproper/platform-go/v9/cache/config"
	cacheredis "github.com/primandproper/platform-go/v9/cache/redis"
	capitalismcfg "github.com/primandproper/platform-go/v9/capitalism/config"
	circuitbreakingcfg "github.com/primandproper/platform-go/v9/circuitbreaking/config"
	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	distributedlockcfg "github.com/primandproper/platform-go/v9/distributedlock/config"
	pglock "github.com/primandproper/platform-go/v9/distributedlock/postgres"
	emailcfg "github.com/primandproper/platform-go/v9/email/config"
	"github.com/primandproper/platform-go/v9/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v9/featureflags/config"
	idempotencycfg "github.com/primandproper/platform-go/v9/idempotency/config"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/messagequeue/redis"
	notificationscfg "github.com/primandproper/platform-go/v9/notifications/mobile/config"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v9/observability/logging/config"
	logotelgrpc "github.com/primandproper/platform-go/v9/observability/logging/otelgrpc"
	metricscfg "github.com/primandproper/platform-go/v9/observability/metrics/config"
	"github.com/primandproper/platform-go/v9/observability/metrics/otelgrpc"
	profilingcfg "github.com/primandproper/platform-go/v9/observability/profiling/config"
	"github.com/primandproper/platform-go/v9/observability/profiling/pprof"
	tracingcfg "github.com/primandproper/platform-go/v9/observability/tracing/config"
	"github.com/primandproper/platform-go/v9/observability/tracing/oteltrace"
	"github.com/primandproper/platform-go/v9/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v9/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"
	"github.com/primandproper/platform-go/v9/server/http"
	uploadscfg "github.com/primandproper/platform-go/v9/uploads/config"
	"github.com/primandproper/platform-go/v9/uploads/objectstorage"
)

const (
	dockerComposeWorkerQueueAddress = "worker_queue:6379"
	localOAuth2TokenEncryptionKey   = debugCookieHashKey
)

var (
	localdevPostgresDBConnectionDetails = databasecfg.ConnectionDetails{
		Username:   "dbuser",
		Password:   "hunter2",
		Database:   "dinner-done-better",
		Host:       "pgdatabase",
		Port:       5432,
		DisableSSL: true,
	}

	localObservabilityConfig = observability.Config{
		Logging: loggingcfg.Config{
			ServiceName: otelServiceName,
			Level:       logging.DebugLevel,
			Provider:    loggingcfg.ProviderOtelSlog,
			OtelSlog: &logotelgrpc.Config{
				CollectorEndpoint: "otel_collector:4317",
				Insecure:          true,
				Timeout:           time.Second * 3,
			},
		},
		Metrics: metricscfg.Config{
			ServiceName: otelServiceName,
			Otel: &otelgrpc.Config{
				Insecure:           true,
				CollectorEndpoint:  "otel_collector:4317",
				CollectionInterval: time.Second,
			},
			Provider: metricscfg.ProviderOtel,
			Enabled:  true,
		},
		Tracing: tracingcfg.Config{
			Provider:                  tracingcfg.ProviderOtel,
			ServiceName:               otelServiceName,
			SpanCollectionProbability: 1,
			Otel: &oteltrace.Config{
				Insecure:          true,
				CollectorEndpoint: "otel_collector:4317",
			},
		},
		Profiling: profilingcfg.Config{
			ServiceName: otelServiceName,
			Provider:    profilingcfg.ProviderPprof,
			Pprof: &pprof.Config{
				Port: pprof.DefaultPort,
			},
		},
	}

	localRoutingConfig = routingcfg.Config{
		Provider: routingcfg.ProviderChi,
		Chi: &chi.Config{
			ServiceName:            otelServiceName,
			EnableCORSForLocalhost: true,
			SilenceRouteLogging:    false,
		},
	}
)

func buildLocalDevConfig() *config.APIServiceConfig {
	uploadsConfig := uploadscfg.Config{
		Debug: true,
		Storage: objectstorage.Config{
			Provider:   objectstorage.FilesystemProvider,
			BucketName: "avatars",
			FilesystemConfig: &objectstorage.FilesystemConfig{
				RootDirectory: "/uploads",
			},
		},
	}

	return &config.APIServiceConfig{
		Routing: localRoutingConfig,
		// Localdev has a Redis, so the record store is shared and the interceptor means
		// something. Prod does not yet; see the prod config.
		Idempotency: config.IdempotencyConfig{
			Enabled: true,
			Manager: idempotencycfg.Config{
				KeyPrefix: "dinner_done_better.idempotency.",
				Cache: cachecfg.Config{
					Provider: cachecfg.ProviderRedis,
					Redis: &cacheredis.Config{
						Addresses: []string{dockerComposeWorkerQueueAddress},
						Namespace: "idempotency:",
					},
				},
				Lock: distributedlockcfg.Config{
					Provider: distributedlockcfg.PostgresProvider,
					Postgres: &pglock.Config{ConnWaitTimeout: 5 * time.Second},
				},
				TTL:         24 * time.Hour,
				InFlightTTL: 2 * time.Minute,
			},
		},
		Queues: queuescfg.Config{
			DataChangesTopicName:              dataChangesTopicName,
			OutboundEmailsTopicName:           outboundEmailsTopicName,
			SearchIndexRequestsTopicName:      searchIndexRequestsTopicName,
			MobileNotificationsTopicName:      mobileNotificationsTopicName,
			UserDataAggregationTopicName:      userDataAggregationTopicName,
			WebhookExecutionRequestsTopicName: webhookExecutionRequestsTopicName,
		},
		Meta: config.MetaSettings{
			Debug:   true,
			RunMode: developmentEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Events: msgconfig.Config{
			Consumer: msgconfig.MessageQueueConfig{
				Provider: msgconfig.ProviderRedis,
				Redis: redis.Config{
					QueueAddresses: []string{dockerComposeWorkerQueueAddress},
				},
			},
			Publisher: msgconfig.MessageQueueConfig{
				Provider: msgconfig.ProviderRedis,
				Redis: redis.Config{
					QueueAddresses: []string{dockerComposeWorkerQueueAddress},
				},
			},
		},
		FeatureFlags: featureflagscfg.Config{
			// we're using a noop version of this in localdev right now, but it still tries to instantiate a circuit breaker.
			Provider: featureflagscfg.ProviderNoop,
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Email: emailcfg.Config{
			// Nothing sends mail in this environment; v9 makes that say so rather than assume it.
			Provider: emailcfg.ProviderNoop,
		},
		Analytics: analyticscfg.Config{
			SourceConfig: analyticscfg.SourceConfig{
				// Analytics are off here, which the provider now says rather than
				// leaving it to an unset value. The circuit breaker is still built.
				Provider: analyticscfg.ProviderNoop,
				CircuitBreaker: circuitbreakingcfg.Config{
					Name:                   "feature_flagger",
					ErrorRate:              .5,
					MinimumSampleThreshold: 100,
				},
			},
		},
		TextSearch: textsearchcfg.Config{
			// localdev has no Algolia account and never had one, so the index was
			// being built against empty credentials. Off, and said so.
			Provider: textsearchcfg.ProviderNoop,
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "dev_text_searcher",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		HTTPServer: http.Config{
			Port:            defaultHTTPPort,
			StartupDeadline: time.Minute,
			AppleAppSiteAssociation: &http.AppleAppSiteAssociationConfig{
				TeamID:   appleTeamID,
				BundleID: appleBundleID,
			},
		},
		Database: dbcfg.Config{
			Config: databasecfg.Config{
				Provider:        databasecfg.ProviderPostgres,
				Debug:           true,
				RunMigrations:   true,
				LogQueries:      true,
				MaxPingAttempts: maxAttempts,
				PingWaitPeriod:  time.Second,
				MaxIdleConns:    5,
				MaxOpenConns:    7,
				ConnMaxLifetime: 30 * time.Minute,
				ReadConnection:  localdevPostgresDBConnectionDetails,
				WriteConnection: localdevPostgresDBConnectionDetails,
			},
			Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			OAuth2TokenEncryptionKey: localOAuth2TokenEncryptionKey,
		},
		Observability: localObservabilityConfig,
		Services: config.ServicesConfig{
			// The capitalism provider is named rather than left empty: platform-go treats an
			// unset provider as an error precisely so that "we forgot to configure billing"
			// cannot masquerade as "we chose not to bill".
			Payments: paymentscfg.Config{
				Capitalism: capitalismcfg.Config{
					Provider: capitalismcfg.NoopProvider,
				},
			},
			Auth: authservice.Config{
				OAuth2: authservice.OAuth2Config{
					Domain:               "http://localhost:9000",
					AccessTokenLifespan:  time.Hour,
					RefreshTokenLifespan: time.Hour,
					Debug:                false,
				},
				Debug:                 true,
				EnableUserSignup:      true,
				MinimumUsernameLength: 3,
				MinimumPasswordLength: 8,
				TokenLifetime:         5 * time.Minute,
				Tokens: authcfg.TokensConfig{
					Config: tokenscfg.Config{
						Provider:                tokenscfg.ProviderPASETO,
						Issuer:                  "dinner-done-better",
						Audience:                "https://api.dinnerdonebetter.dev",
						Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
					},
				},
			},
			DataPrivacy: dataprivacycfg.Config{
				Uploads: uploadscfg.Config{
					Storage: objectstorage.Config{
						FilesystemConfig: &objectstorage.FilesystemConfig{RootDirectory: "/tmp"},
						BucketName:       "userdata",
						Provider:         objectstorage.FilesystemProvider,
					},
					Debug: false,
				},
			},
			Users: identitycfg.Config{
				PublicMediaURLPrefix: "http://localhost:8000/uploads",
				Uploads:              uploadsConfig,
			},
			UploadedMedia: uploadedmediacfg.Config{
				Uploads: uploadsConfig,
			},
		},
		PushNotifications: notificationscfg.Config{
			Provider: notificationscfg.ProviderNoop,
		},
	}
}
