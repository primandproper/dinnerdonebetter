package main

import (
	"encoding/base64"
	"time"

	authcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/branding"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	authservice "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/config"
	mealplanningcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/config"
	oauthcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/oauth/config"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	analyticscfg "github.com/primandproper/platform-go/v8/analytics/config"
	analyticsposthog "github.com/primandproper/platform-go/v8/analytics/posthog"
	tokenscfg "github.com/primandproper/platform-go/v8/authentication/tokens/config"
	circuitbreakingcfg "github.com/primandproper/platform-go/v8/circuitbreaking/config"
	encryptioncfg "github.com/primandproper/platform-go/v8/cryptography/encryption/config"
	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	emailcfg "github.com/primandproper/platform-go/v8/email/config"
	"github.com/primandproper/platform-go/v8/email/resend"
	"github.com/primandproper/platform-go/v8/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v8/featureflags/config"
	"github.com/primandproper/platform-go/v8/featureflags/posthog"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/messagequeue/pubsub"
	notificationscfg "github.com/primandproper/platform-go/v8/notifications/mobile/config"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v8/observability/logging/config"
	logotelgrpc "github.com/primandproper/platform-go/v8/observability/logging/otelgrpc"
	metricscfg "github.com/primandproper/platform-go/v8/observability/metrics/config"
	"github.com/primandproper/platform-go/v8/observability/metrics/otelgrpc"
	profilingcfg "github.com/primandproper/platform-go/v8/observability/profiling/config"
	"github.com/primandproper/platform-go/v8/observability/profiling/pyroscope"
	tracingcfg "github.com/primandproper/platform-go/v8/observability/tracing/config"
	"github.com/primandproper/platform-go/v8/observability/tracing/oteltrace"
	"github.com/primandproper/platform-go/v8/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v8/routing/config"
	"github.com/primandproper/platform-go/v8/search/text/algolia"
	textsearchcfg "github.com/primandproper/platform-go/v8/search/text/config"
	"github.com/primandproper/platform-go/v8/server/grpc"
	"github.com/primandproper/platform-go/v8/server/http"
	uploadscfg "github.com/primandproper/platform-go/v8/uploads/config"
	"github.com/primandproper/platform-go/v8/uploads/objectstorage"
)

const (
	prodGCPProject            = "dinner-done-better-prod"
	prodMediaBucket           = "media.dinnerdonebetter.com"
	prodUserDataBucket        = "userdata.dinnerdonebetter.com"
	prodOtelCollectorEndpoint = "otel-collector-svc.prod.svc.cluster.local:4317"
	prodOAuth2Domain          = "https://dinnerdonebetter.com"
	prodTokensAudience        = "https://http-api.dinnerdonebetter.com" //nolint:gosec // G101: audience URL, not a credential
	iosTeamID                 = "K8R2Q5UWQS"
	iosBundleID               = "com.dinnerdonebetter.ios"
)

func buildProdConfig() *config.APIServiceConfig {
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
		Routing: routingcfg.Config{
			Provider: routingcfg.ProviderChi,
			Chi: &chi.Config{
				ServiceName:            otelServiceName,
				EnableCORSForLocalhost: false,
				SilenceRouteLogging:    false,
			},
		},
		Queues: msgconfig.QueuesConfig{
			DataChangesTopicName:              dataChangesTopicName,
			OutboundEmailsTopicName:           outboundEmailsTopicName,
			SearchIndexRequestsTopicName:      searchIndexRequestsTopicName,
			MobileNotificationsTopicName:      mobileNotificationsTopicName,
			UserDataAggregationTopicName:      userDataAggregationTopicName,
			WebhookExecutionRequestsTopicName: webhookExecutionRequestsTopicName,
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
			Debug:           false,
			Port:            defaultHTTPPort,
			StartupDeadline: 60 * time.Second,
			AppleAppSiteAssociation: &http.AppleAppSiteAssociationConfig{
				TeamID:   appleTeamID,
				BundleID: appleBundleID,
			},
		},
		Database: databasecfg.Config{
			Provider:                 databasecfg.ProviderPostgres,
			Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
			OAuth2TokenEncryptionKey: "",
			Debug:                    false,
			RunMigrations:            true,
			LogQueries:               false,
			MaxPingAttempts:          maxAttempts,
			PingWaitPeriod:           time.Second,
			MaxIdleConns:             5,
			MaxOpenConns:             7,
			ConnMaxLifetime:          30 * time.Minute,
			ReadConnection: databasecfg.ConnectionDetails{
				Username:   "api_db_user",
				Password:   "REPLACE_AT_DEPLOY", /* #nosec G101 */
				Database:   "dinner-done-better",
				Host:       "REPLACE_AT_DEPLOY",
				Port:       5432,
				DisableSSL: false,
			},
			WriteConnection: databasecfg.ConnectionDetails{
				Username:   "api_db_user",
				Password:   "REPLACE_AT_DEPLOY", /* #nosec G101 */
				Database:   "dinner-done-better",
				Host:       "REPLACE_AT_DEPLOY",
				Port:       5432,
				DisableSSL: false,
			},
		},
		Observability: prodObservabilityConfig,
		Email: emailcfg.Config{
			Provider: emailcfg.ProviderResend,
			Resend: &resend.Config{
				APIToken: "placeholder", // overridden by env from api-service-config secret
			},
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "prod_emailer",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Analytics: analyticscfg.Config{
			ProxySources: analyticscfg.ProxySourcesConfig{
				IOS: &analyticscfg.SourceConfig{
					Provider: analyticscfg.ProviderPostHog,
					Posthog:  &analyticsposthog.Config{APIKey: "placeholder"}, // overridden by env from api-service-config secret
					CircuitBreaker: circuitbreakingcfg.Config{
						Name:                   "ios_analytics",
						ErrorRate:              .5,
						MinimumSampleThreshold: 100,
					},
				},
				Web: &analyticscfg.SourceConfig{
					Provider: analyticscfg.ProviderPostHog,
					Posthog:  &analyticsposthog.Config{APIKey: "placeholder"}, // overridden by env from api-service-config secret
					CircuitBreaker: circuitbreakingcfg.Config{
						Name:                   "web_analytics",
						ErrorRate:              .5,
						MinimumSampleThreshold: 100,
					},
				},
			},
			SourceConfig: analyticscfg.SourceConfig{
				Provider: analyticscfg.ProviderPostHog,
				Posthog:  &analyticsposthog.Config{APIKey: "placeholder"}, // overridden by env from api-service-config secret
				CircuitBreaker: circuitbreakingcfg.Config{
					Name:                   "api_analytics",
					ErrorRate:              .5,
					MinimumSampleThreshold: 100,
				},
			},
		},
		TextSearch: textsearchcfg.Config{
			Provider: textsearchcfg.AlgoliaProvider,
			Algolia:  &algolia.Config{},
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "prod_text_searcher",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		FeatureFlags: featureflagscfg.Config{
			Provider: featureflagscfg.ProviderPostHog,
			PostHog:  &posthog.Config{ProjectAPIKey: "placeholder"}, // overridden by env from CSI secret
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Auth: authcfg.Config{
			Passkey: authcfg.PasskeyConfig{
				RPID:          "dinnerdonebetter.com",
				RPDisplayName: branding.CompanyName,
				RPOrigins:     []string{"https://dinnerdonebetter.com", "https://www.dinnerdonebetter.com", "https://admin.dinnerdonebetter.com"},
			},
			Tokens: tokenscfg.Config{
				Provider:                tokenscfg.ProviderPASETO,
				Issuer:                  "dinner-done-better",
				Audience:                prodTokensAudience,
				Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
			},
			Debug:                 false,
			EnableUserSignup:      true,
			MinimumUsernameLength: 3,
			MinimumPasswordLength: 8,
		},
		Services: config.ServicesConfig{
			Auth: authservice.Config{
				OAuth2: authservice.OAuth2Config{
					Domain:               prodOAuth2Domain,
					AccessTokenLifespan:  time.Hour,
					RefreshTokenLifespan: time.Hour,
					Debug:                false,
				},
				Debug:                 false,
				EnableUserSignup:      true,
				MinimumUsernameLength: 3,
				MinimumPasswordLength: 8,
				TokenLifetime:         5 * time.Minute,
				Tokens: tokenscfg.Config{
					Provider:                tokenscfg.ProviderPASETO,
					Issuer:                  "dinner-done-better",
					Audience:                prodTokensAudience,
					Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
				},
			},
			DataPrivacy: dataprivacycfg.Config{
				Uploads: uploadscfg.Config{
					Storage: gcpUserDataStorage,
					Debug:   false,
				},
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
			APNs: &notificationscfg.APNsConfig{
				AuthKeyPath: "/mnt/apns/apns-auth-key.p8", // mounted from K8s secret apns-credentials
				TeamID:      iosTeamID,
				BundleID:    iosBundleID,
				Production:  true,
			},
			FCM: &notificationscfg.FCMConfig{
				// CredentialsPath empty: uses Application Default Credentials (GCP workload identity)
			},
		},
	}
}
