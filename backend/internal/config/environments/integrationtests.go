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
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	oauth2database "github.com/primandproper/platform-go/v14/authentication/oauth2serverstore"
	oauth2servercfg "github.com/primandproper/platform-go/v14/authentication/oauth2serverstore/config"
	webauthncfg "github.com/primandproper/platform-go/v14/authentication/webauthnsessions/config"
	analyticscfg "github.com/primandproper/primitives-go/v2/analytics/config"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"
	platformwebauthn "github.com/primandproper/primitives-go/v2/authentication/webauthn"
	capitalismcfg "github.com/primandproper/primitives-go/v2/capitalism/config"
	circuitbreakingcfg "github.com/primandproper/primitives-go/v2/circuitbreaking/config"
	encryptioncfg "github.com/primandproper/primitives-go/v2/cryptography/encryption/config"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	emailcfg "github.com/primandproper/primitives-go/v2/email/config"
	"github.com/primandproper/primitives-go/v2/encoding"
	featureflagscfg "github.com/primandproper/primitives-go/v2/featureflags/config"
	msgconfig "github.com/primandproper/primitives-go/v2/messagequeue/config"
	"github.com/primandproper/primitives-go/v2/messagequeue/redis"
	notificationscfg "github.com/primandproper/primitives-go/v2/notifications/mobile/config"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	tracingcfg "github.com/primandproper/primitives-go/v2/observability/tracing/config"
	"github.com/primandproper/primitives-go/v2/pointer"
	"github.com/primandproper/primitives-go/v2/routing/backends/chi"
	routingcfg "github.com/primandproper/primitives-go/v2/routing/config"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"
	"github.com/primandproper/primitives-go/v2/server/grpc"
	"github.com/primandproper/primitives-go/v2/server/http"
	uploadscfg "github.com/primandproper/primitives-go/v2/uploads/config"
	"github.com/primandproper/primitives-go/v2/uploads/objectstorage"
)

// BuildIntegrationTestsConfig returns the configuration the integration test environment runs with.
func BuildIntegrationTestsConfig() *config.APIServiceConfig {
	uploadsConfig := uploadscfg.Config{
		Debug: false,
		Storage: objectstorage.Config{
			Provider:   "memory",
			BucketName: "avatars",
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
				EnableCORSForLocalhost: true,
				SilenceRouteLogging:    false,
			},
		},
		Meta: config.MetaSettings{
			Debug:   false,
			RunMode: testingEnv,
		},
		Queues: queuescfg.Config{
			DataChangesTopicName:         dataChangesTopicName,
			OutboundEmailsTopicName:      outboundEmailsTopicName,
			SearchIndexRequestsTopicName: searchIndexRequestsTopicName,
			MobileNotificationsTopicName: mobileNotificationsTopicName,
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
				Name:                   featureFlaggerSource,
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		FeatureFlags: featureflagscfg.Config{
			// we're using a noop version of this in dev right now, but it still tries to instantiate a circuit breaker.
			Provider: featureflagscfg.ProviderNoop,
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   featureFlaggerSource,
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Email: emailcfg.Config{
			// Nothing sends mail in this environment; v9 makes that say so rather than assume it.
			Provider: emailcfg.ProviderNoop,
		},
		Analytics: analyticscfg.Config{
			// The multisource reporter resolves a source name to a reporter and refuses one
			// it does not know, so the sources the clients send have to be declared even
			// where nothing is actually reported. In v9 an unconfigured source fell through
			// to the ambient reporter; in v10 it is ErrUnknownSource, which surfaces to the
			// caller as a failed TrackEvent.
			ProxySources: analyticscfg.ProxySourcesConfig{
				iosPlatform: {
					Provider: analyticscfg.ProviderNoop,
					CircuitBreaker: circuitbreakingcfg.Config{
						Name:                   iosAnalyticsSource,
						ErrorRate:              .5,
						MinimumSampleThreshold: 100,
					},
				},
				webPlatform: {
					Provider: analyticscfg.ProviderNoop,
					CircuitBreaker: circuitbreakingcfg.Config{
						Name:                   webAnalyticsSource,
						ErrorRate:              .5,
						MinimumSampleThreshold: 100,
					},
				},
			},
			SourceConfig: analyticscfg.SourceConfig{
				// Analytics are off here, which the provider now says rather than
				// leaving it to an unset value. The circuit breaker is still built.
				Provider: analyticscfg.ProviderNoop,
				CircuitBreaker: circuitbreakingcfg.Config{
					Name:                   featureFlaggerSource,
					ErrorRate:              .5,
					MinimumSampleThreshold: 100,
				},
			},
		},
		// The suite runs a real passkey ceremony against a virtual authenticator, so the
		// relying party has to be named here: the authenticator signs for this RPID and
		// claims this origin, and either one disagreeing is a ceremony that fails
		// verification. The store is the table, which is what the suite is checking.
		Auth: authcfg.Config{
			Sessions: authcfg.SessionsConfig{
				AbsoluteTimeout: sessionAbsoluteTimeout,
				IdleTimeout:     sessionIdleTimeout,
				TouchInterval:   sessionTouchInterval,
			},
			Passkey: webauthncfg.Config{
				Provider: webauthncfg.ProviderDatabase,
				RelyingParty: platformwebauthn.Config{
					RPID:            branding.LocalDevRPID,
					RPDisplayName:   branding.CompanyName,
					RPOrigins:       branding.LocalDevWebAppOrigins(),
					CeremonyTimeout: passkeyCeremonyTimeout,
				},
			},
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
					Issuer:   localAPIPublicURL,
					// The audience every access token carries, and the identifier the gRPC
					// interceptor checks a presented token against. Named explicitly rather
					// than defaulted so that a token minted for the MCP server — which
					// shares this database, and therefore this store — is not spendable
					// here.
					Resources: []string{localAPIPublicURL},
					// Meant as "no sweeper of its own": `ddb job db-cleaner` calls Sweep
					// instead, one pass for the fleet rather than one per replica, each
					// running the same full-table delete on its own timer.
					//
					// Explicitly none, because the db-cleaner job sweeps for the fleet. v14
					// made this a *time.Duration, so an explicit zero now means "no sweeper"
					// instead of being rewritten to the ten-minute default — platform-go#456.
					SweepInterval: pointer.To(time.Duration(0)),
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
				Encryption:            encryptioncfg.Config{Provider: encryptioncfg.ProviderAES, CurrentKeyID: "v1"},
				ArtifactEncryptionKey: localDisclosureArtifactEncryptionKey,
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
