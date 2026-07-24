package main

import (
	"encoding/base64"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	authservice "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/config"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	analyticscfg "github.com/primandproper/platform-go/v6/analytics/config"
	tokenscfg "github.com/primandproper/platform-go/v6/authentication/tokens/config"
	circuitbreakingcfg "github.com/primandproper/platform-go/v6/circuitbreaking/config"
	encryptioncfg "github.com/primandproper/platform-go/v6/cryptography/encryption/config"
	databasecfg "github.com/primandproper/platform-go/v6/database/config"
	"github.com/primandproper/platform-go/v6/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v6/featureflags/config"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	"github.com/primandproper/platform-go/v6/messagequeue/redis"
	notificationscfg "github.com/primandproper/platform-go/v6/notifications/mobile/config"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v6/observability/logging/config"
	logotelgrpc "github.com/primandproper/platform-go/v6/observability/logging/otelgrpc"
	metricscfg "github.com/primandproper/platform-go/v6/observability/metrics/config"
	"github.com/primandproper/platform-go/v6/observability/metrics/otelgrpc"
	profilingcfg "github.com/primandproper/platform-go/v6/observability/profiling/config"
	"github.com/primandproper/platform-go/v6/observability/profiling/pprof"
	tracingcfg "github.com/primandproper/platform-go/v6/observability/tracing/config"
	"github.com/primandproper/platform-go/v6/observability/tracing/oteltrace"
	"github.com/primandproper/platform-go/v6/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v6/routing/config"
	"github.com/primandproper/platform-go/v6/search/text/algolia"
	textsearchcfg "github.com/primandproper/platform-go/v6/search/text/config"
	"github.com/primandproper/platform-go/v6/server/http"
	uploadscfg "github.com/primandproper/platform-go/v6/uploads/config"
	"github.com/primandproper/platform-go/v6/uploads/objectstorage"
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
		Queues: msgconfig.QueuesConfig{
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
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Analytics: analyticscfg.Config{
			// we're using a noop version of this in localdev right now, but it still tries to instantiate a circuit breaker.
			SourceConfig: analyticscfg.SourceConfig{
				CircuitBreaker: circuitbreakingcfg.Config{
					Name:                   "feature_flagger",
					ErrorRate:              .5,
					MinimumSampleThreshold: 100,
				},
			},
		},
		TextSearch: textsearchcfg.Config{
			Algolia:  &algolia.Config{},
			Provider: textsearchcfg.AlgoliaProvider,
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "dev_text_searcher",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		HTTPServer: http.Config{
			Debug:           true,
			Port:            defaultHTTPPort,
			StartupDeadline: time.Minute,
		},
		Database: databasecfg.Config{
			Provider:                 databasecfg.ProviderPostgres,
			Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			OAuth2TokenEncryptionKey: localOAuth2TokenEncryptionKey,
			Debug:                    true,
			RunMigrations:            true,
			LogQueries:               true,
			MaxPingAttempts:          maxAttempts,
			PingWaitPeriod:           time.Second,
			MaxIdleConns:             5,
			MaxOpenConns:             7,
			ConnMaxLifetime:          30 * time.Minute,
			ReadConnection:           localdevPostgresDBConnectionDetails,
			WriteConnection:          localdevPostgresDBConnectionDetails,
		},
		Observability: localObservabilityConfig,
		Services: config.ServicesConfig{
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
				Tokens: tokenscfg.Config{
					Provider:                tokenscfg.ProviderPASETO,
					Issuer:                  "dinner-done-better",
					Audience:                "https://api.dinnerdonebetter.dev",
					Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
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
