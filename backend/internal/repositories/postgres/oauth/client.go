package oauth

import (
	"context"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth/generated"

	"github.com/primandproper/platform-go/v9/cryptography/encryption"
	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "oauth_db_client"
)

// repository is the oauth2 client and token repo implemenation.
type repository struct {
	database.Client
	tracer                   tracing.Tracer
	logger                   logging.Logger
	generatedQuerier         generated.Querier
	auditLogEntryRepo        audit.Repository
	oauth2ClientTokenEncDec  encryption.EncryptorDecryptor
	readDB                   database.SQLQueryExecutor
	writeDB                  database.SQLQueryExecutor
	oauth2ClientTokenHashKey []byte
}

// ProvideOAuthRepository provides a new repository.
func ProvideOAuthRepository(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	cfg *dbcfg.Config,
	client database.Client,
) oauth.Repository {
	encDec, err := encryptioncfg.NewEncryptorDecryptor(
		ctx,
		&cfg.Encryption,
		[]byte(cfg.OAuth2TokenEncryptionKey),
		encryptioncfg.WithLogger(logger),
		encryptioncfg.WithTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil
	}

	c := &repository{
		Client:                   client,
		readDB:                   client.Reader(),
		writeDB:                  client.Writer(),
		tracer:                   tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:         generated.New(),
		auditLogEntryRepo:        auditLogEntryRepo,
		oauth2ClientTokenEncDec:  encDec,
		oauth2ClientTokenHashKey: []byte(cfg.OAuth2TokenEncryptionKey),
		logger:                   logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
