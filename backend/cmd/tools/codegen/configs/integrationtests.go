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
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	tokenscfg "github.com/primandproper/platform-go/v9/authentication/tokens/config"
	circuitbreakingcfg "github.com/primandproper/platform-go/v9/circuitbreaking/config"
	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	emailcfg "github.com/primandproper/platform-go/v9/email/config"
	"github.com/primandproper/platform-go/v9/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v9/featureflags/config"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/messagequeue/redis"
	notificationscfg "github.com/primandproper/platform-go/v9/notifications/mobile/config"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v9/observability/logging/config"
	tracingcfg "github.com/primandproper/platform-go/v9/observability/tracing/config"
	"github.com/primandproper/platform-go/v9/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v9/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"
	"github.com/primandproper/platform-go/v9/server/grpc"
	"github.com/primandproper/platform-go/v9/server/http"
	uploadscfg "github.com/primandproper/platform-go/v9/uploads/config"
	"github.com/primandproper/platform-go/v9/uploads/objectstorage"
)

func buildIntegrationTestsConfig() *config.APIServiceConfig {
	uploadsConfig := uploadscfg.Config{
		Debug: false,
		Storage: objectstorage.Config{
			Provider:   "memory",
			BucketName: "avatars",
		},
	}

	return &config.APIServiceConfig{
		Routing: routingcfg.Config{
			Provider: routingcfg.ProviderChi,
			Chi: &chi.Config{
				ServiceName:            otelServiceName,
				EnableCORSForLocalhost: true,
				SilenceRouteLogging:    false,
			},
		},
		Meta: config.MetaSettings{
			Debug:   false,
			RunMode: testingEnv,
		},
		Queues: queuescfg.Config{
			DataChangesTopicName:              dataChangesTopicName,
			OutboundEmailsTopicName:           outboundEmailsTopicName,
			SearchIndexRequestsTopicName:      searchIndexRequestsTopicName,
			MobileNotificationsTopicName:      mobileNotificationsTopicName,
			UserDataAggregationTopicName:      userDataAggregationTopicName,
			WebhookExecutionRequestsTopicName: webhookExecutionRequestsTopicName,
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
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		HTTPServer: http.Config{
			Port:            defaultHTTPPort,
			StartupDeadline: time.Minute,
		},
		GRPCServer: grpc.Config{
			Port: defaultGRPCPort,
		},
		Database: dbcfg.Config{
			Config: databasecfg.Config{
				Provider:        databasecfg.ProviderPostgres,
				Debug:           true,
				RunMigrations:   true,
				LogQueries:      true,
				MaxPingAttempts: maxAttempts,
				PingWaitPeriod:  1500 * time.Millisecond,
				MaxIdleConns:    5,
				MaxOpenConns:    7,
				ConnMaxLifetime: 30 * time.Minute,
				ReadConnection:  localdevPostgresDBConnectionDetails,
				WriteConnection: localdevPostgresDBConnectionDetails,
			},
			Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			OAuth2TokenEncryptionKey: localOAuth2TokenEncryptionKey,
		},
		Observability: observability.Config{
			Logging: loggingcfg.Config{
				ServiceName: otelServiceName,
				Level:       logging.InfoLevel,
				Provider:    loggingcfg.ProviderSlog,
			},
			Tracing: tracingcfg.Config{
				Provider:                  "", // noop tracer for integration tests (no tracing-server required)
				SpanCollectionProbability: 0.0,
				ServiceName:               otelServiceName,
			},
		},
		TextSearch: textsearchcfg.Config{
			// we're using a noop version of this in dev right now, but it still tries to instantiate a circuit breaker.
			Provider: textsearchcfg.ProviderNoop,
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		FeatureFlags: featureflagscfg.Config{
			// we're using a noop version of this in dev right now, but it still tries to instantiate a circuit breaker.
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
		Services: config.ServicesConfig{
			Auth: authservice.Config{
				OAuth2: authservice.OAuth2Config{
					Domain:               "http://localhost:9000",
					AccessTokenLifespan:  time.Hour,
					RefreshTokenLifespan: time.Hour,
					Debug:                false,
				},
				Debug:                 false,
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
				Uploads: uploadsConfig,
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
